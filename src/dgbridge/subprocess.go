package main

// Big help from:
// https://kevin.burke.dev/kevin/proxying-to-a-subcommand-with-go/
// https://www.yellowduck.be/posts/reading-command-output-line-by-line

import (
	"bufio"
	"dgbridge/src/ext"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// SubprocessContext is a struct that holds all events for reading and writing to a subprocess' streams.
type SubprocessContext struct {
	cmd                 *exec.Cmd
	StdoutLineEvent     ext.EventChannel[string] // Emits when subprocess' stdout emits a line
	StderrLineEvent     ext.EventChannel[string] // Emits when subprocess' stderr emits a line
	WriteStdinLineEvent ext.EventChannel[string] // Listens for data to write to stdin
	ExitEvent           ext.EventChannel[int]    // Emits when subprocess exits
}

// NewSubprocess creates a command handle from the specified system command string and returns a SubprocessContext
// struct.
// The subprocess is not started.
//
// Parameters:
//
//	command: system command string to use to start the process.
func NewSubprocess(command string) SubprocessContext {
	cmd := createCommand(command)
	return SubprocessContext{
		cmd: cmd,
	}
}

// Start starts the subprocess.
// Starts goroutines:
//  1. Read from the stdout
//  2. Write to the stdin
//  3. Wait for subprocess to finish
//  4. Handle signals sent to the subprocess
func (self *SubprocessContext) Start() error {
	err := self.watchStdout()
	if err != nil {
		return err
	}
	err = self.watchStderr()
	if err != nil {
		return err
	}
	err = self.listenStdin()
	if err != nil {
		return err
	}
	err = self.cmd.Start()
	if err != nil {
		return err
	}
	go self.relaySignalsToSubprocessUntilExit()
	go self.watchSubprocessExit()
	return nil
}

// createCommand returns a command handle created from the specified system command string.
// It doesn't run the command.
func createCommand(command string) *exec.Cmd {
	trimmed := strings.TrimSpace(command)
	tokens := strings.Split(trimmed, " ")
	cmd := exec.Command(tokens[0], tokens[1:]...)
	return cmd
}

// watchStdout watches the subprocess' stdout.
// It broadcasts StdoutLineEvent whenever the process emits a line.
func (self *SubprocessContext) watchStdout() error {
	pipe, err := self.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}
	go func() {
		defer func(pipe io.ReadCloser) {
			_ = pipe.Close()
		}(pipe)
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			self.StdoutLineEvent.Broadcast(scanner.Text())
		}
	}()
	return nil
}

// watchStderr watches the subprocess' stderr.
// It broadcasts StderrLineEvent whenever the process emits a line.
func (self *SubprocessContext) watchStderr() error {
	pipe, err := self.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %v", err)
	}
	go func() {
		defer func(pipe io.ReadCloser) {
			_ = pipe.Close()
		}(pipe)
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			self.StderrLineEvent.Broadcast(scanner.Text())
		}
	}()
	return nil
}

// listenStdin writes data to the subprocess' stdin whenever a WriteStdinLineEvent is emitted.
func (self *SubprocessContext) listenStdin() error {
	pipe, err := self.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating stdin pipe: %v", err)
	}
	writer := bufio.NewWriter(pipe)

	go func() {
		defer func(pipe io.WriteCloser) {
			_ = pipe.Close()
		}(pipe)

		lineCh := self.WriteStdinLineEvent.Listen()
		defer self.WriteStdinLineEvent.Off(lineCh)

		for line := range lineCh {
			_, _ = writer.WriteString(line)
			_ = writer.Flush()
		}
	}()
	return nil
}

// watchSubprocessExit waits for the subprocess to exit.
// When the subprocess exits, it emits ExitEvent.
func (self *SubprocessContext) watchSubprocessExit() {
	err := self.cmd.Wait()

	// Subprocess exited
	// Now we can check for the subprocess' exit code, and exit our own process with that same exit code.
	if exitError, ok := err.(*exec.ExitError); ok {
		// Subprocess exited abnormally - copy the exit code.
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		exitCode := waitStatus.ExitStatus()

		log.Printf("[debug] Subprocess exited abnormally with code %d, emitting exit event\n", exitCode)
		self.ExitEvent.Broadcast(exitCode)
	} else if err != nil {
		// Another type of error occurred while waiting for the command.
		// This is probably a programming error.
		log.Panicln("Error: Waiting for subcommand caused an error:", err)
	} else {
		// Subprocess exited normally
		log.Println("[debug] Subprocess exited normally, emitting exit event")
		self.ExitEvent.Broadcast(0)
	}
}

// relaySignalsToSubprocessUntilExit continuously relays the current process' signals to the specified command.
// When ExitEvent is broadcast, the function exits.
func (self *SubprocessContext) relaySignalsToSubprocessUntilExit() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh)

	exitCh := self.ExitEvent.Listen()
	defer self.ExitEvent.Off(exitCh)

	for {
		select {
		case sig := <-sigCh:
			// We received a signal, let's try passing it to the subprocess
			if err := self.cmd.Process.Signal(sig); err != nil {
				// Not clear how we can hit this, but probably not
				// worth terminating the child.
				log.Printf("[debug] Couldn't send signal \"%v\" to subprocess: %v\n", sig, err)
			}
		case <-exitCh:
			break
		}
	}
}
