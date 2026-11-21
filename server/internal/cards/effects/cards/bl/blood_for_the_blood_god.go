package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood for the Blood God! — Instant {8}{B}{B}{R} (Warhammer 40,000
// Commander):
//
//	"This spell costs {1} less to cast for each creature that died this
//	 turn.
//	 Discard your hand, then draw eight cards. Blood for the Blood God!
//	 deals 8 damage to each opponent. Exile Blood for the Blood God!"
//
// The reduction reads the table-wide tally of creatures that died this
// turn (B11CreaturesDiedThisTurn) at CR 601.2f and spends against generic
// mana only, which the printed {8} always covers. The three steps run in
// printed order, each waiting for a prompt the one before asks (a madness
// offer on a discarded card, CR 608.2c), and the spell exiles itself last
// as Rise of the Eldrazi does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6079dff-9479-4665-aba5-2cddefb9a2bc",
		Name:         "Blood for the Blood God!",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 8, DamageEachOpponent: 8},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.CreaturesThatDiedThisTurnAmount,
				"This spell costs {1} less to cast for each creature that died this turn."),
		},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: 8}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DamageEachOpponent{N: 8}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
