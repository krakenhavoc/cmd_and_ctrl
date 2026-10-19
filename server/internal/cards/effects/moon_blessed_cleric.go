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
// One engine-wide convention applies, the same one Path to Exile,
// Assassin's Trophy and Solemn Simulacrum live under: a declined
// "may search" is answered as a search that finds nothing, and that
// still shuffles, where a table that declined outright would not. The
// engine has no separate "decline to search" answer (the search prompt
// carries card picks only), and the difference is visible only to
// a player who wanted to keep a library order they already knew.
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
