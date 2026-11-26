package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bane Alley Broker — Creature — Human Rogue {1}{U}{B}, 0/3 (Gatecrash,
// #5565):
//
//	"{T}: Draw a card, then exile a card from your hand face down.
//	 You may look at cards exiled with this creature.
//	 {U}{B}, {T}: Return a card exiled with this creature to its owner's hand."
//
// Jacob Hauken's draw-then-exile, linked to the Broker as an object
// (CR 607.2a): the card lands face down and hidden, so only its controller
// may look at it (CR 406.3), and a Broker that left and came back has none
// (CR 400.7). The card chosen for the exile may be the one just drawn, and
// an empty hand exiles nothing. The second ability has its controller choose
// one of the cards linked to this object, which goes to its OWNER's hand
// (a card exiled from an opponent's hand by a stolen Broker goes back to
// that opponent).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4993e373-4b76-4582-a789-6a3123e7ca89",
		Name:         "Bane Alley Broker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Draw a card, then exile a card from your hand face down. You may look at cards exiled with this creature.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  baneAlleyBrokerDrawExile,
			},
			{
				Label:   "{U}{B}, {T}: Return a card exiled with this creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{U}{B}"), effects.TapCost()),
				Effect:  baneAlleyBrokerReturn,
			},
		},
	})
}

func baneAlleyBrokerDrawExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	var hand []uuid.UUID
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
	}
	if len(hand) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   item.Controller,
		Source:    item.SourceCardID,
		Question:  "Bane Alley Broker — exile a card from your hand face down",
		Cards:     hand,
		Min:       1,
		Max:       1,
		LiveZone:  game.ZoneHand,
		ThenKey:   baneAlleyBrokerExiled,
		Carry:     []uuid.UUID{item.SourceCardID},
		CarryNums: []int{item.SourceEpoch},
		Asked:     game.NewAskedItem(item),
	})
	return nil
}

// baneAlleyBrokerExiled exiles the chosen card face down, linked to the
// Broker as the object the ability came from. An on-disk key: never renamed,
// never reused.
var baneAlleyBrokerExiled = effects.AskedPick("choose-cards/s82-old-b413-bane-alley-broker-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) > 0 && len(r.Carry) > 0 && len(r.Nums) > 0 {
		src := game.PermissionCardRef{ID: r.Carry[0], Epoch: r.Nums[0]}
		if _, err := ctx.Game.ExileHiddenLinkedForEffect(src, r.Chooser, r.Picked[0]); err != nil {
			return err
		}
	}
	return nil
})

func baneAlleyBrokerReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Bane Alley Broker — return a card exiled with this creature to its owner's hand",
		Cards:    ids,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  baneAlleyBrokerReturned,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// baneAlleyBrokerReturned puts the chosen card into its owner's hand. An
// on-disk key: never renamed, never reused.
var baneAlleyBrokerReturned = effects.AskedPick("choose-cards/s82-old-b413-bane-alley-broker-return", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(r.Picked[:1], game.KeyedTail{})
})
