package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigean Graftmage — Creature — Vedalken Wizard Mutant {2}{U}, 0/0 (Dissension, #5598):
//
//	"Graft 2 (This creature enters with two +1/+1 counters on it. Whenever
//	 another creature enters, you may move a +1/+1 counter from this
//	 creature onto it.)
//	 {1}{U}: Untap target creature with a +1/+1 counter on it."
//
// Graft is disGraft: Cytoplast Manipulator's entry replacement plus
// cmdr2B23Graft (CR 702.58a).
// The untap is untapTheTarget over disWithPlusOneCounter, judged again
// as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	replacements, triggered := effects.DisGraft("Vigean Graftmage", 2)
	effects.Register(effects.Spec{
		OracleID:     "4c209165-a159-4cd7-8e99-1e533d3a22a6",
		Name:         "Vigean Graftmage",
		Completeness: effects.CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Untap target creature with a +1/+1 counter on it.",
			Cost:    effects.ManaCost("{1}{U}"),
			Targets: effects.DisWithPlusOneCounter(),
			Effect:  effects.UntapTheTarget,
		}},
	})
}
