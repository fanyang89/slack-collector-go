package main

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/urfave/cli/v3"
	"gorm.io/gorm/clause"
)

var collectUserCmd = &cli.Command{
	Name: "user",
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
		users, err := collector.GetUsers()
		if err != nil {
			return err
		}

		for _, user := range users {
			err = db.
				Clauses(clause.OnConflict{UpdateAll: true}).
				Create(&User{
					ID:             user.ID,
					TeamID:         user.TeamID,
					Name:           user.Name,
					RealName:       user.RealName,
					Deleted:        user.Deleted,
					IsBot:          user.IsBot,
					IsAdmin:        user.IsAdmin,
					IsOwner:        user.IsOwner,
					IsPrimaryOwner: user.IsPrimaryOwner,
				}).Error
			if err != nil {
				return err
			}
		}

		return nil
	},
}
