package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Compleat Devotion — Instant {1}{W} (Phyrexia: All Will Be One, #5479):
//
//	"Target creature you control gets +2/+2 until end of turn. If that
//	 creature has toxic, draw a card."
//
// The draw is conditional on the creature as the spell resolves, after
// the pump. A target that is gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2a6cfb7-e5bd-49c8-aa51-fd5203eca8b8",
		Name:         "Compleat Devotion",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Compleat Devotion — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if !effects.S82OneB05HasToxic(ctx.Game, id) {
					return nil
				}
				return effects.DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}
