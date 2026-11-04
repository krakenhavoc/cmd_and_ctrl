package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3bbad38b-7d88-4ff2-9fd3-5081536e7c56",
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "Grandeur — Discard another card named Page, Loose Leaf: Reveal cards from the top of your library until you reveal an instant or sorcery card. Put that card into your hand and the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    DiscardCardsMatching(1, "another card named Page, Loose Leaf", func(c game.Card) bool { return c.Name == name }),
			Effect: socB30RevealUntilToHandRestBottom(name+" — Grandeur",
				func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() }),
		}},
	})
}
