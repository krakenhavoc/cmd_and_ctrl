package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crush Dissent — Instant {3}{U} (War of the Spark, #5520):
//
//	"Counter target spell unless its controller pays {2}.
//	 Amass Zombies 2. (Put two +1/+1 counters on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
// The counter is the shared CounterUnlessPaid prompt; the amass waits for
// the answer (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ece897d-5b52-4fca-9f97-d7e864134c0b",
		Name:         "Crush Dissent",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Amass: 2},
		Targets:      TargetSpell("target spell"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return counterTheTargetSpellUnlessPaid("Crush Dissent", "{2}")(ctx.Game, ctx.Item)
			}),
			Amass{Subtype: "Zombie", N: 2},
		),
	})
}
