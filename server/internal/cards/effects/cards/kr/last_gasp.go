package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Last Gasp — Instant {1}{B}:
//
//	"Target creature gets -3/-3 until end of turn."
//
// Caustic Exhale's second half without the Dragon: a -3/-3 boost that
// state-based actions turn into a death when toughness reaches zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a82c3860-4dd6-4ffd-aa8f-ab8df687db6c",
		Name:         "Last Gasp",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -3,
				Toughness: -3,
				Label:     "Last Gasp — -3/-3 until end of turn",
			}.Apply(ctx)
		},
	})
}
