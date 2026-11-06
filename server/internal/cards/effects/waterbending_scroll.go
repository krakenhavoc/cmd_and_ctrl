package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waterbending Scroll — Artifact {1}{U}:
//
//	"{6}, {T}: Draw a card. This ability costs {1} less to activate for
//	 each Island you control."
//
// Diary of Dreams' shape with the count read off the battlefield: the
// discount is the Islands you control as the cost is determined
// (CR 601.2f, 602.2b), effective subtypes, so a land something made an
// Island counts. It reduces only the generic {6}; the {T} stays.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "882d032b-db36-401a-8075-a9fabc90c53d",
		Name:         "Waterbending Scroll",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{6}, {T}: Draw a card. This ability costs {1} less to activate for each Island you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{6}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(PermanentsYouControl(Subtype("Island")),
					"This ability costs {1} less to activate for each Island you control."),
			},
			Effect: b36DrawOne,
		}},
	})
}
