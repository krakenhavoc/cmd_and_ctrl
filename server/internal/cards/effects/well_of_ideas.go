package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Well of Ideas — Enchantment {5}{U} (Commander 2014, #5553):
//
//	"When this enchantment enters, draw two cards.
//	 At the beginning of each other player's draw step, that player draws
//	 an additional card.
//	 At the beginning of your draw step, draw two additional cards."
//
// Three triggers. The enters trigger draws two for its controller. The two
// draw-step triggers watch EventBeginDrawStep, whose Actor is the active
// player (Font of Mythos' read): the active player draws one more on
// another player's turn, and you draw two more on yours. None is a "may".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ed24fc6-b1d5-47f8-a1d1-3243372e0c9f",
		Name:         "Well of Ideas",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Well of Ideas — draw two cards", Do(DrawCards{N: 2})),
			On(game.EventBeginDrawStep, wellOfIdeasOtherPlayersDrawStep,
				"Well of Ideas — that player draws an additional card",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(NewContext(g, item))
				}),
			On(game.EventBeginDrawStep, ByYou, "Well of Ideas — draw two additional cards", Do(DrawCards{N: 2})),
		},
	})
}

// wellOfIdeasOtherPlayersDrawStep is "at the beginning of each other
// player's draw step": a draw step whose active player is not the Well's
// controller.
func wellOfIdeasOtherPlayersDrawStep(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor != uuid.Nil && ev.Actor != source.Controller
}
