package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reforge the Soul — Sorcery {3}{R}{R} (#1665):
//
//	"Each player discards their hand, then draws seven cards.
//	 Miracle {1}{R} (You may cast this card for its miracle cost when
//	 you draw it if it's the first card you drew this turn.)"
//
// Wheel of Fortune's body (b10EachPlayerWheels): every discard happens
// before any draw, so discard payoffs queue while the spell is still
// resolving. Five mana cast normally, two on a miracle.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ece854f8-8c60-4f30-894f-2286d3dd61b9",
		Name:             "Reforge the Soul",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{1}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B10EachPlayerWheels(ctx.Game, item)
		},
	})
}
