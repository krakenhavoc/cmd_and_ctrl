package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abstruse Archaic — Creature — Avatar {4}, 3/4 (Commander Masters):
//
//	"Vigilance
//	 {1}, {T}: Copy target activated or triggered ability you control from
//	 a colorless source. You may choose new targets for the copy. (Mana
//	 abilities can't be targeted.)"
//
// Lithoform Engine's copy (copyTargetedAbility: the copy is its controller's,
// with new targets on offer, CR 707.10c) over an ability of the caster's
// from a colorless source. The source's colour is read as the ability is
// chosen, from the permanent as it is now or as it last was (a sacrificed
// colorless artifact's ability still qualifies), by
// gapB98AbilityFromColorlessSource. Mana abilities are not on the stack and
// so are never offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0727dc95-03df-4a2c-aed5-73fb840f3db7",
		Name:            "Abstruse Archaic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Copy target activated or triggered ability you control from a colorless source. You may choose new targets for the copy.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.AbilityOnStack("target activated or triggered ability you control from a colorless source", effects.GapB98AbilityFromColorlessSource()),
			Effect:  effects.CopyTargetedAbility,
		}},
	})
}
