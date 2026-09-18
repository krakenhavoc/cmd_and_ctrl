package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whispersilk Cloak — Artifact — Equipment for {3} (EDHREC rank 331):
//
//	"Equipped creature can't be blocked and has shroud.
//	 Equip {2}"
//
// The commander-damage Equipment: unblockable carries the twenty-one,
// shroud keeps the carrier off the end of a removal spell. Shroud is
// honoured — the S23 targeting gate in CanBeTargetedBy — and it is
// the half that is a real drawback as well as a real protection,
// since it locks YOU out of targeting your own creature too. That
// asymmetry is why the Cloak and the Boots are different cards.
//
// "Can't be blocked" shipped one sprint late. It was held back
// because evasion that is not a keyword ability had nowhere to live
// — CanBlock read the CR 702 evasion keywords and nothing else — and
// the S24 restriction vocabulary is the hook that comment was
// waiting for. It is a restriction on the DEFENDER's options
// (CR 509.1b) carried on the attacker, which is why it is read
// inside the block-pair check (Game.BlockPairRefusalLocked) rather
// than anywhere on the attacking side.
//
// The Cloak is now complete, and the two halves are the same card
// again: unblockable carries the commander damage, shroud keeps the
// carrier off the end of a removal spell — including your own, which
// is the drawback that makes the Cloak and the Boots different
// cards.
func init() {
	Register(Spec{
		OracleID:     "9ad4f730-a18e-4a7c-a468-a926c718c741",
		Name:         "Whispersilk Cloak",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("shroud"),
			RestrictAttached(game.CantBeBlocked),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
