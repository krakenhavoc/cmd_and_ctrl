package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tamiyo, the Moon Sage — Legendary Planeswalker {3}{U}{U}, loyalty 4
// (Jumpstart 2022, #5948):
//
//	"+1: Tap target permanent. It doesn't untap during its controller's
//	 next untap step.
//	 −2: Draw a card for each tapped creature target player controls.
//	 −8: You get an emblem with "You have no maximum hand size" and
//	 "Whenever a card is put into your graveyard from anywhere, you may
//	 return it to your hand.""
//
// The +1 is Chill of the Grave's tap-and-freeze. The −2 counts the tapped
// creatures as it resolves. The emblem is a no-maximum hand-size static and
// a trigger on every card of the owner's arriving in their graveyard from
// anywhere (graveyard_arrival.go), one trigger per card, each a "may".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f88be051-36ae-407e-bd5a-2b029c26414e",
		Name:            "Tamiyo, the Moon Sage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label:    "Tamiyo, the Moon Sage emblem",
			Text:     "You have no maximum hand size. Whenever a card is put into your graveyard from anywhere, you may return it to your hand.",
			HandSize: []game.HandSizeStatic{{Players: game.HandSizeYou, Kind: game.HandSizeNoMaximum}},
			Triggered: []game.TriggeredAbility{
				effects.Optional(
					effects.OnAny(effects.GraveyardArrivalKinds, effects.GraveyardArrivals{}.When(),
						"Tamiyo, the Moon Sage emblem — return that card to your hand",
						func(g *game.Game, item *game.StackItem) error {
							if item.Trigger == nil {
								return nil
							}
							ctx := effects.NewContext(g, item)
							return effects.ReturnFromGraveyard{Target: item.Trigger.Event.CardID, Dest: game.ZoneHand}.Apply(ctx)
						}),
					"Return that card to your hand?"),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Tap target permanent. It doesn't untap during its controller's next untap step.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPermanent("target permanent", effects.Permanent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Tamiyo, the Moon Sage — it doesn't untap during its controller's next untap step"}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Draw a card for each tapped creature target player controls.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetPlayer("target player"),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					n := len(effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.TappedPermanent(), effects.ControlledBy(t.ID))))
					return effects.DrawCards{N: n}.Apply(ctx)
				},
			},
			{
				Label: "−8: You get an emblem with \"You have no maximum hand size\" and \"Whenever a card is put into your graveyard from anywhere, you may return it to your hand.\"",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
