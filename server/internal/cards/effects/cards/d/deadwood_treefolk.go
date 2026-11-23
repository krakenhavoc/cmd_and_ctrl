package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadwood Treefolk — Creature — Treefolk {5}{G}, 3/6 (Dominaria
// Remastered):
//
//	"Vanishing 3
//	 When this creature enters or leaves the battlefield, return another
//	 target creature card from your graveyard to your hand."
//
// Vanishing is Deep Forest Hermit's (three time counters, upkeep removal,
// sacrifice on the last). The enters-or-leaves ability is ONE triggered
// ability with two conditions (CR 603.2), targeting as it goes on the
// stack. "Another" is object identity: when the Treefolk leaves it is in
// the graveyard and is not a candidate. A card that left the graveyard in
// response is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7efcb42-aa52-4d13-8c7c-b2db2dd51afd",
		Name:         "Deadwood Treefolk",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Replacements: []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(3)},
		Triggered: append(effects.Cmdr2B17VanishingTriggers(),
			effects.Targeting(
				effects.WhenThisEntersOrLeaves("Deadwood Treefolk — return another target creature card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnTargetFromGraveyardToHand(item, effects.NewContext(g, item))
					}),
				effects.Another(effects.TargetCardInGraveyard("another target creature card from your graveyard", effects.Creature(), effects.YouOwn())))),
	})
}
