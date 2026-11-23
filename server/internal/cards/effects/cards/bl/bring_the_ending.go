package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bring the Ending — Instant {1}{U}:
//
//	"Counter target spell unless its controller pays {2}.
//	 Corrupted — Counter that spell instead if its controller has three or
//	 more poison counters."
//
// The spell's controller is read as this resolves. A corrupted controller
// gets no chance to pay: the spell is countered outright. Otherwise it is
// the ordinary pay-{2} question (CR 118.12a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57b37690-0fb6-4e78-9210-4a529a6ade7c",
		Name:         "Bring the Ending",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			if s := ctx.Game.StackItemForEffect(targets[0].ID); s != nil && effects.PlayerIsCorrupted(ctx.Game, s.Controller) {
				return effects.CounterTarget{StackID: targets[0].ID}.Apply(ctx)
			}
			return effects.CounterTheTargetSpellUnlessPaid("Bring the Ending", "{2}")(ctx.Game, item)
		},
	})
}
