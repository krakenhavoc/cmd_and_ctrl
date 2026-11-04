package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orysa, Tide Choreographer — Legendary Creature — Merfolk Bard {4}{U}, 2/2
// (Secrets of Strixhaven, #4456):
//
//	"This spell costs {3} less to cast if creatures you control have total
//	 toughness 10 or greater.
//	 When Orysa enters, draw two cards."
//
// The discount is a conditional self cost modifier (Count Nefaria's shape)
// read as the price is worked out (CR 601.2f): the sum of the current
// toughness of the creatures the caster controls (socB30ToughnessAtLeast).
// Orysa herself is on the stack, so she adds nothing to her own total.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf56a5b4-48fa-4d53-a4b9-14afe079c1d3",
		Name:         "Orysa, Tide Choreographer",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if creatures you control have total toughness 10 or greater.",
				socB30ToughnessAtLeast(10)),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Orysa — draw two cards", Do(DrawCards{N: 2})),
		},
	})
}
