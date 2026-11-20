package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tangletrap — Instant {1}{G} (Strixhaven: School of Mages, #5502):
//
//	"Choose one —
//	 • Tangletrap deals 5 damage to target creature with flying.
//	 • Destroy target artifact."
//
// Choose Your Weapon's Archery bullet beside Abrade's artifact bullet,
// each written on its option (ModeDoing) and reading its own target: 5
// damage to a creature that has flying as it is targeted and again as the
// spell resolves, or destroy an artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75effdde-0427-4e5a-bd6d-c3118f301520",
		Name:         "Tangletrap",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Tangletrap deals 5 damage to target creature with flying.",
				TargetCreature("target creature with flying", HasKeyword("flying")),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
				})), ForTargets(DamageToTarget(0, 5))),
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return DestroyTarget{Target: id}.Apply(ctx)
				})),
		),
	})
}
