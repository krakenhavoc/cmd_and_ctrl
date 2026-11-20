package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fated Conflagration — Instant {1}{R}{R}{R} (Born of the Gods,
// #5559):
//
//	"Fated Conflagration deals 5 damage to target creature or
//	 planeswalker. If it's your turn, scry 2."
//
// Jaya's Firenado with the scry behind "if it's your turn", read as
// the spell resolves (IsYourTurn, Feed the Cauldron's check).
//
// No simplification.
func init() {
	purpose := ForTargets(DamageToTarget(0, 5))
	purpose.Scry = 2
	Register(Spec{
		OracleID:     "ac0a110a-0721-4aa6-a891-95e3fcc847df",
		Name:         "Fated Conflagration",
		Completeness: CompletenessFull,
		Purpose:      purpose,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			if !IsYourTurn(ctx.Game, item.Controller) {
				return nil
			}
			return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
