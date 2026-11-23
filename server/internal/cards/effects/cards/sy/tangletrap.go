package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "75effdde-0427-4e5a-bd6d-c3118f301520",
		Name:         "Tangletrap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Tangletrap deals 5 damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
				})), effects.ForTargets(effects.DamageToTarget(0, 5))),
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DestroyTarget{Target: id}.Apply(ctx)
				})),
		),
	})
}
