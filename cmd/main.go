package main

import (
	"github.com/spf13/cobra"
	"rs232/pkg/app"
	"time"
)

func main() {
	params := app.Params{}
	cmd := cobra.Command{
		Use: `rs232-reader`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := app.App(params); err != nil {

			}
		},
	}
	cmd.PersistentFlags().IntVar(&params.BaudRate, `baud-rate`, 38400, `baud rate`)
	cmd.PersistentFlags().StringVar(&params.DevicePath, `device-path`, `/dev/ttyS0`, `path to tty device`)
	cmd.PersistentFlags().StringVar(&params.WorkDir, `work-dir`, `/opt/rs232-reader`, `path to working directory`)
	cmd.PersistentFlags().DurationVar(&params.SessionTimeout, `session-timeout`, time.Second*5, `session timeout`)
	cmd.PersistentFlags().BoolVar(&params.Debug, `debug`, false, `debug mode`)
	cmd.PersistentFlags().BoolVar(&params.Verbose, `verbose`, false, `debug mode`)
	cmd.Execute()

}
