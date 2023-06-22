package main

import "dgbridge/src/lib"

type (
	FileRoot struct {
		Tests     Tests                `validate:"required"`
		UserProps map[string]lib.Props `validate:"dive"`
	}
	Tests struct {
		DiscordToSubprocess []DiscordToSubprocessTest `validate:"required"`
		SubprocessToDiscord []SubprocessToDiscordTest `validate:"required,dive"`
	}
	DiscordToSubprocessTest struct {
		Input     string `validate:"required"`
		Expect    string
		UserProps string `validate:"required"`
	}
	SubprocessToDiscordTest struct {
		Input  string `validate:"required"`
		Expect string
	}
)
