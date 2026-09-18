package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turbulent Moor — Land — Plains Swamp (EDHREC rank 3954):
//
//	"({T}: Add {W} or {B}.)
//	 This land enters tapped unless your opponents control eight or
//	 more lands."
//
// The Orzhov member of the Foundations "catch-up" typed duals, and
// the reason this batch has it is that the cycle only proves itself
// once more than one row is in the catalog: Turbulent Fen (Golgari)
// and Turbulent Springs (Simic) are already registered, and a
// transposed colour pair in a table-driven cycle is invisible until
// somebody plays that exact land.
//
// The two basic land types are the mana. They are printed in
// reminder-text parentheses because the ability is intrinsic to the
// types (CR 305.6), but the engine derives no intrinsic mana ability
// from a NONBASIC land's subtypes, so the pipe ability is declared
// the way every other typed dual in the catalog declares it. The
// printed width is offered whole — this is not a "in your
// commander's colour identity" land, so nothing is narrowed.
//
// The tapped entry is a CR 614 self-replacement, not a trigger: it
// checks, as the land enters, how many lands the controller's
// opponents control between them. Post-layer types, so a land one of
// them animated still counts, and an opponent who has been eliminated
// takes their lands out of the sum with them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "2eb4da30-2600-4a7f-8e6c-6a090faa9a8d",
		Name:          "Turbulent Moor",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(b34OpponentsControlLandsAtLeast(8))},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "B")},
	})
}
