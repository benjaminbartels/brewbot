package handlers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

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
	Repo   dynamo.BrewRepo
	Logger *logrus.Logger
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
	user *discordgo.User, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
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

	if err := h.Repo.Save(ctx, brew); err != nil {
		return errors.Wrap(err, "could not save brew")
	}

	message := fmt.Sprintf("%s brewed %0.2f gallons of %s!", name, floatAmount, style)
	if err := respondToChannel(s, i, message, false); err != nil {
		return errors.Wrap(err, "could not respond with log success message")
	}

	return nil
}

func (h *BrewsHandler) handleList(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
	user *discordgo.User) error {
	brews, err := h.Repo.GetByUserID(ctx, user.ID)
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
		return errors.Wrap(err, "could not respond with list message") // TODO: better errors
	}

	return nil
}

func (h *BrewsHandler) handleDelete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
	user *discordgo.User, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	id := opts[0].StringValue()

	brew, err := h.Repo.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "could not get brew %s", id)
	}

	if brew == nil {
		if err := respondToChannel(s, i, fmt.Sprintf("Brew %s not found", id), true); err != nil {
			return errors.Wrap(err, "could not respond with not found error")
		}

		return nil
	}

	if err := h.Repo.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "could not delete brew %s", id)
	}

	if err := respondToChannel(s, i, fmt.Sprintf("Deleted %s's %s brew", user.Username, id), true); err != nil {
		return errors.Wrap(err, "could not respond with delete success message")
	}

	return nil
}

func (h *BrewsHandler) handleLeaderboard(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate) error {
	volumes := make(map[string]float64)
	counts := make(map[string]int)

	brews, err := h.Repo.GetAll(ctx)
	if err != nil {
		return errors.Wrap(err, "could not get brews")
	}

	if len(brews) == 0 {
		if err := respondToChannel(s, i, "No Brews yet!", true); err != nil {
			return errors.Wrap(err, "could not respond with no brews error")
		}

		return nil
	}

	for _, brew := range brews {
		v, ok := volumes[brew.Username]
		if !ok {
			volumes[brew.Username] = brew.Amount
			counts[brew.Username] = 1
		} else {
			volumes[brew.Username] = v + brew.Amount
			counts[brew.Username]++
		}
	}

	type leaderboardItem struct {
		name   string
		count  int
		volume float64
	}

	leaderboard := make([]leaderboardItem, 0, len(volumes))

	for name, volume := range volumes {
		leaderboard = append(leaderboard, leaderboardItem{name: name, count: counts[name], volume: volume})
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].volume > leaderboard[j].volume
	})

	var builder strings.Builder

	writer := tabwriter.NewWriter(&builder, 0, 5, 2, ' ', 0)

	fmt.Fprintln(writer, "\tName\tCount\tGallons")

	var (
		totalCount  int
		totalVolume float64
	)

	for i, item := range leaderboard {
		fmt.Fprintf(writer, "%d\t%s\t%d\t%6.02f\t\n", i+1, item.name, item.count, item.volume)
		totalCount += item.count
		totalVolume += item.volume
	}

	fmt.Fprintf(writer, "---------------------------------------\n")

	fmt.Fprintf(writer, "Total Batches: %d Total Volume: %6.02f\n", totalCount, totalVolume)

	if err := writer.Flush(); err != nil {
		return errors.Wrap(err, "could not flush to channel")
	}

	message := "Leaderboard:\n"

	message += "```\n" + builder.String() + "```"

	if err := respondToChannel(s, i, message, false); err != nil {
		return errors.Wrap(err, "could not respond with leaderboard")
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
