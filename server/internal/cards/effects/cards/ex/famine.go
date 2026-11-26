package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Famine — Sorcery {3}{B}{B} (Commander 2013, S82 OLD b386):
//
//	"Famine deals 3 damage to each creature and each player."
//
// Psychotic Haze's body at 3: one damage event over every creature and every
// player, the caster and their creatures included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f82fe6b-1ec7-400b-a80d-40921b359723",
		Name:         "Famine",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B23DamageEachCreatureAndEachPlayer(ctx, 3)
		},
	})
}
