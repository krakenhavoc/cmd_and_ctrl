package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twist Allegiance — Sorcery {6}{R} (Betrayers of Kamigawa, #5602):
//
//	"You and target opponent each gain control of all creatures the other
//	 controls until end of turn. Untap those creatures. Those creatures gain
//	 haste until end of turn."
//
// Reins of Power's swap (SwapEachPlayersCreatures): both sets are read
// before anything moves, then each is handed to the other player until end
// of turn, untapped, with haste.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b33f3f0-80d3-4ea2-b73b-74510c9f6071",
		Name:         "Twist Allegiance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SwapEachPlayersCreatures(ctx, "Twist Allegiance")
		},
	})
}
