package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Micromancer — Creature — Human Wizard {3}{U}, 3/3 (Foundations, #4465):
//
//	"When this creature enters, you may search your library for an
//	 instant or sorcery card with mana value 1, reveal it, put it into
//	 your hand, then shuffle."
//
// Spellseeker's tutor with the mana value pinned to exactly 1. "You may
// search" is the search's Optional flag, so the searcher can decline the
// card and the shuffle; the mana value is the printed one, which is what
// a card in a library has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2605782-09f1-4d52-b3e7-5ec1442bd9b5",
		Name:         "Micromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Micromancer — search for an instant or sorcery with mana value 1", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: micromancerFinds,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Micromancer — an instant or sorcery card with mana value 1, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// micromancerFinds is "an instant or sorcery card with mana value 1".
func micromancerFinds(c game.Card) bool {
	return (c.IsInstant() || c.IsSorcery()) && c.ManaValue() == 1
}
