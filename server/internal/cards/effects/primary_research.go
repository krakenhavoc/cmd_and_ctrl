package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primary Research — Enchantment {4}{W} (Secrets of Strixhaven, #4456):
//
//	"When this enchantment enters, return target nonland permanent card with
//	 mana value 3 or less from your graveyard to the battlefield.
//	 At the beginning of your end step, if a card left your graveyard this
//	 turn, draw a card."
//
// The enters trigger is Guardian Scalelord's return with a fixed mana-value
// bound: a target is chosen as the trigger goes on the stack and judged again
// as it resolves (CR 608.2b), and the card comes back under the controller's
// control. The end-step ability has an intervening "if" (CR 603.4), checked
// when the step begins and again on resolution. "A card left your graveyard"
// reads the turn's events for a move out of a graveyard of a card you own
// (cardLeftYourGraveyardThisTurn), so the card this very enchantment returned,
// a card cast from the graveyard, exiled or shuffled away all count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "51af3da5-0970-436e-baa6-3171889a40e0",
		Name:         "Primary Research",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Primary Research — return target nonland permanent card with mana value 3 or less from your graveyard to the battlefield",
					returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target nonland permanent card with mana value 3 or less from your graveyard",
					YouOwn(), NonlandPermanentCard(), ManaValueLE(3))),
			InterveningIf(CondBoard{
				Name: "a card left your graveyard this turn",
				Test: func(ctx *Context) bool { return cardLeftYourGraveyardThisTurn(ctx.Game, ctx.Controller()) },
			}, On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && cardLeftYourGraveyardThisTurn(g, source.Controller)
				},
				"Primary Research — draw a card",
				Do(DrawCards{N: 1}))),
		},
	})
}
