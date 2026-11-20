package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wring Flesh — Instant {B} (Magic 2012, #5573):
//
//	"Target creature gets -3/-1 until end of turn."
//
// Darkblast's body at -3/-1 (boostTheTargetUntilEOT).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45bb536c-c173-4b6e-a5a8-354c74fd93a1",
		Name:         "Wring Flesh",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(HarmUntilEOT(0, -3, -1)), Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(-3, -1, "Wring Flesh — -3/-1"),
	})
}
