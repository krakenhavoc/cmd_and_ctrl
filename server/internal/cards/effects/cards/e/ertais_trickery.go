package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ertai's Trickery — Instant {U} (Planeshift, S82 OLD b644, #5615):
//
//	"Counter target spell if it was kicked."
//
// The check is read as the spell resolves (CR 608.2c) from the record of what the
// target's announcement paid (CR 702.33): a spell cast with a kicker, multikicker
// included, is countered; any other is untouched. A target that is no longer on the
// stack makes this do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93d0691f-17b8-458b-a624-e182fc481f49",
		Name:         "Ertai's Trickery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			card, found := ctx.Game.LookupCardForEffect(t.ID)
			si := ctx.Game.StackItemForEffect(t.ID)
			if !found || si == nil || game.KickedTimesPaid(card, si.Paid.OptionalCosts) == 0 {
				return nil
			}
			return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
		},
	})
}
