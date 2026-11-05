package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d2d84ba5-0d20-4ce3-8f1f-93cd9ef94f5f",
		Name:         "Biogenic Ooze",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Biogenic Ooze — create a 2/2 green Ooze",
				Do(CreateToken{Template: TokenCard("2/2 green Ooze"), N: 1})),
			AtYourEndStep("Biogenic Ooze — put a +1/+1 counter on each Ooze you control",
				putAPlusOneCounterOnEachYouControl(Subtype("Ooze"))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}{G}{G}: Create a 2/2 green Ooze creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{1}{G}{G}{G}"),
			Effect:  Do(CreateToken{Template: TokenCard("2/2 green Ooze"), N: 1}),
		}},
	})
}
