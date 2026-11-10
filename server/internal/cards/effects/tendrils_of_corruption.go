package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "84ee0da2-d886-4603-9aaa-bc5326a424d7",
		Name:         "Tendrils of Corruption",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				x := tendrilsSwamps(ctx)
				if len(ctx.Item.Targets) == 0 || x <= 0 {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: ctx.Item.Targets[0].ID, Amount: x}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return GainLife{Amount: tendrilsSwamps(ctx)}.Apply(ctx) }),
		),
	})
}

// tendrilsSwamps is the number of Swamps the spell's controller controls.
func tendrilsSwamps(ctx *Context) int {
	return cmdrB5CountYouControl(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtypeWord("Swamp") })
}
