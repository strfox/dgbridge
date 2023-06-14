package main

import (
	"dgbridge/src/ext"
	"encoding/json"
	"os"
	"strconv"
)

type (
	Rules struct {
		DiscordToSubprocess []Rule `json:"DiscordToSubprocess"`
		SubprocessToDiscord []Rule `json:"SubprocessToDiscord"`
	}
	Rule struct {
		Match    ext.Regexp `json:"Match"`
		Template string     `json:"Template"`
	}
)

type (
	Props struct {
		Author Author
	}
	Author struct {
		Username      string
		Discriminator string
		AccentColor   int
	}
)

// LoadRules loads a set of rules from a JSON file.
func LoadRules(path string) (*Rules, error) {
	fileContents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rules Rules
	err = json.Unmarshal(fileContents, &rules)
	if err != nil {
		return nil, err
	}
	return &rules, err
}

// ApplyRules applies rules to a string.
// If props are provided, a matching template will be built using those props.
func ApplyRules(rules []Rule, props *Props, input string) string {
	for _, rule := range rules {
		result := ApplyRule(rule, props, input)
		if result != "" {
			return result
		}
	}
	return ""
}

// ApplyRule applies a rule to a given input string if it matches.
//
// Parameters:
// props: If passed, the Rule's template is built with the given Props.
func ApplyRule(rule Rule, props *Props, input string) string {
	if rule.Match.MatchString(input) {
		if props == nil {
			return rule.Match.ReplaceAllString(input, rule.Template)
		}
		return rule.Match.ReplaceAllString(input, buildTemplate(rule.Template, *props))
	}
	return ""
}

// Builds a rule template for Discord -> Process communication.
// It replaces all special combinations in the template with their corresponding properties.
//
// Example:
//   - ^U turns into Username
//   - ^T turns into Discriminator
//
// Returns template with Props applied.
func buildTemplate(template string, props Props) string {
	var result []rune
	runes := []rune(template)
	for i := 0; i < len(runes); i++ {
		currentRune := runes[i]
		if currentRune == '^' && i+1 < len(template) {
			switch template[i+1] {
			case '^':
				// This is an escaped ^
				result = append(result, '^')
				i++
				continue
			case 'U':
				result = append(result, []rune(props.Author.Username)...)
				i++
				continue
			case 'T':
				result = append(result, []rune(props.Author.Discriminator)...)
				i++
				continue
			case 'C':
				result = append(result, []rune(strconv.FormatInt(int64(props.Author.AccentColor), 16))...)
				i++
				continue
			}
		}
		result = append(result, currentRune)
	}
	return string(result)
}
