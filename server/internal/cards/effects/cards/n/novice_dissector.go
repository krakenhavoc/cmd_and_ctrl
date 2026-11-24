package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Novice Dissector — Creature — Troll Warlock {3}{B}, 3/3 (Strixhaven: School
// of Mages, #5502):
//
//	"{1}, Sacrifice another creature: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery."
//
// The sacrifice is a cost (CR 118.3), paid when the ability is activated, so
// the sacrificed creature is not a legal target (it has left by the time the
// target is judged only if it was chosen first; targets are announced before
// costs are paid, CR 601.2c/602.2b, and a target that is the sacrificed
// creature is simply illegal on resolution). The counter goes on the target
// as in Bond Beetle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5d6a41e-84bd-4b10-9734-50d91d523d2f",
		Name:         "Novice Dissector",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, Sacrifice another creature: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       effects.PlusOneCounterOnChosenTarget,
		}},
	})
}
