package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jukai Preserver — Enchantment Creature — Human Druid {3}{G}, 3/3 (Neon
// Dynasty, #5493):
//
//	"When this creature enters, put a +1/+1 counter on target creature you
//	 control.
//	 Channel — {2}{G}, Discard this card: Put a +1/+1 counter on each of up
//	 to two target creatures you control."
//
// The enters trigger is Ride the Shoopuf's counter body, and it may target
// the Preserver itself. The channel is Earth Kingdom Soldier's
// counter-on-each-target over a two-target "up to" clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8ce90be-f0ab-4f8d-896a-cde8aaf24579",
		Name:         "Jukai Preserver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Jukai Preserver — put a +1/+1 counter on target creature you control", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {2}{G}, Discard this card: Put a +1/+1 counter on each of up to two target creatures you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
			Effect:  effects.PutPlusOneCounterOnEachLegalTarget,
		}},
	})
}
