package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baron, Airship Kingdom — Land — Town (EDHREC rank 4532):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {R}."
//
// The plainest card in the batch: an Izzet tapland with no rider at
// all — no life, no ping, no basic types. It gets played because
// three-colour and five-colour Commander decks run out of duals long
// before they run out of slots, and a tapland that costs nothing
// is still a dual.
//
// Registered for the reason every nonbasic land is: the engine's
// synthetic mana ability only fires for lands with the BASIC
// supertype, so an unregistered Baron taps for nothing at all.
// Enters-tapped is the self-replacement (it is a replacement effect,
// not a trigger — nobody gets a window to respond to it, CR 614.1c);
// the mana is the pipe every two-colour land in the catalog uses,
// which opens the same colour picker a Birds of Paradise activation
// would, narrowed to the two printed colours.
//
// Town is a land subtype, not a mechanic — nothing in the engine has
// to know it, and the type line carries it from Scryfall for whatever
// eventually cares.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "cc710da0-5a2e-4bc4-8fdd-d90e7bc1f224",
		Name:          "Baron, Airship Kingdom",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "R")},
	})
}
