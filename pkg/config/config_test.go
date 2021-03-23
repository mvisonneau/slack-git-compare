package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	assert.Equal(t, Config{
		ListenAddress: ":8080",
		Log: Log{
			Level:  "info",
			Format: "text",
		},
	}, NewConfig())
}

func TestValidConfig(t *testing.T) {
	cfg := NewConfig()

	cfg.Providers = Providers{
		Provider{
			Type:   "gitlab",
			Token:  "xxx",
			Owners: []string{"foo"},
		},
	}

	cfg.Slack.Token = "xxx"
	cfg.Slack.SigningSecret = "xxx"

	assert.NoError(t, cfg.Validate())
}
