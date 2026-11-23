package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hammer of Ruin — Artifact — Equipment {2} (Jumpstart 2022, #5948):
//
//	"Equipped creature gets +2/+0.
//	 Whenever equipped creature deals combat damage to a player, you may
//	 destroy target Equipment that player controls.
//	 Equip {2}"
//
// The target is chosen as the trigger goes on the stack, among the Equipment
// the damaged player controls (CR 603.3d); "you may" is answered as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78ba3363-0fc8-42a9-b586-69d0795ca65e",
		Name:         "Hammer of Ruin",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
					}, "Hammer of Ruin — you may destroy target Equipment that player controls", effects.DestroyFirstLegalCardTarget),
					"Hammer of Ruin — destroy target Equipment that player controls?"),
				effects.ControlledByThatPlayer(effects.TargetPermanent("target Equipment that player controls", effects.HasSubtype("Equipment"))),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
