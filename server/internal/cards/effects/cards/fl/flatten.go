package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flatten — Instant {3}{B} (Dragons of Tarkir,
// #5551):
//
//	"Target creature gets -4/-4 until end of turn."
//
// Dark Deed's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc5d5084-77de-4dc5-8769-300eafcbfee7",
		Name:         "Flatten",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, -4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: -4, Toughness: -4, Label: "Flatten — -4/-4 until end of turn"}.Apply(ctx)
		},
	})
}
