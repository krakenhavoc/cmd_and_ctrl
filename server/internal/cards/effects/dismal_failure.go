package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dismal Failure — Instant {2}{U}{U} (Planar Chaos, #5595):
//
//	"Counter target spell. Its controller discards a card."
//
// Dream Fracture's shape: the spell's controller is read before it is
// countered, and that player then discards a card of their choice (CR
// 701.9b). A target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1feeb881-ebe6-4afb-ba31-3d116d8ab7a1",
		Name:         "Dismal Failure",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
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
				return DiscardCards{Player: victim, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
