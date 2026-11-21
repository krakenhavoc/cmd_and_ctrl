package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infest — Sorcery {1}{B}{B} (Ravnica: Clue Edition):
//
//	"All creatures get -2/-2 until end of turn."
//
// The set of creatures is pinned as the spell resolves (CR 611.2c). A
// declared sweep (ADR 0126).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6850616-7db5-4141-9ab0-ae8d1f08114f",
		Name:         "Infest",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Infest — -2/-2 until end of turn"}.Apply(ctx)
			}),
		),
	})
}
