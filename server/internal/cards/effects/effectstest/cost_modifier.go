package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_modifier_test.go — S28 sub-PR 2: the nine cards that modify
// what other spells cost. The engine's own ordering and floor rules
// are pinned in game/cost_modifier_test.go; these tests are about
// the DECLARATIONS — that each card taxes or discounts the right
// spells cast by the right player, which is the half a card file can
// get wrong on its own.

const (
	SphereOfResistanceOracle = "09c96077-3804-4f12-a613-5bebc5e0413f"
	ThornOfAmethystOracle    = "0c6c5336-9233-4ab1-9d55-79f20be7ea57"
	ThaliaGuardianOracle     = "9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee"
	GoblinElectromancerOrcl  = "81f06f84-1580-43c0-89d5-08d34541a519"
	HeartlessSummoningOracle = "fa1be67e-a07e-42ce-88be-446acd643dc6"
	AuraOfSilenceOracle      = "e7faf8eb-e829-4109-8dfe-42865a23ba86"
	DampingSphereOracle      = "fd50d76c-7654-47b9-a5b6-d075874e4357"
	TrinisphereOracle        = "69d994f2-b8f6-425f-9655-977c2144d40c"
	AnimarOracle             = "725880b2-1675-414f-b61b-cf6533797dbf"
)

// priceInHand seeds a spell into `seat`'s hand and returns what the
// board currently charges to cast it, in total mana value.
//
// Routed through the public ApplyCostModifiers surface rather than
// through a real cast, so a test can ask the question without also
// having to satisfy timing, priority and the mana pool.
func PriceInHand(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, manaCost string) int {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", manaCost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(),
			Name:       name,
			TypeLine:   typeLine,
			ManaCost:   manaCost,
			Owner:      seat.ID,
			Controller: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}
