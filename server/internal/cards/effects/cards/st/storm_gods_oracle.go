package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm God's Oracle — Enchantment Creature — Human Shaman {1}{U}{R}, 1/3
// (Modern Horizons 2, #5501):
//
//	"{1}: This creature gets +1/-1 until end of turn.
//	 When this creature dies, it deals 3 damage to any target."
//
// Merfolk Coralsmith's +1/-1 at {1} and Pyre Spawn's dies trigger: the
// damage is dealt by the Oracle as it last existed (CR 608.2h), the
// target chosen as the trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21613fac-d959-44ba-9bcb-0a30b5e6545e",
		Name:         "Storm God's Oracle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Storm God's Oracle — +1/-1 until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Storm God's Oracle — 3 damage to any target", effects.SourceDealsDamageToEachLegalTarget(3)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}
