package handlers

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/benjaminbartels/brewbot/internal/untappd"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	untapddCommand               = "untapdd"
	menusSubCommand              = "menu"
	untapddLeaderboardSubCommand = "leaderboard"
	listVenuesSubCommand         = "venues"
)

type UntapddHandler struct {
	venues map[string]string
}

func NewUntapddHandler() UntapddHandler {
	h := UntapddHandler{
		venues: make(map[string]string),
	}

	h.venues["hbs"] = "homebrewstuff-craft-bottle-shop-and-taproom/960464"
	h.venues["brownbeard"] = "brown-beard-brewing-company/11844307"

	return h
}

func UntapddCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        untapddCommand,
		Description: "Issues untapdd realted commands to BrewBot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        menusSubCommand,
				Description: "Get a venue's current Untappd Menu",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "Name of the venue",
						Required:    true,
					},
				},
			},
			{
				Name:        untapddLeaderboardSubCommand,
				Description: "Get a venue's current Untappd Leaderboard",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "Name of the venue",
						Required:    true,
					},
				},
			},
			{
				Name:        listVenuesSubCommand,
				Description: "List available Untappd venues",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	}
}

func (h *UntapddHandler) UntapddHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx := context.Background()
	subcommand := i.ApplicationCommandData().Options[0].Name
	opts := i.ApplicationCommandData().Options[0].Options

	var err error

	switch subcommand {
	case menusSubCommand:
		err = h.handleMenu(ctx, s, i, opts)
	case untapddLeaderboardSubCommand:
		err = h.handleLeaderboard(ctx, s, i, opts)
	case listVenuesSubCommand:
		err = h.handleListVenues(ctx, s, i)
	}

	if err != nil {
		if err := respondToChannel(s, i, "There was a problem processing your request", true); err != nil {
			return errors.Wrap(err, "could not respond with processing error")
		}

		return errors.Wrap(err, "unhandled error occurred")
	}

	return nil
}

func (h *UntapddHandler) handleMenu(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption,
) error {
	venueName := opts[0].StringValue()

	path, ok := h.venues[venueName]
	if !ok {
		if err := respondToChannel(s, i, fmt.Sprintf("Invalid venue: %s", opts[1].Value), true); err != nil {
			return errors.Wrap(err, "could not respond with invalid amount error")
		}
	}

	menus, _, err := untappd.Scrape(path)
	if err != nil {
		return errors.Wrap(err, "could not get scrape Untapdd")
	}

	var builder strings.Builder

	for _, menu := range menus {
		builder.WriteString(fmt.Sprintf("__**%s**__\n", menu.Name))
		for _, i := range menu.Items {
			builder.WriteString(fmt.Sprintf("**%s** *%s* (%s) %s ABV - %s IBU\n",
				i.Name, i.Brewery, i.Style, i.ABV, i.IBU))
		}
	}

	if err := respondToChannel(s, i, builder.String(), false); err != nil {
		return errors.Wrap(err, "could not respond with menu")
	}

	return nil
}

func (h *UntapddHandler) handleLeaderboard(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption,
) error {
	venueName := opts[0].StringValue()

	path, ok := h.venues[venueName]
	if !ok {
		if err := respondToChannel(s, i, fmt.Sprintf("Invalid venue: %s", opts[1].Value), true); err != nil {
			return errors.Wrap(err, "could not respond with invalid amount error")
		}
	}

	_, patrons, err := untappd.Scrape(path)
	if err != nil {
		return errors.Wrap(err, "could not get scrape Untapdd")
	}

	var builder strings.Builder

	writer := tabwriter.NewWriter(&builder, 0, 4, 2, ' ', 0)

	fmt.Fprintln(writer, "\tName\tCheck-Ins")

	for _, p := range patrons {
		fmt.Fprintf(writer, "%d\t%s\t%d\t\n", p.Rank, p.Name, p.CheckIns)
	}

	if err := writer.Flush(); err != nil {
		return errors.Wrap(err, "could not flush to channel")
	}

	message := fmt.Sprintf("%s Top Check-ins", strings.ToUpper(venueName))

	message += "```\n" + builder.String() + "```"

	if err := respondToChannel(s, i, message, false); err != nil {
		return errors.Wrap(err, "could not respond with leaderboard")
	}

	return nil
}

func (h *UntapddHandler) handleListVenues(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	var builder strings.Builder
	for k := range h.venues {
		builder.WriteString(fmt.Sprintf("%s\n", k))
	}

	if err := respondToChannel(s, i, builder.String(), false); err != nil {
		return errors.Wrap(err, "could not respond with leaderboard")
	}

	return nil
}
