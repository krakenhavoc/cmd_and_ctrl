package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Antagonize — Instant {1}{R} (Streets of New Capenna, #5491):
//
//	"Target creature gets +4/+3 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4e860f1-e637-4640-b257-57c2e42793ce",
		Name:         "Antagonize",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 3)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: 4, Toughness: 3, Label: "Antagonize — +4/+3"}.Apply(ctx)
		},
	})
}
