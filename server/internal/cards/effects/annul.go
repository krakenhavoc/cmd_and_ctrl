package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Annul — Instant {U} (Edge of Eternities, #4461):
//
//	"Counter target artifact or enchantment spell."
//
// Negate's body with a narrower clause: the picker offers only artifact
// and enchantment spells, the engine refuses anything else at announce
// (CR 601.2c), and the target is checked again as Annul resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d08e9784-75f7-4164-ac48-d06160f8c56b",
		Name:         "Annul",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell (#4734)
		Targets:      TargetSpell("target artifact or enchantment spell", Or(Artifact(), Enchantment())),
		OnResolve:    counterTheTargetSpell,
	})
}
