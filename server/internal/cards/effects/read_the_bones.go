package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Read the Bones — Sorcery {2}{B} (EDHREC rank 663):
//
//	"Scry 2, then draw two cards. You lose 2 life."
//
// The draw rides Scry's Then, never the next statement — the ordering
// rule every scry-then card follows (Preordain): Scry only queues a
// prompt, and a draw written as a separate primitive would happen
// before the caster has decided what to keep on top, taking one of
// the still-undecided cards and leaving the prompt unanswerable. The
// life loss has no such ordering dependency and runs unconditionally
// once the spell resolves, matching the printed comma splice ("Scry
// 2, then draw two cards." is one sentence; "You lose 2 life." is a
// separate one with no "then").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bf4d8d9-a2b2-4dba-ac05-9d4470a89db2",
		Name:         "Read the Bones",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			c := ctx.Controller()
			if err := (Scry{Player: c, N: 2, Then: func(g *game.Game) error {
				return g.DrawNForEffect(c, 2)
			}}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), c, -2)
		},
	})
}
