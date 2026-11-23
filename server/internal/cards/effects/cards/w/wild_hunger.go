package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         "9d250088-d5f2-40ed-bdb7-a6decce0fb67",
		Name:             "Wild Hunger",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		Purpose:          effects.ForTargets(effects.BenefitUntilEOT(0, 3, 1, "trample")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(3, 1, "Wild Hunger — +3/+1 and trample", "trample")(ctx.Game, item)
		},
	})
}
