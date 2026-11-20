package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Tempest — Sorcery {5}{R}{R} (Portal, #5630):
//
//	"Fire Tempest deals 6 damage to each creature and each player."
//
// Inferno's text, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ede6a352-52de-4ae0-816f-50140033dedf",
		Name:         "Fire Tempest",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b23DamageEachCreatureAndEachPlayer(ctx, 6)
		},
	})
}
