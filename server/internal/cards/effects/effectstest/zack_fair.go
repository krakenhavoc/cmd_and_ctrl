package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zack_fair_test.go — Zack Fair's "attach an Equipment that was attached
// to Zack Fair to that creature" (#3780): the sacrificed permanent's
// last-known attachments, read after the state-based action has already
// unattached them (CR 704.5n).

func ZackFairActivate(t *testing.T, g *game.Game, zack, target uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	if err := g.ActivateCatalogAbility(me.ID, zack, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

func ZackFairBoard(t *testing.T) (g *game.Game, zack, bear, me uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	seat := g.Seats[0]
	AdvanceToMain(t, g)
	zack = CastAndResolveCreature(t, g, "Zack Fair", "Legendary Creature — Human Soldier", ZackFairOracle)
	PassPriorityAroundTable(t, g)
	return g, zack, SeedBear(g, seat.ID), seat.ID
}
