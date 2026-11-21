package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ac29873e-28c3-4933-b427-61d34ff0afcd",
		Name:            "Jace, Unraveler of Secrets",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &EmblemSpec{
			Label: "Jace, Unraveler of Secrets emblem",
			Text:  "Whenever an opponent casts their first spell each turn, counter that spell.",
			Triggered: []game.TriggeredAbility{
				On(game.EventCast, AllOf(AnOpponentCast(nil), func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return g.CastTallyFor(ev.Actor).Total == 1
				}), "Jace, Unraveler of Secrets emblem — counter that spell", func(g *game.Game, item *game.StackItem) error {
					return CounterTarget{StackID: item.Trigger.Event.CardID}.Apply(NewContext(g, item))
				}),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Scry 1, then draw a card.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Scry: 1, Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return Scry{Player: item.Controller, N: 1, ThenTail: ctx.TailN(thenDrawCards, 1)}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Return target creature to its owner's hand.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCreature("target creature"),
				Effect:  bounceChosenTarget,
			},
			{
				Label: "−8: You get an emblem with \"Whenever an opponent casts their first spell each turn, counter that spell.\"",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
