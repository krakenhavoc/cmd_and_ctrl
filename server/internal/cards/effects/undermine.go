package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undermine — Instant {U}{U}{B}:
//
//	"Counter target spell. Its controller loses 3 life."
//
// Dream Fracture's shape with a life loss in place of the draw. The
// victim is read BEFORE the counter, because once the spell leaves the
// stack there is nothing left to ask whose it was. The loss does not
// depend on the counter taking: a spell that can't be countered stays
// on the stack and its controller still loses 3 life, as printed. With
// the target gone in response the whole spell fizzles (CR 608.2b) and
// nobody loses anything.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6bcfe240-c165-49f2-a90a-be05c05498a1",
		Name:         "Undermine",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				victim := uuid.Nil
				if si := ctx.Game.StackItemForEffect(t.ID); si != nil {
					victim = si.Controller
				}
				if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if victim == uuid.Nil {
					return nil
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -3)
			}
			return nil
		},
	})
}
