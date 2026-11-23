package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Covenant of Blood — Sorcery {6}{B} (Magic 2015, #5556):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Covenant of Blood deals 4 damage to any target and you gain 4 life."
//
// Vampiric Feast's body (the damage, then the life gain) with the convoke
// TapCost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20ca6a15-ba88-4efd-b5bd-da5938e9a43e",
		Name:         "Covenant of Blood",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.S82B48DamageToTheFirstTarget(4)),
			effects.GainLife{Amount: 4},
		),
	})
}
