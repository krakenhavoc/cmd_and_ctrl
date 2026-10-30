package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moon-Blessed Cleric — Creature — Human Elf Cleric {2}{W}, 3/2:
//
//	"Divine Intervention — When this creature enters, you may search
//	 your library for an enchantment card, reveal it, then shuffle and
//	 put that card on top."
//
// The creature-borne Enlightened Tutor: SearchLibrary with ToTop is the
// shared search (find, reveal, shuffle, then the card goes on top), with
// Optional set because the clause says "you may search": the prompt
// always offers "take nothing". A search with a stated quality (an
// enchantment) may come up empty anyway (CR 701.23b); Optional is the
// printed "you MAY" in the data. "Divine Intervention" is an ability
// word, so it rides only on the stack label.
//
// The "may" is a real one (#4134): the prompt carries "Don't search",
// which skips the search outright (no look, no shuffle, no "whenever a
// player searches" trigger), as well as "find nothing", which is a search
// that came up empty and still shuffles (CR 701.23).
//
// The card is revealed, so every seat sees what went on top; the
// shuffle has already buried the rest of the order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89ed0598-a3ee-4d14-9a35-25e0dd80d9fb",
		Name:         "Moon-Blessed Cleric",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Divine Intervention — search for an enchantment card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsEnchantment() },
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Divine Intervention — an enchantment card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
