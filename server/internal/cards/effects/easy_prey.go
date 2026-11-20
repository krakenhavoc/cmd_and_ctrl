package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Easy Prey — Instant {1}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Destroy target creature with mana value 2 or less.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The mana value is read off the target as it is announced and again as
// the spell resolves (CR 608.2b); the destroy is destroyTheTargetPermanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b783246a-e00b-4cdf-9111-fe0afb896a8e",
		Name:         "Easy Prey",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with mana value 2 or less", ManaValueLE(2)),
		OnResolve:    destroyTheTargetPermanent,
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
