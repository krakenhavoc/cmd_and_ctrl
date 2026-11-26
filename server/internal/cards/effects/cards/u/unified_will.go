package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unified Will — Instant {1}{U} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Counter target spell if you control more creatures than that spell's
//	 controller."
//
// The comparison is made as Unified Will resolves, over the creatures each
// player controls then (the layered view): the target is legal whatever the
// counts (CR 608.2b), and the spell is countered only when the caster's
// number is strictly greater than its controller's. Equal counts counter
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f90cc072-750c-4a87-bf6c-37192d15e691",
		Name:         "Unified Will",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				owner, ok := ctx.Game.SpellControllerForEffect(t.ID)
				if !ok {
					return nil
				}
				if effects.CreaturesControlledBy(ctx.Game, ctx.Controller()) <= effects.CreaturesControlledBy(ctx.Game, owner) {
					return nil
				}
				return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
