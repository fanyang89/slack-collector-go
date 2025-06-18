package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

func checkRateLimitError(err error) error {
	var rateLimit *slack.RateLimitedError
	if errors.As(err, &rateLimit) {
		rest := rateLimit.RetryAfter + time.Second // one more second
		log.Info().Err(err).Dur("sleep", rest).Msg("Rate limit exceeded, sleeping")
		time.Sleep(rest)
		return nil
	}
	return err
}

type Collector struct {
	client *slack.Client
}

func (c *Collector) ListChannels(ctx context.Context) (channels []slack.Channel, err error) {
	channels = make([]slack.Channel, 0)
	var nextCursor string
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}

		var cs []slack.Channel
		cs, nextCursor, err = c.client.GetConversations(&slack.GetConversationsParameters{
			Cursor: nextCursor,
		})
		err = checkRateLimitError(err)
		if err != nil {
			return nil, err
		}

		if len(cs) != 0 {
			channels = append(channels, cs...)
		}

		if nextCursor == "" {
			break
		}
	}
	return
}

func (c *Collector) GetUsers() ([]slack.User, error) {
	users, err := c.client.GetUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

type IterMessagesFn = func(message *slack.Message) error

func (c *Collector) IterMessages(ctx context.Context, channelID string, oldest, latest time.Time, callback IterMessagesFn) (err error) {
	var cursor string
	for {
		var rsp *slack.GetConversationHistoryResponse
		rsp, err = c.client.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Cursor:    cursor,
			Oldest:    fmt.Sprintf("%v", oldest.Unix()),
			Latest:    fmt.Sprintf("%v", latest.Unix()),
		})
		err = checkRateLimitError(err)
		if err != nil {
			return
		}

		if rsp != nil {
			for _, m := range rsp.Messages {
				err = callback(&m)
				if err != nil {
					return
				}
			}

			cursor = rsp.ResponseMetadata.Cursor
		}

		if cursor == "" {
			break
		}
	}

	return
}
