package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grim Tutor — Sorcery {1}{B}{B} (EDHREC rank 700):
//
//	"Search your library for a card, put that card into your hand,
//	 then shuffle. You lose 3 life."
//
// Demonic Tutor with a life tax. The search is the S22 chooser; the
// life is lost in the search's continuation, after the card is in
// hand — it is not a cost, so the spell is castable at 3 life or
// less, and the loss happens whether or not a card was taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e62f8d69-a559-4f13-a5c9-5fb750b4af2c",
		Name:         "Grim Tutor",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			source, controller := item.SourceCardID, item.Controller
			return effects.SearchLibrary{
				Player:   controller,
				Dest:     game.ZoneHand,
				Limit:    1,
				Shuffle:  true,
				Reason:   "Grim Tutor — search your library for a card",
				ThenTail: game.KeyedTail{Ref: effects.ThenChangeLife, Player: controller, Source: source, N: -3},
			}.Apply(ctx)
		},
	})
}
