package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Robe of the Archmagi — Artifact — Equipment {2}{U} (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Whenever equipped creature deals combat damage to a player, you draw that
//	 many cards.
//	 Equip {4}
//	 Equip Shaman, Warlock, or Wizard {1}"
//
// The number is the combat damage that creature dealt to that player, read
// off the event that triggered it, so a trigger that outlives the creature
// still draws the right amount. The Robe has two equip abilities (CR 702.6): the
// ordinary {4}, and {1} that can only target a Shaman, Warlock or Wizard you
// control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "570de99a-7c65-4052-bc4b-cfb8caef40b3",
		Name:         "Robe of the Archmagi",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Robe of the Archmagi — draw that many cards", robeOfTheArchmagiDraw),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip Shaman, Warlock, or Wizard {1}", "{1}",
				effects.TargetCreature("target Shaman, Warlock, or Wizard you control", effects.YouControl(),
					effects.Or(effects.Subtype("Shaman"), effects.Subtype("Warlock"), effects.Subtype("Wizard")))),
			effects.EquipAbility("{4}"),
		},
	})
}

func robeOfTheArchmagiDraw(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	n := item.Trigger.Event.Amount
	if n <= 0 {
		return nil
	}
	return effects.DrawCards{Player: item.Controller, N: n}.Apply(effects.NewContext(g, item))
}
