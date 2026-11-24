package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honor Troll — {2}{G} Creature — Troll Druid 2/3 (Strixhaven: School of
// Mages, #5502):
//
//	"Vigilance
//	 If you would gain life, you gain that much life plus 1 instead.
//	 This creature gets +2/+1 as long as you have 25 or more life."
//
// Angel of Vitality's life-gain replacement (CR 614.1a, Leyline of Hope's),
// which reaches every gain of yours, lifelink included, and a self pump gated
// on your life total (LifeGatedPump, read live): the +2/+1 is on at a
// Commander table's 40 starting life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "98c059ed-d749-4f66-bda1-6352976198f0",
		Name:            "Honor Troll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements: []game.ReplacementEffect{
			effects.LifeGainBecomes{
				Amount: func(n int) int { return n + 1 },
				Scope:  effects.LifeGainByController,
				Label:  "Honor Troll: gain that much life plus 1",
			}.Build(),
		},
		Static: []game.StaticAbility{
			effects.LifeGatedPump(effects.SelfWhileYourLifeAtLeast(25), 2, 1),
		},
	})
}
