package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// To Arms! — Instant {1}{W} (Guildpact, #5599):
//
//	"Untap all creatures you control.
//	 Draw a card."
//
// Vitalize's untap (b34UntapAllCreaturesYouControl), then the draw
// (effectThenDrawACard). The untap asks nothing, so the two run in
// printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90c97fc4-597c-45cf-96c6-3ba9558ede0d",
		Name:         "To Arms!",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve:    effectThenDrawACard(b34UntapAllCreaturesYouControl),
	})
}
