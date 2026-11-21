package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thinking Cap — Artifact — Equipment {1} (Murders at Karlov Manor,
// #4469):
//
//	"Equipped creature gets +1/+2.
//	 Equip Detective {1}
//	 Equip {3} (Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// Dúnedain Blade's shape: the anthem on the equipped creature, a
// narrowed equip for a Detective you control (EquipOnlyAbility, so
// Leonin Shikari's permission still finds it) and the ordinary equip.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a54fb3de-0581-461e-be33-dc4f4d16a33e",
		Name:         "Thinking Cap",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip Detective {1}", "{1}",
				effects.TargetCreature("target Detective you control", effects.YouControl(), effects.Subtype("Detective"))),
			effects.EquipAbility("{3}"),
		},
	})
}
