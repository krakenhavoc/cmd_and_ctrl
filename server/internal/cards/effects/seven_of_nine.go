package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seven of Nine — Legendary Artifact Creature — Human Scientist
// {2}{U}{B}, 2/4 (Star Trek, #4973):
//
//	"Whenever you face a dilemma, draw a card. (You face a dilemma as you
//	 choose one or more modes for a spell or ability.)"
//
// "Face a dilemma" is the engine's EventFacedDilemma (batch 7): its
// controller chose one or more modes for a spell or ability (CR 700.2a,
// 700.2b). The draw is a Do step, so the row declares its footprint, and
// the card declares itself a draw engine (ADR 0126).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c037865-d38c-48da-aee9-308975e12638",
		Name:         "Seven of Nine",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventFacedDilemma, YouFacedADilemma,
				"Seven of Nine — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}
