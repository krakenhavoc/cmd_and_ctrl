package effectstest

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// appAnswerOffer answers every link of the open any-player offer in
// turn: a seat in `payers` pays, everyone else declines. It returns the
// seats asked, in order, and settles the stack afterwards.
func AppAnswerOffer(t *testing.T, g *game.Game, payers ...uuid.UUID) []uuid.UUID {
	t.Helper()
	var asked []uuid.UUID
	for i := 0; i < 8; i++ {
		var link *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePayUnless && c.AnyPlayerOffer != nil {
				link = c
				break
			}
		}
		if link == nil {
			PassPriorityAroundTable(t, g)
			return asked
		}
		asked = append(asked, link.Chooser)
		if err := g.ResolvePayUnless(link.ID, link.Chooser, slices.Contains(payers, link.Chooser)); err != nil {
			t.Fatalf("answer link %d: %v", i, err)
		}
	}
	t.Fatal("the offer never ended")
	return nil
}

func AppMana(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}
