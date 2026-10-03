package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kiora, the Crashing Wave — Legendary Planeswalker — Kiora {2}{G}{U},
// loyalty 2:
//
//	"+1: Until your next turn, prevent all damage that would be dealt to
//	 and dealt by target permanent an opponent controls.
//	 −1: Draw a card. You may play an additional land this turn.
//	 −5: You get an emblem with 'At the beginning of your end step, create
//	 a 9/9 blue Kraken creature token.'"
//
// ADR 0108 §7, Delivery PR 7 (#1904): the +1 is Dovin's to-and-by row
// (one record, all damage, until your next turn). The −1 is Explore's
// draw and extra land play; the −5 is an emblem (#623) with one
// end-step trigger.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "58cc2097-f4b4-4695-9076-78666050cdb6",
		Name:            "Kiora, the Crashing Wave",
		Completeness:    CompletenessFull,
		StartingLoyalty: 2,
		Emblem: &EmblemSpec{
			Label: "Kiora, the Crashing Wave emblem",
			Text:  "At the beginning of your end step, create a 9/9 blue Kraken creature token.",
			Triggered: []game.TriggeredAbility{
				AtYourEndStep("Kiora, the Crashing Wave emblem — create a 9/9 blue Kraken",
					Do(CreateToken{Template: TokenCard("9/9 blue Kraken"), N: 1})),
			},
		},
		Activated: []ActivatedAbility{
			untilYourNextTurnToAndByRow("+1: Until your next turn, prevent all damage that would be dealt to and dealt by target permanent an opponent controls.",
				LoyaltyCost(1)),
			{
				Label: "−1: Draw a card. You may play an additional land this turn.",
				Cost:  LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := (DrawCards{Player: item.Controller, N: 1}).Apply(NewContext(g, item)); err != nil {
						return err
					}
					g.GrantAdditionalLandPlayForEffect(item.Controller, 1)
					return nil
				},
			},
			{
				Label:  "−5: You get an emblem with \"At the beginning of your end step, create a 9/9 blue Kraken creature token.\"",
				Cost:   LoyaltyCost(-5),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}
