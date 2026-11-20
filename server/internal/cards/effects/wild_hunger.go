package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild Hunger — Instant {2}{G} (Dark Ascension, #5571):
//
//	"Target creature gets +3/+1 and gains trample until end of turn.
//	 Flashback {3}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Brute Strength's pump through pumpAndGrantFirstTarget, with flashback
// (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9d250088-d5f2-40ed-bdb7-a6decce0fb67",
		Name:             "Wild Hunger",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		Purpose:          ForTargets(BenefitUntilEOT(0, 3, 1, "trample")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(3, 1, "Wild Hunger — +3/+1 and trample", "trample")(ctx.Game, item)
		},
	})
}
