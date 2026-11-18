package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crossroads Candleguide — Artifact Creature — Scarecrow {4}, 3/4
// (Innistrad: Midnight Hunt, S82):
//
//	"When this creature enters, exile up to one target card from a
//	 graveyard.
//	 {2}: Add one mana of any color."
//
// The enters trigger's target is "up to one", so it goes on the stack with
// no target and does nothing then; with a card chosen, it is re-checked at
// resolution (CR 608.2b). The mana ability is Energy Refractor's: a mana
// cost, no tap, any one colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b817184-4965-4186-bfa9-6d9cf198050e",
		Name:         "Crossroads Candleguide",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{s82MidB07ExileUpToOneFromAGraveyard("Crossroads Candleguide")},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Mana: "{2}"}, "{2}: Add one mana of any color"),
		},
	})
}
