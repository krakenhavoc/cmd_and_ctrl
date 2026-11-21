package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infuriate — Instant {R}:
//
//	"Target creature gets +3/+2 until end of turn."
//
// Giant Growth's shape with printed +3/+2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d23119ca-73d1-4a48-a8d0-dab8919d891f",
		Name:         "Infuriate",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: 3, Toughness: 2, Label: "Infuriate — +3/+2"}.Apply(ctx)
		},
	})
}
