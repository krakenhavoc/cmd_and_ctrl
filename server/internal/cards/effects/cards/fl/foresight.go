package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foresight — Sorcery {1}{U} (Alliances, #5634):
//
//	"Search your library for three cards, exile them, then shuffle.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The search is for exactly three cards of any kind (all of them with fewer),
// exiled face up (the cards are public in exile), then the library is shuffled;
// the delayed draw at the next upkeep at the table is scheduled as the spell
// resolves (CR 603.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9640b04-9751-46d5-b7bf-5b8a8ecf145f",
		Name:         "Foresight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DrawACardAtTheNextUpkeep("Foresight — draw a card").Apply(ctx); err != nil {
				return err
			}
			return effects.SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneExile,
				Limit:   3,
				Shuffle: true,
				Reason:  "Foresight — search your library for three cards to exile",
				Source:  item.SourceCardID,
			}.Apply(ctx)
		},
	})
}
