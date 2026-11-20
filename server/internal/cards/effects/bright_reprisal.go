package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bright Reprisal — Instant {4}{W} (Ixalan, #5532):
//
//	"Destroy target attacking creature.
//	 Draw a card."
//
// Kill Shot's destroy, then a draw (CR 608.2c). A target that is no
// longer attacking is illegal and the whole spell is countered
// (CR 608.2b), so nothing is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1de2c216-0921-45eb-9613-a68d18fbee77",
		Name:         "Bright Reprisal",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return DestroyTarget{Target: id}.Apply(ctx) }),
			DrawCards{N: 1},
		),
	})
}
