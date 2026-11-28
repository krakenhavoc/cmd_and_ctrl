package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mist Dragon — Creature — Dragon {4}{U}{U}, 4/4 (Mirage, S82 OLD b808):
//
//	"{0}: This creature gains flying. (This effect lasts indefinitely.)
//	 {0}: This creature loses flying. (This effect lasts indefinitely.)
//	 {3}{U}{U}: This creature phases out. (While it's phased out, it's treated
//	 as though it doesn't exist. It phases in before you untap during your
//	 next untap step.)"
//
// The two {0} abilities are layer 6 effects with no stated duration (CR
// 611.2a), pinned to this object so they end when it leaves (CR 400.7). Each
// takes the timestamp of its own resolution (CR 613.7), so whichever resolved
// last decides: gain then lose is no flying, lose then gain is flying. The
// phase-out is Blink Dog's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "448f5c55-7334-4fa4-9d90-0e57bae19444",
		Name:         "Mist Dragon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{0}: This creature gains flying.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{0}"),
				Effect:  effects.OldB808FlyingIndefinitely(true, "Mist Dragon — gains flying"),
			},
			{
				Label:   "{0}: This creature loses flying.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{0}"),
				Effect:  effects.OldB808FlyingIndefinitely(false, "Mist Dragon — loses flying"),
			},
			{
				Label:   "{3}{U}{U}: This creature phases out.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{3}{U}{U}"),
				Effect:  effects.ThisPhasesOut,
			},
		},
	})
}
