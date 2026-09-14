package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Warhammer — Artifact — Equipment for {3} (EDHREC rank 963):
//
//	"Equipped creature gets +3/+0 and has trample and lifelink.
//	 Equip {3}"
//
// Bonesplitter's big brother, and the card that shows the two
// attachment statics compose: a layer 7c modify and a layer 6 grant
// on the same permanent, scoped by the same relation, moving together
// when the Hammer is re-equipped.
//
// Both keywords are honoured, and the pair is the reason the card is
// worth the six mana it really costs: trample turns the +3/+0 into
// damage that gets through a chump block, and lifelink turns that
// damage into the life total that wins a long game. Neither is a
// badge — trample is read by the combat damage assignment and
// lifelink by the damage application.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dba35ac5-7ad3-488a-a006-6b9a1d54eea5",
		Name:         "Loxodon Warhammer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 0),
			GrantToAttached("trample", "lifelink"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
