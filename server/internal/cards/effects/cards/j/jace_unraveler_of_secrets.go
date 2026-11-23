package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace, Unraveler of Secrets — Legendary Planeswalker — Jace {3}{U}{U},
// loyalty 5 (Innistrad Remastered):
//
//	"+1: Scry 1, then draw a card.
//	 −2: Return target creature to its owner's hand.
//	 −8: You get an emblem with "Whenever an opponent casts their first spell
//	 each turn, counter that spell.""
//
// The scry finishes before the draw (Scry's ThenTail). The emblem's trigger is
// an ordinary trigger of the emblem object (CR 114.4): the per-player cast
// tally is bumped before the cast event fires, so a tally of one is the
// opponent's first spell this turn, and it is countered as the trigger
// resolves (a spell that can't be countered stays).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac29873e-28c3-4933-b427-61d34ff0afcd",
		Name:            "Jace, Unraveler of Secrets",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Jace, Unraveler of Secrets emblem",
			Text:  "Whenever an opponent casts their first spell each turn, counter that spell.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventCast, effects.AllOf(effects.AnOpponentCast(nil), func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return g.CastTallyFor(ev.Actor).Total == 1
				}), "Jace, Unraveler of Secrets emblem — counter that spell", func(g *game.Game, item *game.StackItem) error {
					return effects.CounterTarget{StackID: item.Trigger.Event.CardID}.Apply(effects.NewContext(g, item))
				}),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Scry 1, then draw a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Scry: 1, Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Scry{Player: item.Controller, N: 1, ThenTail: ctx.TailN(effects.ThenDrawCards, 1)}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Return target creature to its owner's hand.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BounceChosenTarget,
			},
			{
				Label: "−8: You get an emblem with \"Whenever an opponent casts their first spell each turn, counter that spell.\"",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
