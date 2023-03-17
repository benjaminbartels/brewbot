package handlers

import (
	"context"
	"fmt"

	"github.com/benjaminbartels/brewbot/internal/styles"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	styleCommand     = "styles"
	randomSubCommand = "rand"
)

type StylesHandler struct {
	StyleRepo styles.StyleRepo
}

func StyleCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        styleCommand,
		Description: "Issues style realted commands to BrewBot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        randomSubCommand,
				Description: "Pick a random style to brew",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	}
}

func (h *StylesHandler) StyleHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx := context.Background()
	subcommand := i.ApplicationCommandData().Options[0].Name
	user := i.Member.User

	var err error

	switch subcommand {
	case randomSubCommand:
		err = h.handleRandom(ctx, s, i, user)
	}

	if err != nil {
		if err := respondToChannel(s, i, "There was a problem processing your request", true); err != nil {
			return errors.Wrap(err, "could not respond with processing error")
		}

		return errors.Wrap(err, "unhandled error occurred")
	}

	return nil
}

func (h *StylesHandler) handleRandom(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, user *discordgo.User,
) error {
	style := h.StyleRepo.Random(ctx)

	name := user.Username
	if i.Member.Nick != "" {
		name = i.Member.Nick
	}

	message := fmt.Sprintf("%s should brew a %s (%s - %s).", name, style.Name, style.Number, style.CategoryName)

	if err := respondToChannel(s, i, message, false); err != nil {
		return errors.Wrap(err, "could not respond with leaderboard")
	}

	return nil
}
