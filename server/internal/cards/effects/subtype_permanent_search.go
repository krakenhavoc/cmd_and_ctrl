package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// searchForASubtypePermanentWithManaValueAtMost is the Mercadian Masques
// recruiters' ability (#5619): "Search your library for a <subtype>
// permanent card with mana value N or less, put it onto the
// battlefield, then shuffle." The Ramosian Rebels and the Cateran
// Mercenaries print it at different N.
//
// A permanent card (CR 110.4) of that subtype, read through the lens the
// search lends the source's text (HasSubtypeWord), so a changeling card
// counts and a text change to the source's word is followed. Mana value
// is the card's own, with X as 0 (CR 202.3e). The search names a
// quality, so the searcher may find nothing (CR 701.23b); the library
// is shuffled either way.
func searchForASubtypePermanentWithManaValueAtMost(subtype string, n int, reason string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player: item.Controller,
			Predicate: func(c game.Card) bool {
				return isPermanentCard(c) && c.HasSubtypeWord(subtype) && c.ManaValue() <= n
			},
			Dest:    game.ZoneBattlefield,
			Limit:   1,
			Shuffle: true,
			Reason:  reason,
			Source:  item.SourceCardID,
		}.Apply(NewContext(g, item))
	}
}
