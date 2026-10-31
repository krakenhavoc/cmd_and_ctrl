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
// controller picks a card from hand with a spell of mana value no greater
// than the damage dealt (or none), judged on the face being cast (#4509). The pick opens the cast at once, at any
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
			Effect: Do(
				// The free cast waits for a prompt the draw asks (CR 608.2c, #4155).
				DrawCards{N: 1},
				Instruction(func(ctx *Context) error {
					return busterSwordOfferFreeCast(ctx, ctx.Controller(), thatDamage(ctx.Item))
				}),
			),
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// busterSwordOfferFreeCast is "you may cast a spell from your hand with
// mana value less than or equal to that damage without paying its mana
// cost": a pick among the hand cards with a spell that fits, or none.
func busterSwordOfferFreeCast(ctx *Context, controller uuid.UUID, damage int) error {
	if damage < 0 {
		return nil
	}
	offerFreeHandCast(ctx, controller, busterSwordFreeCast,
		"a spell with mana value "+strconv.Itoa(damage)+" or less",
		spellWithManaValueAtMost(damage))
	return nil
}

// busterSwordFreeCast is the offer's keyed continuation (#4246).
var busterSwordFreeCast = registerFreeHandCastOffer("choose-cards/buster-sword-cast", "Buster Sword")
