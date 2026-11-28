package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zuberi, Golden Feather — Legendary Creature — Griffin {4}{W}, 3/3 (Mirage,
// #5633):
//
//	"Flying
//	 Other Griffin creatures get +1/+1."
//
// Every other Griffin on the battlefield, whoever controls it (a layer 7c effect,
// CR 613.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5a43b164-61a4-4dcb-9397-45356ad4260e",
		Name:            "Zuberi, Golden Feather",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Griffin"}, Others: true}, 1, 1)},
	})
}
