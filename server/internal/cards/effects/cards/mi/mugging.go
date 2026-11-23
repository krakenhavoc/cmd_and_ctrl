package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mugging — Sorcery {R} (Ravnica Remastered, #5948):
//
//	"Mugging deals 2 damage to target creature. That creature can't block
//	 this turn."
//
// Damage first, then the restriction on the same creature (CR 608.2c). A
// target that has left is skipped (CR 608.2b); one that survives the damage
// cannot block until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e4f08c2-1bad-4a85-8aad-9bb34a5579b1",
		Name:         "Mugging",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ids := effects.LegalTargetCards(item, ctx.Game)
			if len(ids) == 0 {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: ids[0], Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.RestrictUntilEOT{Target: ids[0], Restrictions: game.CantBlock, Label: "Mugging — can't block"}.Apply(ctx)
		},
	})
}
