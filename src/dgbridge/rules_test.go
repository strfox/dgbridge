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
			name: "Basic",
			ctx: TemplateContext{
				session: nil,
				message: &discordgo.Message{
					Member: bobMember,
					Author: bobUser,
				},
			},
			input:  "<^U#^T> ${1} ^^ ^A",
			expect: "<Bob^T#1337> ${1} ^ ^A",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.ctx.buildTemplate(test.input)
			assert.Equal(t, test.expect, result)
		})
	}
}
