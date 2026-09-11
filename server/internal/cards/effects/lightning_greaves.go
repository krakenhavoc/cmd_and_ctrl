package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Greaves — Artifact — Equipment for {2} (top-100
// Commander staples rank 13):
//
//	"Equipped creature has haste and shroud.
//	 Equip {0}"
//
// Both halves are real, which was the open question in ADR 0036
// decision 10: haste rides Layer 6 into Effective().Abilities, where
// HasSummoningSickness reads it, and shroud rides the same grant
// into CanBeTargetedBy, the single choke point every targeting path
// goes through — announce, the CR 608.2b re-check, the client's
// legal_targets projection and the bot's move enumerator all at
// once. The shroud gate shipped with the S23 protection-style
// keywords, so this card needed no new engine work at all.
//
// Shroud mattering is the point, not a footnote. It is the
// drawback-adjacent half — it stops YOU targeting your own creature,
// including with your own removal-protection — and a Greaves that
// granted only haste would be strictly STRONGER than printed, which
// inverts this repo's convention that every simplification is
// strictly weaker.
//
// The equip cost really is {0}: Greaves' whole identity is that it
// re-equips for free, so a single copy protects whatever is most
// threatened each turn. The engine's sorcery-speed gate still
// applies — free is not instant-speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "ca204b66-8d0c-431a-8d34-282f7c2d17da",
		Name:     "Lightning Greaves",
		Static: []game.StaticAbility{
			GrantToAttached("haste", "shroud"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{0}"),
		},
	})
}
