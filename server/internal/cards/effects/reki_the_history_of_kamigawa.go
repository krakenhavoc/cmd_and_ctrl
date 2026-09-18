package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reki, the History of Kamigawa — Legendary Creature — Human Shaman
// {2}{G}, 1/2 (EDHREC rank 4323):
//
//	"Whenever you cast a legendary spell, draw a card."
//
// Three mana for a card every time the legend deck does the thing it
// was already going to do. In Commander, where every deck casts at
// least one legendary spell a game and a "superfriends" or Kamigawa
// list casts several a turn, Reki is the cheapest draw engine that
// asks for no deckbuilding concession at all.
//
// "A LEGENDARY spell", not "a legendary creature spell": a legendary
// artifact, enchantment, land — lands are not cast — planeswalker or
// instant all count, which is why the predicate is the bare Legendary()
// supertype read rather than a type filter. Reki himself triggers when
// you cast HIM: the trigger is on the permanent, and the permanent is
// not on the battlefield while its own spell is on the stack, so it
// does not. That is CR 603.2 working, not a gap — Reki has to be on
// the battlefield to see a cast.
//
// Legendary() reads the effective type line, so a spell something made
// legendary on the stack counts and a token copy printed "except it
// isn't legendary" does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a7d68ae-ac43-46a6-9dc8-d6b07cc0333c",
		Name:         "Reki, the History of Kamigawa",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Legendary(), "Reki, the History of Kamigawa — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}
