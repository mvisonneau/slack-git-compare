package slack

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/mvisonneau/slack-git-compare/pkg/config"

	"github.com/slack-go/slack"
)

// Slack holds clients and configuration to perform operations
// upon the Slack API, git providers and storage/cache endpoints
type Slack struct {
	Client        *slack.Client
	SigningSecret string
	CustomUsers   config.Users
}

// NewOptions holds configuration parameters to create a new and
// operational Slack object
type NewOptions struct {
	SigningSecret string
	Token         string
}

// New creates and configures a new Slack object
func New(cfg config.Slack, customUsers config.Users) (s Slack) {
	s.Client = slack.New(cfg.Token)
	s.SigningSecret = cfg.SigningSecret
	s.CustomUsers = customUsers

	return
}

// VerifySigningSecret was taken from the slash example
// https://github.com/slack-go/slack/blob/master/examples/slash/slash.go
func (s Slack) VerifySigningSecret(r *http.Request) error {
	verifier, err := slack.NewSecretsVerifier(r.Header, s.SigningSecret)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Need to use r.Body again when unmarshalling SlashCommand and InteractionCallback
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if _, err = verifier.Write(body); err != nil {
		return err
	}

	return verifier.Ensure()
}

// SendMessage on a given channel ID
func (s Slack) SendMessage(channelID, msg string) (err error) {
	_, _, err = s.Client.PostMessage(channelID,
		slack.MsgOptionText(msg, false),
		slack.MsgOptionAttachments())
	return
}
