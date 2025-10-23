package maplebot

import (
	"testing"
)

func TestRate(t *testing.T) {
	_, _, count := calculateStarForce(false, 0, 12, 160, false, false, false, false)
	t.Log(count)
}
