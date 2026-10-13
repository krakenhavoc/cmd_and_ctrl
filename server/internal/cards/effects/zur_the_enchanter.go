package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zur the Enchanter — Legendary Creature — Human Wizard {1}{W}{U}{B},
// 1/4 (EDHREC rank 2817):
//
//	"Flying
//	 Whenever Zur attacks, you may search your library for an
//	 enchantment card with mana value 3 or less, put it onto the
//	 battlefield, then shuffle."
//
// The enchantment toolbox commander. The attack trigger is a "you
// may" search — the search prompt's decline — for an enchantment card
// of mana value at most 3, which enters the battlefield untapped
// through the search path, so its own ETB triggers and enters-tapped
// clauses fire.
//
// An Aura fetched this way — the Aura half of Zur's toolbox (Shielded
// by Faith, Diplomatic Immunity) — chooses what it enchants as it
// enters (CR 303.4f, #3158); with nothing to enchant it stays in the
// library (CR 303.4g), and the library is shuffled either way.
func init() {
	Register(Spec{
		OracleID:        "d7950018-d744-48a8-81aa-0d8384703f48",
		Name:            "Zur the Enchanter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Zur the Enchanter — you may search for an enchantment with mana value 3 or less", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsEnchantment() && c.ManaValue() <= 3 },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Zur the Enchanter — an enchantment card with mana value 3 or less, onto the battlefield",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
