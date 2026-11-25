package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steelclaw Lance — Artifact — Equipment {B}{R} (Throne of Eldraine, S82 old
// sets):
//
//	"Equipped creature gets +2/+2.
//	 Equip Knight {1}
//	 Equip {3}"
//
// Two equip abilities, Brotherhood Regalia's shape: the cheap one narrowed to
// Knights you control, the plain one open to any creature you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfe06fb5-83b5-4212-86f8-947dfba15b26",
		Name:         "Steelclaw Lance",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip Knight {1}", "{1}",
				effects.TargetCreature("target Knight you control", effects.YouControl(), effects.Subtype("Knight"))),
			effects.EquipAbility("{3}"),
		},
	})
}
