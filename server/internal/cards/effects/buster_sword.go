package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Buster Sword — Artifact — Equipment {3}:
//
//	"Equipped creature gets +3/+2.
//	 Whenever equipped creature deals combat damage to a player, draw
//	 a card, then you may cast a spell from your hand with mana value
//	 less than or equal to that damage without paying its mana cost.
//	 Equip {2}"
//
// The free cast (#3172) is the same offer Rishkar's Expertise and Surtland
// Elementalist make: after the draw, the controller picks a nonland card
// from hand with mana value no greater than the damage dealt (or none),
// and the pick is granted a one-shot "cast it for {0}" permission, so it
// is cast through the ordinary announce path with its targets and modes.
// The damage is read off the trigger's event, so a lifelinking or
// doubled-damage hit counts what was actually dealt.
//
// The draw itself needs nothing new: attachedCreatureDealtCombatDamageToPlayer
// is the Swords' own condition, reused verbatim.
func init() {
	Register(Spec{
		OracleID:     "5e060d58-4d6e-425c-b7d4-727669fcce5b",
		Name:         "Buster Sword",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 2),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Buster Sword — draw a card, then you may cast a spell from your hand for free",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return busterSwordOfferFreeCast(ctx, item.Controller, thatDamage(item))
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// busterSwordOfferFreeCast is "you may cast a spell from your hand with
// mana value less than or equal to that damage without paying its mana
// cost": a pick among the nonland hand cards that fit, or none.
func busterSwordOfferFreeCast(ctx *Context, controller uuid.UUID, damage int) error {
	if damage < 0 {
		return nil
	}
	pred := And(Not(Land()), ManaValueLE(damage))
	var candidates []uuid.UUID
	for _, id := range allHandCardIDs(ctx.Game, controller) {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !pred(ctx.Game, controller, c) {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   ctx.Source(),
		Question: "Buster Sword — cast a spell with mana value " + strconv.Itoa(damage) + " or less from your hand without paying its mana cost?",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			g.GrantCastPermissionOverCardForEffect(picked[0], game.CastPermission{
				Player:   controller,
				Cost:     "{0}",
				CastOnly: true,
				Label:    "Buster Sword — cast it without paying its mana cost",
			})
			return nil
		},
	})
	return nil
}
