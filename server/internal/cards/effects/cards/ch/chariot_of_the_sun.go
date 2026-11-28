package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chariot of the Sun — Artifact {3} (Mirage, S82 OLD b808):
//
//	"{2}, {T}: Until end of turn, target creature you control gains flying and
//	 has base toughness 1."
//
// One effect, one timestamp (CR 613.7): a layer 6 grant of flying and a layer
// 7b base-toughness setting (CR 613.4b), so counters and later pumps still
// apply on top. Locked to the target as the ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0af249d-9255-42f5-bafc-ac61ee281930",
		Name:         "Chariot of the Sun",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Until end of turn, target creature you control gains flying and has base toughness 1.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.OldB808TargetGainsFlyingBaseToughnessOne,
		}},
	})
}
