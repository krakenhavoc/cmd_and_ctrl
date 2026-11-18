package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boon of Safety — Instant {W} (Streets of New Capenna):
//
//	"Put a shield counter on target creature. (If it would be dealt
//	 damage or destroyed, remove a shield counter from it instead.)
//	 Scry 1."
//
// The counter protects through the CR 122.1c built-ins
// (game/shield_counters.go). With its only target illegal the spell
// does nothing on resolution, the scry included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34f0beae-a8e2-4474-9f7f-4735d3f26863",
		Name:         "Boon of Safety",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Scry: 1, Answers: game.AnswerSurvive},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := shieldCounterOnFirstTarget(ctx); err != nil {
				return err
			}
			return Scry{N: 1}.Apply(ctx)
		},
	})
}
