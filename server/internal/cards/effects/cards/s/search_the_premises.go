package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Search the Premises — Enchantment {3}{W}:
//
//	"Whenever a creature attacks you or a planeswalker you control,
//	 investigate."
//
// One trigger per attacking creature, whoever controls it. An attack
// at a battle you protect is not an attack at you (CR 508.3e).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0232418d-1c52-4573-b488-c62345ca1c8d",
		Name:         "Search the Premises",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.CmdrB1CreatureAttacksYouOrYourPlaneswalker,
				"Search the Premises — investigate", effects.Do(effects.CmdrB1Investigate())),
		},
	})
}
