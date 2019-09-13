package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func createCtrlCContext() context.Context {
	fmt.Println("Press Ctrl-C to exit...")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
		<-sigChan
		cancel()
	}()

	return ctx
}
