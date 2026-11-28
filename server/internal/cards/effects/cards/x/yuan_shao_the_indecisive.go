package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuan Shao, the Indecisive — Legendary Creature — Human Soldier {4}{R}, 2/3
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"Each creature you control can't be blocked by more than one creature."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2fc5fffa-2eb1-4a3c-af47-3ff8c7df608d",
		Name:            "Yuan Shao, the Indecisive",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		BlockRules: []game.BlockRule{
			effects.MaxBlockers(effects.OnMatching(effects.And(effects.Creature(), effects.YouControl())), 1),
		},
	})
}
