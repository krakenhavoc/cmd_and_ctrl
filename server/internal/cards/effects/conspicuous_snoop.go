package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conspicuous Snoop — Creature — Goblin Rogue {R}{R}, 2/2:
//
//	"Play with the top card of your library revealed.
//	 You may cast Goblin spells from the top of your library.
//	 As long as the top card of your library is a Goblin card, this
//	 creature has all activated abilities of that card."
//
// The first two halves are Garruk's Horde's shapes: the REVEALED
// visibility strength (every seat sees the top card) and a cast-only
// top-of-library permission narrowed to Goblin cards, at their printed
// cost and timing. A changeling card is a Goblin card (CR 702.73a).
//
// The third is a static borrow from the library top (ADR 0155 §8):
// TopOfLibraryLender reads the top card while it is a Goblin card and
// lends its printed abilities, and DependsOnLibraryTop re-reads it the
// moment the top changes — a draw, a cast from the top, a scry, a
// shuffle. "This creature" in a lent ability is the Snoop (CR 201.5b).
//
// No simplification.
func init() {
	const clause = "As long as the top card of your library is a Goblin card, this creature has all activated abilities of that card."
	borrow := HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: TopOfLibraryLender(ByController(Subtype("Goblin"))),
	})
	borrow.DependsOnLibraryTop = true
	Register(Spec{
		OracleID:          "f5d1bd3c-0e65-4999-a810-881b2389b40e",
		Name:              "Conspicuous Snoop",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{NonLandOnly: true, CreatureType: "Goblin"},
				"Cast a Goblin spell from the top of your library (Conspicuous Snoop)"),
		},
		Static: []game.StaticAbility{borrow},
	})
}
