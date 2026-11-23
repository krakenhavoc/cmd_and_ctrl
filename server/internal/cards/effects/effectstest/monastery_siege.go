package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// priceOpponentsBoltAgainst returns what `caster` is charged to cast a
// {R} instant targeting `target` (a player or, with targetCard set, a
// card), through the same public ApplyCostModifiers surface
// cost_modifier_test.go's priceInHand uses — widened to carry
// Targets, which a target-reading modifier needs to see.
func PriceOpponentsBoltAgainst(t *testing.T, g *game.Game, caster *game.Player, targets []game.TargetRef) int {
	t.Helper()
	base, err := game.ParseCost("{R}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant",
			ManaCost: "{R}", Owner: caster.ID, Controller: caster.ID,
		},
		Controller: caster.ID,
		FromZone:   game.ZoneHand,
		Targets:    targets,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}
