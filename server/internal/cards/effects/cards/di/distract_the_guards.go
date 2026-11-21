package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Distract the Guards — Sorcery {1}{W}{W}:
//
//	"Freerunning {1}{W} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)
//	 Create three 1/1 white Human Rogue creature tokens."
//
// Freerunning is Eagle Vision's alternative cost; the tokens are the printed
// Human Rogue of Assassin's Creed.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5b3687e5-546f-4a70-aa5c-b7db3300a3f4",
		Name:             "Distract the Guards",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tokens: 3},
		AlternativeCosts: []game.AlternativeCost{effects.Freerunning("{1}{W}", "Assassin")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Human Rogue"), N: 3}.Apply(ctx)
		},
	})
}
