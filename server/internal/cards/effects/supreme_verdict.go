package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Supreme Verdict — Sorcery {1}{W}{W}{U}:
//
//	"This spell can't be countered.
//	 Destroy all creatures."
//
// A Wrath of God that costs one more and a second colour, bought
// with the only clause that matters at a table holding up blue mana.
// Both halves are real here: the sweep is wrathDestroyAllCreatures,
// and CantBeCountered is honoured at the engine's counter choke
// point, where a Counterspell aimed at this resolves and does
// nothing rather than fizzling (CR 701.6a). See
// server/internal/game/cant_be_countered.go.
//
// No simplification. Reviewed for #2741: it prints no "can't be
// regenerated", so the sweep is wrathDestroyAllCreatures and not the
// NoRegen form, and s23_boardwipes_test.go pins both halves.
func init() {
	Register(Spec{
		OracleID:        "0230de18-8d15-4cfa-9d42-7ccddd9f9570",
		Name:            "Supreme Verdict",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		CantBeCountered: true,
		OnResolve:       wrathDestroyAllCreatures,
	})
}
