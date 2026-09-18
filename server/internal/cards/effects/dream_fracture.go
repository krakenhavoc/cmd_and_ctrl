package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dream Fracture — Instant {1}{U}{U} (EDHREC rank 4390):
//
//	"Counter target spell. Its controller draws a card.
//	 Draw a card."
//
// The counterspell that refuses to two-for-one anybody: the victim
// replaces the spell, you replace the Fracture, and what you actually
// bought is the tempo. In a four-player game that is a real bargain,
// which is why a card printed as a limited common is in the roadmap's
// top five thousand at all. Batch 42 (#449), "no new machinery".
//
// The victim's controller is read BEFORE the counter, because once
// the item leaves the stack there is nothing left to ask. Reading it
// afterwards would silently drop the second sentence.
//
// Order is as printed: counter, they draw, you draw. It matters for
// an empty library — the player who draws first is the player who
// dies to CR 704.5b first, and the printed order says that is them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5843d13-855f-41dd-ac13-3c7f7e18bb39",
		Name:         "Dream Fracture",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim := uuid.Nil
			if len(item.Targets) > 0 {
				if si := ctx.Game.StackItemForEffect(item.Targets[0].ID); si != nil {
					victim = si.Controller
				}
				if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if victim != uuid.Nil {
				if err := (DrawCards{Player: victim, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
