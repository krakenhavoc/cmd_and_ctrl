package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swiftfoot Boots — Artifact — Equipment for {2} (top-100 Commander
// staples rank 12, the single highest-ranked Equipment in the
// format):
//
//	"Equipped creature has hexproof and haste.
//	 Equip {1}"
//
// The Greaves' more expensive and more usable twin: hexproof instead
// of shroud means the creature is protected from your opponents and
// still targetable by YOU, so the Boots can carry a commander that
// you intend to keep pumping. Both keywords are honoured — hexproof
// is enforced in CanBeTargetedBy against the caster's identity
// (CR 702.11b: opponents only), so the asymmetry that makes this
// card better than Lightning Greaves is a real asymmetry here too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "c8b143ad-43ec-4e0d-a440-e348daa31391",
		Name:     "Swiftfoot Boots",
		Static: []game.StaticAbility{
			GrantToAttached("hexproof", "haste"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
