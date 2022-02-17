package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type HandlerFunc func(s *discordgo.Session, i *discordgo.InteractionCreate) error

type Bot struct {
	session  *discordgo.Session
	guildID  string
	handlers map[string]HandlerFunc
	logger   *logrus.Logger
}

func NewBot(session *discordgo.Session, guildID string, logger *logrus.Logger) *Bot {
	bot := &Bot{
		session:  session,
		guildID:  guildID,
		logger:   logger,
		handlers: make(map[string]HandlerFunc),
	}

	// session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
	// 	logger.Info("Bot is up!")
	// })

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := bot.handlers[i.ApplicationCommandData().Name]; ok {
			if err := handler(s, i); err != nil {
				logger.WithError(err).Errorf("could not handle '%s' command", i.ApplicationCommandData().Name)
			}
		}
	})

	return bot
}

func (b *Bot) AddCommand(command *discordgo.ApplicationCommand) error {
	if _, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.guildID, command); err != nil {
		return errors.Wrapf(err, "could not create Discord command %s", command.Name)
	}

	return nil
}

func (b *Bot) AddHandler(name string, handlerFunc HandlerFunc) {
	b.handlers[name] = handlerFunc
}

func (b *Bot) RemoveAllCommands() error {
	commands, err := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
	if err != nil {
		return errors.Wrapf(err, "could not get Discord commands to delete")
	}

	for _, command := range commands {
		if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, command.ID); err != nil {
			return errors.Wrapf(err, "could not delete Discord command %s", command.Name)
		}
	}
	return nil
}
