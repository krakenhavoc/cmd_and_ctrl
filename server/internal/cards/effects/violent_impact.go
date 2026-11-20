package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Violent Impact — Sorcery {3}{R}
// (Amonkhet, #5535):
//
//	"Destroy target artifact or land.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Sunder from Within, plus cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d48d9e93-72a2-48f2-8164-8032615ee217",
		Name:         "Violent Impact",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or land", Or(Artifact(), Land())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    destroyTheTargetPermanent,
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
