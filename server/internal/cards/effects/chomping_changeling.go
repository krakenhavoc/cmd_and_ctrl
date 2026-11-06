package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chomping Changeling — Creature — Shapeshifter {2}{G}, 1/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 When this creature enters, destroy up to one target artifact or
//	 enchantment."
//
// Changeling is printed and honoured in every zone (Irregular Cohort).
// The entry trigger is Hatut Zeraze Strike Force's: "up to one" is a
// clause of zero or one pick, so the trigger goes on the stack with no
// target when there is none or its controller picks none, and a target
// gone by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d740dbd9-8e90-4121-8d53-c6ddf5178d58",
		Name:            "Chomping Changeling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Chomping Changeling — destroy up to one target artifact or enchantment", func(g *game.Game, item *game.StackItem) error {
				return destroyEachLegalTarget(item, NewContext(g, item))
			}),
				TargetPermanent("up to one target artifact or enchantment", Or(Artifact(), Enchantment())).WithCount(0, 1)),
		},
	})
}
