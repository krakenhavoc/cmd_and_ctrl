package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        oracle,
		Name:            "Realm-Cloaked Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Cast Off",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), Not(HasSubtype("Giant")))}.Apply(ctx)
		},
	})
}
