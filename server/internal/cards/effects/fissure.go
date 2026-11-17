package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fissure — Instant {3}{R}{R} (The Dark, #5638):
//
//	"Destroy target creature or land. It can't be regenerated."
//
// Terror's body on a creature-or-land clause: "It can't be
// regenerated" rides the destroy and a regeneration shield declines
// it. The clause reads the current types at announce and again at
// resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8b1e9f3-b014-4e57-b278-6d84a7e88b23",
		Name:         "Fissure",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      TargetPermanent("target creature or land", Or(Creature(), Land())),
		OnResolve:    destroyTheTargetPermanentNoRegen,
	})
}
