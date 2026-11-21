package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shacklegeist — Creature — Spirit {1}{U}, 2/2:
//
//	"Flying
//	 This creature can block only creatures with flying.
//	 Tap two untapped Spirits you control: Tap target creature you don't
//	 control."
//
// Air Bladder's block rule on itself, Catapult Squad's tap-two cost over
// Spirits.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a1e4880-9abb-4018-84bd-91b5ec702c51",
		Name:            "Shacklegeist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap two untapped Spirits you control: Tap target creature you don't control.",
			Cost:    effects.TapTwoUntapped("two untapped Spirits you control", "Spirit"),
			Targets: effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
			Effect:  effects.B36TapChosenCreature,
		}},
	})
}
