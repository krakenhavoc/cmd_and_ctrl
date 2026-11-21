package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haze of Pollen — Instant {1}{G}
// (Amonkhet, #5535):
//
//	"Prevent all combat damage that would be dealt this turn.
//	 Cycling {3} ({3}, Discard this card: Draw a card.)"
//
// Fog, plus cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0140f00-5414-493b-962d-33312d49f6ea",
		Name:         "Haze of Pollen",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PreventAllCombatDamageThisTurn{Label: "Haze of Pollen: prevent combat damage"}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
