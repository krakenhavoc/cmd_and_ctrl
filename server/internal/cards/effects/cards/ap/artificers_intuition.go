package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "be14b10a-7cac-47fc-a954-685091990291",
		Name:         "Artificer's Intuition",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Discard an artifact card: Search your library for an artifact card with mana value 1 or less, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.DiscardCardsMatching(1, "an artifact card", effects.MatchArtifact)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsArtifact() && c.ManaValue() <= 1 },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Artificer's Intuition — an artifact card with mana value 1 or less, to your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
