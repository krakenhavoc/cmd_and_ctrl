package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inspired Tinkering — Sorcery {4}{R} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP #5948):
//
//	"Exile the top three cards of your library. Until the end of your next turn,
//	 you may play those cards.
//	 Create three Treasure tokens."
//
// Impulse's exile-and-play (the top three go to exile face up, each with a play
// permission until the end of the caster's next turn: lands and spells, ending
// with that turn), then three Treasures for the caster whatever the library
// held.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3088b09b-1dce-4fc9-99de-2ca9c752eb12",
		Name:         "Inspired Tinkering",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g, you := ctx.Game, item.Controller
			if _, err := g.ExileTopWithPermissionForEffect(you, you, 3, game.CastPermission{
				Duration: g.UntilEndOfYourNextTurnDuration(you),
			}); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TreasureToken(), N: 3}.Apply(ctx)
		},
	})
}
