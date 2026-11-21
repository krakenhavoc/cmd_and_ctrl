package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82Walker puts a planeswalker onto the battlefield as an ordinary entry.
func S82Walker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Planeswalker — Test", OracleID: oracle,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	g.RunStateChecksForTest()
	return id
}

func S82Activate(t *testing.T, g *game.Game, who, walker uuid.UUID, idx int, p game.ActivateAbilityParams) {
	t.Helper()
	p.Strict = true
	if err := g.ActivateCatalogAbility(who, walker, idx, p); err != nil {
		t.Fatalf("activate %d: %v", idx, err)
	}
}

func S82TargetCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
