package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optionFrom is the ID of the ordering prompt's option whose source is
// `src`.
func OptionFrom(t *testing.T, g *game.Game, order *game.PendingChoice, src uuid.UUID) game.ReplacementEffectID {
	t.Helper()
	for _, id := range order.ReplacementEffectIDs {
		var got uuid.UUID
		g.ReadSnapshot(func() { _, got = g.ReplacementOptionMetaForEffect(id) })
		if got == src {
			return id
		}
	}
	t.Fatalf("no option on the prompt comes from %s: %v", src, order.ReplacementEffectIDs)
	return 0
}

// shieldedBear puts a 2/2 Bear under the opponent's control and gives it
// one regeneration shield; it returns the Bear.
func ShieldedBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	bear := B12Creature(g, owner, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.RegenerateForEffect(bear) })
	return bear
}

// destroyAndOrder destroys `victim` and returns the opponent's CR 616
// ordering prompt: the replacer against the victim's protection.
func DestroyAndOrder(t *testing.T, g *game.Game, victim uuid.UUID, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	order := B06ReplacementOrderFor(g, chooser)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("want the victim's controller asked to order two replacements (CR 616.1), got %+v", g.PendingChoices)
	}
	return order
}

func AnswerOrder(t *testing.T, g *game.Game, order *game.PendingChoice, chooser uuid.UUID, first game.ReplacementEffectID) {
	t.Helper()
	if err := g.ResolveReplacementOrder(order.ID, chooser, OrderWith(order, first)); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if left := B06ReplacementOrderFor(g, chooser); left != nil {
		t.Errorf("a second ordering prompt is open: %+v", left)
	}
}
