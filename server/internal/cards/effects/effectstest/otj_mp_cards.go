package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otjMPMercenaries lists the Mercenary tokens `owner` controls.
func OtjMPMercenaries(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Mercenary" && c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// otjMPOneMercenary asserts `owner` controls exactly one Mercenary token,
// printed as the set makes it, and returns it.
func OtjMPOneMercenary(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	ids := OtjMPMercenaries(g, owner)
	if len(ids) != 1 {
		t.Fatalf("%d Mercenary tokens, want 1", len(ids))
	}
	c, ok := BattlefieldCard(g, ids[0])
	if !ok {
		t.Fatal("the Mercenary is not on the battlefield")
	}
	if game.TokenTextForCard(c) == "" {
		t.Error("the Mercenary ships no printed text")
	}
	if !c.IsCreature() || !c.HasSubtype("Mercenary") || len(c.Colors) != 1 || c.Colors[0] != "R" {
		t.Errorf("Mercenary %q colors %v, want a red Mercenary creature", c.TypeLine, c.Colors)
	}
	if p, tt := PtOf(t, g, ids[0]); p != 1 || tt != 1 {
		t.Errorf("Mercenary is %d/%d, want 1/1", p, tt)
	}
	return ids[0]
}
