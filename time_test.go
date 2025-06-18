package main

import (
	"testing"
)

func TestTimeParse(t *testing.T) {
	s := "1750056914.623559"
	mustParseTime(s)
}
