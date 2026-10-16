package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ureni's Rebuff — Sorcery {1}{U} (#3931):
//
//	"Return target creature to its owner's hand.
//	 Harmonize {5}{U}"
//
// Unsummon at sorcery speed, and harmonize (CR 702.180).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c0b7e989-8194-484c-93aa-a44df0764a77",
		Name:             "Ureni's Rebuff",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{5}{U}")},
		Targets:          TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				return BounceToHand{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
