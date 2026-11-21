package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flaring Pain — Instant {1}{R}:
//
//	"Damage can't be prevented this turn.
//	 Flashback {R}"
//
// The whole card is ADR 0107 §5's turn grant (ModDamageCantBePrevented,
// CR 615.12), which covers every damage event until cleanup, from any
// source (CR 611.2c). Flashback is the shared alternative cost.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "cad19ad2-3ea9-43f5-bb5a-b0793402fab0",
		Name:             "Flaring Pain",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageCantBePreventedThisTurn{}.Apply(ctx)
		},
	})
}
