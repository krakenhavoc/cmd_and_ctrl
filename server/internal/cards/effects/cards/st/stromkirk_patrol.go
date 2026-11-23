package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stromkirk Patrol — Creature — Vampire Soldier {4}{B}, 4/3 (Innistrad, #5572):
//
//	"Whenever this creature deals combat damage to a player, put a
//	 +1/+1 counter on it."
//
// Hexgold Slith's combat-damage trigger: the counter goes on it only
// while it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c5fbb54-f0e3-4afe-9158-ff9eadaa338d",
		Name:         "Stromkirk Patrol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Stromkirk Patrol — put a +1/+1 counter on it", effects.ThisStillHere(effects.PlusOneCountersOnThis(1))),
		},
	})
}
