package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Outflank — Instant {W}:
//
//	"Outflank deals damage to target attacking or blocking creature equal
//	 to the number of creatures you control."
//
// The creatures are counted as the spell resolves, after the target is
// re-checked (CR 608.2b); a creature that has left combat takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3436d70-230c-4bda-b7da-935a270f135f",
		Name:         "Outflank",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			n := len(effects.PermanentsControlledByMatching(ctx.Game, item.Controller, effects.Creature()))
			return effects.DealDamage{Source: ctx.Source(), Target: legal[0].ID, Amount: n}.Apply(ctx)
		},
	})
}
