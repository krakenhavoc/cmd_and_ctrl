package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saskia the Unyielding — Legendary Creature — Human Soldier {B}{R}{G}{W},
// 3/4 (Commander 2016, #5537):
//
//	"Vigilance, haste
//	 As Saskia enters, choose a player.
//	 Whenever a creature you control deals combat damage to a player, it
//	 deals that much damage to the chosen player."
//
// The choice is Sewer Nemesis's CR 614.12 "as enters" pick among every
// player still in the game, made again each time Saskia enters (a Saskia
// that returns does not remember the old pick). The trigger is one per
// damage event (Bident of Thassa's gate): the creature that dealt the
// combat damage deals that much again to the chosen player, read as the
// ability triggers so a Saskia that leaves in response still sends it.
// That second damage is not combat damage, so it does not trigger Saskia
// again; a creature that hits the chosen player hits them twice, as the
// ruling says, and a chosen player who has left the game takes nothing.
//
// No simplification.
func init() {
	const label = "Saskia the Unyielding — the creature deals that much damage to the chosen player"
	trigger := effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
	}, label, saskiaDamageTheChosenPlayer)
	// The chosen player is fixed as the ability triggers, off Saskia as
	// she is then; a fill-in Build records it.
	trigger.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Player = effects.ChosenPlayerOf(g, source.InstanceID)
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "e54d207c-51b8-458e-86a1-2633ac064c8e",
		Name:            "Saskia the Unyielding",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		AsEnters:        effects.ChoosePlayerAsEnters("Saskia the Unyielding", effects.Players),
		Triggered:       []game.TriggeredAbility{trigger},
	})
}

func saskiaDamageTheChosenPlayer(g *game.Game, item *game.StackItem) error {
	chosen := item.Params.Player
	if chosen == uuid.Nil || item.Trigger == nil {
		return nil
	}
	if p := g.PlayerByIDForEffect(chosen); p == nil || p.Eliminated {
		return nil
	}
	ev := item.Trigger.Event
	return effects.DealDamage{Source: ev.Source, Target: chosen, Amount: ev.Amount}.Apply(effects.NewContext(g, item))
}
