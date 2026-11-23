package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voyager Staff — Artifact {1} (Ravnica Remastered, #5948):
//
//	"{2}, Sacrifice this artifact: Exile target creature. Return the exiled
//	 card to the battlefield under its owner's control at the beginning of
//	 the next end step."
//
// Getaway Glamer's delayed flicker: one exile now, one delayed trigger that
// returns the card as a new object (CR 400.7). The sacrifice is a cost, so
// the staff is gone before the ability resolves; a token exiled this way
// does not come back (CR 111.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4573f044-7011-412e-81bd-ae8d0c21283b",
		Name:         "Voyager Staff",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: Exile target creature. Return the exiled card to the battlefield under its owner's control at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Voyager Staff — return the exiled card")
			},
		}},
	})
}
