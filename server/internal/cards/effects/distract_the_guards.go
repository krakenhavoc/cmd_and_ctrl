package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "5b3687e5-546f-4a70-aa5c-b7db3300a3f4",
		Name:             "Distract the Guards",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 3},
		AlternativeCosts: []game.AlternativeCost{Freerunning("{1}{W}", "Assassin")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("1/1 white Human Rogue"), N: 3}.Apply(ctx)
		},
	})
}
