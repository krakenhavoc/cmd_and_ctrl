package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fishing Gear — Artifact — Equipment {3} (Final Fantasy, #4462):
//
//	"Whenever equipped creature deals combat damage to a player, exile the
//	 top card of that player's library. If it's a permanent card, you may
//	 put it onto the battlefield under your control. If you don't, create a
//	 1/1 blue Fish creature token.
//	 Equip {2}"
//
// The player is the damage event's target. The exile is the mill-to-exile
// primitive, face up, and everything after it waits for the card to have
// landed (CR 608.2c): a permanent card is offered, and a nonpermanent card,
// an empty library, or a "no" all mean "you don't", so the Fish is made. A
// yes that does not bring the card onto the battlefield (a replacement kept
// it out) also means you didn't put it, so it makes the Fish too. The card
// enters under the controller's control and stays its owner's (CR 108.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d309b18-8dc5-41ee-888c-e275761029e7",
		Name:         "Fishing Gear",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Fishing Gear — exile the top card of that player's library; you may put a permanent card onto the battlefield, otherwise create a Fish",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim := ctx.Trigger().Event.Target
					return effects.MillToZone{
						Player:   victim,
						N:        1,
						To:       game.ZoneExile,
						ThenTail: ctx.ItemTail(fishingGearAfterExile),
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}

// fishingGearAfterExile offers the exiled permanent card, or makes the
// Fish at once when there is nothing to offer. An on-disk key: never
// renamed, never reused.
var fishingGearAfterExile = effects.CardTail("fishing-gear/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return fishingGearFish(ctx)
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || !effects.IsPermanentCard(c) {
		return fishingGearFish(ctx)
	}
	return effects.MayChoice{
		Question: "Fishing Gear — put " + c.Name + " onto the battlefield under your control? (If you don't, create a 1/1 blue Fish.)",
		YesLabel: "Put it onto the battlefield",
		NoLabel:  "Create a Fish",
		Then:     fishingGearAnswered,
		Carry:    []uuid.UUID{o.IDs[0]},
	}.Apply(ctx)
})

// fishingGearAnswered puts the card on a yes and makes the Fish on a no,
// or when the card did not arrive. An on-disk key: never renamed, never
// reused.
var fishingGearAnswered = effects.MayYesNo("confirm/fishing-gear-put",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if len(r.Carry) == 0 {
			return fishingGearFish(ctx)
		}
		return effects.ReturnFromExile{
			Target:     r.Carry[0],
			Controller: ctx.Controller(),
			ThenTail:   ctx.ItemTail(fishingGearEntered),
		}.Apply(ctx)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error { return fishingGearFish(ctx) })

// fishingGearEntered makes the Fish when the card did not enter. An
// on-disk key: never renamed, never reused.
var fishingGearEntered = effects.EnteredTail("fishing-gear/entered", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered != uuid.Nil {
		return nil
	}
	return fishingGearFish(ctx)
})

// fishingGearFish is "create a 1/1 blue Fish creature token".
func fishingGearFish(ctx *effects.Context) error {
	return effects.CreateToken{Template: effects.TokenCard("1/1 blue Fish"), N: 1}.Apply(ctx)
}
