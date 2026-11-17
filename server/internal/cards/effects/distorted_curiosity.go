package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Distorted Curiosity — Sorcery {2}{U} (Phyrexia: All Will Be One, #5479):
//
//	"Corrupted — This spell costs {2} less to cast if an opponent has three
//	 or more poison counters.
//	 Draw two cards."
//
// Corrupted is a self cost reduction read as the cast is priced (CR 601.2f);
// it only takes generic mana off, and the mana value stays 3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ebbf025-2996-46f2-ac23-992909251525",
		Name:         "Distorted Curiosity",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "Corrupted — This spell costs {2} less to cast if an opponent has three or more poison counters.", s82OneB08Corrupted),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
