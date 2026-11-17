package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morale — Instant {1}{W}{W} (The Dark, #5638):
//
//	"Attacking creatures get +1/+1 until end of turn."
//
// Iroh, Firebending Instructor's pump as a spell: every attacking
// creature, whoever controls it, as it resolves. Outside combat
// nothing is attacking, so it does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8af9d623-efa3-4cbc-a903-aaec32d78b0a",
		Name:         "Morale",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{Match: AttackingCreature(), Power: 1, Toughness: 1, Label: "Morale — attacking creatures get +1/+1"}.Apply(ctx)
		},
	})
}
