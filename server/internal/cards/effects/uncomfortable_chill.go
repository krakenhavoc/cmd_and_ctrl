package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Uncomfortable Chill — Instant {2}{U} (Foundations Jumpstart, #5929):
//
//	"Creatures your opponents control get -2/-0 until end of turn.
//	 Draw a card."
func init() {
	Register(Spec{
		OracleID:     "d6596336-8177-4475-95b9-a21858563276",
		Name:         "Uncomfortable Chill",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: ResolveSteps(
			BoostUntilEOT{
				Match: And(Creature(), OpponentControls()), Power: -2,
				Label: "Uncomfortable Chill — creatures your opponents control get -2/-0",
			},
			DrawCards{N: 1},
		),
	})
}
