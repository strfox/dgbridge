package main

import (
	"dgbridge/src/ext"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"os"
	"strconv"
)

type Rules struct {
	DiscordToSubprocess []Rule `json:"DiscordToSubprocess"`
	SubprocessToDiscord []Rule `json:"SubprocessToDiscord"`
}

type Rule struct {
	Match    ext.Regexp `json:"Match"`
	Template string     `json:"Template"`
}

// LoadRules loads a set of rules from a file.
//
// Parameters:
//
//	path: Path of the file to load
//
// Returns:
//
//	If an error occurs while reading the file, it returns nil an error.
//	Otherwise, it returns a pointer to the Rules struct and a nil
func LoadRules(path string) (*Rules, error) {
	fileContents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var payload Rules
	err = json.Unmarshal(fileContents, &payload)
	if err != nil {
		return nil, err
	}
	return &payload, err
}

// ApplyRules applies a set of rules to a given input string.
// Each rule is applied to the input string using the Apply method of the Rule
// struct. If the Apply method returns a non-empty string, the function returns
// that string. If none of the rules return a non-empty string, the function
// returns an empty string.
func ApplyRules(rules []Rule, in string, ctx *TemplateContext) string {
	for _, rule := range rules {
		result := rule.Apply(in, ctx)
		if result != "" {
			return result
		}
	}
	return ""
}

// Apply applies a rule to a given input string.
// It checks if the input string matches the rule's Match regular expression.
// If there is a match, it replaces the match with the Template string and
// returns the modified input string. If there is no match, it returns an empty
// string.
func (r *Rule) Apply(in string, ctx *TemplateContext) string {
	if r.Match.MatchString(in) {
		return r.Match.ReplaceAllString(in, ctx.buildTemplate(r.Template))
	}
	return ""
}

type TemplateContext struct {
	session *discordgo.Session
	message *discordgo.Message
}

func (ctx *TemplateContext) buildTemplate(template string) string {
	var result []rune
	runes := []rune(template)
	for i := 0; i < len(runes); i++ {
		iRune := runes[i]
		if iRune == '^' && i+1 < len(template) {
			switch template[i+1] {
			case '^':
				// This is an escaped ^
				result = append(result, '^')
				i++
				continue
			case 'U':
				result = append(result, []rune(ctx.message.Author.Username)...)
				i++
				continue
			case 'T':
				result = append(result, []rune(ctx.message.Author.Discriminator)...)
				i++
				continue
			case 'C':
				result = append(result, []rune(strconv.FormatInt(int64(ctx.message.Author.AccentColor), 16))...)
				i++
				continue
			}
		}
		result = append(result, iRune)
	}
	return string(result)
}
