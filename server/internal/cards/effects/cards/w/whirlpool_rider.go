package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whirlpool Rider — Creature — Merfolk {1}{U}, 1/1 (Apocalypse, S82 OLD, #5614):
//
//	"When this creature enters, shuffle the cards from your hand into your
//	 library, then draw that many cards."
//
// Winds of Change's shuffle-and-redraw over the controller alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13a3f2c1-0454-4661-a462-542254f8360d",
		Name:         "Whirlpool Rider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Whirlpool Rider — shuffle the cards from your hand into your library, then draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.WindsOfChangeShuffleFrom(g, []uuid.UUID{item.Controller}, nil)
				}),
		},
	})
}
