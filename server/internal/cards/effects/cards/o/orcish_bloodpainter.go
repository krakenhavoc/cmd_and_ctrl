package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Bloodpainter — Creature — Orc Shaman {2}{R}, 2/1 (Coldsnap):
//
//	"{T}, Sacrifice a creature: This creature deals 1 damage to any target."
//
// The creature sacrificed may be the Bloodpainter itself; its damage is then
// dealt by the source as it last existed (CR 608.2h). The {T} needs the
// Bloodpainter free of summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "259bba2f-8c68-4267-8f51-0b2144778653",
		Name:         "Orcish Bloodpainter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a creature: This creature deals 1 damage to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DamageToFirstTarget(1),
		}},
	})
}
