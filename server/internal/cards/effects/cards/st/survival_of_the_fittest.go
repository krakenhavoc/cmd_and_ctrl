package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Survival of the Fittest — Enchantment {1}{G} (Exodus, S82 OLD batch 742, #5626):
//
//	"{G}, Discard a creature card: Search your library for a creature card,
//	 reveal that card, put it into your hand, then shuffle."
//
// Fauna Shaman's ability without the {T}: the creature card is discarded as a
// cost at announce (CR 602.2b) and the search is the plain "reveal it, put it
// into your hand, then shuffle" (CR 701.23). It was a declared skip in
// batch22_test.go until #660's discard-a-card cost landed; that entry is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "119d719d-e965-45b4-9bc9-ac03211b10c2",
		Name:         "Survival of the Fittest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Discard a creature card: Search your library for a creature card, reveal that card, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.DiscardCardsMatching(1, "a creature card", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.MatchCreature,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Survival of the Fittest — a creature card, to your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
