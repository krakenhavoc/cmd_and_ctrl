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
// The free cast (#3172, #3270) is the same offer Rishkar's Expertise and
// Surtland Elementalist make (offerFreeHandCast): after the draw, the
// controller picks a nonland card from hand with mana value no greater
// than the damage dealt (or none). The pick opens the cast at once, at any
// timing (CR 608.2g), through the ordinary announce path with its
// targets, modes and additional costs; passing priority without casting
// ends the chance and the card stays in hand.
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
	offerFreeHandCast(ctx, controller, "Buster Sword",
		"a spell with mana value "+strconv.Itoa(damage)+" or less",
		ManaValueLE(damage))
	return nil
}
