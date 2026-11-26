package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruthless Invasion — Sorcery {3}{R/P}:
//
//	"({R/P} can be paid with either {R} or 2 life.)
//	 Nonartifact creatures can't block this turn."
//
// "Can't block" changes no characteristic, so the set is a live rule
// (CR 611.2c): a creature that arrives later this turn, or stops being an
// artifact, is judged as blockers are declared (CR 509.1b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88bbbc00-0704-4dc0-897f-93184b2e1eed",
		Name:         "Ruthless Invasion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.RestrictUntilEOT{
				Scope:        game.ScopeCreaturesMatching,
				Restrictions: game.CantBlock,
				Queries:      []game.PermanentQuery{{Types: []string{"creature"}, WithoutTypes: []string{"artifact"}}},
				Label:        "Ruthless Invasion — nonartifact creatures can't block",
			}.Apply(ctx)
		},
	})
}
