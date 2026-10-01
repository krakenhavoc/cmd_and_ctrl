package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// transmute.go — transmute (CR 702.53), first needed by Shred Memory
// (#1807).
//
// CR 702.53a: "Transmute is an activated ability that functions only
// while the card with transmute is in a player's hand. 'Transmute
// [cost]' means '[Cost], Discard this card: Search your library for a
// card with the same mana value as the discarded card, reveal that
// card, and put it into your hand. Then shuffle your library. Activate
// only as a sorcery.'"
//
// It is cycling's shape (cycling.go) with a different effect and
// sorcery timing: an ordinary CR 602 activation from the hand through
// ActivatedAbilityShape.Zones, paying with DiscardThis(). It is NOT a
// cycling ability, so it leaves the Cycling bit unset and a "whenever
// you cycle" watcher does not see it.

// Transmute is "Transmute <cost>". `cost` is the printed mana cost of
// the ability; the discard, the hand-only zone and the sorcery timing
// come with the keyword.
//
// The mana value is the DISCARDED card's (CR 702.53a), read off the
// card where it now is: a card's printed cost does not change between
// the hand and the graveyard. A card whose cost the engine cannot read
// matches nothing, as every mana-value comparison in the engine does
// (game.Card.ParsedManaValue); the search still shuffles.
func Transmute(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label: "Transmute " + cost + " (" + cost + ", Discard this card: Search your library for a card with the same mana value as this card, " +
			"reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)",
		Cost:         Plus(ManaCost(cost), DiscardThis()),
		Zones:        []game.ZoneKind{game.ZoneHand},
		SorcerySpeed: true,
		Effect: func(g *game.Game, item *game.StackItem) error {
			discarded, found := g.LookupCardForEffect(item.SourceCardID)
			mv, known := discarded.ParsedManaValue()
			return SearchLibrary{
				Player: item.Controller,
				Predicate: func(c game.Card) bool {
					if !found || !known {
						return false
					}
					cmv, ok := c.ParsedManaValue()
					return ok && cmv == mv
				},
				Dest:    game.ZoneHand,
				Limit:   1,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Transmute — a card with mana value " + numberWord(mv) + ", to your hand",
				Source:  item.SourceCardID,
			}.Apply(NewContext(g, item))
		},
	}
}
