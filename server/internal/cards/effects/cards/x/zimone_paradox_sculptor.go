package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zimone, Paradox Sculptor — Legendary Creature — Human Wizard {2}{G}{U},
// 1/4 (Foundations, #4465):
//
//	"At the beginning of combat on your turn, put a +1/+1 counter on each
//	 of up to two target creatures you control.
//	 {G}{U}, {T}: Double the number of each kind of counter on up to two
//	 target creatures and/or artifacts you control."
//
// The trigger is a targeted beginning-of-combat trigger with up to two
// targets (CR 603.3d: with none chosen it still triggers and does nothing).
// Each target is re-read as it resolves (CR 608.2b), one that left or
// stopped being a creature you control getting nothing. The activation is
// Vorel of the Hull Clade's doubling over up to two targets: each kind of
// counter is read once before the first placement, then added again through
// the ordinary counter path, so Hardened Scales and Doubling Season apply
// as the ruling says. A creature source, so the tap waits out summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9dd674a7-becf-4106-b53f-bca88426d92d",
		Name:         "Zimone, Paradox Sculptor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Zimone — put a +1/+1 counter on each of up to two target creatures you control",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
			),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{U}, {T}: Double the number of each kind of counter on up to two target creatures and/or artifacts you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{G}{U}"), effects.TapCost()),
			Targets: effects.TargetPermanent("up to two target creatures and/or artifacts you control",
				effects.Or(effects.Creature(), effects.Artifact()), effects.YouControl()).WithCount(0, 2),
			Effect: effects.TheThingDouble,
		}},
	})
}
