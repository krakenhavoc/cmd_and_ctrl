package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metalhead — {4}{U} Legendary Artifact Creature — Robot Turtle 4/4
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"When Metalhead enters, return up to one other target artifact or
//	 creature to its owner's hand.
//	 {R}, Sacrifice another artifact: Put a +1/+1 counter on Metalhead.
//	 He gains menace and haste until end of turn."
//
// The enter trigger is Matterbending Mage's "up to one other" bounce over
// an artifact or a creature, either player's. The activation is Edgar,
// Ancient Bloodlord's shape: the sacrifice is a cost and never Metalhead
// himself; the counter and the two keywords go on Metalhead only while
// he is still the permanent the ability came from (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7df45a2a-75bc-4343-8dad-3a642efd13ee",
		Name:         "Metalhead",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Metalhead — return up to one other target artifact or creature to its owner's hand", bounceChosenTarget),
				Another(TargetPermanent("up to one other target artifact or creature", Or(Artifact(), Creature())).WithCount(0, 1)),
			),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice another artifact: Put a +1/+1 counter on Metalhead. He gains menace and haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet | game.AnswerCombatGrant},
			Cost:    Plus(ManaCost("{R}"), SacrificeAnotherN(1, "another artifact", Artifact())),
			Effect:  plusOneCounterOnThisThenItGains("Metalhead — menace and haste until end of turn", "menace", "haste"),
		}},
	})
}
