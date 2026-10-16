package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "99eb4f63-841f-4b6b-9794-82bc2ed7b758",
		Name:          "Spectral Shift",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Change the text of target spell or permanent by replacing all instances of one basic land type with another.",
				TargetSpellOrPermanent("target spell or permanent", nil, nil),
				spectralShiftChange(textLandWords)),
			ModeDoing("Change the text of target spell or permanent by replacing all instances of one color word with another.",
				TargetSpellOrPermanent("target spell or permanent", nil, nil),
				spectralShiftChange(textColorWords)),
		)),
	})
}

// spectralShiftChange is one bullet: the change made to the bullet's
// own target, if it is still legal (CR 608.2b).
func spectralShiftChange(kinds []game.TextWordKind) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		return ChangeTextWord{Target: t.ID, Kinds: kinds}.Apply(ctx)
	}
}
