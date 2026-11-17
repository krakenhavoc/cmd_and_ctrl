package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shadow Rift — Instant {U} (Tempest, #5628):
//
//	"Target creature gains shadow until end of turn. (It can block or be
//	 blocked by only creatures with shadow.)
//	 Draw a card."
//
// Leap with shadow in place of flying: a layer-6 grant on the target
// (CR 613.1f) that ends in the cleanup step (CR 514.2), then the draw
// (CR 608.2c). Any creature may be the target, an opponent's included,
// which stops it blocking your creatures without shadow. A target gone
// in response means the spell doesn't resolve and nothing is drawn
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66810ef7-ea3a-4afa-92f0-b21ec52cc8d0",
		Name:         "Shadow Rift",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Draws:   1,
			Answers: game.AnswerCombatGrant,
			Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "shadow")),
		},
		Targets: TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"shadow"}, Label: "Shadow Rift — shadow until end of turn"}.Apply(ctx)
			}),
			DrawCards{N: 1},
		),
	})
}
