package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Panopticon — Plane — Mirrodin (Commander Masters, #4870):
//
//	"When you planeswalk to Panopticon, draw a card.
//	 At the beginning of your draw step, draw an additional card.
//	 Whenever chaos ensues, draw a card."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The planeswalk trigger (CR 901.11) and the chaos ability (CR 311.7) each
// draw one card; the draw-step trigger is the planar controller's, who is
// the active player, so it draws on every player's turn in their own draw
// step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd8ce6f1-9753-40a7-a625-5ed63c15c29e",
		Name:         "Panopticon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("Panopticon — draw a card", Do(DrawCards{N: 1})),
			On(game.EventBeginDrawStep, ByYou, "Panopticon — draw an additional card", Do(DrawCards{N: 1})),
			WheneverChaosEnsues("Panopticon — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
