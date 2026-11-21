package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Biogenic Ooze — {3}{G}{G} Creature — Ooze 2/2 (reprinted in Teenage
// Mutant Ninja Turtles Commander, #5011):
//
//	"When this creature enters, create a 2/2 green Ooze creature token.
//	 At the beginning of your end step, put a +1/+1 counter on each Ooze
//	 you control.
//	 {1}{G}{G}{G}: Create a 2/2 green Ooze creature token."
//
// The end-step counters go on every permanent its controller controls
// with the subtype Ooze as the trigger resolves, the Ooze itself and
// every token included (Nazgûl's shape, putAPlusOneCounterOnEachYouControl).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2d84ba5-0d20-4ce3-8f1f-93cd9ef94f5f",
		Name:         "Biogenic Ooze",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Biogenic Ooze — create a 2/2 green Ooze",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Ooze"), N: 1})),
			effects.AtYourEndStep("Biogenic Ooze — put a +1/+1 counter on each Ooze you control",
				effects.PutAPlusOneCounterOnEachYouControl(effects.Subtype("Ooze"))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}{G}{G}: Create a 2/2 green Ooze creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{1}{G}{G}{G}"),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Ooze"), N: 1}),
		}},
	})
}
