package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Blessing — Sorcery {2}{W} (Portal, #5630):
//
//	"Target creature gets +3/+3 and gains flying until end of turn. (It can't be blocked except by creatures with flying or reach.)"
//
// Mighty Leap's body with +3/+3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3758fca-0522-4b5a-a1cc-3b2b3ab299ba",
		Name:         "Angelic Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "flying")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(3, 3, "Angelic Blessing — +3/+3 and flying", "flying")(ctx.Game, item)
		},
	})
}
