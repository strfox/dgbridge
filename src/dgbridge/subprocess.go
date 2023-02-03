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

// SubprocessCtx is a struct that holds the event emitter for reading and
// writing to a subprocess' streams.
type SubprocessCtx struct {
	cmd                 *exec.Cmd
	StdoutLineEvent     ext.Event[string]
	StderrLineEvent     ext.Event[string]
	WriteStdinLineEvent ext.Event[string]
	ExitEvent           ext.Event[int]
}

// CreateSubprocess creates a command handle from the specified system command
// string and returns a SubprocessCtx struct. The subprocess is not started.
//
// Parameters:
//
//	command: system command string to use to start the process.
func CreateSubprocess(command string) SubprocessCtx {
	cmd := createCommand(command)
	return SubprocessCtx{
		cmd: cmd,
	}
}

// Start starts the subprocess. It starts goroutines to read from the stdout
// and write to the stdin of the subprocess, as well as a goroutine to wait for
// the subprocess to finish and handle signals sent to the subprocess.
// If an error occurs while starting the subprocess, the function returns the
// error.
func (ctx *SubprocessCtx) Start() error {
	err := ctx.startReadStdout()
	if err != nil {
		return err
	}

	err = ctx.startReadStderr()
	if err != nil {
		return err
	}

	err = ctx.startListenToWriteStdinEvents()
	if err != nil {
		return err
	}

	err = ctx.cmd.Start()
	if err != nil {
		return err
	}

	go ctx.relaySignalsToSubprocessUntilExit()
	go ctx.watchSubprocessExit()

	return nil
}

// createCommand creates a command handle from the specified system command
// string. It does not start the command automatically.
//
// Returns:
//
//	*exec.Cmd struct representing the command
func createCommand(command string) *exec.Cmd {
	trimmed := strings.TrimSpace(command)
	tokens := strings.Split(trimmed, " ")
	cmd := exec.Command(tokens[0], tokens[1:]...)
	return cmd
}

// startReadStdout reads the stdout from the subprocess and broadcasts
// SubprocessCtx.StdoutLineEvent whenever it emits a line.
// If there is an error creating the stdout pipe, it returns an error.
func (ctx *SubprocessCtx) startReadStdout() error {
	pipe, err := ctx.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}
	go func() {
		defer func(pipe io.ReadCloser) {
			_ = pipe.Close()
		}(pipe)

		scanner := bufio.NewScanner(pipe)

		for scanner.Scan() {
			ctx.StdoutLineEvent.Broadcast(scanner.Text())
		}
	}()
	return nil
}

// startReadStderr reads the stderr from the subprocess and broadcasts
// SubprocessCtx.StderrLineEvent whenever it emits a line.
// If there is an error creating the stderr pipe, it returns an error.
func (ctx *SubprocessCtx) startReadStderr() error {
	pipe, err := ctx.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %v", err)
	}
	go func() {
		defer func(pipe io.ReadCloser) {
			_ = pipe.Close()
		}(pipe)

		scanner := bufio.NewScanner(pipe)

		for scanner.Scan() {
			ctx.StderrLineEvent.Broadcast(scanner.Text())
		}
	}()
	return nil
}

// startListenToWriteStdinEvents writes data to the subprocess' stdin. It listens
// to stdin write events, and writes the event's payload string to the
// subprocess' stdin.
// If an error occurs while opening the pipe, it returns an error.
func (ctx *SubprocessCtx) startListenToWriteStdinEvents() error {
	pipe, err := ctx.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating stdin pipe: %v", err)
	}
	writer := bufio.NewWriter(pipe)

	go func() {
		defer func(pipe io.WriteCloser) {
			_ = pipe.Close()
		}(pipe)

		lineCh := ctx.WriteStdinLineEvent.Listen()
		defer ctx.WriteStdinLineEvent.Off(lineCh)

		for line := range lineCh {
			_, _ = writer.WriteString(line)
			_ = writer.Flush()
		}
	}()
	return nil
}

// watchSubprocessExit waits for the subprocess to exit.
func (ctx *SubprocessCtx) watchSubprocessExit() {
	err := ctx.cmd.Wait()

	// Subprocess exited
	// Now we can check for the subprocess' exit code, and exit our own
	// process with that same exit code.
	if exitError, ok := err.(*exec.ExitError); ok {
		// Subprocess exited abnormally - copy the exit code.
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		exitCode := waitStatus.ExitStatus()

		log.Printf("[debug] Subprocess exited abnormally with code %d, emitting exit event\n", exitCode)
		ctx.ExitEvent.Broadcast(exitCode)
	} else if err != nil {
		// Another type of error occurred while waiting for the command.
		// This is probably a programming error.
		log.Panicln("Error: Waiting for subcommand caused an error:", err)
	} else {
		// Subprocess exited normally
		log.Println("[debug] Subprocess exited normally, emitting exit event")
		ctx.ExitEvent.Broadcast(0)
	}
}

// relaySignalsToSubprocessUntilExit continuously relays the current process'
// signals to the specified command. When SubprocessCtx.ExitEvent is
// broadcast, the function exits.
func (ctx *SubprocessCtx) relaySignalsToSubprocessUntilExit() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh)

	exitCh := ctx.ExitEvent.Listen()
	defer ctx.ExitEvent.Off(exitCh)

	for {
		select {
		case sig := <-sigCh:
			log.Println("[debug] Received signal:", sig)
			// We received a signal, let's try passing it to the subprocess
			if err := ctx.cmd.Process.Signal(sig); err != nil {
				// Not clear how we can hit this, but probably not
				// worth terminating the child.
				log.Printf("[debug] Couldn't send signal \"%v\" to subprocess: %v\n", sig, err)
			}
		case <-exitCh:
			break
		}
	}
}
