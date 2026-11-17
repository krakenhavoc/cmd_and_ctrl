package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const takeFlightGrant = "take-flight/attack-draw" // an on-disk identity: never renamed

// Take Flight — Enchantment — Aura {3}{U} (The Brothers' War):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 and has flying and "Whenever this
//	 creature attacks, draw a card.""
//
// The trigger is the creature's, so its controller draws, not the Aura's
// controller (Giant Inheritance's grant shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4cf61e5-b2b7-4fb9-addf-db84dec49681",
		Name:         "Take Flight",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("flying"),
			GrantAbilitiesToAttached(takeFlightGrant),
		},
		Grants: []AbilityGrant{{
			Key: takeFlightGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Take Flight — draw a card", Do(DrawCards{N: 1})),
			},
			Text: "Whenever this creature attacks, draw a card.",
		}},
	})
}
