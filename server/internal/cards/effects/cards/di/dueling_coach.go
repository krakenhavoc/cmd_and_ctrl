package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dueling Coach — Creature — Human Monk {3}{W}, 2/2 (Strixhaven: School of
// Mages, #5502):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 {4}{W}, {T}: Put a +1/+1 counter on each creature you control with a
//	 +1/+1 counter on it."
//
// The enters trigger is Bond Beetle's. The activated ability reads the set as
// it resolves (CR 608.2c) and puts one counter on each creature you control
// that has a +1/+1 counter on it then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aae4dd38-3b35-4c92-a0d4-5d37cfd93df8",
		Name:         "Dueling Coach",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Dueling Coach — put a +1/+1 counter on target creature", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Label:   "{4}{W}, {T}: Put a +1/+1 counter on each creature you control with a +1/+1 counter on it.",
			Cost:    effects.Plus(effects.ManaCost("{4}{W}"), effects.TapCost()),
			Effect: effects.PutPlusOneCountersOnEachYouControlThat(1, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return effects.HasCounterOfKind(c, game.CounterPlusOne)
			}),
		}},
	})
}
