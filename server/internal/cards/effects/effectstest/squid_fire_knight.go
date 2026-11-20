package effectstest

import (
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sfkPrompt is the open dice prompt with the key, or nil.
func SfkPrompt(g *game.Game, key string) *game.PendingChoice {
	var out *game.PendingChoice
	g.WithWriteLock(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick && c.PickThen == key {
				cp := *c
				out = &cp
				return
			}
		}
	})
	return out
}

// sfkRolls reads the dice values off an ignore prompt's options.
func SfkRolls(t *testing.T, c *game.PendingChoice) []int {
	t.Helper()
	out := make([]int, 0, len(c.PickOptions))
	for _, o := range c.PickOptions {
		f := strings.Fields(o.Label) // "Ignore the 7 (die 2)"
		if len(f) < 3 {
			t.Fatalf("option %q", o.Label)
		}
		v, err := strconv.Atoi(f[2])
		if err != nil {
			t.Fatalf("option %q: %v", o.Label, err)
		}
		out = append(out, v)
	}
	return out
}

func SfkEventCount(g *game.Game) int {
	n := 0
	g.WithWriteLock(func() { n = len(g.Events) })
	return n
}

func SfkAnswer(t *testing.T, g *game.Game, c *game.PendingChoice, by uuid.UUID, index int) {
	t.Helper()
	if err := g.ResolveOptionPick(c.ID, by, index); err != nil {
		t.Fatalf("answer %q: %v", c.Reason, err)
	}
}
