package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unquenchable Fury — Enchantment — Aura {2}{R} (Neon Dynasty Commander,
// #5493):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature attacks, it deals X
//	 damage to defending player, where X is the number of cards in their
//	 hand."
//	 When this Aura is put into your graveyard from the battlefield, return
//	 it to your hand."
//
// The attack trigger is a granted bundle (ADR 0093): the enchanted creature
// is the source of the damage, "defending player" is the player attacked or
// the controller of the planeswalker or battle attacked, and X is their hand
// as the trigger resolves. The return is Rancor's: any trip to the graveyard
// from the battlefield, but only into the graveyard of the Aura's controller
// (the Aura's owner, when it is also its controller).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c69a41e-122a-42c1-98d4-56035d68cfc8",
		Name:         "Unquenchable Fury",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: unquenchableFuryAttack,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Unquenchable Fury — it deals X damage to defending player, where X is the number of cards in their hand",
					func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil {
							return nil
						}
						defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
						p := g.PlayerByIDForEffect(defender)
						if p == nil || p.Eliminated {
							return nil
						}
						return effects.DealDamage{Source: item.SourceCardID, Target: defender, Amount: p.Hand.Size()}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever this creature attacks, it deals X damage to defending player, where X is the number of cards in their hand.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(unquenchableFuryAttack)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source) && source.Owner == source.Controller
			}, "Unquenchable Fury — return it to your hand", func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

const unquenchableFuryAttack = "unquenchable-fury/attack"
