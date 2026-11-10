package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Search the Premises — Enchantment {3}{W}:
//
//	"Whenever a creature attacks you or a planeswalker you control,
//	 investigate."
//
// One trigger per attacking creature, whoever controls it. An attack
// at a battle you protect is not an attack at you (CR 508.3e).
func init() {
	Register(Spec{
		OracleID:     "0232418d-1c52-4573-b488-c62345ca1c8d",
		Name:         "Search the Premises",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, cmdrB1CreatureAttacksYouOrYourPlaneswalker,
				"Search the Premises — investigate", Do(cmdrB1Investigate())),
		},
	})
}
