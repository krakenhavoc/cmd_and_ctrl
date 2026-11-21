package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4573f044-7011-412e-81bd-ae8d0c21283b",
		Name:         "Voyager Staff",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: Exile target creature. Return the exiled card to the battlefield under its owner's control at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSurvive},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return exileTargetsThenScheduleReturn(NewContext(g, item), "Voyager Staff — return the exiled card")
			},
		}},
	})
}
