package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skittering Crustacean — Creature — Crab {2}{U}, 2/3 (Conspiracy: Take
// the Crown, S82 older sets batch 271, #5540):
//
//	"{6}{U}: Monstrosity 4. (If this creature isn't monstrous, put four
//	 +1/+1 counters on it and it becomes monstrous.)
//	 As long as this creature is monstrous, it has hexproof."
//
// Fleecemane Lion's pair: the shared Monstrosity activation and the layer 6
// self-grant behind the monstrous gate, so before the designation the Crab
// is an ordinary 2/3 anyone can target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a6d8d37-fcf9-4974-89f5-3500e9da3518",
		Name:         "Skittering Crustacean",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{6}{U}"), 4)},
		Static:       []game.StaticAbility{effects.MonstrousKeywords("hexproof")},
	})
}
