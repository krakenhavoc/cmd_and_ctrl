package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Erdwal Ripper — Creature — Vampire {1}{R}{R}, 2/1 (Dark Ascension, #5571):
//
//	"Haste
//	 Whenever this creature deals combat damage to a player, put a
//	 +1/+1 counter on it."
//
// Falkenrath Marauders' trigger with one counter. A Ripper that has left
// the battlefield gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6a207599-188f-49fc-8932-1974fc716f0b",
		Name:            "Erdwal Ripper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Erdwal Ripper — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
