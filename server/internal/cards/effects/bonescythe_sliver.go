package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonescythe Sliver — Creature — Sliver {3}{W}, 2/2 (EDHREC rank
// 4317):
//
//	"Sliver creatures you control have double strike."
//
// The Sliver that turns a board into a lethal one. Note the wording
// against Squirrel Sovereign's: there is no "other", so the Bonescythe
// grants double strike to ITSELF as well, and there is a "you
// control", so the modern printing does not hand the keyword to the
// Sliver deck sitting across the table. Both halves are the printed
// text and both are one field on TribeFilter.
//
// Double strike is a canonical engine keyword read by the combat
// damage steps, so the grant is real: a board of Slivers under this
// deals its damage twice, first-strike step and regular.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d95ad32-768f-40a1-a461-e0562326b2b7",
		Name:         "Bonescythe Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "double strike"),
		},
	})
}
