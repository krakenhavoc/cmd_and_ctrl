package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Attrition — Enchantment {1}{B}{B} (Urza's Destiny, #5621):
//
//	"{B}, Sacrifice a creature: Destroy target nonblack creature."
//
// Stronghold Assassin's row on an enchantment, with {B} in place of
// {T}: it can be activated any number of times. The creature is
// sacrificed as part of the cost (CR 701.21a), so it is gone even if
// the target becomes illegal; a black target is refused at
// activation, and one gone by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a85ee0cf-39ce-49f2-874a-f6f54ca1775f",
		Name:         "Attrition",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice a creature: Destroy target nonblack creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeACreature()),
			Targets: effects.TargetCreature("target nonblack creature", effects.NonBlack()),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
