package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellspur Brute — Creature — Minotaur Mercenary {4}{R}, 5/4 (Outlaws of Thunder Junction):
//
//	"Affinity for outlaws (This spell costs {1} less to cast for each
//	 Assassin, Mercenary, Pirate, Rogue, and/or Warlock you control.)
//	 Trample"
//
// Affinity for a batch (b40Outlaw reads effective subtypes, so a changeling
// counts): the shared self cost modifier reduces only the generic part.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "81e5c62c-5aed-41e2-8bd1-610022e880bd",
		Name:            "Hellspur Brute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for outlaws", b40Outlaw()),
		},
	})
}
