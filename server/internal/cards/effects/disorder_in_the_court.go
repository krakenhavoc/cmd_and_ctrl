package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disorder in the Court — Instant {X}{W}{U} (Murders at Karlov Manor
// Commander, #5368):
//
//	"Exile X target creatures, then investigate X times. Return the
//	 exiled cards to the battlefield tapped under their owners' control
//	 at the beginning of the next end step."
//
// Waterbender's Restoration's shape: the X targets are counted by the
// announced X, the exile is one event, and the return is a delayed
// trigger (CR 603.7) over the cards that actually reached exile, so it
// re-triggers their enters abilities. The return is tapped
// (Morningtide's Light's returnExiledToOwnersTappedBody), for every card. Tokens
// exiled do not come back (CR 111.8). The Clues are made after the
// exile, as printed, X of them whatever came back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e7586dd-d42e-4ac6-adb8-d46b6cdef923",
		Name:         "Disorder in the Court",
		Completeness: CompletenessFull,
		Targets:      targetsCountedByX(TargetCreature("X target creatures")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := exileTargetsThenScheduleReturnBody(ctx, "Disorder in the Court — return the exiled cards tapped", returnExiledToOwnersTappedBody); err != nil {
				return err
			}
			return Investigate{N: ctx.X()}.Apply(ctx)
		},
	})
}
