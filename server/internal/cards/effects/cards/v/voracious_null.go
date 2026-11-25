package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Voracious Null — Creature — Zombie {2}{B}, 2/2 (Battle for Zendikar):
//
//	"{1}{B}, Sacrifice another creature: Put two +1/+1 counters on this
//	 creature. Activate only as a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f92daf5-520d-472a-a835-e6680a92fdab",
		Name:         "Voracious Null",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}, Sacrifice another creature: Put two +1/+1 counters on this creature. Activate only as a sorcery.",
			Purpose:      game.Purpose{SourceCounters: 2},
			Cost:         effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
