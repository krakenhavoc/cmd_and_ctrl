package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mire Shade — Creature — Shade {1}{B}, 1/1 (Mirage, S82 OLD b805):
//
//	"{B}, Sacrifice a Swamp: Put a +1/+1 counter on this creature. Activate
//	 only as a sorcery."
//
// Heartwood Giant's sacrifice-a-land-by-subtype cost joined to a mana cost,
// at sorcery speed (CR 602.5d), with the counter put on the Shade itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b77a7f4a-1b3a-440e-977c-1ba39e450581",
		Name:         "Mire Shade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{B}, Sacrifice a Swamp: Put a +1/+1 counter on this creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{B}"), effects.SacrificeN(1, "a Swamp", effects.Subtype("Swamp"))),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
