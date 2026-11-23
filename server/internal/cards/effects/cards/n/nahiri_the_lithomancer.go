package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri, the Lithomancer — Legendary Planeswalker — Nahiri {3}{W}{W},
// loyalty 3 (Commander Masters, #5948):
//
//	"+2: Create a 1/1 white Kor Soldier creature token. You may attach an
//	 Equipment you control to it.
//	 −2: You may put an Equipment card from your hand or graveyard onto the
//	 battlefield.
//	 −10: Create a colorless Equipment artifact token named Stoneforged
//	 Blade. It has indestructible, "Equipped creature gets +5/+5 and has
//	 double strike," and equip {0}.
//	 Nahiri, the Lithomancer can be your commander."
//
// The +2 creates the Soldier through the keyed creation; its tail then asks
// the controller for up to one Equipment they control that can legally be
// attached to the new token (CR 701.3b) and attaches it. The −2 is Bonny
// Pall's pick over both zones (Equipment cards, never tokens), the chosen
// card entering from whichever zone holds it, unattached. The −10 is a
// catalog token of its own (effects.StoneforgedBladeToken). "Can be your
// commander" is deck-building and is read from the oracle text by the deck
// validator, not here.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a554aa2c-d947-4fc6-929d-d5891f467615",
		Name:            "Nahiri, the Lithomancer",
		StartingLoyalty: 3,
		Completeness:    effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Create a 1/1 white Kor Soldier creature token. You may attach an Equipment you control to it.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups: []game.TokenGroup{{
							Template: ctx.TextToken(effects.TokenCard("1/1 white Kor Soldier")),
							Count:    1,
						}},
					}, ctx.Tail(nahiriAttachEquipment))
				},
			},
			{
				Label:  "−2: You may put an Equipment card from your hand or graveyard onto the battlefield.",
				Cost:   effects.LoyaltyCost(-2),
				Effect: nahiriPutEquipment,
			},
			{
				Label:   "−10: Create a colorless Equipment artifact token named Stoneforged Blade.",
				Cost:    effects.LoyaltyCost(-10),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.StoneforgedBladeToken(), N: 1}),
			},
		},
	})
}

// nahiriAttachEquipment is the +2's tail: the Equipment the controller may
// attach to the Soldier that was just created. An on-disk key.
var nahiriAttachEquipment = effects.CardTail("then/nahiri-the-lithomancer-attach-equipment", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	g, soldier := ctx.Game, o.IDs[0]
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != ctx.Controller() || !c.HasSubtype("Equipment") || c.IsAttachedTo(soldier) {
			continue
		}
		if g.AttachLegalForEffect(c.InstanceID, soldier) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Nahiri, the Lithomancer — you may attach an Equipment you control to the Kor Soldier",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		ThenKey:  nahiriAttachPicked,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{soldier},
	})
	return nil
})

// nahiriAttachPicked attaches the picked Equipment to the Soldier riding the
// prompt. An on-disk key: never renamed, never reused.
var nahiriAttachPicked = game.RegisterChooseCardsThen("choose-cards/nahiri-the-lithomancer-attach-equipment", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	return g.AttachForEffect(r.Picked[0], game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]})
})

// nahiriPutEquipment is the −2: the pick over the Equipment cards in the
// controller's hand and graveyard.
func nahiriPutEquipment(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if c.HasSubtype("Equipment") && !c.IsToken() {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Nahiri, the Lithomancer — you may put an Equipment card from your hand or graveyard onto the battlefield",
		Cards:    ids,
		Min:      0,
		Max:      1,
		ThenKey:  nahiriEquipmentPicked,
	})
	return nil
}

// nahiriEquipmentPicked puts the picked Equipment onto the battlefield from
// the zone that holds it now. An on-disk key: never renamed, never reused.
var nahiriEquipmentPicked = game.RegisterChooseCardsThen("choose-cards/nahiri-the-lithomancer-put-equipment", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	switch z.Kind {
	case game.ZoneHand:
		return g.PutFromHandOntoBattlefieldKeyedForEffect(id, game.HandEntryOptions{Controller: r.Chooser}, game.KeyedTail{})
	case game.ZoneGraveyard:
		return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, r.Chooser, false, game.KeyedTail{})
	}
	return nil
})
