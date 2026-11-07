package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hamlet Glutton — Creature — Giant {5}{G}{G}, 6/6 (Wilds of Eldraine,
// #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 This spell costs {2} less to cast if it's bargained.
//	 Trample
//	 When this creature enters, you gain 3 life."
//
// Bargain is the shared optional cost (g2Bargain), and the reduction is
// Ice Out's CostsLessIfBargained: the bargain is announced before the
// total cost is locked in (CR 601.2b, 601.2f), so a bargained Glutton
// costs {3}{G}{G}. Trample is the engine's keyword; the life gain
// happens on every entry, bargained or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aa1b7cdd-0296-418c-ac3b-7ad3c51f1760",
		Name:            "Hamlet Glutton",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		OptionalCosts:   []game.AdditionalCost{g2Bargain()},
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfBargained(2, "This spell costs {2} less to cast if it's bargained."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hamlet Glutton — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
