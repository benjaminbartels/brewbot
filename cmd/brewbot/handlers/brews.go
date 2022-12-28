package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/benjaminbartels/brewbot/internal/dynamo"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	brewCommand           = "brew"
	logSubCommand         = "log"
	listSubCommand        = "list"
	deleteSubCommand      = "delete"
	leaderboardSubCommand = "leaderboard"
)

type BrewsHandler struct {
	BrewRepo          dynamo.BrewRepo
	LeaderboardRepo   dynamo.LeaderboardRepo
	LeaderboardCutoff time.Time
	Logger            *logrus.Logger
}

func BrewCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        brewCommand,
		Description: "Issues commands to BrewBot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        logSubCommand,
				Description: "Log your homebrew",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "style",
						Description: "Beer style",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "amount",
						Description: "How many gallons?",
						Required:    true,
					},
				},
			},
			{
				Name:        listSubCommand,
				Description: "List your homebrews",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        deleteSubCommand,
				Description: "Delete a homebrew",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "id",
						Description: "ID of homebrew",
						Required:    true,
					},
				},
			},
			{
				Name:        leaderboardSubCommand,
				Description: "Show the leaderboard",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	}
}

func (h *BrewsHandler) BrewHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx := context.Background()
	subcommand := i.ApplicationCommandData().Options[0].Name
	user := i.Member.User
	opts := i.ApplicationCommandData().Options[0].Options

	var err error

	switch subcommand {
	case logSubCommand:
		err = h.handleLog(ctx, s, i, user, opts)
	case listSubCommand:
		err = h.handleList(ctx, s, i, user)
	case deleteSubCommand:
		err = h.handleDelete(ctx, s, i, user, opts)
	case leaderboardSubCommand:
		err = h.handleLeaderboard(ctx, s, i)
	}

	if err != nil {
		if err := respondToChannel(s, i, "There was a problem processing your request", true); err != nil {
			return errors.Wrap(err, "could not respond with processing error")
		}

		return errors.Wrap(err, "unhandled error occurred")
	}

	return nil
}

func (h *BrewsHandler) handleLog(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
	user *discordgo.User, opts []*discordgo.ApplicationCommandInteractionDataOption,
) error {
	style := opts[0].StringValue()

	floatAmount, err := strconv.ParseFloat(opts[1].StringValue(), 64)
	if err != nil {
		if err := respondToChannel(s, i, fmt.Sprintf("Invalid amount: %s", opts[1].Value), true); err != nil {
			return errors.Wrap(err, "could not respond with invalid amount error")
		}

		return nil
	}

	name := user.Username
	if i.Member.Nick != "" {
		name = i.Member.Nick
	}

	brew := &dynamo.Brew{
		UserID:   user.ID,
		Username: name,
		Style:    style,
		Amount:   floatAmount,
	}

	if err := h.BrewRepo.Save(ctx, brew); err != nil {
		return errors.Wrap(err, "could not save brew")
	}

	if err := h.refreshLeaderboard(ctx, user.ID); err != nil {
		return errors.Wrapf(err, "could not refresh leaderboard for user %s", user.ID)
	}

	message := fmt.Sprintf("%s brewed %0.2f gallons of %s!", name, floatAmount, style)
	if err := respondToChannel(s, i, message, false); err != nil {
		return errors.Wrap(err, "could not respond with log success message")
	}

	return nil
}

func (h *BrewsHandler) handleList(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
	user *discordgo.User,
) error {
	brews, err := h.BrewRepo.GetByUserID(ctx, user.ID, h.LeaderboardCutoff.String())
	if err != nil {
		return errors.Wrapf(err, "could not get brews for user %s", user.Username)
	}

	if len(brews) == 0 {
		if err := respondToChannel(s, i, "No Brews yet!", true); err != nil {
			return errors.Wrap(err, "could not respond with no brews error")
		}

		return nil
	}

	var builder strings.Builder

	writer := tabwriter.NewWriter(&builder, 0, 5, 2, ' ', 0)

	fmt.Fprintln(writer, "Style\tAmount\tID\t")

	for _, brew := range brews {
		fmt.Fprintf(writer, "%s\t%6.02f\t%s\t\n", brew.Style, brew.Amount, brew.ID)
	}

	if err := writer.Flush(); err != nil {
		return errors.Wrap(err, "could not flush to writer")
	}

	message := "Your Brew Log:"

	message += "```\n" + builder.String() + "```"

	if err := respondToChannel(s, i, message, true); err != nil {
		return errors.Wrap(err, "could not respond with list message")
	}

	return nil
}

