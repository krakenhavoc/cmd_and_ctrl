package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Parting Thoughts — Sorcery {2}{B} (Commander 2016, #5537):
//
//	"Destroy target creature. You draw X cards and you lose X life,
//	 where X is the number of counters on that creature."
//
// X counts counters of every kind (Nils, Discipline Enforcer's reading),
// as the creature last existed on the battlefield (CR 608.2h): read live
// when it is still there (an indestructible or regenerated creature
// survives, and the draw and the life loss happen all the same, as the
// ruling says), and from its last-known counters once it has left. The
// draw and the loss are separate steps, so the loss waits for a prompt a
// draw asks (CR 608.2c). A creature with no counters means nothing more
// happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57fd34d3-3e30-4467-bace-3f4dc2a85c80",
		Name:         "Parting Thoughts",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyFirstLegalTarget(ctx.Game, ctx.Item) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: partingThoughtsX(ctx)}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if x := partingThoughtsX(ctx); x > 0 {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -x)
				}
				return nil
			}),
		),
	})
}

// partingThoughtsX is the number of counters on the target creature: on
// the permanent while it is on the battlefield, else as it last existed
// there. The spell has one target and resolved, so that target was legal
// as it began to resolve.
func partingThoughtsX(ctx *effects.Context) int {
	if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetCard {
		return 0
	}
	// The live permanent, or its last-known record once it has left
	// (departedCreaturePower's read): the trigger-time counter map
	// (LastKnownCountersForEffect) is gone by the next step.
	ref, ok := ctx.Game.PermanentRefForEffect(ctx.Item.Targets[0].ID)
	if !ok {
		return 0
	}
	info, ok := ctx.Game.PermanentForEffect(ref)
	if !ok {
		return 0
	}
	return effects.Cmdr2B25CounterTotal(info.Counters)
}
