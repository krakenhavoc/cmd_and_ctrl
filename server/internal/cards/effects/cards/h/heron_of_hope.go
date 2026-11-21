package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heron of Hope — Creature — Bird {3}{W}, 2/3 (Innistrad: Crimson Vow,
// #5495):
//
//	"Flying
//	 If you would gain life, you gain that much life plus 1 instead.
//	 {1}{W}: This creature gains lifelink until end of turn."
//
// The replacement is Angel of Vitality's (Leyline of Hope's) life-gain
// replacement (CR 614.1a), which reaches every gain of yours, lifelink
// included. The activation has no {T}, so it works with summoning sickness
// and as often as you can pay, and gives this creature lifelink until end
// of turn (Aethersphere Harvester's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f297648b-0dda-4c58-a3eb-fdabdfb42177",
		Name:            "Heron of Hope",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.LifeGainBecomes{
				Amount: func(n int) int { return n + 1 },
				Scope:  effects.LifeGainByController,
				Label:  "Heron of Hope: gain that much life plus 1",
			}.Build(),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: This creature gains lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("lifelink", "Heron of Hope — lifelink until end of turn"),
		}},
	})
}
