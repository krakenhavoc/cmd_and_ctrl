package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d9356199-4ce3-44fa-9463-90fc134088f9",
		Name:         "Tamiyo's Logbook",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}, {T}: Draw a card. This ability costs {1} less to activate for each other artifact you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{5}{U}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.CmdrB3OtherArtifactsYouControl,
					"This ability costs {1} less to activate for each other artifact you control."),
			},
			Effect: effects.B36DrawOne,
		}},
	})
}
