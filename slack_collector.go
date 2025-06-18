package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fioepq9/pzlog"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

var rootCmd = &cli.Command{
	Name: "slack-collector",
	Commands: []*cli.Command{
		collectCmd,
	},
}

var collectCmd = &cli.Command{
	Name: "collect",
	Commands: []*cli.Command{
		collectChannelCmd,
		collectMessageCmd,
		collectUserCmd,
	},
}

func main() {
	log.Logger = zerolog.New(pzlog.NewPtermWriter()).With().Timestamp().Caller().Stack().Logger()

	err := godotenv.Load(".env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = rootCmd.Run(ctx, os.Args)
	if err != nil {
		log.Error().Err(err).Msg("failed to run command")
	}
}
