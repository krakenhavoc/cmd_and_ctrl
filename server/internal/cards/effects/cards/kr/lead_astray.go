package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lead Astray — Instant {1}{W} (Judgment):
//
//	"Tap up to two target creatures."
//
// Elder Deep-Fiend's tap of every still-legal target (CR 608.2b), with up to
// two chosen as the spell is cast; none is a legal cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e8e5ad4-05a3-4757-8797-c37bf49506d5",
		Name:         "Lead Astray",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
	})
}
