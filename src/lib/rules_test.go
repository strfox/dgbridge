package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildTemplate(t *testing.T) {
	tests := []struct {
		Name   string
		Props  Props
		Input  string
		Expect string
	}{
		{
			Name: "All parameters",
			Props: Props{
				Author: Author{
					Username:      "Bob^T",
					Discriminator: "1337",
					AccentColor:   0xFFFF00,
				},
			},
			Input:  "<^U#^T> ${1} ^^ ^A ^C",
			Expect: "<Bob^T#1337> ${1} ^ ^A ffff00",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := buildTemplate(test.Input, test.Props)
			assert.Equal(t, test.Expect, result)
		})
	}
}
