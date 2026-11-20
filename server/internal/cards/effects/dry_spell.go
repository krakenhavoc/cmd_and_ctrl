package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dry Spell — Sorcery {1}{B} (Portal, #5630):
//
//	"Dry Spell deals 1 damage to each creature and each player."
//
// Psychotic Haze's sweep without the madness: 1 damage to each
// creature and each player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "980a9957-6b52-45c6-b847-f84974d5a653",
		Name:         "Dry Spell",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b23DamageEachCreatureAndEachPlayer(ctx, 1)
		},
	})
}