func (h *BrewsHandler) handleDelete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
	user *discordgo.User, opts []*discordgo.ApplicationCommandInteractionDataOption,
) error {
	id := opts[0].StringValue()

	brew, err := h.BrewRepo.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "could not get brew %s", id)
	}

	if brew == nil {
		if err := respondToChannel(s, i, fmt.Sprintf("Brew %s not found", id), true); err != nil {
			return errors.Wrap(err, "could not respond with not found error")
		}

		return nil
	}

	if err := h.BrewRepo.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "could not delete brew %s", id)
	}

	if err := h.refreshLeaderboard(ctx, brew.UserID); err != nil {
		return errors.Wrapf(err, "could not refresh leaderboard for user %s", brew.UserID)
	}

	if err := respondToChannel(s, i, fmt.Sprintf("Deleted %s's %s brew", user.Username, id), true); err != nil {
		return errors.Wrap(err, "could not respond with delete success message")
	}

	return nil
}

func (h *BrewsHandler) handleLeaderboard(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	leaderboardEntries, err := h.LeaderboardRepo.GetAll(ctx)
	if err != nil {
		return errors.Wrap(err, "could not get brews")
	}

	if len(leaderboardEntries) == 0 {
		if err := respondToChannel(s, i, "No Brews yet!", true); err != nil {
			return errors.Wrap(err, "could not respond with no brews error")
		}

		return nil
	}

	var builder strings.Builder

	writer := tabwriter.NewWriter(&builder, 0, 5, 2, ' ', 0)

	fmt.Fprintln(writer, "\tName\tCount\tGallons")

	var (
		totalCount  int
		totalVolume float64
	)

	for i, entry := range leaderboardEntries {
		fmt.Fprintf(writer, "%d\t%s\t%d\t%6.02f\t\n", i+1, entry.Username, entry.Count, entry.Volume)
		totalCount += entry.Count
		totalVolume += entry.Volume
	}

	fmt.Fprintf(writer, "---------------------------------------\n")

	fmt.Fprintf(writer, "Total Batches: %d Total Volume: %6.02f\n", totalCount, totalVolume)

	if err := writer.Flush(); err != nil {
		return errors.Wrap(err, "could not flush to channel")
	}

	fmt.Fprintf(writer, "Total Batches: %d\t Total Volume: %6.02f\t\n", totalCount, totalVolume)

	message := "Leaderboard:\n"

	message += "```\n" + builder.String() + "```"

	if err := respondToChannel(s, i, message, false); err != nil {
		return errors.Wrap(err, "could not respond with leaderboard")
	}

	return nil
}

func (h *BrewsHandler) refreshLeaderboard(ctx context.Context, userID string) error {
	entry, err := h.LeaderboardRepo.Get(ctx, userID)
	if err != nil {
		return errors.Wrapf(err, "could not get brew for user %s", userID)
	}

	entry.Count = 0
	entry.Volume = 0

	brews, err := h.BrewRepo.GetByUserID(ctx, userID, h.LeaderboardCutoff.String())
	if err != nil {
		return errors.Wrapf(err, "could not get brew for user %s", userID)
	}

	for _, brew := range brews {
		entry.Count++

		entry.Volume += brew.Amount
	}

	if err := h.LeaderboardRepo.Save(ctx, entry); err != nil {
		return errors.Wrapf(err, "could not get save LeaderboardEntry for %s", userID)
	}

	return nil
}

func respondToChannel(s *discordgo.Session, i *discordgo.InteractionCreate, message string, isEphemeral bool) error {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	}

	if isEphemeral {
		//nolint: gomnd
		response.Data.Flags = 1 << 6
	}

	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		return errors.Wrap(err, "could not send interaction response")
	}

	return nil
}
