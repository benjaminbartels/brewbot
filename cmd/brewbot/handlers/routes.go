package handlers

import (
	"time"

	"github.com/benjaminbartels/brewbot/internal/dynamo"
	"github.com/benjaminbartels/brewbot/internal/platform/discord"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewAPI(bot *discord.Bot, brewRepo dynamo.BrewRepo, leaderboardRepo dynamo.LeaderboardRepo,
	leaderboardCutoff time.Time, logger *logrus.Logger,
) error {
	brewsHandler := &BrewsHandler{
		BrewRepo:          brewRepo,
		LeaderboardRepo:   leaderboardRepo,
		LeaderboardCutoff: leaderboardCutoff,
		Logger:            logger,
	}

	if err := bot.AddCommand(BrewCommand()); err != nil {
		return errors.Wrap(err, "could not add 'log' command")
	}

	bot.AddHandler("brew", brewsHandler.BrewHandler)

	return nil
}
