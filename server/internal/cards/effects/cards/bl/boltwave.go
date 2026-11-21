package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boltwave — Sorcery {R} (EDHREC rank 1438):
//
//	"Boltwave deals 3 damage to each opponent."
//
// A Lightning Bolt to every opponent at once. Damage from the spell,
// through the damage pipeline, so a doubler doubles it and a shield
// stops it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb8cccc8-e56f-42fa-a0d7-fc889b5e1c7b",
		Name:         "Boltwave",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageToEachOpponent(ctx.Game, item, 3)
		},
	})
}
