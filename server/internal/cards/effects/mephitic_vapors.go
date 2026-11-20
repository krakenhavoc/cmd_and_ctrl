package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5731ecb5-6ddc-4452-990c-b69986367f25",
		Name:         "Mephitic Vapors",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		OnResolve: ResolveSteps(
			BoostUntilEOT{Match: Creature(), Power: -1, Toughness: -1, Label: "Mephitic Vapors — -1/-1"},
			Surveil{N: 2},
		),
	})
}
