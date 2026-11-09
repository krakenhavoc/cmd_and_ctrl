package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tamiyo's Logbook — Artifact — Book {2}{U} (Foundations Commander):
//
//	"{5}{U}, {T}: Draw a card. This ability costs {1} less to activate for
//	 each other artifact you control."
//
// Diary of Dreams' shape: the discount is the ability's own cost modifier,
// counted as the cost is determined (CR 602.2b, 601.2f), {1} per OTHER
// artifact the activator controls. It reduces generic mana only, so {U}
// is always paid (printed ruling), and the {T} is still paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9356199-4ce3-44fa-9463-90fc134088f9",
		Name:         "Tamiyo's Logbook",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}{U}, {T}: Draw a card. This ability costs {1} less to activate for each other artifact you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{5}{U}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(cmdrB3OtherArtifactsYouControl,
					"This ability costs {1} less to activate for each other artifact you control."),
			},
			Effect: b36DrawOne,
		}},
	})
}
