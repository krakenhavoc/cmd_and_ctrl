package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planeswalker's Mischief — Enchantment {2}{U} (Planeshift, S82 OLD #5615):
//
//	"{3}{U}: Target opponent reveals a card at random from their hand. If it's
//	 an instant or sorcery card, exile it. You may cast it without paying its
//	 mana cost for as long as it remains exiled. At the beginning of the next
//	 end step, if you haven't cast it, return it to its owner's hand. Activate
//	 only as a sorcery."
//
// The random reveal as Planeswalker's Favor's, then an instant or sorcery card
// is exiled and its controller is given a free cast that lasts while it stays
// in exile (CR 608.2g). A delayed trigger at the next end step returns it to
// its owner's hand only while it is still the object that was exiled: a cast
// card has left exile, and a card that left and came back is a new object
// (CR 400.7). Other cards are revealed and stay in hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78e87805-19e3-415b-ad0b-3275183d7297",
		Name:         "Planeswalker's Mischief",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{U}: Target opponent reveals a card at random from their hand. If it's an instant or sorcery card, exile it. You may cast it without paying its mana cost for as long as it remains exiled. At the beginning of the next end step, if you haven't cast it, return it to its owner's hand. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{U}"),
			Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.Instruction(effects.OldB639MischiefReveal)),
		}},
	})
}
