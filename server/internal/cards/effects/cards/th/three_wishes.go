package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Three Wishes — Instant {1}{U}{U} (Visions, S82 older sets batch 792, #5631):
//
//	"Exile the top three cards of your library face down. You may look at those
//	 cards for as long as they remain exiled. Until your next turn, you may play
//	 those cards. At the beginning of your next upkeep, put any of those cards
//	 you didn't play into your graveyard."
//
// Gonti's face-down impulse exile over three cards, the window being "until your
// next turn" (the play permission ends as that turn begins); the cards still in
// exile as the same objects go to their owner's graveyard at the beginning of
// the caster's next upkeep (CR 603.7), and a card that was played has left exile
// and is not touched. A library of fewer than three exiles what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b3110a5-fd38-4d09-9073-c8a6f3f17840",
		Name:         "Three Wishes",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Old792ThreeWishes(ctx.Game, item)
		},
	})
}
