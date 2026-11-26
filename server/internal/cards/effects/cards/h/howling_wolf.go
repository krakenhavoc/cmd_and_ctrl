package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howling Wolf — Creature — Wolf {2}{G}{G}, 2/2 (Conspiracy, S82 OLD batch
// 364, #5557):
//
//	"When this creature enters, you may search your library for up to three
//	 cards named Howling Wolf, reveal them, put them into your hand, then
//	 shuffle."
//
// Battalion Foot Soldier's search with a limit of three: the searcher may
// decline, and may take fewer than three; the cards are revealed and put into
// hand, and the library is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d95d0b24-2e6f-4f61-8668-26effeaaf2fb",
		Name:         "Howling Wolf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Howling Wolf — search for up to three cards named Howling Wolf",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.Name == "Howling Wolf" },
						Dest:      game.ZoneHand,
						Limit:     3,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Howling Wolf — you may search for up to three cards named Howling Wolf",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
