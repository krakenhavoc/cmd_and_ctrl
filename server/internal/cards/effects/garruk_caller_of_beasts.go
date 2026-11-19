package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garruk, Caller of Beasts — Legendary Planeswalker — Garruk {4}{G}{G},
// loyalty 4 (Foundations Jumpstart, #5929):
//
//	"+1: Reveal the top five cards of your library. Put all creature cards
//	 revealed this way into your hand and the rest on the bottom of your
//	 library in any order.
//	 −3: You may put a green creature card from your hand onto the
//	 battlefield.
//	 −7: You get an emblem with "Whenever you cast a creature spell, you may
//	 search your library for a creature card, put it onto the battlefield,
//	 then shuffle.""
//
// The +1 is Ajani Unyielding's reveal-and-take-all over five cards and
// creature cards. The −3 is Akul the Unrepentant's put from hand, matching
// green creature cards. The emblem's trigger is Garruk, Unleashed's search
// moved from the end step to a creature cast (CR 114.4: it is the owner's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8e2fbc26-5373-4fd7-8e9e-41af5bb9ff18",
		Name:            "Garruk, Caller of Beasts",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Garruk, Caller of Beasts emblem",
			Text:  "Whenever you cast a creature spell, you may search your library for a creature card, put it onto the battlefield, then shuffle.",
			Triggered: []game.TriggeredAbility{
				WheneverYouCast(Creature(), "Garruk, Caller of Beasts emblem — search for a creature",
					func(g *game.Game, item *game.StackItem) error {
						return SearchLibrary{Player: item.Controller, Predicate: MatchCreature, Dest: game.ZoneBattlefield,
							Limit: 1, Optional: true, Shuffle: true,
							Reason: "Garruk, Caller of Beasts emblem — a creature onto the battlefield"}.Apply(NewContext(g, item))
					}),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Reveal the top five cards of your library. Put all creature cards revealed this way into your hand and the rest on the bottom of your library in any order.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 5),
						Match:  Creature(),
						All:    true,
						Label:  "Garruk, Caller of Beasts — reveal the top five cards, creatures to your hand and the rest on the bottom",
						Key:    takeRevealRestBottomAny,
					}.Apply(ctx)
				},
			},
			{
				Label: "−3: You may put a green creature card from your hand onto the battlefield.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Match:    And(Creature(), OfColor("G")),
						Optional: true,
						Label:    "Garruk, Caller of Beasts — you may put a green creature card from your hand onto the battlefield",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "−7: You get an emblem with \"Whenever you cast a creature spell, you may search your library for a creature card, put it onto the battlefield, then shuffle.\"",
				Cost:   LoyaltyCost(-7),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}
