package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- fixtures ------------------------------------------------------------------

// tdmTap taps a battlefield permanent directly.
func TdmTap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// tdmPlaceWarriors answers every "which player does this token attack"
// prompt with the first offer, as the mobilize tests do.
func TdmPlaceWarriors(t *testing.T, g *game.Game, me uuid.UUID) {
	t.Helper()
	for c := LatestChoiceOfKind(g, game.PendingChoicePayAmount); c != nil; c = LatestChoiceOfKind(g, game.PendingChoicePayAmount) {
		if err := g.ResolvePayAmount(c.ID, me, c.PayAmount.Max); err != nil {
			t.Fatalf("place the Warriors: %v", err)
		}
	}
}

// tdmExileFromGraveyard exiles a graveyard card by an effect `by`
// controls.
func TdmExileFromGraveyard(t *testing.T, g *game.Game, by, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: by, Owner: by})
		if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
}
