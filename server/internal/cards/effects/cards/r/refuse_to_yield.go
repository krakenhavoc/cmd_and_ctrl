package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Refuse to Yield — Instant {1}{W} (Streets of New Capenna, #5491):
//
//	"Target creature gets +2/+7 until end of turn. Untap it."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d05aeada-2a9f-464d-8a66-b6e90f43bae3",
		Name:         "Refuse to Yield",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 7)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 7, Label: "Refuse to Yield — +2/+7"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
