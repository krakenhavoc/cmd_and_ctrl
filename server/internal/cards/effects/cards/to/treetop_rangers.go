package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treetop Rangers — Creature — Elf Ranger {2}{G}, 2/2 (Urza's Saga, S82 OLD
// batch 731, #5624):
//
//	"This creature can't be blocked except by creatures with flying."
//
// A block restriction (CR 509.1b) read as blockers are declared, through the
// layers: only a creature that has flying as it blocks may block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4606bc2-faaa-4dd8-9109-55ce3440cf5b",
		Name:         "Treetop Rangers",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.HasKeyword("flying"), "creatures with flying"),
		},
	})
}
