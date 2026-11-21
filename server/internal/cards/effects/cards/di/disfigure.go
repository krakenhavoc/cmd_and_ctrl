package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disfigure — Instant {B}: "Target creature gets -2/-2 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77eafe49-b9c5-461d-89c5-cec217dd2974",
		Name:         "Disfigure",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -2,
				Toughness: -2,
				Label:     "Disfigure — -2/-2",
			}.Apply(ctx)
		},
	})
}
