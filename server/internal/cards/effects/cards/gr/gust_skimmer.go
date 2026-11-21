package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gust-Skimmer — Artifact Creature — Insect {2}, 2/1 (Mirrodin Besieged,
// #5576):
//
//	"{U}: This creature gains flying until end of turn."
//
// Vector Asp's self grant (somB1SelfGainsUntilEOT) with flying: it ends in
// the cleanup step (CR 514.2), and activated after blockers are declared it
// doesn't undo a block already made (CR 509.1h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f49dfa6-d0d4-4b58-8417-ed11c062e220",
		Name:         "Gust-Skimmer",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SomB1SelfGainsUntilEOT("Gust-Skimmer", "{U}", "flying")},
	})
}
