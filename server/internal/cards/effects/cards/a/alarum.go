package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alarum — Instant {1}{W} (Mirage, #5633):
//
//	"Untap target nonattacking creature. It gets +1/+3 until end of turn."
//
// The target is judged nonattacking at announce and again as the spell resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18e96674-6fd8-4046-a3fe-7ba80138d85d",
		Name:         "Alarum",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nonattacking creature", effects.Not(effects.AttackingCreature())),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 3, Label: "Alarum — +1/+3"}.Apply(ctx)
		},
	})
}
