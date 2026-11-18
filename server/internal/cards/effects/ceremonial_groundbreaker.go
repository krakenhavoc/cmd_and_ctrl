package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ceremonial Groundbreaker — Artifact — Equipment {1}{G}{W} (Streets of New
// Capenna, #5491):
//
//	"Equipped creature gets +2/+1 and has trample.
//	 Equip Citizen {1}
//	 Equip {3}"
//
// Dúnedain Blade's two equip abilities: the Citizen-only one is listed first
// with its tighter target clause and cheaper cost, the plain {3} one follows,
// and the restriction is checked while activating (CR 702.6c). Trample is a
// grant to the equipped creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bb05573-7e8f-471f-908d-04bccead79e3",
		Name:         "Ceremonial Groundbreaker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
			GrantToAttached("trample"),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip Citizen {1}", "{1}",
				TargetCreature("target Citizen you control", YouControl(), Subtype("Citizen"))),
			EquipAbility("{3}"),
		},
	})
}
