package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hmlCreature puts a catalog creature with real P/T on the battlefield,
// not summoning sick.
func HmlCreature(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	id := PushCatalogPermanent(g, owner, name, "Creature — Test", oracle, false)
	c := FindBattlefieldCardForTest(g, id)
	c.Power, c.Toughness = power, toughness
	return id
}

// hmlUntapSkips is the permanent's next-untap-step marks.
func HmlUntapSkips(t *testing.T, g *game.Game, id uuid.UUID) []game.UntapSkip {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	return c.NextUntapSkips
}
