package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entomb — Instant for {B}:
//
//	"Search your library for a card, put that card into your
//	 graveyard, then shuffle."
//
// One black mana that puts any card in the deck exactly where a
// reanimation spell wants it. The card does nothing on its own,
// which is the whole design.
//
// "For a card" — no predicate at all. A nil Predicate matches
// everything, which is literally what the card says: lands,
// instants, anything. The searcher picks, because a real library
// holds more than one card and the S22 search chooser queues a
// prompt whenever there is more than Limit to choose from.
//
// Dest is ZoneGraveyard, which is a real search destination (unlike
// ZoneLibrary, which the engine treats as "leave it where it is"
// because a put-on-top is not modelled). Reveal is false: Entomb
// does not say "reveal it", and a card arriving in a graveyard is
// public anyway once it lands.
//
// Shuffle is true and is not optional — it is the third clause of
// the sentence, not a rider on finding something. Searching and
// finding nothing still shuffles, which is what
// SearchLibraryThenForEffect does on a zero-match search.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "299fc083-0834-4064-8344-f895aff68867",
		Name:         "Entomb",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:  item.Controller,
				Dest:    game.ZoneGraveyard,
				Limit:   1,
				Shuffle: true,
				Reason:  "Entomb — put a card into your graveyard",
			}.Apply(ctx)
		},
	})
}
