package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rain of Embers — Sorcery {1}{R} (Ravnica: City of Guilds, #5600):
//
//	"Rain of Embers deals 1 damage to each creature and each player."
//
// Dry Spell's body: 1 damage to each creature and each player, all from
// this spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "041e2a01-3652-46c1-a9d1-d38ae7ee0551",
		Name:         "Rain of Embers",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b23DamageEachCreatureAndEachPlayer(ctx, 1)
		},
	})
}
