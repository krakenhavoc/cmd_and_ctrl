package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sword of War and Peace — Artifact — Equipment {3}:
//
//	"Equipped creature gets +2/+2 and has protection from red and from white.
//	 Whenever equipped creature deals combat damage to a player, this Equipment
//	 deals damage to that player equal to the number of cards in their hand and
//	 you gain 1 life for each card in your hand.
//	 Equip {2}"
//
// The Swords' shared shape: the pump and ADR 0072's layer-6 protection grant
// on the equipped creature, and a combat-damage trigger matched through the
// attachment. Both counts are read as the trigger resolves, before anything is
// dealt: the damaged player's hand for the damage, the Sword's controller's
// hand for the life. The damage is the Equipment's (red and colourless, so it
// is not stopped by the wearer's protection); a player with no cards in hand
// takes none, and one that left the game takes none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b10ad415-0aae-4d70-82e6-a4055a40cbe4",
		Name:         "Sword of War and Peace",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("protection from red", "protection from white"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key:    "Sword of War and Peace — damage to that player equal to the cards in their hand, and you gain 1 life for each card in your hand",
			Effect: swordOfWarAndPeaceResolve,
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

func swordOfWarAndPeaceResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	victim := item.Trigger.Event.Target
	damage, life := 0, 0
	if p := g.PlayerByIDForEffect(victim); p != nil && p.Hand != nil {
		damage = p.Hand.Size()
	}
	if p := g.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Hand != nil {
		life = p.Hand.Size()
	}
	if damage > 0 {
		if err := (effects.DealDamage{Source: item.SourceCardID, Target: victim, Amount: damage}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.GainLife{Amount: life}.Apply(ctx)
}
