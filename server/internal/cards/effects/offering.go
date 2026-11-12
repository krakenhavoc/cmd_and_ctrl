package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// offering.go — offering, CR 702.48 (#5657).
//
// CR 702.48a: "[Quality] offering" means "As an additional cost to cast
// this spell, you may sacrifice a [quality] permanent. If you chose to
// pay the additional cost, this spell's total cost is reduced by the
// sacrificed permanent's mana cost, and you may cast this spell any time
// you could cast an instant."
//
// One optional additional cost with three parts, each the engine's own:
//
//   - the sacrifice is the optional cost's Sacrifice clause, chosen with
//     the spell's other choices (CR 702.48b, 601.2b) and paid with the
//     spell already on the stack (601.2h), so the permanent's "when this
//     is put into a graveyard" trigger goes above the spell;
//   - the timing is AsThoughFlash (CR 601.3c): the instant-speed window
//     opens for the cast that announces the offering and no other;
//   - the price is ReducedBySacrificedManaCost (game/offering.go): the
//     permanent's mana cost comes off the total cost symbol by symbol
//     (CR 702.48c), after the commander tax and any increase, never
//     below {0}.
//
// It is an additional cost, not an alternative one, so an offering cast
// may still claim an alternative cost (CR 118.9a) and the reduction
// applies to that (CR 118.9d).

// OfferingKey is the optional cost's identity on the card.
const OfferingKey = "offering"

// Offering is "[Quality] offering". `quality` is the printed word, used
// in the label and the picker prompt; the predicates are what the
// sacrificed permanent must match. Blast-Furnace Hellkite's "Artifact
// offering" is
//
//	OptionalCosts: []game.AdditionalCost{Offering("artifact", Artifact())},
func Offering(quality string, preds ...CardPredicate) game.AdditionalCost {
	what := "a " + quality
	if strings.ContainsRune("aeiou", rune(quality[0])) {
		what = "an " + quality
	}
	return game.AdditionalCost{
		Optional:                    true,
		Key:                         OfferingKey,
		Label:                       strings.ToUpper(quality[:1]) + quality[1:] + " offering",
		Sacrifice:                   sacrificeSpec(what, preds...),
		AsThoughFlash:               true,
		ReducedBySacrificedManaCost: true,
	}
}
