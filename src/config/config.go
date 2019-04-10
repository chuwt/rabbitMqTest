package config

import (
    "github.com/jinzhu/configor"
    "github.com/rs/zerolog/log"
)

var Config = struct {
    RabbitMQ struct {
        URL string `required:"true" env:"URL"`
        ErrorRepeat int `env:"ErrorRepeat" default:"3"`
    }
}{}

func init() {
    if err := configor.New(&configor.Config{ENVPrefix: ""}).Load(&Config, "./src/config/config.yaml"); err != nil {
        log.Error().Str("err", err.Error()).Msg("init log error")
    }
}