package main

import (
	"dgbridge/src/ext"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"sync"
)

// BotParameters holds data to be passed to StartDiscordBot.
type BotParameters struct {
	// Discord authentication token
	Token string

	// For information on the below fields, see their documentation on
	// BotCtx.
	RelayChannelId string
	Subprocess     *SubprocessCtx
	Rules          Rules
}

// StartDiscordBot starts the discord bot. This function is non-blocking.
//
// Returns:
//
//	a function that when called will close the discord bot session, or an
//	error if an error occurs while starting the bot
func StartDiscordBot(params BotParameters) (func(), error) {
	dg, err := discordgo.New("Bot " + params.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %v", err)
	}
	ctx := BotCtx{
		relayChannelId: params.RelayChannelId,
		subprocess:     params.Subprocess,
		rules:          params.Rules,
		readyOnce:      sync.Once{},
	}
	dg.AddHandler(ctx.ready())
	dg.AddHandler(ctx.messageCreate())
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %v", err)
	}
	return func() {
		_ = dg.Close()
	}, nil
}

// BotCtx holds the context of a discordgo bot, including any additional data
// needed by the bot.
type BotCtx struct {
	// relayChannelId is the ID of the Discord channel to bridge the subprocess
	// with
	relayChannelId string

	// The context of a subprocess that the bot bridges with.
	subprocess *SubprocessCtx

	// The rules used for conversion of Discord messages to server messages and
	// vice versa
	rules Rules

	// Tracks if discordgo.Ready was emitted so that initialization logic
	// only executes once
	readyOnce sync.Once
}

// ready handles a discordgo.Ready event.
// It sets up the jobs to relay the subprocess' streams to Discord.
func (ctx *BotCtx) ready() func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		ctx.readyOnce.Do(func() {
			go ctx.relayStreamToDiscord(s, &ctx.subprocess.StdoutLineEvent)
			go ctx.relayStreamToDiscord(s, &ctx.subprocess.StderrLineEvent)
		})
	}
}

// relayStdoutToDiscord relays the output of a subprocess to a discord channel.
// It continuously listens to the subprocess' output event and each the line
// to the specified discord channel.
//
// The line is passed through a set of rules defined in
// 'Rules.SubprocessToDiscord' before being sent to the discord channel.
//
// If an error occurs when sending a message to Discord, error is simply
// logged to stdout.
//
// Parameters:
//
//	s: a pointer to a discordgo session, used to send the message to discord
//	channel.
//	event: which event to listen to
func (ctx *BotCtx) relayStreamToDiscord(s *discordgo.Session, event *ext.Event[string]) {
	lineCh := event.Listen()
	defer event.Off(lineCh)
	for line := range lineCh {
		line = ApplyRules(ctx.rules.SubprocessToDiscord, line, &TemplateContext{
			session: s,
			message: nil,
		})
		if line == "" {
			// No rules matched.
			continue
		}
		_, err := s.ChannelMessageSend(ctx.relayChannelId, line)
		if err != nil {
			log.Printf("error sending message to discord: %v", err)
		}
	}
}

// messageCreate listens for Discord message creation events and relays the
// contents of those messages to the subprocess.
//
// The line is passed through a set of rules defined in
// 'Rules.DiscordToSubprocess' before being sent to the discord channel.
func (ctx *BotCtx) messageCreate() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Is it the bot's own message?
		if m.Author.ID == s.State.User.ID {
			return
		}
		// Is it the relay channel?
		if !(m.ChannelID == ctx.relayChannelId) {
			return
		}

		msg := m.Content
		msg = ApplyRules(ctx.rules.DiscordToSubprocess, msg, &TemplateContext{
			session: s,
			message: m.Message,
		})
		if msg == "" {
			// No rules matched.
			return
		}

		ctx.subprocess.WriteStdinLineEvent.Broadcast(msg + "\n")
	}
}
