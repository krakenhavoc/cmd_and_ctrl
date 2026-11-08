package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seize Opportunity — Instant {2}{R} (Tarkir: Dragonstorm, #4463):
//
//	"Choose one —
//	 • Exile the top two cards of your library. Until the end of your
//	   next turn, you may play those cards.
//	 • Up to two target creatures each get +2/+1 until end of turn."
//
// The first bullet is Reckless Impulse's body
// (b19ExileTopTwoUntilEndOfNextTurn): the permission ends with your next
// turn whoever's turn it is now. The second is Unbounded Potential's "up
// to two target creatures" bullet with a +2/+1 boost in place of the
// counter: none, one or two creatures, and only those still legal as it
// resolves get the boost (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0458ffe-9a61-443f-8e80-3d2dde224a4d",
		Name:         "Seize Opportunity",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Exile the top two cards of your library. Until the end of your next turn, you may play those cards.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return b19ExileTopTwoUntilEndOfNextTurn(ctx.Game, item)
				}),
			ModeWithPurpose(ModeDoing("Up to two target creatures each get +2/+1 until end of turn.",
				TargetCreature("up to two target creatures").WithCount(0, 2),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1,
							Label: "Seize Opportunity — +2/+1"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}), ForTargets(BenefitUntilEOT(0, 2, 1))),
		),
	})
}
