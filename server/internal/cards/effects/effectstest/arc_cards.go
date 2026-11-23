package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Sorcerer's Strongbox ---------------------------------------------------

// arcStrongboxFlip activates the Strongbox as me, resolves the ability,
// runs `before` while the coin call is open, calls heads and reports
// whether the flip was won.
func ArcStrongboxFlip(t *testing.T, g *game.Game, me *game.Player, box uuid.UUID, before func()) bool {
	t.Helper()
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(box) })
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, box, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	PassPriorityAroundTable(t, g)
	ask := LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if ask == nil {
		t.Fatal("no coin call was asked")
	}
	if before != nil {
		before()
	}
	flips := len(RandomEvents(g, game.EventFlipCoin))
	if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	PassPriorityAroundTable(t, g)
	evs := RandomEvents(g, game.EventFlipCoin)
	if len(evs) != flips+1 {
		t.Fatalf("%d coins were flipped, want 1", len(evs)-flips)
	}
	return evs[len(evs)-1].Won
}
