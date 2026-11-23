package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelsong — Instant {1}{W} (Shards of Alara, #5588):
//
//	"Prevent all combat damage that would be dealt this turn.
//	 Cycling {2}"
//
// Fog's text (a CR 615.1 prevention effect for the rest of the turn)
// with the shared Cycling row (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe7bad80-f853-4af9-82e2-5ccf8038d93b",
		Name:         "Angelsong",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PreventAllCombatDamageThisTurn{Label: "Angelsong: prevent combat damage"}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
