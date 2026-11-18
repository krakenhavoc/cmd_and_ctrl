package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jukai Trainee — Creature — Human Samurai {1}{G}, 2/2 (Neon Dynasty):
//
//	"Whenever this creature blocks or becomes blocked, it gets +1/+1
//	 until end of turn."
//
// Dromosaur's trigger with +1/+1: "blocks" triggers once however many
// attackers it blocks (CR 509.3a), and "becomes blocked" once however many
// creatures block it (CR 509.3c).
//
// No simplification.
func init() {
	const label = "Jukai Trainee — +1/+1 until end of turn"
	Register(Spec{
		OracleID:     "cb9fa34d-1d5f-461c-acad-bc573d41ceb2",
		Name:         "Jukai Trainee",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, thisBlocksOrBecomesBlocked,
				label, thisGetsUntilEndOfTurn(1, 1, label)),
		},
	})
}
