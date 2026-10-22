package effects

import "github.com/google/uuid"

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
		// Each kicker's effect waits for a prompt the damage before it asks (a CR 616 order;
		// CR 608.2c, #4155). The creature's controller is read before the damage and filed
		// for the player damage behind it.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				creature, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				target, ok := ctx.Game.LookupCardForEffect(creature)
				if !ok {
					return nil
				}
				ctx.RecordStep(illuminateRecord, []uuid.UUID{target.Controller})
				return DealDamage{Source: ctx.Item.SourceCardID, Target: creature, Amount: ctx.X()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				ids := ctx.StepRecord(illuminateRecord).IDs
				if len(ids) == 0 || !ctx.KickedWith("{2}{R}") {
					return nil
				}
				return DealDamage{Source: ctx.Item.SourceCardID, Target: ids[0], Amount: ctx.X()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if len(ctx.StepRecord(illuminateRecord).IDs) == 0 || !ctx.KickedWith("{3}{U}") {
					return nil
				}
				return DrawCards{N: ctx.X()}.Apply(ctx)
			}),
		),
	})
}

// illuminateRecord files the controller of the creature the damage was dealt to.
const illuminateRecord = "illuminate/controller"
