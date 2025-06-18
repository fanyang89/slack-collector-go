package main

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/cespare/xxhash"
)

type Message struct {
	ID        uint64 `gorm:"primary_key;autoIncrement"`
	Hash      string
	Type      string
	Timestamp time.Time
	Channel   string
	User      string
	Text      string
	IsStared  bool
	Team      string
}

func hashString(s string) string {
	h := xxhash.New()
	_, err := io.WriteString(h, s)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", h.Sum64())
}

func (m *Message) String() string {
	return fmt.Sprintf("%s,%s:%s:%s:%s", m.Timestamp.Format(time.RFC3339),
		m.Type, m.Channel, m.User, hashString(m.Text))
}

func (m *Message) WithHash() *Message {
	m.Hash = hashString(m.String())
	return m
}

func mustParseTime(s string) time.Time {
	x, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return time.Time{}
	}
	return time.UnixMicro(int64(x * 1000000))
}

type CollectedType = string

const (
	CollectedMessage CollectedType = "Message"
)

type Collected struct {
	ID      uint64 `gorm:"primary_key;autoIncrement"`
	Channel string
	Type    CollectedType
	Day     time.Time
}

type Channel struct {
	ID          string `gorm:"primary_key"`
	Name        string
	Creator     string
	IsArchived  bool
	IsChannel   bool
	IsGeneral   bool
	IsMember    bool
	IsGroup     bool
	IsExtShared bool
	IsIM        bool
	IsMpIM      bool
	IsPrivate   bool
	IsReadOnly  bool
	IsShared    bool
	NumMembers  int
}

type User struct {
	ID             string `gorm:"primary_key"`
	Name           string
	Deleted        bool
	RealName       string
	IsBot          bool
	IsAdmin        bool
	IsOwner        bool
	IsPrimaryOwner bool
	TeamID         string
}

type Config struct{}
