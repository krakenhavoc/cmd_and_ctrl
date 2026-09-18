package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystalline Sliver — Creature — Sliver {W}{U}, 2/2 (EDHREC rank
// 4302):
//
//	"All Slivers have shroud. (They can't be the targets of spells or
//	 abilities.)"
//
// The old-border Sliver, and the one whose wording is the opposite of
// the three modern ones in this batch: "ALL Slivers", with no "you
// control". It really does give shroud to the Slivers across the
// table as well — Crystalline Sliver is a symmetrical effect, and a
// mirror match under it is two boards nobody can target, including
// their own controllers. TribeFilter with neither Others nor
// YoursOnly set is exactly that sentence.
//
// The drawback is the point of the card as much as the protection:
// shroud is enforced at the targeting choke point against everyone,
// so you cannot put an Aura or an Equipment's equip ability on your
// own Slivers either. That is why the Sliver deck's other pump
// effects are statics.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba3aa1eb-722a-47d3-83be-96daddb50265",
		Name:         "Crystalline Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "shroud"),
		},
	})
}
