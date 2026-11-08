package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Vitality — Creature — Angel {2}{W}, 2/2 (Foundations, #4465):
//
//	"Flying
//	 If you would gain life, you gain that much life plus 1 instead.
//	 This creature gets +2/+2 as long as you have 25 or more life."
//
// Leyline of Hope's life-gain replacement (CR 614.1a), which reaches
// every gain of yours, lifelink included, and a self pump gated on your
// life total (LifeGatedPump, read live). At a Commander table's 40
// starting life the Angel starts at 4/4.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "635fe908-6f58-4c9e-ac55-0775a7c6f278",
		Name:            "Angel of Vitality",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			LifeGainBecomes{
				Amount: func(n int) int { return n + 1 },
				Scope:  LifeGainByController,
				Label:  "Angel of Vitality: gain that much life plus 1",
			}.Build(),
		},
		Static: []game.StaticAbility{
			LifeGatedPump(SelfWhileYourLifeAtLeast(25), 2, 2),
		},
	})
}
