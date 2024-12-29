package handlers

import (
	"time"

	"github.com/benjaminbartels/brewbot/internal/dynamo"
	"github.com/benjaminbartels/brewbot/internal/platform/discord"
	"github.com/benjaminbartels/brewbot/internal/styles"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewAPI(bot *discord.Bot, brewRepo dynamo.BrewRepo, leaderboardRepo dynamo.LeaderboardRepo,
	stylesRepo styles.StyleRepo, leaderboardCutoff time.Time, logger *logrus.Logger,
) error {
	brewsHandler := &BrewsHandler{
		BrewRepo:          brewRepo,
		LeaderboardRepo:   leaderboardRepo,
		LeaderboardCutoff: leaderboardCutoff,
		Logger:            logger,
	}

	stylesHandler := &StylesHandler{
		StyleRepo: stylesRepo,
	}

	if err := bot.AddCommand(BrewCommand()); err != nil {
		return errors.Wrap(err, "could not add 'brew' command")
	}

	if err := bot.AddCommand(StyleCommand()); err != nil {
		return errors.Wrap(err, "could not add 'style' command")
	}

	if err := bot.AddCommand(UntapddCommand()); err != nil {
		return errors.Wrap(err, "could not add 'untapdd' command")
	}

	untapddHandler := NewUntapddHandler()

	bot.AddHandler("brew", brewsHandler.BrewHandler)
	bot.AddHandler("styles", stylesHandler.StyleHandler)
	bot.AddHandler("untapdd", untapddHandler.UntapddHandler)

	return nil
}
