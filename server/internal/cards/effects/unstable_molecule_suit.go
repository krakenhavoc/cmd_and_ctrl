package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2bc0821a-5a8f-465d-8532-3fb5c5d11d8f",
		Name:         "Unstable Molecule Suit",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("indestructible"),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip commander {2}", "{2}",
				TargetCreature("target creature you control that's your commander", YouControl(), YourCommander())),
			EquipAbility("{4}"),
		},
	})
}
