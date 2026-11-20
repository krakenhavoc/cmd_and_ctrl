package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "57fd34d3-3e30-4467-bace-3f4dc2a85c80",
		Name:         "Parting Thoughts",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return destroyFirstLegalTarget(ctx.Game, ctx.Item) }),
			Instruction(func(ctx *Context) error {
				return DrawCards{N: partingThoughtsX(ctx)}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
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
func partingThoughtsX(ctx *Context) int {
	if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetCard {
		return 0
	}
	id := ctx.Item.Targets[0].ID
	if onBattlefield(ctx.Game, id) {
		if c, found := ctx.Game.LookupCardForEffect(id); found {
			return socB9CountersOn(c)
		}
		return 0
	}
	return cmdr2B25CounterTotal(ctx.Game.LastKnownCountersForEffect(id))
}
