package main

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/urfave/cli/v3"
	"gorm.io/gorm/clause"
)

var collectChannelCmd = &cli.Command{
	Name: "channel",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "dsn",
			Value: "slack-collector.db",
		},
		&cli.StringFlag{
			Name:    "slack-token",
			Sources: cli.NewValueSourceChain(cli.EnvVar("SLACK_TOKEN")),
		},
	},
	Action: func(ctx context.Context, command *cli.Command) error {
		dsn := command.String("dsn")
		slackToken := command.String("slack-token")

		db, err := openDB(dsn)
		if err != nil {
			return err
		}

		collector := Collector{client: slack.New(slackToken)}

		channels, err := collector.ListChannels(ctx)
		if err != nil {
			return err
		}

		for _, c := range channels {
			db.Clauses(clause.OnConflict{UpdateAll: true}).
				Create(&Channel{
					ID:          c.ID,
					Name:        c.Name,
					Creator:     c.Creator,
					NumMembers:  c.NumMembers,
					IsArchived:  c.IsArchived,
					IsChannel:   c.IsChannel,
					IsExtShared: c.IsExtShared,
					IsGroup:     c.IsGroup,
					IsIM:        c.IsIM,
					IsMember:    c.IsMember,
					IsGeneral:   c.IsGeneral,
					IsMpIM:      c.IsMpIM,
					IsPrivate:   c.IsPrivate,
					IsReadOnly:  c.IsReadOnly,
					IsShared:    c.IsShared,
				})
		}

		return nil
	},
}
