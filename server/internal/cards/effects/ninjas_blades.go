package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ninja's Blades — Artifact — Equipment {2}{B} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+1, is a Ninja in addition to its other
//	 types, and has "Whenever this creature deals combat damage to a
//	 player, draw a card, then discard a card. That player loses life
//	 equal to the discarded card's mana value."
//	 Mutsunokami — Equip {2}"
//
// Black Mage's Rod's job-select shape with a granted combat-damage trigger
// (ADR 0093): the trigger belongs to the equipped creature, so the creature
// is the source and "you" is its controller. The draw waits for any prompt
// it asks, then the discard is the controller's choice (a keyed discard
// that carries the trigger's event, so "that player" is still the player
// who was dealt damage when the card has left the hand). The life loss
// reads the discarded card's mana value in the graveyard, with X as 0
// (CR 202.3e); no card discarded (an empty hand) means no life loss.
// "Mutsunokami" is a flavor word; the ability is an ordinary equip.
//
// No simplification.
func init() {
	const grantKey = "ninjas-blades/combat-damage-loot"
	loot := WheneverThisDealsCombatDamageToAPlayer("Ninja's Blades — draw a card, then discard a card; that player loses life equal to its mana value",
		Do(
			DrawCards{N: 1},
			Instruction(func(ctx *Context) error {
				return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player: ctx.Controller(),
					Source: ctx.Source(),
					N:      1,
					Asked:  game.NewAskedItem(ctx.Item),
				}, ninjasBladesDiscarded, nil)
			}),
		))
	Register(Spec{
		OracleID:     "cdaffb13-c2a4-4bc5-a9d6-c1a0d1686a81",
		Name:         "Ninja's Blades",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{loot},
			Text:      "Whenever this creature deals combat damage to a player, draw a card, then discard a card. That player loses life equal to the discarded card's mana value.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			AttachedIsAlso("Ninja"),
			GrantAbilitiesToAttached(grantKey),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Ninja's Blades")},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// ninjasBladesDiscarded is the discard answered: the player dealt damage
// loses life equal to the discarded card's mana value.
var ninjasBladesDiscarded = game.RegisterDiscardThen("choose-cards/ninjas-blades-loot", func(g *game.Game, r game.DiscardsDone) error {
	cards := r.Discarded.Cards()
	if len(cards) == 0 || r.Item == nil || r.Item.Trigger == nil {
		return nil
	}
	card, ok := g.LookupCardForEffect(cards[0])
	if !ok {
		return nil
	}
	victim := r.Item.Trigger.Event.Target
	if mv := card.ManaValue(); mv > 0 && g.PlayerByIDForEffect(victim) != nil {
		return g.ChangePlayerLifeForEffect(r.Source, victim, -mv)
	}
	return nil
})
