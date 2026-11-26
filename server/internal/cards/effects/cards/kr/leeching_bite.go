package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leeching Bite — Instant {1}{G}:
//
//	"Target creature gets +1/+1 until end of turn. Another target creature
//	 gets -1/-1 until end of turn."
//
// Each half is judged on its own target at resolution (CR 608.2b): losing one
// does not stop the other.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42e4a55b-12c8-4ad0-a27d-28bbdbb06940",
		Name:         "Leeching Bite",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1), effects.HarmUntilEOT(1, -1, -1)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Leeching Bite — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return effects.BoostUntilEOT{Target: t.ID, Power: -1, Toughness: -1, Label: "Leeching Bite — -1/-1"}.Apply(ctx)
			}
			return nil
		},
	})
}
