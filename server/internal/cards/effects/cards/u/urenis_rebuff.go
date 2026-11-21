package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ureni's Rebuff — Sorcery {1}{U} (#3931):
//
//	"Return target creature to its owner's hand.
//	 Harmonize {5}{U}"
//
// Unsummon at sorcery speed, and harmonize (CR 702.180).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c0b7e989-8194-484c-93aa-a44df0764a77",
		Name:             "Ureni's Rebuff",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Harmonize("{5}{U}")},
		Targets:          effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.BounceToHand{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
