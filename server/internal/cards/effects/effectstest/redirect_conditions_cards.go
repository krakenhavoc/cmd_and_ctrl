package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// redirect_conditions_cards_test.go — ADR 0108 §9, amendment 2026-10-16
// (#2056): Blood of the Martyr (an optional redirection from every
// creature), Kjeldoran Royal Guard (from unblocked creatures) and Ascent
// of the Worthy (from your creatures, never you), each against the damage
// it must move and the damage it must leave alone.

const (
	RcBloodOfTheMartyr    = "0e69e6c2-fdc5-4c1a-bd96-c4283b82743a"
	RcKjeldoranRoyalGuard = "8137e765-4df9-469f-a527-dee91d58fb7e"
	RcAscentOfTheWorthy   = "d748bad4-dd4c-4553-9fcc-e462260f6ff3"
)

// rcMayPrompt is the open optional_replacement prompt for `chooser`.
func RcMayPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement {
			if c.Chooser != chooser {
				t.Fatalf("the may is asked of %v, want %v", c.Chooser, chooser)
			}
			return c
		}
	}
	t.Fatal("no optional_replacement prompt open")
	return nil
}
