package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flame Jab — Sorcery {R}:
//
//	"Flame Jab deals 1 damage to any target.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: `CastableZones` opens the graveyard and
// `Retrace(cost)` prices it (ADR 0066, 2026-10-07 amendment). The spell goes
// back to the graveyard when it resolves, so it can be retraced again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "05eccdee-84f8-42d5-b79c-36d081656915",
		Name:             "Flame Jab",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{R}")},
		Targets:          TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return DealDamage{
				Source: ctx.Source(),
				Target: item.Targets[0].ID,
				Amount: 1,
			}.Apply(ctx)
		},
	})
}
