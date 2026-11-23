package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain Sisay — Legendary Creature — Human Soldier {2}{G}{W}, 2/2
// (EDHREC rank 4522):
//
//	"{T}: Search your library for a legendary card, reveal that card,
//	 put it into your hand, then shuffle."
//
// The original Selesnya toolbox commander. A repeatable tutor with no
// mana cost attached is the strongest thing a four-drop can offer,
// and "legendary CARD" — not "legendary permanent", not "legendary
// creature" — is what makes it a deck rather than a value engine:
// every legendary land, artifact, enchantment, planeswalker and
// sorcery in the format is a Sisay target.
//
// The activation is a bare tap with no mana, so summoning sickness is
// the only gate (CR 302.6, which the engine applies to any {T} cost
// on a creature). The search is the catalog-wide deterministic
// chooser: the engine offers every legendary card in the library and
// the activating player picks one. The card is REVEALED, so the whole
// table sees what was found, as printed.
//
// "Legendary" is the supertype read off the card in the library, so a
// permanent made legendary on the battlefield is irrelevant here and
// a legendary instant is a legal find.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e7a9ea9-f76c-4c12-950f-c613fa16cfa8",
		Name:         "Captain Sisay",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Search your library for a legendary card, reveal that card, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsLegendary() },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Choose a legendary card to put into your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
