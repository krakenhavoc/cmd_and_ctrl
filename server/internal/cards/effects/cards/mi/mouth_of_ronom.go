package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mouth of Ronom — Snow Land (Tales of Middle-earth Commander):
//
//	"{T}: Add {C}.
//	 {4}{S}, {T}, Sacrifice this land: It deals 4 damage to target creature.
//	 ({S} can be paid with one mana from a snow source.)"
//
// A colourless tap for mana and a snow-cost activation (Boreal Griffin's {S}),
// with the land sacrificed as a cost: the damage is dealt by it as it last
// existed on the battlefield (CR 608.2h), so its deathtouch or lifelink, if
// it somehow had any, would apply. The target is a creature, chosen as the
// ability goes on the stack and judged again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7c05d239-39fc-4d34-a853-e3d591f4a235",
		Name:          "Mouth of Ronom",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.TapForColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{S}, {T}, Sacrifice this land: It deals 4 damage to target creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
			Cost:    effects.Plus(effects.ManaCost("{4}{S}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.B33DamageChosenTargetFromSource(4),
		}},
	})
}
