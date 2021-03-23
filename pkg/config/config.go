package config

import (
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

type Provider struct {
	Type   string `validate:"oneof=github gitlab"`
	URL    string
	Token  string   `validate:"required"`
	Owners []string `validate:"gt=0"`
}

type Providers []Provider

type Log struct {
	Level  string `default:"info" validate:"required,oneof=trace debug info warning error fatal panic"`
	Format string `default:"text" validate:"oneof=text json"`
}

type Slack struct {
	Token         string `validate:"required"`
	SigningSecret string `validate:"required" json:"signing_secret" yaml:"signing_secret"`
}

type User struct {
	Email   string   `validate:"required,email"`
	Aliases []string `validate:"gt=0"`
}

type Users []User

type Config struct {
	Providers     Providers `validate:"gt=0,unique=Type"`
	ListenAddress string    `default:":8080" validate:"required"`
	Log           Log
	Slack         Slack
	Users         Users
}

func NewConfig() (cfg Config) {
	defaults.Set(&cfg)
	return
}

func (c Config) Validate() error {
	if validate == nil {
		validate = validator.New()
	}
	return validate.Struct(c)
}
