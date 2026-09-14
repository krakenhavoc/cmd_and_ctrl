package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudshredder Sliver — Creature — Sliver {R}{W}, 1/1 (EDHREC rank
// 3395):
//
//	"Sliver creatures you control have flying and haste."
//
// The first Sliver lord in the catalog. Two Layer 6 keyword grants
// over "Sliver creatures you control" — the printed clause DOES say
// "you control", unlike the old Slivers, so the grant is YoursOnly;
// it does not say "other", so the Cloudshredder grants itself flying
// and haste too, as printed. Effective subtypes, so a changeling you
// control is a Sliver and flies.
//
// No simplification.
func init() {
	slivers := TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "f943c005-9b77-411a-b522-1182e22724e1",
		Name:         "Cloudshredder Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(slivers, "flying"),
			TribalKeywordGrant(slivers, "haste"),
		},
	})
}
