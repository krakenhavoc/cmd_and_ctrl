package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surveyor's Scope — Artifact {2} (Commander 2013, S82 OLD b380, #5560):
//
//	"{T}, Exile this artifact: Search your library for up to X basic land cards,
//	 where X is the number of players who control at least two more lands than
//	 you. Put those cards onto the battlefield, then shuffle."
//
// X is counted as the ability resolves (S82OldB380PlayersWithTwoMoreLands),
// the lands put onto the battlefield untapped. With X = 0 the library is still
// shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9633730d-c41f-4597-b806-55ce2dd848e9",
		Name:         "Surveyor's Scope",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Exile this artifact: Search your library for up to X basic land cards, where X is the number of players who control at least two more lands than you. Put those cards onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, Tutors: 1},
			Cost:    effects.Plus(effects.TapCost(), effects.ExileThis()),
			Effect:  effects.S82OldB380SurveyorSearch,
		}},
	})
}
