package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_restriction.go — #1985 (ADR 0109 amendment of 2026-10-15):
// constructors for Spec.EntryRestrictions, "… can't enter the
// battlefield" (CR 101.2).
//
// One constructor per printed SHAPE, as cast_restriction.go and
// land_play_restriction.go have, because the shape decides which object
// the clause is about:
//
//   - LandsCantEnter is about the PERMANENT the object would be (CR
//     614.12's look-ahead): a creature entering as a copy of an artifact
//     land is a land, a land manifested face down is not, and a token
//     land is one.
//   - CardsInZonesCantEnter is about the CARD in the zone it is leaving
//     ("creature cards in graveyards"): its characteristics there, and
//     never a token, which is not a card and comes from no zone.
//
// The predicate runs under g.mu inside the CR 614 window. Read-only.

// LandsCantEnter is Worms of the Earth's "Lands can't enter the
// battlefield." From anywhere, under anyone's control, a token included:
// a land play is refused before it is made, a fetched land stays in the
// library (which is still shuffled), a reanimated one stays in the
// graveyard, a flickered one stays in exile, and a land token is not
// created.
func LandsCantEnter(label string) game.EntryRestriction {
	return game.EntryRestriction{
		Label: label,
		Forbids: func(q game.EntryQuery) bool {
			return q.WouldBe().IsLand()
		},
	}
}

// CardsInZonesCantEnter is Grafdigger's Cage's "Creature cards in
// graveyards and libraries can't enter the battlefield" and its kin
// (Kunoros, Soulless Jailer, Weathered Runestone): a card leaving one of
// `zones` for the battlefield, judged as it is in that zone by `match`.
// A token is never refused (it is not a card).
func CardsInZonesCantEnter(label string, match func(c game.Card) bool, zones ...game.ZoneKind) game.EntryRestriction {
	from := make(map[game.ZoneKind]bool, len(zones))
	for _, z := range zones {
		from[z] = true
	}
	return game.EntryRestriction{
		Label: label,
		Forbids: func(q game.EntryQuery) bool {
			return !q.Token && from[q.FromZone] && match(q.Card)
		},
	}
}
