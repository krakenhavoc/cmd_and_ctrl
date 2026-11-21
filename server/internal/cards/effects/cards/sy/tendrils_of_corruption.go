package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tendrils of Corruption — Instant {3}{B}:
//
//	"Tendrils of Corruption deals X damage to target creature and you
//	 gain X life, where X is the number of Swamps you control."
//
// X is counted once, as the spell resolves (CR 608.2h); the life is
// gained after the damage, and not at all when the target is illegal and
// the spell does not resolve.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84ee0da2-d886-4603-9aaa-bc5326a424d7",
		Name:         "Tendrils of Corruption",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				x := tendrilsSwamps(ctx)
				if len(ctx.Item.Targets) == 0 || x <= 0 {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: ctx.Item.Targets[0].ID, Amount: x}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.GainLife{Amount: tendrilsSwamps(ctx)}.Apply(ctx) }),
		),
	})
}

// tendrilsSwamps is the number of Swamps the spell's controller controls.
func tendrilsSwamps(ctx *effects.Context) int {
	return effects.CmdrB5CountYouControl(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtypeWord("Swamp") })
}
