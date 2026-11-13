package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grid Monitor — Artifact Creature — Construct {4}, 4/6 (Mirrodin,
// #5606):
//
//	"You can't cast creature spells."
//
// Nullhide Ferox's restriction turned around: a cast restriction on the
// Monitor's controller only (YouCantCastUnless with no escape), read
// from the battlefield, so it ends the moment the Monitor leaves and an
// opponent who gains control of it is the one restricted. The "can't"
// wins over any permission to cast (CR 101.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b947b9a7-7a42-447d-b3fd-e20dba8a8f41",
		Name:         "Grid Monitor",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			YouCantCastUnless("You can't cast creature spells.", Creature(), nil),
		},
	})
}
