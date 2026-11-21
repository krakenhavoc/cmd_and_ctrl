package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disruptive Student — Creature — Human Wizard {2}{U}, 1/1 (Urza's Saga):
//
//	"{T}: Counter target spell unless its controller pays {1}."
//
// Spiketail Hatchling's tax with a tap cost in place of the sacrifice.
// The spell's controller is asked as the ability resolves; a spell that
// has left the stack by then makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74315869-46e3-43a2-abc0-d86d0a70d343",
		Name:         "Disruptive Student",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Counter target spell unless its controller pays {1}.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.CounterTheTargetSpellUnlessPaid("Disruptive Student", "{1}"),
		}},
	})
}
