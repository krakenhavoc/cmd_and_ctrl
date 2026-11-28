package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wave Elemental — Creature — Elemental {2}{U}{U}, 2/3 (Mirage, S82 OLD b808):
//
//	"{U}, {T}, Sacrifice this creature: Tap up to three target creatures
//	 without flying."
//
// The sacrifice is a cost (CR 118.12), so the Elemental is gone before the
// ability resolves; the up-to-three targets are chosen from creatures without
// flying and each one still legal as it resolves is tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8eed5f7-8724-45c7-b6fc-dcfcc400f09a",
		Name:         "Wave Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}, Sacrifice this creature: Tap up to three target creatures without flying.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("up to three target creatures without flying", effects.WithoutKeyword("flying")).WithCount(0, 3),
			Effect:  effects.TapLegalCardTargets,
		}},
	})
}
