package lib

import (
	"os"
	"os/signal"
	"syscall"
)

// Panic panics with the given error if it is not nil
// It is a convenience function to reduce boilerplate code for handling errors
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

// WaitForQuitSignal waits for a SIGINT (Ctrl+C) or SIGTERM signal to be received
// When the signal is received, the function unblocks and continues execution
// This function is useful for gracefully shutting down the application when a termination signal is received
func WaitForQuitSignal() {
	// Create a channel to receive the quit signals (SIGINT and SIGTERM)
	quitCh := make(chan os.Signal, 2)

	// Notify the quit channel when SIGINT (Ctrl+C) or SIGTERM is received
	signal.Notify(quitCh, syscall.SIGINT)
	signal.Notify(quitCh, syscall.SIGTERM)

	// Wait for a signal to be received on the quit channel
	<-quitCh
}
