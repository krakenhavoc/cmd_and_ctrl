package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turbulent Fen — Land — Swamp Forest (EDHREC rank 3589):
//
//	"({T}: Add {B} or {G}.)
//	 This land enters tapped unless your opponents control eight or
//	 more lands."
//
// The Golgari member of the "catch-up" typed duals: the two basic
// land types are the mana (the shockland's pipe ability, since the
// engine derives no intrinsic ability from a nonbasic's types), and
// the tapped entry is a CR 614 self-replacement whose condition adds
// up every land the controller's opponents control — at a four-player
// table that is usually true from turn three on, which is the whole
// design. Post-layer types, so an animated land still counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "114dd40d-5ad8-4913-a08f-572b9521eb5b",
		Name:          "Turbulent Fen",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(b34OpponentsControlLandsAtLeast(8))},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "G")},
	})
}
