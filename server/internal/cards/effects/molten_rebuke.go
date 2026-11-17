package effects

import "github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "af05cc39-725b-4535-919a-ddbcb89c2210",
		Name:         "Molten Rebuke",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Molten Rebuke deals 5 damage to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return DealDamage{Source: ctx.Source(), Target: id, Amount: 5}.Apply(ctx)
				})),
			ModeDoing("Destroy target Equipment.",
				TargetPermanent("target Equipment", HasSubtype("Equipment")),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return DestroyTarget{Target: id}.Apply(ctx)
				})),
		),
	})
}
