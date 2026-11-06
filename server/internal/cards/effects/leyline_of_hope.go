package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of Hope — Enchantment {2}{W}{W} (Duskmourn: House of Horror,
// #4466):
//
//	"If this card is in your opening hand, you may begin the game with it
//	 on the battlefield.
//	 If you would gain life, you gain that much life plus 1 instead.
//	 As long as you have at least 7 life more than your starting life
//	 total, creatures you control get +2/+2."
//
// The opening-hand clause is Spec.OpeningHand (ADR 0133, CR 103.6a), as
// on Leyline of Sanctity. The life is Pest Rescuer's replacement (CR
// 614.1a): a gain of at least 1 by you becomes that much plus 1. The
// anthem is Righteous Valkyrie's LifeGatedPump, read live against the
// table's starting life total (ADR 0075), so it is 47 or more at a
// Commander table.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7ee79c3-a273-45b4-b3a4-548d9b2b883c",
		Name:         "Leyline of Hope",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		Replacements: []game.ReplacementEffect{
			LifeGainBecomes{
				Amount: func(n int) int { return n + 1 },
				Scope:  LifeGainByController,
				Label:  "Leyline of Hope: gain that much life plus 1",
			}.Build(),
		},
		Static: []game.StaticAbility{
			LifeGatedPump(YourCreaturesWhileYourLifeAboveStarting(7), 2, 2),
		},
	})
}
