package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mephitic Vapors — Sorcery {2}{B} (Guilds of Ravnica, #5525):
//
//	"All creatures get -1/-1 until end of turn.
//	 Surveil 2."
//
// Two steps: Glistening Deluge's -1/-1 to every creature until cleanup,
// then surveil 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5731ecb5-6ddc-4452-990c-b69986367f25",
		Name:         "Mephitic Vapors",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{Match: effects.Creature(), Power: -1, Toughness: -1, Label: "Mephitic Vapors — -1/-1"},
			effects.Surveil{N: 2},
		),
	})
}
