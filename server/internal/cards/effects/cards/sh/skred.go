package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skred — Instant {R} (Coldsnap):
//
//	"Skred deals damage to target creature equal to the number of snow
//	 permanents you control."
//
// The count is read once as Skred resolves (CR 608.2h), and is the snow
// permanents its controller controls then (CR 205.4g). A target that is no
// longer legal fizzles the spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "026318be-e3dd-4d5c-bd77-53ce23a2e7cc",
		Name:         "Skred",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := len(effects.PermanentsControlledByMatching(ctx.Game, item.Controller, effects.Snow()))
			if n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: n}.Apply(ctx)
		},
	})
}
