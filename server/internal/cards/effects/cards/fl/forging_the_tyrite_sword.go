package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forging the Tyrite Sword — Enchantment — Saga {1}{R}{W} (Kaldheim, #5505):
//
//	"I, II — Create a Treasure token.
//	 III — Search your library for a card named Halvar, God of Battle or an
//	 Equipment card, reveal it, put it into your hand, then shuffle."
//
// Chapters I and II are one Treasure each. Chapter III is a search for one
// card that has the name Halvar, God of Battle (a double-faced card in the
// library has its front face's name, CR 712.8a) or the Equipment subtype,
// revealed, then the shuffle. The searcher may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	const name = "Forging the Tyrite Sword"
	treasure := effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})
	effects.Register(effects.Spec{
		OracleID:     "47a8f3a8-253b-4df9-9cfc-ff7df65d850c",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a Treasure token"), treasure),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create a Treasure token"), treasure),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "search your library for a card named Halvar, God of Battle or an Equipment card, reveal it, put it into your hand, then shuffle"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return game.HasName(c, "Halvar, God of Battle") || c.HasSubtype("Equipment")
						},
						Dest:    game.ZoneHand,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Reason:  name + " — search for Halvar, God of Battle or an Equipment card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
