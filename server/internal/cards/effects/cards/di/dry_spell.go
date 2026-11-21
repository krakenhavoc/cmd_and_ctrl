package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dry Spell — Sorcery {1}{B} (Portal, #5630):
//
//	"Dry Spell deals 1 damage to each creature and each player."
//
// Psychotic Haze's sweep without the madness: 1 damage to each
// creature and each player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "980a9957-6b52-45c6-b847-f84974d5a653",
		Name:         "Dry Spell",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B23DamageEachCreatureAndEachPlayer(ctx, 1)
		},
	})
}
