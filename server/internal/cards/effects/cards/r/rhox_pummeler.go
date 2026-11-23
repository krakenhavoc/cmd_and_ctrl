package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhox Pummeler — Creature — Rhino Soldier {5}{G}, 6/3 (Streets of New
// Capenna):
//
//	"This creature enters with a shield counter on it. (If it would be
//	 dealt damage or destroyed, remove a shield counter from it
//	 instead.)
//	 This creature has trample as long as it has a shield counter on
//	 it."
//
// Chaos Imps' layer-6 self grant read off the live counters: the
// trample goes the moment the counter is spent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "898c8415-8769-48c4-ad16-a4ba8c3d5d7d",
		Name:         "Rhox Pummeler",
		Completeness: effects.CompletenessFull,
		Replacements: effects.Snc6Entry("Rhox Pummeler"),
		Static:       []game.StaticAbility{effects.ThisHasWhileItHasCounter(game.CounterShield, "trample")},
	})
}
