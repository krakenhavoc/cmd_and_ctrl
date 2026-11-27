package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulshok Gauntlets — Artifact — Equipment {2} (Mirrodin, #5606):
//
//	"Equipped creature gets +4/+2 and doesn't untap during its controller's
//	 untap step.
//	 Equip {3}"
//
// Heavy Arbalest's untap restriction (CR 502.3) with a plain pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8b1672c-e5b1-41e7-9f34-c822f1587369",
		Name:         "Vulshok Gauntlets",
		Completeness: effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{
			{Label: "equipped creature doesn't untap", Restricts: effects.AttachedToSource},
		},
		Static:    []game.StaticAbility{effects.PumpAttached(4, 2)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
