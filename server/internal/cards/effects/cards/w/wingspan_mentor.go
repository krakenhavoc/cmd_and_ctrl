package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wingspan Mentor — Creature — Human Wizard {2}{U}, 1/3 (S82 reprint pass):
//
//	"When this creature enters, put a flying counter on target non-Human
//	 creature you control.
//	 {2}{U}, {T}: Put a +1/+1 counter on each creature you control with
//	 flying."
//
// The enter trigger targets (Abigale's shape) and puts the engine's flying
// keyword counter. The activated ability reads flying through the layered
// view as it resolves, so a flying counter, a granted flying and a printed
// one all count; one counter each, as one placement per creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b3b0e56-0b83-403b-bdfd-b5b7c0accf1b",
		Name:         "Wingspan Mentor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Wingspan Mentor — put a flying counter on target non-Human creature you control",
					effects.PutAFlyingCounterOnTheTarget),
				effects.TargetCreature("target non-Human creature you control", effects.YouControl(), effects.Not(effects.OfSubtype("Human")))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Put a +1/+1 counter on each creature you control with flying.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Effect:  effects.PutPlusOneCountersOnEachYouControlThat(1, effects.HasKeyword("flying")),
		}},
	})
}
