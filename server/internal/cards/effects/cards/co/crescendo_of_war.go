package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crescendo of War — Enchantment {3}{W} (Commander 2011, S82 older sets batch
// 454, #5574):
//
//	"At the beginning of each upkeep, put a strife counter on this
//	 enchantment.
//	 Attacking creatures get +1/+0 for each strife counter on this
//	 enchantment.
//	 Blocking creatures you control get +1/+0 for each strife counter on this
//	 enchantment."
//
// Two layer-7c statics reading the strife counters live: every attacking
// creature (anyone's) and only your blocking creatures. The upkeep trigger
// fires on each player's upkeep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c6a8ad8-9b40-4136-9cfd-2b4cf9a9d900",
		Name:         "Crescendo of War",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB454CombatAnthem("Attacking creatures get +1/+0 for each strife counter on this enchantment.", "strife", true, false),
			effects.S82OldB454CombatAnthem("Blocking creatures you control get +1/+0 for each strife counter on this enchantment.", "strife", false, true),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Crescendo of War — put a strife counter on this enchantment",
				effects.Do(effects.CounterOnThis{Kind: "strife", N: 1})),
		},
	})
}
