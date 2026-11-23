package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnat Alley Creeper — Creature — Human Rogue {2}{R}, 3/1 (Dissension, #5598):
//
//	"This creature can't be blocked by creatures with flying."
//
// Skyscythe Engulfer's block rule on itself: a flying blocker, read
// through the layered view as the block is declared (CR 509.1b), is
// refused.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17f930a9-64ee-46f9-8a7c-8877e532c94f",
		Name:         "Gnat Alley Creeper",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.HasKeyword("flying"), "creatures with flying"),
		},
	})
}
