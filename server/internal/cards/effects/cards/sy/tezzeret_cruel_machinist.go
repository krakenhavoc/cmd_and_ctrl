package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret, Cruel Machinist — Legendary Planeswalker — Tezzeret {4}{U}{U},
// loyalty 4 (Core Set 2019):
//
//	"+1: Draw a card.
//	 0: Until your next turn, target artifact you control becomes a 5/5 creature
//	 in addition to its other types.
//	 −7: Put any number of cards from your hand onto the battlefield face down.
//	 They're 5/5 artifact creatures."
//
// The 0 is Wrenn and Realmbreaker's +1: one record pinned to the artifact until
// the controller's next turn, adding the Creature type and setting base 5/5.
// The −7 asks for any number of cards in hand (a nonpermanent card is legal
// face down, CR 708.2), then puts them onto the battlefield as ONE entry with
// the LISTED body of a 5/5 artifact creature (CR 708.2, Cybership's shape): no
// enter trigger or "as enters" choice runs for a face-down entry (CR 708.2a)
// and only their controller may look at them (CR 708.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e5a95145-dc29-4bfc-bb14-b7a269268366",
		Name:            "Tezzeret, Cruel Machinist",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card.",
				Purpose: game.Purpose{Draws: 1},
				Cost:    effects.LoyaltyCost(1),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "0: Until your next turn, target artifact you control becomes a 5/5 creature in addition to its other types.",
				Cost:    effects.LoyaltyCost(0),
				Targets: effects.TargetPermanent("target artifact you control", effects.And(effects.Artifact(), effects.YouControl())),
				Effect:  tezzeretCruelMachinistAnimate,
			},
			{
				Label:  "−7: Put any number of cards from your hand onto the battlefield face down. They're 5/5 artifact creatures.",
				Cost:   effects.LoyaltyCost(-7),
				Effect: tezzeretCruelMachinistMinusSeven,
			},
		},
	})
}

func tezzeretCruelMachinistAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: t.ID,
		Mods: append([]game.Mod{game.AddTypesMod("Creature")},
			game.SetBasePTMods(5, 5)...),
		Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Tezzeret, Cruel Machinist — becomes a 5/5 creature until your next turn",
	}.Apply(ctx)
}

func tezzeretCruelMachinistMinusSeven(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	hand := effects.HandCardIDs(g, ctx.Controller())
	if len(hand) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     ctx.Controller(),
		FromPlayer:  ctx.Controller(),
		Source:      item.SourceCardID,
		Question:    "Tezzeret, Cruel Machinist — put any number of cards from your hand onto the battlefield face down",
		Cards:       hand,
		Min:         0,
		Max:         len(hand),
		Zone:        game.ZoneHand,
		Destination: game.ChooseOntoBattlefield,
		ThenKey:     tezzeretCruelMachinistPicked,
		Asked:       game.NewAskedItem(item),
	})
	return nil
}

// tezzeretCruelMachinistBody is the LISTED body of the cards: 5/5 artifact
// creatures with no name, colour or other characteristic (CR 708.2).
func tezzeretCruelMachinistBody() *game.FaceDownListing {
	return &game.FaceDownListing{Types: []string{"Artifact", "Creature"}, Power: 5, Toughness: 5}
}

// tezzeretCruelMachinistPicked puts the picked cards onto the battlefield face
// down as 5/5 artifact creatures, in one entry. An on-disk key: never renamed,
// never reused.
var tezzeretCruelMachinistPicked = effects.ItemChoice("choose-cards/tezzeret-cruel-machinist-face-down", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.PutCardsFromHandOntoBattlefieldKeyedForEffect(r.Picked, game.HandEntryOptions{
		Controller:     ctx.Controller(),
		FaceDown:       game.FaceDownTurned,
		FaceDownListed: tezzeretCruelMachinistBody(),
	}, game.KeyedTail{})
})
