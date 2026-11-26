package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Respite — Instant {1}{G} (Conspiracy):
//
//	"Prevent all combat damage that would be dealt this turn. You gain 1 life
//	 for each attacking creature."
//
// The prevention is Root Snare's shield. The life is counted as the spell
// resolves, over every attacking creature on the table, not only yours.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46137f00-f7f5-430a-857c-ff2dbf2d2579",
		Name:         "Respite",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.PreventAllCombatDamageThisTurn{Label: "Respite: prevent combat damage"}).Apply(ctx); err != nil {
				return err
			}
			n := len(effects.AttackingCreatures(ctx.Game))
			if n == 0 {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), n)
		},
	})
}
