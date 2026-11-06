package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drill Too Deep — Instant {1}{R} (Edge of Eternities, #4461):
//
//	"Choose one —
//	 • Put five charge counters on target Spacecraft or Planet you control.
//	 • Destroy target artifact."
//
// Each bullet carries its own target clause and body (ModeDoing), so the
// picker offers only what the chosen bullet can target. Spacecraft and
// Planet are artifact subtypes read through the layers. The charge counters
// are one placement of five, so a counter replacement applies once, and a
// station threshold they cross takes effect at once (CR 721.2a). A target
// that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0413e0e4-c8f8-44c3-a749-c799766bc8cd",
		Name:         "Drill Too Deep",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // the destroy bullet; the counters answer nothing
		Modes: ChooseOne(
			ModeDoing("Put five charge counters on target Spacecraft or Planet you control.",
				TargetPermanent("target Spacecraft or Planet you control",
					Or(Subtype("Spacecraft"), Subtype("Planet")), YouControl()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return AddCounter{Target: t.ID, Kind: game.CounterCharge, N: 5}.Apply(ctx)
				}),
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				titanOfIndustryDestroy), // destroy the bullet's target, if still legal
		),
	})
}
