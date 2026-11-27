package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invoke the Firemind — Sorcery {X}{U}{U}{R} (Duel Decks: Izzet vs. Golgari):
//
//	"Choose one —
//	 • Draw X cards.
//	 • Invoke the Firemind deals X damage to any target."
//
// Fascination's X-scaled modes: X is the announced value (CR 107.3), read as
// each bullet resolves. The damage bullet has its own target, chosen only when
// that bullet is (CR 601.2c), and is dealt only to a target still legal as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2037659f-2efe-4321-afaf-961bbec35e9e",
		Name:         "Invoke the Firemind",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Draw X cards.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DrawCards{N: ctx.X()}.Apply(ctx)
				}),
			effects.ModeDoing("Invoke the Firemind deals X damage to any target.", effects.TargetAny(),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: ctx.X()}.Apply(ctx)
				})),
		),
	})
}
