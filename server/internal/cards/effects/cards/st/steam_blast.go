package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steam Blast — Sorcery {2}{R} (Urza's Saga, S82 OLD b723):
//
//	Steam Blast deals 2 damage to each creature and each player.
//
// Blockbuster's shape: every creature and every living player is dealt the damage as one instance, the spell being the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92f4bdfd-1c84-4541-bc46-2a65c39de717",
		Name:         "Steam Blast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatchingAndEachPlayerThen(ctx, effects.Creature(), 2, effects.RecipientThen{})
		},
	})
}
