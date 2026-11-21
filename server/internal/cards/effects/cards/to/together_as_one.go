package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Together as One — Sorcery {6} (Secrets of Strixhaven, #4456):
//
//	"Converge — Target player draws X cards, Together as One deals X
//	 damage to any target, and you gain X life, where X is the number of
//	 colors of mana spent to cast this spell."
//
// Converge is an ability word (CR 207.2c). X is the number of distinct
// colours that paid for the spell (Context.ColorsSpentCount, Painful
// Truths' read): colourless mana is not a colour, so six generic paid
// with {C} is X = 0 and the spell does nothing. The spell reads its own
// payment (ReadsOwnPayment), so a waived cast asks what was paid outside
// the pool, and the auto-tapper prefers distinct colours.
//
// Two target clauses, a player and any target, which may be the same
// player. The three parts happen in printed order, each waiting for any
// question the one before it asked (a dredge offer during the draw),
// and a part whose target has become illegal is skipped while the
// others still happen (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "234f1c4b-dd0c-4c4e-b3c1-3b9ccfc42226",
		Name:                "Together as One",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        effects.CompletenessFull,
		WantsDistinctColors: true,
		Targets:             effects.Clauses(effects.TargetPlayer("target player"), effects.TargetAny()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return effects.DrawCards{Player: t.ID, N: ctx.ColorsSpentCount()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: ctx.ColorsSpentCount()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Amount: ctx.ColorsSpentCount()}.Apply(ctx)
			}),
		),
	})
}
