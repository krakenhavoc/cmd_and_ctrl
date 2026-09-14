package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turbulent Springs — Land — Island Mountain (EDHREC rank 3799):
//
//	"({T}: Add {U} or {R}.)
//	 This land enters tapped unless your opponents control eight or
//	 more lands."
//
// The Izzet member of the "catch-up" typed duals — Turbulent Fen's
// shape exactly: the two basic land types are the mana (the
// shockland's pipe ability, since the engine derives no intrinsic
// ability from a nonbasic's types), and the tapped entry is a CR 614
// self-replacement whose condition adds up every land the
// controller's opponents control. Post-layer types, so an animated
// land still counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "9aef7510-9f06-4939-8cae-f71330d1105e",
		Name:          "Turbulent Springs",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(b34OpponentsControlLandsAtLeast(8))},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "R")},
	})
}
