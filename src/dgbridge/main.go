package main

import (
	"bufio"
	"dgbridge/src/lib"
	"fmt"
	"github.com/alexflint/go-arg"
	"log"
	"os"
)

type CliArgs struct {
	Token     string `arg:"required,-t,--token" help:"Discord authentication token"`
	ChannelId string `arg:"required,-i,--channel_id" help:"Discord channel ID"`
	RulesFile string `arg:"required,-r,--rules" help:"Path to the file with translation rules"`
	Command   string `arg:"required,positional"`
}

func main() {
	fmt.Printf("Dgbridge (%v)\n", lib.Version)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var args CliArgs
	arg.MustParse(&args)

	rules, err := lib.LoadRules(args.RulesFile)
	if err != nil {
		log.Fatalf("error loading rules: %v\n", err)
	}

	subprocess := NewSubprocess(args.Command)

	go relaySubprocessStdout(&subprocess)
	go relaySubprocessStderr(&subprocess)
	go relayStdinToSubprocessStdin(&subprocess)

	// Create a goroutine that will wait for the subprocess to emit an exit event.
	go func() {
		log.Println("[debug] Waiting for child to exit")
		exitCh := subprocess.ExitEvent.Listen()
		defer subprocess.ExitEvent.Off(exitCh)
		exitCode := <-exitCh
		os.Exit(exitCode)
	}()

	err = subprocess.Start()
	if err != nil {
		log.Fatalln("[fatal] error starting command:", err)
	}

	freeBotFunc, err := StartDiscordBot(BotParameters{
		Token:          args.Token,
		RelayChannelId: args.ChannelId,
		Subprocess:     &subprocess,
		Rules:          *rules,
	})
	if err != nil {
		// This is a non-fatal error. We want the server to run even if the
		// Discord connection failed.
		log.Println("[error] failed to start Discord bot:", err)
	}
	defer freeBotFunc()

	// Block forever
	select {}
}

// relayStdinToSubprocessStdin continuously relays os.Stdin to the subprocess' stdin.
func relayStdinToSubprocessStdin(ctx *SubprocessContext) {
	// Relay os.Stdin to the subprocess' stdin.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// It is required to append a newline to the line because  it is
		// not included in Text().
		ctx.WriteStdinLineEvent.Broadcast(scanner.Text() + "\n")
	}
}

// relaySubprocessStdout continuously relays the subprocess' stdout to os.Stdout.
func relaySubprocessStdout(ctx *SubprocessContext) {
	lineCh := ctx.StdoutLineEvent.Listen()
	defer ctx.StdoutLineEvent.Off(lineCh)
	for line := range lineCh {
		_, _ = os.Stdout.WriteString(line + "\n")
	}
}

// relaySubprocessStderr continuously relays the subprocess' stderr to os.Stderr.
func relaySubprocessStderr(ctx *SubprocessContext) {
	lineCh := ctx.StderrLineEvent.Listen()
	defer ctx.StderrLineEvent.Off(lineCh)
	for line := range lineCh {
		_, _ = os.Stderr.WriteString(line + "\n")
	}
}
