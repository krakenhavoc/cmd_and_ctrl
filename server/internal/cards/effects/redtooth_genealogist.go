package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Redtooth Genealogist — Creature — Elf Advisor {2}{G}, 2/3:
//
//	"When this creature enters, create a Royal Role token attached to
//	 another target creature you control. (If you control another Role
//	 on it, put that one into the graveyard. Enchanted creature gets
//	 +1/+1 and has ward {1}.)"
//
// "Another" excludes the Genealogist by object; with no other creature
// the trigger is removed (CR 603.3d). No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f6cb9c5a-601e-4580-9fab-903e9b00e35e",
		Name:         "Redtooth Genealogist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Redtooth Genealogist — create a Royal Role token attached to another target creature you control",
					createRoleOnFirstTarget(RoleRoyal)),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
