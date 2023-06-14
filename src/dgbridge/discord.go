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
	Token          string             // Discord auth token
	RelayChannelId string             // Saved in BotContext
	Subprocess     *SubprocessContext // Saved in BotContext
	Rules          Rules              // Saved in BotContext
}

type BotContext struct {
	relayChannelId string             // ID of destination Discord channel
	subprocess     *SubprocessContext // Subprocess context
	rules          Rules              // Message conversion rules
	readyOnce      sync.Once          // Tracks if bot was initialized
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
	context := BotContext{
		relayChannelId: params.RelayChannelId,
		subprocess:     params.Subprocess,
		rules:          params.Rules,
		readyOnce:      sync.Once{},
	}
	dg.AddHandler(context.ready())
	dg.AddHandler(context.messageCreate())
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %v", err)
	}
	return func() {
		_ = dg.Close()
	}, nil
}

// Handles a discordgo.Ready event.
// Sets up the jobs to relay text to Discord.
func (self *BotContext) ready() func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		self.readyOnce.Do(func() {
			go self.startRelayJob(s, &self.subprocess.StdoutLineEvent)
			go self.startRelayJob(s, &self.subprocess.StderrLineEvent)
		})
	}
}

// Relays the output of a subprocess to a discord channel.
// It continuously listens to the specified event for data to relay.
//
// If an error occurs when sending a message to Discord, error is simply
// logged to stdout.
//
// Parameters:
//
//	s:
//		A pointer to a discordgo session, used to send the message to discord
//		channel.
//	event:
//		Which subprocess event to listen to
func (self *BotContext) startRelayJob(session *discordgo.Session, event *ext.EventChannel[string]) {
	lineCh := event.Listen()
	defer event.Off(lineCh)
	for line := range lineCh {
		line = ApplyRules(self.rules.SubprocessToDiscord, nil, line)
		if line == "" {
			// No rules matched.
			continue
		}
		_, err := session.ChannelMessageSend(self.relayChannelId, line)
		if err != nil {
			log.Printf("error sending message to discord: %v", err)
		}
	}
}

// Listens for Discord message creation events and relays the
// contents of those messages to the subprocess.
func (self *BotContext) messageCreate() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			// Is bot's own message
			return
		}
		if !(m.ChannelID == self.relayChannelId) {
			// Is not relay channel
			return
		}
		msg := m.Content
		msg = ApplyRules(self.rules.DiscordToSubprocess, &Props{
			Author: Author{
				Username:      m.Author.Username,
				Discriminator: m.Author.Discriminator,
				AccentColor:   m.Author.AccentColor,
			}}, msg)
		if msg == "" {
			// No rules matched.
			return
		}
		self.subprocess.WriteStdinLineEvent.Broadcast(msg + "\n")
	}
}
