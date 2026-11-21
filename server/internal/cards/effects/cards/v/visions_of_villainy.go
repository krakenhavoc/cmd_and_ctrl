package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3b46800b-711b-4cbb-b296-26115b4f6839",
		Name:         "Visions of Villainy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Villain.",
				effects.YouControlA(effects.HasSubtype("Villain"))),
		},
		OnResolve: effects.ResolveSteps(
			// The life loss waits for a prompt the draw asks (CR 608.2c, #4155).
			effects.DrawCards{N: 2},
			effects.GainLife{Amount: -2},
		),
	})
}
