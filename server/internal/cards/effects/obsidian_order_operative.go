package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obsidian Order Operative — Creature — Cardassian Officer Spy {3}{B},
// 2/5 (Star Trek, #4973):
//
//	"Whenever you face a dilemma, each opponent loses 1 life and you gain
//	 1 life. (You face a dilemma as you choose one or more modes for a
//	 spell or ability.)"
//
// "Face a dilemma" is the engine's EventFacedDilemma (batch 7): its
// controller chose one or more modes for a spell or ability (CR 700.2a,
// 700.2b). The drain is a life change with no damage, written with Do
// steps so the row declares its footprint.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7145e1b7-b991-4651-b21d-0620cb017090",
		Name:         "Obsidian Order Operative",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventFacedDilemma, YouFacedADilemma,
				"Obsidian Order Operative — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}
