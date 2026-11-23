package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ec763f36-3504-4bfe-8481-1a7d0bff8d8e",
		Name:         "Genesis Effect",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
		),
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 6)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 6}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TrekB16NewPlanetToken()).EntersTapped(), N: 1}.Apply(ctx)
		},
	})
}
