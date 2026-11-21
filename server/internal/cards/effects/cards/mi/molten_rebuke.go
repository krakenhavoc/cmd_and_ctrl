package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molten Rebuke — Sorcery {4}{R} (Phyrexia: All Will Be One):
//
//	"Choose one or both —
//	 • Molten Rebuke deals 5 damage to target creature or planeswalker.
//	 • Destroy target Equipment."
//
// #5479, S82. Amazing Acrobatics' "choose one or both" (CR 700.2): each
// bullet reads its own target, in printed order (CR 608.2c), and a target
// that is no longer legal skips only its own bullet (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af05cc39-725b-4535-919a-ddbcb89c2210",
		Name:         "Molten Rebuke",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Molten Rebuke deals 5 damage to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
				})),
			effects.ModeDoing("Destroy target Equipment.",
				effects.TargetPermanent("target Equipment", effects.HasSubtype("Equipment")),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DestroyTarget{Target: id}.Apply(ctx)
				})),
		),
	})
}
