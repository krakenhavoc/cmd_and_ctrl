package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1943cbb0-2b12-4450-90af-060b8c8627c3",
		Name:         "Atarka's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Your opponents can't gain life this turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return PlayersCantGainLifeThisTurn{Opponents: true, Label: "Atarka's Command — your opponents can't gain life this turn"}.Apply(ctx)
				}),
			ModeDoing("Atarka's Command deals 3 damage to each opponent.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return damageToEachOpponent(ctx.Game, item, 3)
				}),
			ModeDoing("You may put a land card from your hand onto the battlefield.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return MayPutALandFromHand("Atarka's Command").Apply(ctx)
				}),
			ModeDoing("Creatures you control get +1/+1 and gain reach until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Atarka's Command — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"reach"}, Label: "Atarka's Command — reach"}.Apply(ctx)
				}),
		),
	})
}
