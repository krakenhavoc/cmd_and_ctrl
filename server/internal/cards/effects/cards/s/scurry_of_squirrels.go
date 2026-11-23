package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scurry of Squirrels — Creature — Squirrel Scout {2}{G}, 2/2 (Bloomburrow
// Commander):
//
//	"Myriad, myriad (...)
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on target creature you control."
//
// Two instances of myriad (CR 702.116b): each is its own triggered ability
// and triggers separately, so each opponent other than the defending player
// is asked twice and may get two copies. The second instance is the shared
// myriad trigger under its own key.
//
// No simplification.
func init() {
	second := effects.GrantedMyriad("Scurry of Squirrels")
	second.Key += " (second instance)"
	effects.Register(effects.Spec{
		OracleID:     "3beed054-a099-4887-b1bd-39cf9e514e94",
		Name:         "Scurry of Squirrels",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GrantedMyriad("Scurry of Squirrels"),
			second,
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Scurry of Squirrels — put a +1/+1 counter on target creature you control",
				effects.Cmdr2S46PlusOneCounterOnClauseTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
