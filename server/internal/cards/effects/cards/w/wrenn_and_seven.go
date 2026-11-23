package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wrenn and Seven — Legendary Planeswalker — Wrenn {3}{G}{G}, starting
// loyalty 5:
//
//	"+1: Reveal the top four cards of your library. Put all land cards
//	 revealed this way into your hand and the rest into your graveyard.
//	 0: Put any number of land cards from your hand onto the battlefield
//	 tapped.
//	 −3: Create a green Treefolk creature token with reach and "This
//	 token's power and toughness are each equal to the number of lands you
//	 control."
//	 −8: Return all permanent cards from your graveyard to your hand. You
//	 get an emblem with "You have no maximum hand size.""
//
// The +1 reveals the cards, then moves the lands to the hand and the rest
// to the graveyard; nothing is chosen. The emblem's one ability is a
// maximum-hand-size static (EmblemSpec.HandSize), folded with every other
// effect on the owner's maximum hand size.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63a509ea-cedc-40e9-b4e8-0ff4fc356485",
		Name:            "Wrenn and Seven",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label:    "Wrenn and Seven emblem",
			Text:     "You have no maximum hand size.",
			HandSize: []game.HandSizeStatic{{Kind: game.HandSizeNoMaximum}},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Reveal the top four cards of your library. Put all land cards revealed this way into your hand and the rest into your graveyard.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 4,
						"Wrenn and Seven — reveal the top four cards of your library")
					var lands, rest []uuid.UUID
					for _, id := range revealed {
						if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
							lands = append(lands, id)
						} else {
							rest = append(rest, id)
						}
					}
					g.TakeFromLibraryToHandForEffect(item.Controller, lands)
					return effects.RestIntoGraveyard(g, effects.CardsStillInALibrary(g, rest))
				},
			},
			{
				Label: "0: Put any number of land cards from your hand onto the battlefield tapped.",
				Cost:  effects.LoyaltyCost(0),
				Effect: effects.Do(effects.PutFromHandOntoBattlefield{
					Match:    effects.Land(),
					Tapped:   true,
					Optional: true,
					Max:      1000,
					Label:    "Wrenn and Seven — put any number of land cards from your hand onto the battlefield tapped",
				}),
			},
			{
				Label:   "−3: Create a green Treefolk creature token with reach and \"This token's power and toughness are each equal to the number of lands you control.\"",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.Cmdr2S51TreefolkToken(), N: 1}),
			},
			{
				Label: "−8: Return all permanent cards from your graveyard to your hand. You get an emblem with \"You have no maximum hand size.\"",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, id := range effects.GraveyardCardIDs(ctx, item.Controller, effects.IsPermanentCard) {
						if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.CreateEmblem{}.Apply(ctx)
				},
			},
		},
	})
}
