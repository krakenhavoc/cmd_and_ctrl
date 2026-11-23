package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Callous Dismissal — Sorcery {1}{U} (War of the Spark, #5520):
//
//	"Return target nonland permanent to its owner's hand.
//	 Amass Zombies 1. (Put a +1/+1 counter on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
// The bounce is Void Squall's; the amass waits for any prompt it asks
// (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9219b7d5-642d-4470-a7c4-f3e56bfae8c7",
		Name:         "Callous Dismissal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 1},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.BounceTheTarget(ctx.Item, ctx) }),
			effects.Amass{Subtype: "Zombie", N: 1},
		),
	})
}
