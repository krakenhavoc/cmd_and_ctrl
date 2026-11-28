package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Purgatory — Enchantment {2}{W}{B} (Mirage, S82 OLD batch 809, #5633):
//
//	"Whenever a nontoken creature is put into your graveyard from the
//	 battlefield, exile that card.
//	 At the beginning of your upkeep, you may pay {4} and 2 life. If you do,
//	 return a card exiled with this enchantment to the battlefield."
//
// The first ability exiles the card linked to this object (ExileTarget.Linked,
// CR 607.2a), only while it is still the card that died (CR 400.7). The
// second pays {4} through the prompt and the 2 life inside the paid branch
// (Ripples of Undeath's split, CR 119.4 checked before the offer), then has
// its controller choose one linked card, which returns under their control.
// No linked card, or under 2 life, is no offer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fb8e100-58a1-4dc6-bc38-ad7f115e3e46",
		Name:         "Purgatory",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !dead.IsToken() && dead.Owner == source.Controller
			}, "Purgatory — exile that card", effects.ExileTheDeadLinked),
			effects.AtYourUpkeep("Purgatory — you may pay {4} and 2 life; if you do, return a card exiled with this enchantment to the battlefield",
				purgatoryUpkeepOffer),
		},
	})
}

// purgatoryExiledWith lists the cards exiled with the resolving ability's
// source object (CR 607.2a), oldest first.
func purgatoryExiledWith(ctx *effects.Context) []uuid.UUID {
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// purgatoryPaid pays the 2 life, then has its controller choose the card. An
// on-disk key: never renamed, never reused.
var purgatoryPaid = effects.PayUnlessBranches("pay-unless/purgatory-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if err := ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), 2); err != nil {
			return err
		}
		ids := purgatoryExiledWith(ctx)
		switch len(ids) {
		case 0:
			return nil
		case 1:
			return effects.ReturnFromExile{Target: ids[0], Controller: ctx.Controller()}.Apply(ctx)
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Purgatory — return a card exiled with this enchantment to the battlefield",
			Cards:    ids,
			Min:      1,
			Max:      1,
			LiveZone: game.ZoneExile,
			ThenKey:  purgatoryReturned,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}, nil)

// purgatoryReturned puts the chosen card onto the battlefield. An on-disk key:
// never renamed, never reused.
var purgatoryReturned = effects.AskedPick("choose-cards/s82-old-b809-purgatory-return", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromExile{Target: r.Picked[0], Controller: ctx.Controller()}.Apply(ctx)
})

// purgatoryUpkeepOffer is the upkeep offer: made only with 2 life to pay
// (CR 119.4) and a card to return.
func purgatoryUpkeepOffer(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Life < 2 || len(purgatoryExiledWith(ctx)) == 0 {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{4}",
		Question: "Purgatory — pay {4} and 2 life to return a card exiled with it to the battlefield?",
		Then:     purgatoryPaid,
	}.Apply(ctx)
}
