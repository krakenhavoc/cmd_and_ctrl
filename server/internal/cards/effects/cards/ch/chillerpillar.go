package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chillerpillar — Snow Creature — Insect {3}{U}, 3/3 (Jumpstart 2022, #5948):
//
//	"{4}{S}{S}: Monstrosity 2. (If this creature isn't monstrous, put two
//	 +1/+1 counters on it and it becomes monstrous. {S} can be paid with one
//	 mana from a snow source.)
//	 As long as this creature is monstrous, it has flying."
//
// Domesticated Hydra's monstrosity row with a fixed N and a snow cost, and its
// monstrous keyword grant (CR 701.37).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5607bc18-f3b7-40b0-b97d-2e06affa99ac",
		Name:         "Chillerpillar",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{4}{S}{S}"), 2)},
		Static:       []game.StaticAbility{effects.MonstrousKeywords("flying")},
	})
}
