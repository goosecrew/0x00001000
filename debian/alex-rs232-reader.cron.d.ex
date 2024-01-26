#
# Regular cron jobs for the alex-rs232-reader package
#
0 4	* * *	root	[ -x /usr/bin/alex-rs232-reader_maintenance ] && /usr/bin/alex-rs232-reader_maintenance
