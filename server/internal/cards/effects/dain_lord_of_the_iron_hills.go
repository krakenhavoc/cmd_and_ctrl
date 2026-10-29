package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dáin, Lord of the Iron Hills — Legendary Creature — Dwarf Noble {1}{W},
// 2/2:
//
//	"Vigilance
//	 Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 As long as you have an enduring story, creatures can't attack you
//	 unless their controller pays {1} for each of those creatures."
//
// Storied is the engine's (CR 702.195a). The tax is Ghostly Prison's
// CR 508.1h attack tax at {1} a creature, charged only while Dáin's
// controller has the story, and only for attacks on that player (not
// their planeswalkers): the engine's pricer gives a tax to its own
// controller alone.
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:        "5b79b68a-db71-4638-824e-ab89b1716c86",
		Name:            "Dáin, Lord of the Iron Hills",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", game.KeywordStoried},
		AttackTaxes: []game.AttackTax{
			AttackTaxCounting(func(q game.AttackTaxQuery) int {
				if !YouHaveAnEnduringStory(q.Game, q.Defender) {
					return 0
				}
				return 1
			}, "As long as you have an enduring story, creatures can't attack you unless their controller pays {1} for each of those creatures."),
		},
	})
}
