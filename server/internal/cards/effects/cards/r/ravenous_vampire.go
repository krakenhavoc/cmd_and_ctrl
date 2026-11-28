package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Vampire — Creature — Vampire {3}{B}{B}, 3/3 (Mirage):
//
//	"Flying
//	 At the beginning of your upkeep, you may sacrifice a nonartifact creature.
//	 If you do, put a +1/+1 counter on this creature. If you don't, tap this
//	 creature."
//
// Crovax the Cursed's upkeep offer with a nonartifact creature to sacrifice (the
// Vampire itself qualifies) and a tap in place of the counter's removal. A
// controller with no such creature is not asked (CR 118.3) and "doesn't". Each
// branch acts only while the Vampire is still the permanent the trigger came
// from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c029419-3601-4158-b10f-45a19de7faa3",
		Name:            "Ravenous Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Ravenous Vampire — you may sacrifice a nonartifact creature; if you do, put a +1/+1 counter on this creature, if you don't, tap it", effects.OldB804RavenousVampireOffer),
		},
	})
}
