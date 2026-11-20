package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atarka's Command — Instant {R}{G}:
//
//	"Choose two —
//	 • Your opponents can't gain life this turn.
//	 • Atarka's Command deals 3 damage to each opponent.
//	 • You may put a land card from your hand onto the battlefield.
//	 • Creatures you control get +1/+1 and gain reach until end of
//	   turn."
//
// The first bullet is ADR 0107 §5's "this turn" grant (#1880): a rule,
// read at every life gain until cleanup (CR 514.2), covering the
// caster's opponents only. Chosen with the damage, it is applied first
// (CR 608.2c, printed order), which is what stops a lifelink blocker or
// a Soul Warden trigger answering the damage.
//
// The pump is locked to the creatures you control as it resolves
// (CR 611.2c); the land drop is the shared optional put from hand and
// does not use a land play (CR 305.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1943cbb0-2b12-4450-90af-060b8c8627c3",
		Name:         "Atarka's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Your opponents can't gain life this turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.PlayersCantGainLifeThisTurn{Opponents: true, Label: "Atarka's Command — your opponents can't gain life this turn"}.Apply(ctx)
				}),
			effects.ModeDoing("Atarka's Command deals 3 damage to each opponent.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageToEachOpponent(ctx.Game, item, 3)
				}),
			effects.ModeDoing("You may put a land card from your hand onto the battlefield.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.MayPutALandFromHand("Atarka's Command").Apply(ctx)
				}),
			effects.ModeDoing("Creatures you control get +1/+1 and gain reach until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Atarka's Command — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"reach"}, Label: "Atarka's Command — reach"}.Apply(ctx)
				}),
		),
	})
}
