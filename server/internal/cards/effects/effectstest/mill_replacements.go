package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_replacements_test.go — #569 from the card side: the CR 614
// window on the mill AMOUNT, watched through the ordinary
// Spec.Replacements slot.
//
// Two printed cards, deliberately, so the kind is not shaped around
// one of them: Bruvac the Grandiloquent's ×2 and The Water Crystal's
// +4 are the same sentence with different arithmetic, and CR 616.1's
// ordering question between them has two different answers.
//
// It is also the first printed board on which #982's rule is
// observable. Both replacements belong to whoever controls the
// artifacts; the affected player is the OPPONENT being milled, so the
// ordering prompt goes to them.

const (
	BruvacOracle          = "274b999f-f193-48fd-9a4a-0fdaf535e6c3"
	TheWaterCrystalOracle = "f8d2a94f-7be1-4b17-8556-398dde531360"
)

// pushMillReplacement puts one of the two cards onto the battlefield
// under `owner`.
func PushMillReplacement(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return PushCatalogPermanent(g, owner, name, typeLine, oracle, false)
}

// millSeat runs a fire-and-forget mill of n cards against `seat` and
// reports how many cards reached their graveyard.
func MillSeat(t *testing.T, g *game.Game, seat *game.Player, n int) int {
	t.Helper()
	before := seat.Graveyard.Size()
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(seat.ID, n); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	return seat.Graveyard.Size() - before
}

// stockLibraryFor makes sure a seat has at least n cards to mill, so a
// count test measures the replacement and not the deck size.
func StockLibraryFor(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for p.Library.Size() < n {
			p.Library.PushTop(game.NewCard("basic-filler", p.ID))
		}
	})
}

// replacementOrderByLabel turns prompt labels into the ID permutation
// ResolveReplacementOrder takes.
func ReplacementOrderByLabel(t *testing.T, g *game.Game, c *game.PendingChoice, labels ...string) []game.ReplacementEffectID {
	t.Helper()
	byLabel := make(map[string]game.ReplacementEffectID, len(c.ReplacementEffectIDs))
	for _, id := range c.ReplacementEffectIDs {
		label, _ := g.ReplacementOptionMetaForEffect(id)
		byLabel[label] = id
	}
	out := make([]game.ReplacementEffectID, 0, len(labels))
	for _, want := range labels {
		id, ok := byLabel[want]
		if !ok {
			t.Fatalf("prompt has no option labelled %q (has %v)", want, byLabel)
		}
		out = append(out, id)
	}
	return out
}
