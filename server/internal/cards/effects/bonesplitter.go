package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonesplitter — Artifact — Equipment for {1}:
//
//	"Equipped creature gets +2/+0.
//	 Equip {1}"
//
// The whole of Equipment in five lines, and the reason it is in this
// batch: it is the shape every other Equipment is a variation on, so
// if the relation, the equip ability and the layer-7c static are
// wired correctly, this card is complete with nothing else to say
// about it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "452e3f5f-ce17-4682-966b-5cc100210aee",
		Name:         "Bonesplitter",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
