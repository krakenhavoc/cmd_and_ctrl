package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skirsdag Cultist — {2}{R}{R} Creature — Human Shaman 2/2 (Innistrad, S82 OLD
// b448, #5572):
//
//	"{R}, {T}, Sacrifice a creature: This creature deals 2 damage to any
//	 target."
//
// The sacrifice is a cost, so the Cultist may sacrifice itself (CR 602.2b);
// the damage is still dealt by it, from its last-known information.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7fe5473-fae5-490e-8600-b7e497b9989e",
		Name:         "Skirsdag Cultist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}, Sacrifice a creature: This creature deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.DamageToFirstTarget(2),
		}},
	})
}
