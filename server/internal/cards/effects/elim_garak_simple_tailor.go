package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elim Garak, Simple Tailor — Legendary Creature — Cardassian Rogue
// {2}{B}, 3/3 (Star Trek, #4973):
//
//	"Menace
//	 Whenever you face a dilemma, each opponent loses 2 life and you gain
//	 2 life. (You face a dilemma as you choose one or more modes for a
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
		OracleID:        "c83465bd-f95f-4c57-b9ce-925a85b45276",
		Name:            "Elim Garak, Simple Tailor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventFacedDilemma, YouFacedADilemma,
				"Elim Garak, Simple Tailor — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}
