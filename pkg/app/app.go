package app

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Params struct {
	DevicePath     string
	BaudRate       int
	WorkDir        string
	SessionTimeout time.Duration
	Verbose        bool
	Debug          bool
}

func App(params Params) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go CatchSigTerm(ctx, cancel)
	log.SetLevel(log.ErrorLevel)
	if params.Verbose {
		log.SetLevel(log.InfoLevel)
	}
	if params.Debug {
		log.SetLevel(log.DebugLevel)
	}
	stat, err := os.Stat(params.WorkDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(params.WorkDir, os.ModeDir); err != nil {
				return errors.Wrapf(err, `create workdir error`)
			}
		} else {
			return errors.Wrap(err, `workdir error`)
		}
	} else {
		if !stat.IsDir() {
			return fmt.Errorf(`workdir path is not a catalog(directory)`)
		}
	}
	c := &serial.Config{Name: params.DevicePath, Baud: params.BaudRate}
	c.Parity = serial.ParityNone
	c.StopBits = serial.Stop1
	c.Size = serial.DefaultSize
	port, err := serial.OpenPort(c)
	if err != nil {
		return errors.Wrapf(err, `serial.OpenPort error`)
	}

	var bufChan chan []byte
	var errChan chan error
	var fatalError error

	go func(bufChan chan []byte, errChan chan error) {
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := port.Read(buf)
				if err != nil {
					errChan <- errors.Wrapf(err, `serial port read error`)
					return
				}
				log.Debugln(111)
				bufChan <- buf[:n]
			}
		}
	}(bufChan, errChan)

	go func(errChan chan error) {
		select {
		case <-ctx.Done():
			return
		case fatalError = <-errChan:
			cancel()
			return
		}
	}(errChan)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup, bufChan chan []byte, errChan chan error) {
		defer func() {
			wg.Done()
			fmt.Println(`exit `)
		}()
		var fd *os.File
		var lastScan time.Time
		var bytesReceived int
		var buf20kb int
		for {
			log.Debugln(222)
			select {
			case <-ctx.Done():
				return
			case buf := <-bufChan:
				log.Debugln(333)
				dur := time.Now().Sub(lastScan)
				if dur.Seconds() > params.SessionTimeout.Seconds() {
					if fd != nil {
						log.Infof(`close %s`, fd.Name())
						fd.Close()
						fd = nil
						bytesReceived = 0
						buf20kb = 0
					}
				}
				lastScan = time.Now()
				if fd == nil {
					path := fmt.Sprintf(`%s/%s.log`, params.WorkDir, lastScan.String())
					fd, err = os.Create(path)
					if err != nil {
						errChan <- errors.Wrapf(err, `create %s error`, fmt.Sprintf(`%s.log`, lastScan.String()))
						return
					}
					log.Infof(`create %s`, path)
				}
				if params.Debug {
					fmt.Print(string(buf))
				}
				bytesReceived += len(buf)
				buf20kb += len(buf)
				if buf20kb > 20*1024 {
					log.Infof(`%dkb received`, bytesReceived/1024)
					buf20kb = 0
				}
				_, err := fd.Write(buf)
				if err != nil {
					errChan <- errors.Wrapf(err, `file write error`)
					return
				}
			}

		}
	}(&wg, bufChan, errChan)
	wg.Wait()
	return fatalError
}

func CatchSigTerm(ctx context.Context, cancel context.CancelFunc) {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-ctx.Done():
			return
		case <-cancelChan:
			cancel()
			return
		}
	}
}
