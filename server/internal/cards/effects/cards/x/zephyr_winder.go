package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephyr Winder — Creature — Elemental {1}{U}, 2/1 (March of the Machine,
// S82 MOM batch 13):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, untap up to
//	 one target creature."
//
// Cold-eyed Selkie's trigger on the combat-damage-to-a-player event, with
// an "up to one" creature target chosen as the trigger goes on the stack.
// A target that has left by resolution is illegal and nothing is untapped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13c4437e-48a9-4c51-9b76-9e83e3ec362f",
		Name:            "Zephyr Winder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
					"Zephyr Winder — untap up to one target creature", effects.UntapEachLegalTarget),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}
