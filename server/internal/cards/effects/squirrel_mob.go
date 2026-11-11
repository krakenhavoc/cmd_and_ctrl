package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squirrel Mob — Creature — Squirrel {1}{G}{G}, 2/2:
//
//	"This creature gets +1/+1 for each other Squirrel on the
//	 battlefield."
//
// A layer 7c self-modifier counting every other Squirrel, whoever
// controls it, read live on every layer pass through effective
// subtypes (so a changeling counts).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "451486df-6160-4ff7-b47c-dd3760417a31",
		Name:         "Squirrel Mob",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{cmdr2B17SquirrelMobPT()},
	})
}
