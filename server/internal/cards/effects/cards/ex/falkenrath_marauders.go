package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Falkenrath Marauders — Creature — Vampire Warrior {3}{R}{R}, 2/2:
//
//	"Flying
//	 Haste
//	 Whenever this creature deals combat damage to a player, put two +1/+1
//	 counters on it."
//
// Arcbound Slith's trigger with two counters; the counters go on only if the
// Marauders is still on the battlefield as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f819b134-24cd-4863-a46a-0a068e34ca9e",
		Name:            "Falkenrath Marauders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Falkenrath Marauders — put two +1/+1 counters on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
