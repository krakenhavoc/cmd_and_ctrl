package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "bcbe879a-4609-4493-bd7d-47c26d27e83c",
		Name:         "Bandit's Talent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bandit's Talent — each opponent discards two cards unless they discard a nonland card",
				func(g *game.Game, item *game.StackItem) error {
					for _, opp := range effects.NewContext(g, item).Opponents() {
						if err := effects.AskDiscardTwoUnlessOne(g, opp, item.SourceCardID, 2,
							"Bandit's Talent — discard two cards, or one nonland card", banditsTalentDiscard); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.AtLevel(2, effects.InterveningIf(effects.CondBoard{
				Name: "that player has one or fewer cards in hand",
				Test: func(ctx *effects.Context) bool { return effects.CardsInHand(ctx.Game, ctx.Trigger().Event.Actor) <= 1 },
			}, effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByAnOpponent, banditsTalentFewCards),
				"Bandit's Talent — that player loses 2 life (one or fewer cards in hand)",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.GainLife{Player: ctx.Trigger().Event.Actor, Amount: -2}.Apply(ctx)
				}))))),
			effects.AtLevel(3, effects.On(game.EventBeginDrawStep, effects.ByYou,
				"Bandit's Talent — draw an additional card for each opponent with one or fewer cards in hand",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.DrawCards{N: banditsTalentOpponentsWithFewCards(ctx)}.Apply(ctx)
				})))),
		},
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{B}")),
			effects.LevelUp(3, effects.ManaCost("{3}{B}")),
		},
	})
}

// banditsTalentDiscard is the keyed discard: two cards, or one nonland
// card. An on-disk key: never renamed, never reused.
var banditsTalentDiscard = effects.DiscardTwoUnlessOneRule("choose-cards/bandits-talent-discard",
	func(c game.Card) bool { return !c.IsLand() })

// banditsTalentFewCards is "if that player has one or fewer cards in hand",
// read off the player whose upkeep began.
func banditsTalentFewCards(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor != uuid.Nil && effects.CardsInHand(g, ev.Actor) <= 1
}

// banditsTalentOpponentsWithFewCards counts the opponents with one or fewer
// cards in hand.
func banditsTalentOpponentsWithFewCards(ctx *effects.Context) int {
	n := 0
	for _, opp := range ctx.Opponents() {
		if effects.CardsInHand(ctx.Game, opp) <= 1 {
			n++
		}
	}
	return n
}
