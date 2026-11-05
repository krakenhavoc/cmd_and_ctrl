package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Genesis Effect — Instant {4}{R} (Star Trek, #4973):
//
//	"Genesis Effect deals 6 damage to target creature. Destroy target land
//	 you control. Create a tapped Planet land token named New Planet with
//	 "{T}: Add one mana of any color.""
//
// Two target clauses, each re-checked on its own at resolution (CR
// 608.2b): a creature that left in response is not dealt damage and the
// land is still destroyed, and the other way round. The token is made
// either way, unless every target is illegal and the spell does nothing
// at all. The New Planet matches the printed token (ttrk 27).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec763f36-3504-4bfe-8481-1a7d0bff8d8e",
		Name:         "Genesis Effect",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			TargetPermanent("target land you control", Land(), YouControl()),
		),
		Purpose: ForTargets(DamageToTarget(0, 6)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 6}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return CreateTokenAdvanced{Spec: Token(trekB16NewPlanetToken()).EntersTapped(), N: 1}.Apply(ctx)
		},
	})
}
