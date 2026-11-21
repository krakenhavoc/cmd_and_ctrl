package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blade Sliver — Creature — Sliver {2}{R}, 2/2 (Legions, #5609):
//
//	"All Sliver creatures get +1/+0."
//
// A tribal anthem with neither "other" nor "you control": every Sliver
// creature at the table gets +1/+0, the Blade Sliver itself and an
// opponent's Slivers included. Layer 7c, so it stacks with the other
// Sliver lords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "110e5695-a9df-4201-9230-99f6abe9978f",
		Name:         "Blade Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}}, 1, 0),
		},
	})
}
