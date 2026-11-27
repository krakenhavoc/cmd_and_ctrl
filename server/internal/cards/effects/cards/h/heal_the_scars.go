package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heal the Scars — Instant {3}{G} (Lorwyn):
//
//	"Regenerate target creature. You gain life equal to that creature's
//	 toughness."
//
// The regeneration shield is set first; the life is the target's toughness as
// the spell resolves (CR 608.2h), shield or not. A target that is illegal on
// resolution fizzles the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc65a99f-ac7f-4ded-bb00-a5daf51e4bcd",
		Name:         "Heal the Scars",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			toughness := 0
			if c, found := ctx.Game.LookupCardForEffect(id); found {
				toughness = max(c.CurrentToughness(), 0)
			}
			if err := (effects.Regenerate{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: toughness}.Apply(ctx)
		},
	})
}
