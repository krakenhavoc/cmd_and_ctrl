package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Page, Loose Leaf — Legendary Artifact Creature — Construct {2}, 0/2
// (Secrets of Strixhaven, #4456):
//
//	"{T}: Add {C}.
//	 Grandeur — Discard another card named Page, Loose Leaf: Reveal cards
//	 from the top of your library until you reveal an instant or sorcery
//	 card. Put that card into your hand and the rest on the bottom of your
//	 library in a random order."
//
// Grandeur is an ordinary activated ability of the permanent whose cost
// discards a card with this name from hand (Oriss, Samite Guardian's shape);
// "another" is any such card, since this one is on the battlefield. The whole
// run is revealed (token cards in a library never count as the instant or
// sorcery, CR 108.2); with no instant or sorcery in the library the whole
// library goes to the bottom in a random order and nothing reaches the hand.
//
// No simplification.
func init() {
	const name = "Page, Loose Leaf"
	effects.Register(effects.Spec{
		OracleID:     "3bbad38b-7d88-4ff2-9fd3-5081536e7c56",
		Name:         name,
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Grandeur — Discard another card named Page, Loose Leaf: Reveal cards from the top of your library until you reveal an instant or sorcery card. Put that card into your hand and the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.DiscardCardsMatching(1, "another card named Page, Loose Leaf", func(c game.Card) bool { return c.Name == name }),
			Effect: effects.SocB30RevealUntilToHandRestBottom(name+" — Grandeur",
				func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() }),
		}},
	})
}
