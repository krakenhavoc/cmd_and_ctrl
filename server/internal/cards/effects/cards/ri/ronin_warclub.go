package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ronin Warclub — Artifact — Equipment {3} (Betrayers of Kamigawa, #5602):
//
//	"Equipped creature gets +2/+1.
//	 Whenever a creature you control enters, attach this Equipment to that
//	 creature.
//	 Equip {5}"
//
// The pump is a layer 7c modify (PumpAttached). The enters trigger is
// Quintessential Katana's, mandatory and for every creature you control that
// enters: it attaches the Warclub to the creature, through the door equip uses
// (AttachSourceForEffect), so an attach that cannot happen (the Warclub or the
// creature has gone, CR 701.3b) does nothing. Equip {5} is the ordinary
// sorcery-speed equip.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11b969f2-ff68-4c9e-b3f3-839736084303",
		Name:         "Ronin Warclub",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			Key: "Ronin Warclub — attach this Equipment to that creature",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID})
			},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{5}")},
	})
}
