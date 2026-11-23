package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saruman's Trickery — Instant {1}{U}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Counter target spell.
//	 Amass Orcs 1."
//
// The counter, then the keyword action (CR 701.47) as the next step of the same
// resolution. If the target spell has already left the stack the spell is
// countered on resolution (CR 608.2b) and does not amass; a spell that can't be
// countered is still "countered" for the instruction's purposes and the Army
// still grows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b415806-a3e3-4d87-b4c1-ccd11a8ee589",
		Name:         "Saruman's Trickery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		Purpose:      game.Purpose{Amass: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
			}),
			effects.Amass{Subtype: "Orc", N: 1},
		),
	})
}
