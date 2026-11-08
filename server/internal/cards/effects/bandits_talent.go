package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bandit's Talent — Enchantment — Class {1}{B} (Bloomburrow, #4467):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, each opponent discards two cards unless they
//	 discard a nonland card.
//	 {B}: Level 2
//	 At the beginning of each opponent's upkeep, if that player has one
//	 or fewer cards in hand, they lose 2 life.
//	 {3}{B}: Level 3
//	 At the beginning of your draw step, draw an additional card for each
//	 opponent who has one or fewer cards in hand."
//
// ADR 0071 Class machinery (Hunter's Talent): LevelUp for each level, and
// the level-2 and level-3 triggers gated with AtLevel (CR 716.2a).
//
//   - Level 1 is Compulsive Research's "discards two cards unless they
//     discard a <kind> card" with a nonland card as the kind, asked of each
//     opponent in turn: two cards of any kind, or one nonland card. A hand
//     of one card discards it, whatever it is, and an empty hand discards
//     nothing (CR 701.9a).
//   - Level 2 is Iron Maiden's "each opponent's upkeep" with an intervening
//     if (CR 603.4): that player's hand is counted as their upkeep begins
//     and again as the ability resolves, and the loss is theirs.
//   - Level 3 is Lord Skitter's Blessing's draw-step trigger. The opponents
//     with one or fewer cards in hand are counted as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bcbe879a-4609-4493-bd7d-47c26d27e83c",
		Name:         "Bandit's Talent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bandit's Talent — each opponent discards two cards unless they discard a nonland card",
				func(g *game.Game, item *game.StackItem) error {
					for _, opp := range NewContext(g, item).Opponents() {
						if err := askDiscardTwoUnlessOne(g, opp, item.SourceCardID, 2,
							"Bandit's Talent — discard two cards, or one nonland card", banditsTalentDiscard); err != nil {
							return err
						}
					}
					return nil
				}),
			AtLevel(2, InterveningIf(CondBoard{
				Name: "that player has one or fewer cards in hand",
				Test: func(ctx *Context) bool { return cardsInHand(ctx.Game, ctx.Trigger().Event.Actor) <= 1 },
			}, On(game.EventBeginUpkeep, AllOf(ByAnOpponent, banditsTalentFewCards),
				"Bandit's Talent — that player loses 2 life (one or fewer cards in hand)",
				Do(Instruction(func(ctx *Context) error {
					return GainLife{Player: ctx.Trigger().Event.Actor, Amount: -2}.Apply(ctx)
				}))))),
			AtLevel(3, On(game.EventBeginDrawStep, ByYou,
				"Bandit's Talent — draw an additional card for each opponent with one or fewer cards in hand",
				Do(Instruction(func(ctx *Context) error {
					return DrawCards{N: banditsTalentOpponentsWithFewCards(ctx)}.Apply(ctx)
				})))),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{B}")),
			LevelUp(3, ManaCost("{3}{B}")),
		},
	})
}

// banditsTalentDiscard is the keyed discard: two cards, or one nonland
// card. An on-disk key: never renamed, never reused.
var banditsTalentDiscard = discardTwoUnlessOneRule("choose-cards/bandits-talent-discard",
	func(c game.Card) bool { return !c.IsLand() })

// banditsTalentFewCards is "if that player has one or fewer cards in hand",
// read off the player whose upkeep began.
func banditsTalentFewCards(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor != uuid.Nil && cardsInHand(g, ev.Actor) <= 1
}

// banditsTalentOpponentsWithFewCards counts the opponents with one or fewer
// cards in hand.
func banditsTalentOpponentsWithFewCards(ctx *Context) int {
	n := 0
	for _, opp := range ctx.Opponents() {
		if cardsInHand(ctx.Game, opp) <= 1 {
			n++
		}
	}
	return n
}
