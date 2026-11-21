package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of Secret Streams — Creature — Merfolk Warrior {3}{U}, 2/3 (Lost
// Caverns of Ixalan Commander, #5368):
//
//	"Creatures you control with +1/+1 counters on them can't be blocked."
//
// A block rule over the creatures its controller controls, judged live on
// the attacker (Thieves' Tools' shape): the counter is read as the block
// is declared, so a creature that loses its last counter can be blocked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "366a218e-84d3-4cf9-bbc8-f2f8ecce92a3",
		Name:         "Herald of Secret Streams",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnMatching(effects.And(effects.Creature(), effects.YouControl())), effects.Cmdr2B23CreatureYouControlWithPlusOne),
		},
	})
}
