package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explosive Derailment — Instant {R}:
//
//	"Spree (Choose one or more additional costs.)
//	 + {2} — Explosive Derailment deals 4 damage to target creature.
//	 + {2} — Destroy target artifact."
//
// Spree proof card #2 (CR 702.172a, ADR 0065's 2026-09-23 amendment):
// two bullets, each independently targeted, each priced identically
// at {2} — casting both announces {R} + {2} + {2} = {4}{R} and a
// creature target alongside an artifact target, neither of which
// constrains the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b23dc81d-01bb-4bf0-9932-5d32a6b22cf7",
		Name:         "Explosive Derailment",
		Completeness: effects.CompletenessFull,
		Modes: effects.Spree(
			effects.ModeWithPurpose(effects.SpreeModeDoing("Explosive Derailment deals 4 damage to target creature.", "{2}",
				effects.TargetCreature("target creature"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
				}), effects.ForTargets(effects.DamageToTarget(0, 4))),
			effects.SpreeModeDoing("Destroy target artifact.", "{2}",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
		),
	})
}
