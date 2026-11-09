package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cerebral Download — Instant {4}{U} (Edge of Eternities, #4461):
//
//	"Surveil X, where X is the number of artifacts you control. Then draw
//	 three cards."
//
// X is counted as the spell resolves, through the layers. The draw is the
// surveil's keyed tail, so it waits for the player's answer (the
// warning on Surveil.ThenTail): a card put into the graveyard is not
// drawn, and one left on top is. X of zero still draws (CR 701.25c: no surveil
// event occurs, and the instruction goes on).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f27889b9-40d9-4cc7-b8af-8716c53c44f5",
		Name:         "Cerebral Download",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			x := countControlled(ctx.Game, player, func(c game.Card) bool { return c.IsArtifact() })
			return Surveil{Player: player, N: x, ThenTail: ctx.Tail(cerebralDownloadDraw)}.Apply(ctx)
		},
	})
}

// cerebralDownloadDraw is "then draw three cards". An on-disk key: never
// renamed, never reused.
var cerebralDownloadDraw = game.RegisterTail("then/cerebral-download-draw", func(g *game.Game, o game.TailOutcome) error {
	return g.DrawNForEffect(o.Tail.Player, 3)
})
