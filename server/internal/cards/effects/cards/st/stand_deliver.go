package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stand // Deliver — split card (Invasion, #5616):
//
//	Stand — Instant {W}: "Prevent the next 2 damage that would be dealt
//	        to target creature this turn."
//	Deliver — Instant {2}{U}: "Return target permanent to its owner's
//	          hand."
//
// Each half is an ADR 0034 face (ADR 0103): Stand under the bare oracle
// ID, Deliver under "#1". Only one half can be cast (CR 709.3; the card
// has no fuse). Stand is the charged shield of Benevolent Ancestor
// (PreventNextDamage, CR 615.7) at 2, gone at end of turn; Deliver is
// Void Squall's bounceTheTarget on any permanent. A target gone by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6343aa78-2df6-41f3-9366-4933f9e6a93e",
		Name:         "Stand",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			return effects.PreventNextDamage{Target: ts[0].ID, Amount: 2, Label: "Stand: prevent the next 2 damage"}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "6343aa78-2df6-41f3-9366-4933f9e6a93e#1",
		Name:         "Deliver",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve:    effects.BounceTheTarget,
	})
}
