package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildfire Howl — Sorcery {1}{R}{R} (Bloomburrow, #4467):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Wildfire Howl deals 2 damage to each creature. If the gift was
//	 promised, instead Wildfire Howl deals 1 damage to any target and 2
//	 damage to each creature."
//
// The gift is the shared seam (gift.go): the opponent draws before the
// rest of the spell. The "any target" exists only when the gift was
// promised, so it is the gift's own clause (GiftACard().Instead), chosen
// as the spell is cast and re-checked as it resolves. Then the spell's
// damage is one damage instance: 1 to that target, if it is still legal,
// and 2 to each creature on the battlefield as it resolves, so a creature
// that is the target takes 3 at once. A promised spell whose target has
// become illegal does not resolve at all, as for any spell whose every
// target is gone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04fd8c1f-80e5-4b3d-b63e-7af6ccafbbaa",
		Name:         "Wildfire Howl",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		Gift:         GiftACard().Instead(TargetAny()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.DamageInstanceForEffect(func() error {
				if ctx.GiftPromised() {
					for _, t := range ctx.LegalTargets() {
						if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return damageEachMatching(ctx, Creature(), 2)
			})
		},
	})
}
