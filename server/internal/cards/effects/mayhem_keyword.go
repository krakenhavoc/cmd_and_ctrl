package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mayhem_keyword.go — #4628 (CR 702.187, ADR 0025 amendment 2026-10-30).

// AltCostKeyMayhem is mayhem's offer key. It rides cast_spell as
// `alternative_cost` and lands on StackItem.AltCost. An on-disk
// identity: never renamed.
const AltCostKeyMayhem = "mayhem"

// Mayhem is "Mayhem [cost] (You may cast this card from your graveyard
// for [cost] if you discarded it this turn. Timing rules still apply.)"
// — CR 702.187b, an alternative cost (CR 601.2b, 601.2f-h):
//
//	CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
//	AlternativeCosts: []game.AlternativeCost{Mayhem("{4}{R}")},
//
// Flashback's place and price without its exile: the offer is bound to
// the graveyard, a card that opens the zone and prints a bound offer
// can be cast from there only by claiming it, and the spell goes to
// the graveyard like any other when it leaves the stack, to be cast
// again next time it is discarded. The zone opens only while the card
// was discarded this turn (MayhemCastableZone) and the offer is
// claimable only then (RequiresDiscardedThisTurn), both judged by
// Game.DiscardedThisTurn, which is per object (CR 400.7).
func Mayhem(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:                       AltCostKeyMayhem,
		Label:                     "Mayhem " + cost,
		ManaCost:                  cost,
		FromZone:                  game.ZoneGraveyard,
		RequiresDiscardedThisTurn: true,
	}
}

// MayhemCastableZone is the graveyard opening every mayhem card
// declares in Spec.CastableZonesIf: "as long as you discarded this card
// this turn" (CR 702.187b), asked of the card's owner.
func MayhemCastableZone() game.ConditionalCastZone {
	return game.ConditionalCastZone{Zone: game.ZoneGraveyard, Cond: discardedThisTurn}
}

// discardedThisTurn is MayhemCastableZone's condition: a pure read of
// the turn tally for the card object being asked about.
func discardedThisTurn(g *game.Game, _ uuid.UUID, card game.Card) bool {
	return g.DiscardedThisTurn(card.InstanceID)
}
