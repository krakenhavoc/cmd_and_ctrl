package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// World Champion, Celestial Weapon — the back face of Sidequest: Play
// Blitzball, Legendary Artifact — Equipment (Final Fantasy, #4462):
//
//	"Double Overdrive — Equipped creature gets +2/+0 and has double strike.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// Dragoon's Lance's shape: a layer 7c pump and a layer 6 keyword grant over
// the attached creature, and the ordinary equip (CR 702.6a). Double
// Overdrive is an ability word. Registered under "<oracle_id>#1".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     sidequestPlayBlitzballOracleID + "#1",
		Name:         "World Champion, Celestial Weapon",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("double strike"),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
