package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armory of Iroas — Artifact — Equipment {2} (Commander Legends):
//
//	"Whenever equipped creature attacks, put a +1/+1 counter on it.
//	 Equip {2}"
//
// "It" is the creature that attacked (the trigger's object), so a counter
// goes on it even if the Armory has been moved in response; a creature that
// has left the battlefield gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ac1aab0-9d2d-4fb7-b0bb-0af3e6d665db",
		Name:         "Armory of Iroas",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, "Armory of Iroas — put a +1/+1 counter on the equipped creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.AddCounter{Target: item.Trigger.Event.CardID, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
