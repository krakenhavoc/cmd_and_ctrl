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
// Sandbox simplification, declared: an Aura fetched this way enters
// UNATTACHED and stays that way. CR 303.4f lets Zur's controller
// choose what it enchants as it enters; the search path has no such
// prompt (#478), and the attachment state-based action sweeps only an
// Aura attached to something illegal, not one attached to nothing.
// Weaker than printed (the Aura does nothing), never stronger — and
// it is the Aura half of Zur's toolbox (Shielded by Faith, Diplomatic
// Immunity) that this loses.
func init() {
	Register(Spec{
		OracleID:        "d7950018-d744-48a8-81aa-0d8384703f48",
		Name:            "Zur the Enchanter",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"An Aura fetched this way enters unattached and stays that way — you don't get to choose what it enchants."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Zur the Enchanter — you may search for an enchantment with mana value 3 or less", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsEnchantment() && manaValueOf(c) <= 3 },
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
