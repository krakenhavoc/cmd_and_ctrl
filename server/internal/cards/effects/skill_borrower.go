package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skill Borrower — Artifact Creature — Human Wizard {2}{U}, 1/3:
//
//	"Play with the top card of your library revealed.
//	 As long as the top card of your library is an artifact or creature
//	 card, this creature has all activated abilities of that card. (If
//	 any of the abilities use that card's name, use this creature's name
//	 instead.)"
//
// Conspicuous Snoop's two halves without the cast permission: the
// REVEALED visibility strength, and a static borrow from the library top
// (ADR 0155 §8) while it is an artifact or creature card, re-read the
// moment the top changes (DependsOnLibraryTop). The reminder text is CR
// 201.5b, which a lent row's label already applies.
//
// No simplification.
func init() {
	const clause = "As long as the top card of your library is an artifact or creature card, this creature has all activated abilities of that card."
	borrow := HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: TopOfLibraryLender(ByController(Or(Artifact(), Creature()))),
	})
	borrow.DependsOnLibraryTop = true
	Register(Spec{
		OracleID:          "9e862e4a-254d-4d39-b2cf-4beed6db761f",
		Name:              "Skill Borrower",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		Static:            []game.StaticAbility{borrow},
	})
}
