package handlers

import (
	"github.com/benjaminbartels/brewbot/internal/dynamo"
	"github.com/benjaminbartels/brewbot/internal/platform/discord"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewAPI(bot *discord.Bot, brewRepo dynamo.BrewRepo, logger *logrus.Logger) error {
	brewsHandler := &BrewsHandler{
		Repo:   brewRepo,
		Logger: logger,
	}

	if err := bot.AddCommand(BrewCommand()); err != nil {
		return errors.Wrap(err, "could not add 'log' command")
	}

	bot.AddHandler("brew", brewsHandler.BrewHandler)

	return nil
}
