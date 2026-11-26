package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meletis Charlatan — Creature — Human Wizard {2}{U}, 2/3 (Theros):
//
//	"{2}{U}, {T}: The controller of target instant or sorcery spell copies it.
//	 That player may choose new targets for the copy."
//
// The copy is made and controlled by the targeted spell's controller, not by
// the Charlatan's (CR 707.10), with the CR 707.10c prompt put to that player.
// A target gone by resolution is nothing to copy (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11dfe74c-bc8e-4d58-b526-41ea5b409135",
		Name:         "Meletis Charlatan",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: The controller of target instant or sorcery spell copies it. That player may choose new targets for the copy.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell"),
			Effect:  effects.CopyTargetedSpellByItsController,
		}},
	})
}
