package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/urfave/cli/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func isCollected(db *gorm.DB, channelName string, collectType CollectedType, day time.Time) (bool, error) {
	err := db.Where(&Collected{Channel: channelName, Type: collectType, Day: day}).
		First(&Collected{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

var collectMessageCmd = &cli.Command{
	Name: "message",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "dsn",
			Value: "slack-collector.db",
		},
		&cli.StringFlag{
			Name:    "slack-token",
			Sources: cli.NewValueSourceChain(cli.EnvVar("SLACK_TOKEN")),
		},
		&cli.TimestampFlag{
			Name:  "latest",
			Value: initNow,
			Config: cli.TimestampConfig{
				Timezone: time.Local,
				Layouts:  []string{time.DateOnly},
			},
		},
		&cli.TimestampFlag{
			Name:  "oldest",
			Value: initNow,
			Config: cli.TimestampConfig{
				Timezone: time.Local,
				Layouts:  []string{time.DateOnly},
			},
		},
		&cli.IntFlag{
			Name:  "batch",
			Value: 10,
		},
	},
	Action: func(ctx context.Context, command *cli.Command) error {
		dsn := command.String("dsn")
		slackToken := command.String("slack-token")
		latest := command.Timestamp("latest")
		oldest := command.Timestamp("oldest")
		batch := command.Int("batch")
		log.Info().Time("latest", latest).Time("oldest", oldest).Msgf("Collecting messages")

		if latest == oldest {
			log.Info().Msg("Nothing to do, please enter valid time range")
			return nil
		}

		db, err := openDB(dsn)
		if err != nil {
			return err
		}

		var channels []Channel
		err = db.Find(&channels).Error
		if len(channels) == 0 || err != nil {
			log.Error().Err(err).Msg("Get channels failed, please collect channels first")
			return nil
		}

		collector := Collector{client: slack.New(slackToken)}

		from := oldest
		for _, channel := range channels {
			cnt := uint64(0)
			fmt.Printf("Collecting channel: %s\n", channel.Name)
			bar := newProgressBar(-1, "")

			to := from.Add(time.Duration(batch) * oneDay)
			if to.After(latest) {
				to = latest
			}
			if from == to || from.After(to) {
				break
			}

			bar.Describe(fmt.Sprintf("Collecting from %s to %s",
				from.Format(time.DateOnly), to.Format(time.DateOnly)))

			if ok, err := isCollected(db, CollectedMessage, channel.ID, from); ok {
				from.Add(oneDay)
				continue
			} else if err != nil {
				return err
			}

			err = collector.IterMessages(ctx, channel.ID, from, to, func(m *slack.Message) error {
				mm := &Message{
					Channel:   m.Channel,
					User:      m.User,
					Text:      m.Text,
					Timestamp: mustParseTime(m.Timestamp),
					Type:      m.Type,
					IsStared:  m.IsStarred,
					Team:      m.Team,
				}
				mm = mm.WithHash()

				err = db.Clauses(clause.OnConflict{UpdateAll: true}).Create(mm).Error
				if err != nil {
					return err
				}

				cnt++
				return bar.Add(1)
			})
			if err != nil {
				return err
			}

			from.Add(oneDay)
			err = db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&Collected{
				Channel: channel.ID,
				Type:    CollectedMessage,
				Day:     from,
			}).Error
			if err != nil {
				return err
			}
			_ = bar.Finish()
			fmt.Printf("Collected %d messages\n", cnt)
		}

		return nil
	},
}
