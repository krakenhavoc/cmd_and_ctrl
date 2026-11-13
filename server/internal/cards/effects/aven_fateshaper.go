package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aven Fateshaper — Creature — Bird Wizard {6}{U}, 4/5 (Onslaught):
//
//	"Flying
//	 When this creature enters, look at the top four cards of your
//	 library, then put them back in any order.
//	 {4}{U}: Look at the top four cards of your library, then put them
//	 back in any order."
//
// Sensei's Divining Top's look-and-reorder (LookAtTop) twice: once as an
// enters trigger and once on a mana-only activated ability, which has no
// {T} and so may be used the turn it arrives. The activated row has no
// target and answers nothing (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "739d93c4-8d7b-4a54-bb6e-1d283e333392",
		Name:            "Aven Fateshaper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aven Fateshaper — look at the top four cards of your library, then put them back in any order",
				Do(LookAtTop{N: 4})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{U}: Look at the top four cards of your library, then put them back in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{4}{U}"),
			Effect:  Do(LookAtTop{N: 4}),
		}},
	})
}
