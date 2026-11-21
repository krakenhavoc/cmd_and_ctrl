package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unstable Molecule Suit — Artifact — Equipment {3} (Marvel Super Heroes
// Commander, #4455):
//
//	"Equipped creature gets +2/+2 and has indestructible.
//	 Equip commander {2}
//	 Equip {4}"
//
// Commander's Plate's pair of equip abilities: the cheap one is narrowed
// to a creature that is your commander, the other is the ordinary equip.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2bc0821a-5a8f-465d-8532-3fb5c5d11d8f",
		Name:         "Unstable Molecule Suit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("indestructible"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip commander {2}", "{2}",
				effects.TargetCreature("target creature you control that's your commander", effects.YouControl(), effects.YourCommander())),
			effects.EquipAbility("{4}"),
		},
	})
}
