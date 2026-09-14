package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Behemoth Sledge — Artifact — Equipment for {1}{G}{W} (EDHREC rank
// 3687):
//
//	"Equipped creature gets +2/+2 and has trample and lifelink.
//	 Equip {3}"
//
// Loxodon Warhammer's coloured cousin: a smaller bonus for a much
// smaller cast cost, on exactly the same two statics. It is in the
// batch because a Selesnya Voltron deck wants both — the Sledge on
// turn three and the Hammer later — and because two cards sharing a
// shape is how you find out the shape composes.
//
// Both keywords are honoured. Trample turns the +2/+2 into damage a
// chump block does not absorb; lifelink turns it into life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00573e77-8ff6-4acb-8683-8827d965288f",
		Name:         "Behemoth Sledge",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("trample", "lifelink"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
