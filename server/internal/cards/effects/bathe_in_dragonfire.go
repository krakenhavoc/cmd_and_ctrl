package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bathe in Dragonfire — Sorcery {2}{R} (Foundations Jumpstart, #5929):
//
//	"Bathe in Dragonfire deals 4 damage to target creature."
//
// Bombard's body at sorcery speed: 4 damage to one creature, skipped if the
// target is no longer legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72d793e8-1d19-4ca1-894c-8851a2ebecde",
		Name:         "Bathe in Dragonfire",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(DamageToTarget(0, 4)),
		},
		OnResolve: damageToFirstTarget(4),
	})
}
