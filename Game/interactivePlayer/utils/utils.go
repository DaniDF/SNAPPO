package utils

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

/*
Credits: Gemini
Reads a single char from stdin without echo and NewLine
*/
func ReadKey() (byte, error) {
	fd := int(os.Stdin.Fd())

	// 1. Verify standard input is actually a terminal
	if !term.IsTerminal(fd) {
		return 0, errors.New("Not interactive terminal")
	}

	// 2. Put terminal into raw mode
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, err
	}

	// Create a cleanup function so we don't duplicate code
	restore := func() {
		_ = term.Restore(fd, oldState)
	}
	defer restore()

	// 3. Catch interrupt signals so terminal gets restored on Ctrl+C/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		restore() // Restore terminal before exiting!
		os.Exit(1)
	}()

	// Read 1 byte
	var buf [1]byte
	_, err = os.Stdin.Read(buf[:])
	if err != nil {
		return 0, err
	}

	return buf[0], nil
}
