package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razaketh's Rite — Sorcery {3}{B}{B} (Hour of Devastation, #5534):
//
//	"Search your library for a card, put that card into your hand, then
//	 shuffle.
//	 Cycling {B} ({B}, Discard this card: Draw a card.)"
//
// Demonic Tutor's search (the shared demonicTutorOnResolve; the card is
// not revealed), and the shared Cycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ab0be8e-d6be-411b-827f-584bb5d03766",
		Name:         "Razaketh's Rite",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve:    demonicTutorOnResolve,
		Activated:    []ActivatedAbility{Cycling("{B}")},
	})
}
