package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thundertrap Trainer — Creature — Otter Wizard {1}{U}, 1/2 (Bloomburrow,
// #4467):
//
//	"Offspring {4} (You may pay an additional {4} as you cast this spell.
//	 If you do, when this creature enters, create a 1/1 token copy of
//	 it.)
//	 When this creature enters, look at the top four cards of your
//	 library. You may reveal a noncreature, nonland card from among them
//	 and put it into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// Offspring is the shared keyword (offspring.go), and the token copy
// brings its own entry trigger, so a paid offspring looks twice. The
// look is Dragonologist's LookAtTopThenMayTakeToHand over a noncreature,
// nonland card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "d0be5969-0442-4557-a5fc-bee56e1835c2",
		Name:          "Thundertrap Trainer",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{4}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Thundertrap Trainer"),
			WhenThisEnters("Thundertrap Trainer — look at the top four cards; you may reveal a noncreature, nonland card and put it into your hand",
				LookAtTopThenMayTakeToHand(4, And(Noncreature(), Nonland()), 1,
					"Thundertrap Trainer — reveal a noncreature, nonland card and put it into your hand")),
		},
	})
}
