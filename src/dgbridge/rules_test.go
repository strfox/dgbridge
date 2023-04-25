package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildTemplate(t *testing.T) {
	bobUser := &discordgo.User{
		Username:      "Bob^T",
		Discriminator: "1337",
		AccentColor:   0xFFFF00,
	}
	bobMember := &discordgo.Member{
		Nick: "bobby",
	}
	tests := []struct {
		name   string
		ctx    TemplateContext
		input  string
		expect string
	}{
		{
			name: "All parameters",
			ctx: TemplateContext{
				session: nil,
				message: &discordgo.Message{
					Member: bobMember,
					Author: bobUser,
				},
			},
			input:  "<^U#^T> ${1} ^^ ^A ^C",
			expect: "<Bob^T#1337> ${1} ^ ^A ffff00",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.ctx.buildTemplate(test.input)
			assert.Equal(t, test.expect, result)
		})
	}
}
