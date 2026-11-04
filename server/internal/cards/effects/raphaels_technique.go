package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raphael's Technique — Instant {4}{R}{R} (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Sneak {2}{R} (You may cast this spell for {2}{R} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step.)
//	 Each player may discard their hand and draw seven cards."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The wheel is
// the shared "each player may" wheel: every player chooses in APNAP order,
// then the players who said yes discard their hands, then each of them
// draws seven (CR 101.4, each_player_may_wheel.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f58a5f8b-254c-4206-bdbe-55a23798ed2d",
		Name:             "Raphael's Technique",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return raphaelsTechniqueMayWheel.Start(ctx)
		},
	})
}

var raphaelsTechniqueMayWheel = newMayWheel("confirm/raphaels-technique-may-wheel",
	"Raphael's Technique — discard your hand and draw seven cards?")
