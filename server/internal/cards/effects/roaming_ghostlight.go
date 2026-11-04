package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roaming Ghostlight — Creature — Spirit {3}{U}{U}, 3/2 (Jumpstart
// Azorius deck request, #4894):
//
//	"Flying
//	 When this creature enters, return up to one target non-Spirit
//	 creature to its owner's hand."
//
// "Up to one" may be left empty, so the trigger always goes on the
// stack. "Non-Spirit" is the creature's effective subtypes, judged at
// announce and again at resolution (CR 608.2b): a creature that became
// a Spirit in response stays where it is. The Ghostlight is itself a
// Spirit, so it can never return itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0d5eb0b9-9b08-4d26-b141-a9d612a680a6",
		Name:            "Roaming Ghostlight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Roaming Ghostlight — return up to one target non-Spirit creature to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return b35BounceChosenPermanents(NewContext(g, item))
				}),
				TargetCreature("up to one target non-Spirit creature", Not(HasSubtype("Spirit"))).WithCount(0, 1)),
		},
	})
}
