package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Visions of Villainy — {2}{B} Instant (Marvel Super Heroes, #4455):
//
//	"This spell costs {1} less to cast if you control a Villain.
//	 You draw two cards and lose 2 life."
//
// Night's Whisper's body at instant speed behind Wizard's Retort's
// conditional self cost modifier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b46800b-711b-4cbb-b296-26115b4f6839",
		Name:         "Visions of Villainy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a Villain.",
				YouControlA(HasSubtype("Villain"))),
		},
		OnResolve: ResolveSteps(
			// The life loss waits for a prompt the draw asks (CR 608.2c, #4155).
			DrawCards{N: 2},
			GainLife{Amount: -2},
		),
	})
}
