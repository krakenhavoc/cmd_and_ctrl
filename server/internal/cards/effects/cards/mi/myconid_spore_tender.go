package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myconid Spore Tender — Creature — Fungus {3}{G}, 4/1 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Infesting Spores — When this creature enters, destroy up to one target
//	 artifact or enchantment."
//
// The entry trigger is Chomping Changeling's: "up to one" is a clause of zero
// or one pick, so the trigger goes on the stack with no target when there is
// none, and a target gone by resolution is skipped (CR 608.2b). "Infesting
// Spores" is an ability word with no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adf59918-a021-4b4e-b672-d82be960726e",
		Name:         "Myconid Spore Tender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Myconid Spore Tender — destroy up to one target artifact or enchantment", func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
			}),
				effects.TargetPermanent("up to one target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 1)),
		},
	})
}
