package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corrupted Resolve — Instant {1}{U} (New Phyrexia, S82 OLD batch 463,
// #5575):
//
//	"Counter target spell if its controller is poisoned."
//
// Any target spell; whether its controller has a poison counter is
// judged as this resolves (CR 608.2b), from the spell's controller as it
// is on the stack. Bring the Ending's reading with the threshold of one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87837a94-05db-48fe-b810-26d8f92cfc58",
		Name:         "Corrupted Resolve",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			s := ctx.Game.StackItemForEffect(targets[0].ID)
			if s == nil {
				return nil
			}
			if p := ctx.Game.PlayerByIDForEffect(s.Controller); p == nil || p.Poison < 1 {
				return nil
			}
			return effects.CounterTarget{StackID: targets[0].ID}.Apply(ctx)
		},
	})
}
