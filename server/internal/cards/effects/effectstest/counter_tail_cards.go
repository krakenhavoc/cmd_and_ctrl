package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_tail_cards_test.go — #1282 on real cards. Every card here
// sequences a clause behind a counter placement that READS the
// counters, and every pausing test builds a real pausing board: two
// different counter replacements on the same placement, which queues
// the CR 616 ordering prompt and returns with nothing placed.

const EarthshapeOracle = "300f17a2-1594-40cb-9187-952898a83622"

// answerOrderBySource answers the one open CR 616 prompt in the order
// of the permanents that contributed each replacement.
func AnswerOrderBySource(t *testing.T, g *game.Game, sources ...uuid.UUID) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the one CR 616 prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, game.PendingChoiceReplacementOrder)
	}
	order := make([]game.ReplacementEffectID, 0, len(sources))
	for _, src := range sources {
		order = append(order, ReplacementIDForSource(t, g, p.ReplacementEffectIDs, src))
	}
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

func HasKeywordOnBattlefield(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return game.HasKeyword(c, kw)
}

func PlayerHasHexproof(g *game.Game, p *game.Player) bool {
	var has bool
	g.WithWriteLock(func() { has = g.PlayerHasKeywordLocked(p, game.KeywordHexproof) })
	return has
}
