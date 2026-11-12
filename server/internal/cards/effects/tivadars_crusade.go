package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tivadar's Crusade — Sorcery {1}{W}{W} (The Dark, #5638):
//
//	"Destroy all Goblins."
//
// Zombie Apocalypse's "destroy all Humans" with Goblin: every
// permanent with the Goblin subtype, read from its current types, so a
// noncreature Kindred Goblin is destroyed too. Regeneration and
// indestructible work as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5cf4d98e-f2f4-435b-a633-990b2e654f3a",
		Name:         "Tivadar's Crusade",
		Completeness: CompletenessFull,
		// ADR 0126 §6: only Goblins, and a Kindred Goblin may be a
		// noncreature permanent, so the class is the upper bound.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: HasSubtype("Goblin")}.Apply(ctx)
		},
	})
}
