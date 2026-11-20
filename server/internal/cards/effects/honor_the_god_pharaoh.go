package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honor the God-Pharaoh — Sorcery {2}{R} (War of the Spark, #5520):
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards. Amass Zombies 1. (Put a +1/+1 counter on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
// The discard is an additional cost (Tormenting Voice's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "5ce7780d-8a32-44c2-b779-aae732f5b05d",
		Name:           "Honor the God-Pharaoh",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Draws: 2, Amass: 1},
		AdditionalCost: DiscardCost(1),
		OnResolve:      ResolveSteps(DrawCards{N: 2}, Amass{Subtype: "Zombie", N: 1}),
	})
}
