package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glider Staff — Artifact — Equipment {2}{W}:
//
//	"When this Equipment enters, airbend up to one target creature.
//	 Equipped creature gets +1/+1 and has flying.
//	 Equip {2}"
//
// Airbend is the shared keyword action (airbend.go): the creature is
// exiled and its OWNER may cast it for {2} while it stays exiled.
// "Up to one" may answer zero, and a target that became illegal is
// left alone (CR 608.2b). No simplification.
func init() {
	Register(Spec{
		OracleID:     "823006a8-098c-457a-8546-77ae2457c459",
		Name:         "Glider Staff",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Glider Staff — airbend up to one target creature", airbendFirstLegalTarget),
				TargetCreature("up to one target creature").WithCount(0, 1),
			),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
