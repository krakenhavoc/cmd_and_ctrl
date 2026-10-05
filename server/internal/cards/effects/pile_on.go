package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pile On — Instant {3}{B}:
//
//	"Convoke
//	 Destroy target creature or planeswalker. Surveil 2."
//
// Convoke (CR 702.51) rides Spec.TapCost, so the creatures tapped to
// help pay are real taps and a summoning-sick creature may be tapped
// for it.
//
// The surveil is the spell's own instruction, not a consequence of the
// destruction: it happens even if the destroy does nothing (an
// indestructible target). It only QUEUES the prompt, which is why it
// follows the destroy and not the other way round. If the single
// target is gone at resolution the whole spell fizzles (CR 608.2b) and
// there is no surveil, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "361b0d7f-1e43-45c5-92c2-92baacaf326f",
		Name:         "Pile On",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyFirstLegalCardTarget(ctx.Game, item); err != nil {
				return err
			}
			return Surveil{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
