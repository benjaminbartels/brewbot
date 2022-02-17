package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/benjaminbartels/brewbot/cmd/brewbot/handlers"
	"github.com/benjaminbartels/brewbot/internal/dynamo"
	c "github.com/benjaminbartels/brewbot/internal/platform/context"
	"github.com/benjaminbartels/brewbot/internal/platform/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const localDynamoEndpoint = "http://dynamo:8000"

type config struct {
	AWSRegion      string `default:"us-west-2"`
	BrewTableName  string `default:"brewbot-brews"`
	UseLocalDynamo bool   `default:"false"`
	DiscordToken   string `required:"true"`
	DiscordGuildID string `required:"true"`
	Debug          bool   `default:"false"`
}

func main() {
	logger := logrus.New()

	var cfg config

	if err := envconfig.Process("brewbot", &cfg); err != nil {
		logger.WithError(err).Error("could not process env vars")
	}

	if err := run(logger, cfg); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

func run(logger *logrus.Logger, cfg config) error {
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	ctx, interruptCancel := c.WithInterrupt(context.Background())
	defer interruptCancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return errors.Wrap(err, "could not create Discord session")
	}

	if err := session.Open(); err != nil {
		return errors.Wrap(err, "could not open Discord session")
	}

	defer func() {
		session.Close()
	}()

	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if cfg.UseLocalDynamo {
				return aws.Endpoint{
					URL: localDynamoEndpoint,
				}, nil
			}

			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(cfg.AWSRegion),
		awsConfig.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		return errors.Wrap(err, "could not load AWS SDK config")
	}

	if cfg.UseLocalDynamo {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider("test", "test", "")
	}

	brewRepo := dynamo.NewBrewRepo(dynamodb.NewFromConfig(awsCfg), cfg.BrewTableName)

	bot := discord.NewBot(session, cfg.DiscordGuildID, logger)

	if err := handlers.NewAPI(bot, brewRepo, logger); err != nil {
		return errors.Wrap(err, "could not create new API")
	}

	logger.Infof("brewbot started")

	defer logger.Info("brewbot stopped 👋!")

	<-ctx.Done()

	if err := bot.RemoveAllCommands(); err != nil {
		return errors.Wrap(err, "could not remove all commands")
	}

	return nil
}
