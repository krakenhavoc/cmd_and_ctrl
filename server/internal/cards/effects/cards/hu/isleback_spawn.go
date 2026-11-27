package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isleback Spawn — Creature — Kraken {5}{U}{U} 4/8:
//
//	"Shroud
//	 This creature gets +4/+8 as long as a library has twenty or fewer cards in
//	 it."
//
// A layer 7c static of the source alone (Thresher Lizard's shape) over ANY
// seated player's library, not only its controller's. Library size is not on
// the battlefield, so the static declares DependsOnLibrarySize and the layer
// pass is redone whenever a library's size changes (a draw, a mill, a search).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "05ad7ec4-f48d-4cbe-be58-f871dc19702b",
		Name:            "Isleback Spawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Static:          []game.StaticAbility{effects.SelfPumpWhileALibraryHasAtMost(20, 4, 8)},
	})
}
