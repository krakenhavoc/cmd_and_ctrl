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
// ONE SIMPLIFICATION, strictly weaker: "can't be blocked" is not
// granted. Evasion that is not a keyword ability lives in the
// declare-blockers legality check, and the engine has no restriction
// vocabulary there — CanBlock reads the CR 702 evasion keywords
// (flying, and reach against it) and nothing else, and the same gap
// is what keeps Pacifism's "can't attack or block" out of the
// catalog. So the Cloak ships as a shroud-granter that costs three,
// which is a worse card than the printed one and never a better one.
// It becomes whole the day a blocking-restriction hook lands.
func init() {
	Register(Spec{
		OracleID:     "9ad4f730-a18e-4a7c-a468-a926c718c741",
		Name:         "Whispersilk Cloak",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"\"Can't be blocked\" isn't granted — the equipped creature blocks and is blocked normally."},
		Static: []game.StaticAbility{
			GrantToAttached("shroud"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
