package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artificer's Intuition — Enchantment {1}{U} (Fifth Dawn, #5604):
//
//	"{U}, Discard an artifact card: Search your library for an artifact
//	 card with mana value 1 or less, reveal it, put it into your hand,
//	 then shuffle."
//
// Fauna Shaman's tutor with no tap, an artifact card discarded and an
// artifact card with mana value 1 or less found. The discard is part of
// the cost, picked from the hand at activation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be14b10a-7cac-47fc-a954-685091990291",
		Name:         "Artificer's Intuition",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}, Discard an artifact card: Search your library for an artifact card with mana value 1 or less, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{U}"), DiscardCardsMatching(1, "an artifact card", MatchArtifact)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsArtifact() && c.ManaValue() <= 1 },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Artificer's Intuition — an artifact card with mana value 1 or less, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
