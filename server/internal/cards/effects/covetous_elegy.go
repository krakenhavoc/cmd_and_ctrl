package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Covetous Elegy — Sorcery {4}{W}{B} (Ravnica: Clue Edition):
//
//	"Each player chooses up to two creatures they control, then sacrifices
//	 the rest. Then you create a tapped Treasure token for each creature your
//	 opponents control."
//
// The keep-set shape of Archfiend of Depravity, asked of every seat in turn
// order from the active player (CR 101.4) and settled together. The
// Treasures are counted once the sacrifices are done (CR 608.2c), so only
// the opponents' survivors count. See s82_bns_b22_helpers.go.
func init() {
	Register(Spec{
		OracleID:     "dd69ef80-d30c-4b15-aec8-4154dc20cb25",
		Name:         "Covetous Elegy",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve:    covetousElegyResolve,
	})
}
