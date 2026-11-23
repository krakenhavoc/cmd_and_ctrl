package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "d740dbd9-8e90-4121-8d53-c6ddf5178d58",
		Name:            "Chomping Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Chomping Changeling — destroy up to one target artifact or enchantment", func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
			}),
				effects.TargetPermanent("up to one target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 1)),
		},
	})
}
