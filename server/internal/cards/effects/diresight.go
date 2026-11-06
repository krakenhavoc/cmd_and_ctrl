package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diresight — Sorcery {2}{B} (Avatar: The Last Airbender Eternal, #5075):
//
//	"Surveil 2, then draw two cards. You lose 2 life."
//
// Risky Research's text, word for word, so it is Risky Research's
// resolution: the surveil prompt carries the same keyed tail
// (riskyResearchThen), so the cards drawn come from the library the
// player arranged and the life loss follows the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7b12df9-fddd-4cd4-8c09-d54a47867ab6",
		Name:         "Diresight",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(riskyResearchThen)}.Apply(ctx)
		},
	})
}
