package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wilt — Instant {1}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Destroy target artifact or enchantment.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Naturalize's target and body (destroyTheTargetPermanent), and cycling
// from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77b6c3bb-bd05-4547-8f3e-9c50879973f1",
		Name:         "Wilt",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve:    destroyTheTargetPermanent,
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
