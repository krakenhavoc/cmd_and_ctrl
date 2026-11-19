package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jadelight Ranger — Creature — Merfolk Scout Ranger {1}{G}{G}, 2/1
// (Foundations Jumpstart, BNS batch 32):
//
//	"When this creature enters, it explores, then it explores again."
//
// Two explores in a row (ExploresTimes, Jadelight Spelunker's seam,
// CR 701.44): the second reveals the card the first left on top, and each
// is a separate explore for "whenever a creature you control explores".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b5d174c-7d4d-4c31-a582-8169d3bc932a",
		Name:         "Jadelight Ranger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jadelight Ranger — it explores, then it explores again", func(g *game.Game, item *game.StackItem) error {
				return ExploresTimes{Explorer: item.SourceCardID, N: 2}.Apply(NewContext(g, item))
			}),
		},
	})
}
