package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant's Skewer — Artifact — Equipment {1}{B} (Throne of Eldraine, S82 old
// sets):
//
//	"Equipped creature gets +2/+1.
//	 Whenever equipped creature deals combat damage to a creature, create a
//	 Food token.
//	 Equip {3}"
//
// The pump is PumpAttached. The trigger is on the Equipment, keyed on the
// creature it is attached to (Baseball Bat's attachment test) and on combat
// damage to a creature (Kaldra Compleat's, with the attachment swapped in): it
// fires once per damage event, so a creature that deals combat damage to two
// creatures makes two Foods.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaa8377e-21de-4c63-a26a-c25d986a272d",
		Name:         "Giant's Skewer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.OldB148AttachedCreatureDealtCombatDamageToACreature,
				"Giant's Skewer — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
