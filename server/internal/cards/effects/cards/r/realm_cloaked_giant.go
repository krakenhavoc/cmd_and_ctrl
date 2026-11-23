package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Realm-Cloaked Giant // Cast Off — adventure creature, oracle
// a1d17244-9c92-4094-9843-d7ee31c85ea4:
//
//	Realm-Cloaked Giant {5}{W}{W} Creature — Giant 7/7: "Vigilance"
//	Cast Off {3}{W}{W} Sorcery — Adventure: "Destroy all non-Giant
//	 creatures."
//
// Two specs, Cast Off under "<oracle_id>#1" as Bonecrusher Giant's Stomp is;
// the adventure's exile and later cast are the engine's. Cast Off is a
// Wrath of God without the regeneration rider, sparing every Giant (read
// through the layered view, so a changeling is a Giant).
//
// No simplification.
func init() {
	const oracle = "a1d17244-9c92-4094-9843-d7ee31c85ea4"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Realm-Cloaked Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Cast Off",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Giant")))}.Apply(ctx)
		},
	})
}
