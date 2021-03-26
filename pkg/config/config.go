package config

import (
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Cache holds the configuration regarding the scheduling of cache updates
type Cache struct {
	Providers CacheProviders
	Slack     CacheSlack
}

// CacheProviders ..
type CacheProviders struct {
	UpdateRepositories     CacheProvidersUpdateRepositories     `json:"update_repositories" yaml:"update_repositories"`
	UpdateRepositoriesRefs CacheProvidersUpdateRepositoriesRefs `json:"update_repositories_refs" yaml:"update_repositories_refs"`
}

// CacheSlack ..
type CacheSlack struct {
	UpdateUsersEmails CacheSlackUpdateUsersEmails `json:"update_users_emails" yaml:"update_users_emails"`
}

// CacheProvidersUpdateRepositories ..
type CacheProvidersUpdateRepositories struct {
	OnStart      bool `default:"true" json:"on_start" yaml:"on_start"`
	EverySeconds int  `default:"3600" json:"every_seconds" yaml:"on_schedule"`
}

// CacheProvidersUpdateRepositoriesRefs ..
type CacheProvidersUpdateRepositoriesRefs struct {
	OnStart      bool `default:"false" json:"on_start" yaml:"on_start"`
	EverySeconds int  `json:"every_seconds" yaml:"on_schedule"`
}

// CacheSlackUpdateUsersEmails ..
type CacheSlackUpdateUsersEmails struct {
	OnStart      bool `default:"true" json:"on_start" yaml:"on_start"`
	EverySeconds int  `default:"86400" json:"every_seconds" yaml:"on_schedule"`
}

// Provider holds the configuration of a git provider
type Provider struct {
	Type   string `validate:"oneof=github gitlab"`
	URL    string
	Token  string   `validate:"required"`
	Owners []string `validate:"gt=0"`
}

// Providers is a slice of Provider
type Providers []Provider

// Log holds runtime logging configuration
type Log struct {
	Level  string `default:"info" validate:"required,oneof=trace debug info warning error fatal panic"`
	Format string `default:"text" validate:"oneof=text json"`
}

// Slack holds Slack related configuration
type Slack struct {
	Token         string `validate:"required"`
	SigningSecret string `validate:"required" json:"signing_secret" yaml:"signing_secret"`
}

// User can be used to alias email addresses for a Slack user
type User struct {
	Email   string   `validate:"required,email"`
	Aliases []string `validate:"gt=0"`
}

// Users is a slice of User
type Users []User

// Config represents all the parameters required for the app to be configured properly
type Config struct {
	Cache         Cache
	Providers     Providers `validate:"gt=0,unique=Type"`
	ListenAddress string    `default:":8080" validate:"required"`
	Log           Log
	Slack         Slack
	Users         Users
}

// NewConfig returns a new Config with default values
func NewConfig() (cfg Config) {
	_ = defaults.Set(&cfg)
	return
}

// Validate will throw an error if the Config parameters are whether incomplete or incorrects
func (c Config) Validate() error {
	if validate == nil {
		validate = validator.New()
	}
	return validate.Struct(c)
}
