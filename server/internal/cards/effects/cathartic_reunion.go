package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cathartic Reunion — Sorcery {1}{R}:
//
//	"As an additional cost to cast this spell, discard two cards.
//	 Draw three cards."
//
// The two-for-three rate the discard-and-draw cycle scales to at
// {1}{R}. Same cost shape as Thrill of Possibility, just N=2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "0f3c3e5f-6af3-4af2-8703-4ccc8ed8f675",
		Name:           "Cathartic Reunion",
		Completeness:   CompletenessFull,
		AdditionalCost: DiscardCost(2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
		},
	})
}
