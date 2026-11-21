package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unnatural Speed — Instant — Arcane {R} (Champions of Kamigawa, #5603):
//
//	"Target creature gains haste until end of turn."
//
// Expedite without the draw; the haste ends in the cleanup step (CR
// 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa46ced5-b701-4aa3-b4a1-36fb15b80696",
		Name:         "Unnatural Speed",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"haste"},
				Label:    "Unnatural Speed — haste until end of turn",
			}.Apply(ctx)
		},
	})
}
