package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deepchannel Mentor — Creature — Merfolk Rogue {5}{U}, 2/2 (Shadowmoor, #5591):
//
//	"Blue creatures you control can't be blocked."
//
// A block rule on every blue creature its controller controls, itself
// included, read from current colours as each block is declared (CR 509.1b):
// a creature recoloured blue gets it and one recoloured away loses it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53b121df-3040-4ad0-a25d-587a299b4f8b",
		Name:         "Deepchannel Mentor",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnMatching(effects.And(effects.Creature(), effects.YouControl(), effects.OfColor("U"))),
				func(*game.Game, uuid.UUID, game.Card) bool { return true }),
		},
	})
}
