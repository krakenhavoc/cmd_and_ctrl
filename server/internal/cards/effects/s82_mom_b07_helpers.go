package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b07_helpers.go — S82, March of the Machine batch 7: the cards
// that read the creatures that convoked them.

// CountersPerConvoked is "this permanent enters with <per> <kind>
// counters on it for each creature that convoked it" (CR 702.51,
// CR 614.1c): Ancient Imperiosaur. The count is the creatures tapped to
// convoke the spell (PaidCost.Convoked), none for a spell that was not
// convoked or a copy of one (CR 707.10).
func CountersPerConvoked(kind string, per int) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast game.CastCounts) int { return per * cast.Convoked },
	}
}

// convokersPayload is the creatures that convoked the spell that became
// `source`, as the Payload of a trigger item built while it is still on
// the battlefield (CR 400.7d), so the trigger resolves the same if the
// permanent has since left.
func ConvokersPayload(source *game.Card) []game.TargetRef {
	var out []game.TargetRef
	for _, id := range source.ConvokedBy() {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

// activatedPlaneswalkerOrTapped is Cut Short's target: "planeswalker that
// was activated this turn or tapped creature". A planeswalker has been
// activated when one of its loyalty abilities was (CR 606), which is the
// per-permanent tally Game.LoyaltyActivatedThisTurn keeps.
func ActivatedPlaneswalkerOrTapped() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		if c.IsPlaneswalker() && g.LoyaltyActivatedThisTurn[c.InstanceID] {
			return true
		}
		return c.IsCreature() && c.Tapped
	}
}
