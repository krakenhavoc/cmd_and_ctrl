package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oracle of the Alpha — Creature — Bird Wizard {2}{U}, 2/3 (Mystery
// Booster Commander, #5948):
//
//	"Flying
//	 When Oracle of the Alpha enters the battlefield, conjure the Power
//	 Nine into your library, then shuffle.
//	 Whenever Oracle of the Alpha attacks, scry 1."
//
// The Power Nine are nine real cards (PowerNine; ADR 0099 amendment
// 2026-11-24), then one shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b77dcee1-f42e-4c9d-b6b7-6d5885826060",
		Name:            "Oracle of the Alpha",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Oracle of the Alpha — conjure the Power Nine into your library, then shuffle",
				effects.Do(effects.CheckConjure("Oracle of the Alpha", effects.Conjure{Names: effects.PowerNine, Zone: game.ZoneLibrary, Shuffle: true}))),
			effects.WheneverThisAttacks("Oracle of the Alpha — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
