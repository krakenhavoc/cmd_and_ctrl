package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a8ce90be-f0ab-4f8d-896a-cde8aaf24579",
		Name:         "Jukai Preserver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Jukai Preserver — put a +1/+1 counter on target creature you control", PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {2}{G}, Discard this card: Put a +1/+1 counter on each of up to two target creatures you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{2}{G}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("up to two target creatures you control", YouControl()).WithCount(0, 2),
			Effect:  putPlusOneCounterOnEachLegalTarget,
		}},
	})
}
