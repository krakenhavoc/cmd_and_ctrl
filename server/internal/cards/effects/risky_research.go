package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Risky Research — Sorcery {2}{B} (Marvel's Spider-Man, #4460):
//
//	"Surveil 2, then draw two cards. You lose 2 life."
//
// The draw rides the surveil's continuation: the prompt is answered
// first, so the cards drawn come from the library the player arranged.
// The life loss follows the draw, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47d52cd7-d72f-455f-aad3-70f08d54691f",
		Name:         "Risky Research",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(riskyResearchThen)}.Apply(ctx)
		},
	})
}

// riskyResearchThen is "then draw two cards. You lose 2 life", keyed
// (#4253): an on-disk tail the surveil prompt carries.
var riskyResearchThen = game.RegisterTail("then/risky-research-draw-two-lose-two", func(g *game.Game, o game.TailOutcome) error {
	if err := g.DrawNForEffect(o.Tail.Player, 2); err != nil {
		return err
	}
	return g.ChangePlayerLifeForEffect(o.Tail.Source, o.Tail.Player, -2)
})
