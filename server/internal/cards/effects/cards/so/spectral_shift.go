package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Shift — Instant {1}{U}:
//
//	"Choose one —
//	 • Change the text of target spell or permanent by replacing all
//	   instances of one basic land type with another. (This effect lasts
//	   indefinitely.)
//	 • Change the text of target spell or permanent by replacing all
//	   instances of one color word with another. (This effect lasts
//	   indefinitely.)
//	 Entwine {2} (Choose both if you pay the entwine cost.)"
//
// Magical Hack and Sleight of Mind as two bullets (CR 612.2, ADR 0160's
// 2026-10-15 amendment), with entwine (CR 702.42, #3866): paying {2}
// more takes both, each with its own target — the same object may be
// both, one target word per bullet (CR 115.3). Entwined, the land-type
// change is made first, its two words chosen before the colour change
// is asked (CR 702.42b, the printed order; the walk waits on the
// prompts, #2789).
//
// Graded Full on the same footing as the two cards it joins: the words
// the substitution cannot reach are the target's gaps.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "99eb4f63-841f-4b6b-9794-82bc2ed7b758",
		Name:          "Spectral Shift",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Change the text of target spell or permanent by replacing all instances of one basic land type with another.",
				effects.TargetSpellOrPermanent("target spell or permanent", nil, nil),
				spectralShiftChange(effects.TextLandWords)),
			effects.ModeDoing("Change the text of target spell or permanent by replacing all instances of one color word with another.",
				effects.TargetSpellOrPermanent("target spell or permanent", nil, nil),
				spectralShiftChange(effects.TextColorWords)),
		)),
	})
}

// spectralShiftChange is one bullet: the change made to the bullet's
// own target, if it is still legal (CR 608.2b).
func spectralShiftChange(kinds []game.TextWordKind) func(*game.StackItem, *effects.Context, int) error {
	return func(_ *game.StackItem, ctx *effects.Context, occ int) error {
		t, ok := effects.ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		return effects.ChangeTextWord{Target: t.ID, Kinds: kinds}.Apply(ctx)
	}
}
