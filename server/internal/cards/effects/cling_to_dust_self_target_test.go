package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4869, CR 601.2a: escaped Cling to Dust is on the stack before its
// "target card from a graveyard" is chosen, so it cannot exile itself.
// Before the fix the cast was accepted and the spell named itself.
func TestClingToDustEscapeCannotTargetItself(t *testing.T) {
	const oracle = "9c44e556-4c7a-48bf-a108-c584f69c2cfa"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Cling to Dust", "Instant", oracle)
	pay := seedGraveyardFodder(t, g, 5)

	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "escape",
		AltCostIDs:      pay,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: id}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("escaped Cling to Dust naming itself: got %v, want ErrIllegalTarget", err)
	}
	if !active.Graveyard.Contains(id) {
		t.Errorf("the refused cast moved Cling to Dust out of the graveyard")
	}
}
