package docker

import (
	"os"
	"os/signal"
	"syscall"
)

var stop chan os.Signal

// Run main func with signal notification
func Run(main func()) chan os.Signal {
	stop = make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go main()

	return stop
}

// Exit with code
func Exit(code int) {
	os.Exit(code)

	stop <- syscall.SIGTERM
}