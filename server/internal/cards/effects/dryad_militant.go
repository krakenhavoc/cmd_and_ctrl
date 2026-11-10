package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dryad Militant — Creature — Dryad Soldier {G/W}, 2/1 (Foundations,
// #4465):
//
//	"If an instant or sorcery card would be put into a graveyard from
//	 anywhere, exile it instead."
//
// Rest in Peace's replacement (GraveyardBecomesExile, CR 614.1a) narrowed
// to instant and sorcery cards (OnlyTypes) of every player, from every zone:
// a spell that finishes resolving or is countered (CR 608.2n), a discard
// (which still happened, so madness works and the card can be found in exile),
// a mill, a search. It leaves tokens and every other card alone, and it only
// works while the Militant is on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8ca5877-ac9e-4b15-8c23-c70f61b01895",
		Name:         "Dryad Militant",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{GraveyardBecomesExile{
			CardsOnly: true,
			OnlyTypes: []string{"instant", "sorcery"},
			Label:     "Dryad Militant: instant and sorcery cards are exiled instead of put into a graveyard",
		}.Build()},
	})
}
