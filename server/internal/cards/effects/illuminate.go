package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illuminate — Sorcery {X}{R}:
//
//	"Kicker {2}{R} and/or {3}{U} (You may pay an additional {2}{R}
//	 and/or {3}{U} as you cast this spell.)
//	 Illuminate deals X damage to target creature. If this spell was
//	 kicked with its {2}{R} kicker, it deals X damage to that
//	 creature's controller. If this spell was kicked with its {3}{U}
//	 kicker, you draw X cards."
//
// Two kicker costs (CR 702.33b, #2153), each linked to one clause of the
// spell (CR 702.33f), read with ctx.KickedWith. The creature's
// controller is read BEFORE the damage, since a creature this kills is
// in the graveyard by the time the second clause runs; if the creature
// has left by resolution the spell has no legal target and does nothing
// (CR 608.2b), so neither kicker clause happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "f944c245-08ff-41a8-bdfc-7d21a6a63eae",
		Name:          "Illuminate",
		Completeness:  CompletenessFull,
		XMatters:      true,
		OptionalCosts: Kickers("{2}{R}", "{3}{U}"),
		Targets:       TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			creature, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			target, ok := ctx.Game.LookupCardForEffect(creature)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: item.SourceCardID, Target: creature, Amount: x}).Apply(ctx); err != nil {
				return err
			}
			if ctx.KickedWith("{2}{R}") {
				if err := (DealDamage{Source: item.SourceCardID, Target: target.Controller, Amount: x}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.KickedWith("{3}{U}") {
				return DrawCards{N: x}.Apply(ctx)
			}
			return nil
		},
	})
}
