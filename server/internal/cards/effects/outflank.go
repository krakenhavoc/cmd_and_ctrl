package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c3436d70-230c-4bda-b7da-935a270f135f",
		Name:         "Outflank",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			n := len(permanentsControlledByMatching(ctx.Game, item.Controller, Creature()))
			return DealDamage{Source: ctx.Source(), Target: legal[0].ID, Amount: n}.Apply(ctx)
		},
	})
}
