package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duskmantle Operative — Creature — Human Rogue {1}{B}, 2/2 (War of the Spark, #5520):
//
//	"This creature can't be blocked by creatures with power 4 or greater."
//
// Beloved Princess's block rule with a power floor of 4.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7734ee04-1d99-4807-be68-b889e75a7551",
		Name:         "Duskmantle Operative",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerGE(4), "creatures with power 4 or greater"),
		},
	})
}
