package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rain of Embers — Sorcery {1}{R} (Ravnica: City of Guilds, #5600):
//
//	"Rain of Embers deals 1 damage to each creature and each player."
//
// Dry Spell's body: 1 damage to each creature and each player, all from
// this spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "041e2a01-3652-46c1-a9d1-d38ae7ee0551",
		Name:         "Rain of Embers",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B23DamageEachCreatureAndEachPlayer(ctx, 1)
		},
	})
}
