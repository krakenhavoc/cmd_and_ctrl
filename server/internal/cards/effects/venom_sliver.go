package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venom Sliver — Creature — Sliver {1}{G}, 1/1 (EDHREC rank 4346):
//
//	"Sliver creatures you control have deathtouch. (Any amount of
//	 damage a creature with deathtouch deals to a creature is enough
//	 to destroy it.)"
//
// The third of the batch's modern "you control" Slivers. Deathtouch
// on a whole board is what makes attacking into a Sliver deck a bad
// idea even when the Slivers are 1/1s, and it is the keyword whose
// consumer is furthest from the grant: the combat damage step's
// lethal check, not the declaration gates the other two ride.
//
// No "other", so the Venom Sliver has deathtouch itself; "you
// control", so an opposing Sliver deck gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffe05d4b-0ec9-4319-bb11-1366dc091224",
		Name:         "Venom Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "deathtouch"),
		},
	})
}
