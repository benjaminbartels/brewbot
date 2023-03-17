package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/benjaminbartels/brewbot/internal/styles"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	styleCommand     = "styles"
	randomSubCommand = "rand"
	infoSubCommand   = "info"
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
				Description: "Pick a random beer style to brew from the 2021 BJCP style guide",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        infoSubCommand,
				Description: "Get information about a beer style from the 2021 BJCP style guide",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "number",
						Description: "BJCP category number",
						Required:    true,
					},
				},
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
	case infoSubCommand:
		err = h.handleInfo(ctx, s, i, user)
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

	message := fmt.Sprintf("%s should brew a %s (%s - %s).", name, style.Name, style.Number, style.Category)

	if err := respondToChannel(s, i, message, false); err != nil {
		return errors.Wrap(err, "could not respond with leaderboard")
	}

	return nil
}

func (h *StylesHandler) handleInfo(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, user *discordgo.User, opts []*discordgo.ApplicationCommandInteractionDataOption,
) error {
	number := opts[0].StringValue()

	style := h.StyleRepo.Get(ctx, number)
	if style == nil {
		if err := respondToChannel(s, i, fmt.Sprintf("Style %s not found", number), true); err != nil {
			return errors.Wrap(err, "could not respond with not found error")
		}

		return nil
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("**%s** (%s)\n", style.Name, style.Number))
	builder.WriteString(fmt.Sprintf("**Category:** %s (%s)\n", style.Category, style.CategoryNumber))
	builder.WriteString("__**Overall Impression:**__\n")
	builder.WriteString(fmt.Sprintf("%s\n", style.OverallImpression))
	builder.WriteString("__**Aroma:**__\n")
	builder.WriteString(fmt.Sprintf("%s\n", style.Aroma))
	builder.WriteString("__**Appearance:**__\n")
	builder.WriteString(fmt.Sprintf("%s\n", style.Appearance))
	builder.WriteString("__**Flavor:**__\n")
	builder.WriteString(fmt.Sprintf("%s\n", style.Flavor))
	builder.WriteString("__**Mouthfeel:**__\n")
	builder.WriteString(fmt.Sprintf("%s\n", style.Mouthfeel))
	builder.WriteString("__**Examples:**__\n")
	builder.WriteString(fmt.Sprintf("%s\n", style.CommercialExamples))

	if err := respondToChannel(s, i, builder.String(), true); err != nil {
		return errors.Wrap(err, "could not respond with leaderboard")
	}

	return nil
}
