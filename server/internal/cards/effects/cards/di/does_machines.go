package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "df54b79d-f31d-4582-86c0-1d34ec77be0f",
		Name:         "Does Machines",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{1}{U}")),
			effects.LevelUp(3, effects.ManaCost("{4}{U}")),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Does Machines — mill two cards, draw two cards, then discard two cards",
				effects.Do(effects.MillCards{N: 2}, effects.Instruction(func(ctx *effects.Context) error {
					return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 2, 2, "")
				}))),
			effects.Targeting(
				effects.BecomesLevel(2, "Does Machines — level 2: return up to two target artifact cards from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
					}),
				effects.TargetCardInGraveyard("up to two target artifact cards from your graveyard", effects.Artifact(), effects.YouOwn()).WithCount(0, 2)),
			effects.AtLevel(3, effects.Targeting(
				effects.AtBeginningOfYourCombat("Does Machines — put three +1/+1 counters on target artifact you control",
					effects.TmtB8RobotCountersOnFirstTarget("Does Machines — becomes a 0/0 Robot creature")),
				effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl()))),
		},
	})
}
