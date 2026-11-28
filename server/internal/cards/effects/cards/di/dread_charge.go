package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Charge — Sorcery {3}{B} (Portal, S82 OLD batch 781, #5630):
//
//	"Black creatures you control can't be blocked this turn except by black
//	 creatures."
//
// Each black creature the caster controls as it resolves gets the shared
// "can't be blocked this turn except by" rule, now with a colour allowance
// (CantBeBlockedThisTurnExceptBy.Colors): a creature that becomes black or
// enters later is not affected (CR 611.2c), and the blocker's colour is read
// as blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b7933b3-821a-4d64-9a10-af8a69fa008e",
		Name:         "Dread Charge",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl(), effects.OfColor("B"))) {
				if err := (effects.CantBeBlockedThisTurnExceptBy{
					Target: c.InstanceID,
					Colors: []string{"B"},
					Text:   "black creatures",
					Label:  "Dread Charge — can't be blocked except by black creatures",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
