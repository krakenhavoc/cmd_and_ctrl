package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treetop Snarespinner — Creature — Spider {3}{G}, 1/4 (Foundations,
// #4465):
//
//	"Reach
//	 Deathtouch
//	 {2}{G}: Put a +1/+1 counter on target creature you control.
//	 Activate only as a sorcery."
//
// Two printed keywords, and Hall of Oracles' sorcery-speed counter
// row limited to a creature you control. The counter goes on the
// target if it is still legal as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f57a220-9f64-4143-979a-d6aaf63d49b2",
		Name:            "Treetop Snarespinner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{G}: Put a +1/+1 counter on target creature you control. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{2}{G}"),
			Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
			SorcerySpeed: true,
			Effect:       effects.B36CounterOnChosenAnimal,
		}},
	})
}
