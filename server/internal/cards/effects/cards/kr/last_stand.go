package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Last Stand — Sorcery {W}{U}{B}{R}{G} (Planechase 2012, S82 OLD, #5569):
//
//	"Target opponent loses 2 life for each Swamp you control. Last Stand deals
//	 damage to target creature equal to the number of Mountains you control.
//	 Create a 1/1 green Saproling creature token for each Forest you control.
//	 You gain 2 life for each Plains you control. Draw a card for each Island
//	 you control, then discard that many cards."
//
// Two target clauses (an opponent, a creature) and five instructions in printed
// order (CR 608.2c); each count is of basic land types you control, read as that
// instruction runs (CR 608.2h) through the card's text (CR 612.2), so a Dryad of
// the Ilysian Grove's lands count for all five. A clause whose target is gone is
// skipped and the rest still happen (CR 608.2b). The discard waits for the draw
// and is as many as were counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d2a465e-9ebd-4002-b6cd-e0eab08bad54",
		Name:         "Last Stand",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()).Then(effects.TargetCreature("target creature")),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				n := 2 * effects.S82OldB431CountLands(ctx, "Swamp")
				if n <= 0 {
					return nil
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: effects.S82OldB431CountLands(ctx, "Mountain")}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				n := effects.S82OldB431CountLands(ctx, "Forest")
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Amount: 2 * effects.S82OldB431CountLands(ctx, "Plains")}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				n := effects.S82OldB431CountLands(ctx, "Island")
				if n <= 0 {
					return nil
				}
				return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), n, n, "")
			}),
		),
	})
}
