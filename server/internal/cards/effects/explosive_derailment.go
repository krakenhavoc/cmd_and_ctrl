package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b23dc81d-01bb-4bf0-9932-5d32a6b22cf7",
		Name:         "Explosive Derailment",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Explosive Derailment deals 4 damage to target creature.", "{2}",
				TargetCreature("target creature"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
				}),
			SpreeModeDoing("Destroy target artifact.", "{2}",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
		),
	})
}
