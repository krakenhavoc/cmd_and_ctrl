package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Does Machines — Enchantment — Class {1}{U} (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, mill two cards, draw two cards, then discard
//	 two cards.
//	 {1}{U}: Level 2
//	 When this Class becomes level 2, return up to two target artifact
//	 cards from your graveyard to your hand.
//	 {4}{U}: Level 3
//	 At the beginning of combat on your turn, put three +1/+1 counters on
//	 target artifact you control. If it isn't a creature, it becomes a 0/0
//	 Robot creature in addition to its other types."
//
// A Class (CR 716) written with the designation gates: LevelUp for the
// two level lines, BecomesLevel for level 2's trigger and AtLevel for
// level 3's (Artificer Class). The enters trigger mills first and then
// draws, and the discard prompt opens only once the draws are done
// (drawThenDiscard), so the drawn cards may be discarded. The milled
// cards are in the graveyard in time for level 2's targets. Level 3's
// body is Tezzeret, Cruel Captain's emblem (tmtB8RobotCountersOnFirstTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df54b79d-f31d-4582-86c0-1d34ec77be0f",
		Name:         "Does Machines",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{U}")),
			LevelUp(3, ManaCost("{4}{U}")),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Does Machines — mill two cards, draw two cards, then discard two cards",
				Do(MillCards{N: 2}, Instruction(func(ctx *Context) error {
					return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 2, 2, "")
				}))),
			Targeting(
				BecomesLevel(2, "Does Machines — level 2: return up to two target artifact cards from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return returnLegalGraveyardTargetsToHand(NewContext(g, item))
					}),
				TargetCardInGraveyard("up to two target artifact cards from your graveyard", Artifact(), YouOwn()).WithCount(0, 2)),
			AtLevel(3, Targeting(
				AtBeginningOfYourCombat("Does Machines — put three +1/+1 counters on target artifact you control",
					tmtB8RobotCountersOnFirstTarget("Does Machines — becomes a 0/0 Robot creature")),
				TargetPermanent("target artifact you control", Artifact(), YouControl()))),
		},
	})
}
