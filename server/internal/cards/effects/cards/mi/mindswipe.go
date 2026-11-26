package cardsmi

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindswipe — Instant {X}{U}{R} (Khans of Tarkir):
//
//	"Counter target spell unless its controller pays {X}. Mindswipe deals X
//	 damage to that spell's controller."
//
// Condescend's pay-unless with X announced at the cast, then the damage,
// which waits for the payment prompt (CR 608.2c) and is dealt whether or not
// the spell was countered. The controller is read from the spell's last-known
// information (CR 608.2h) because a countered spell has left the stack by then.
// With the spell's one target illegal the whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4682c39e-c21a-4c8b-bc63-9cbac59dc563",
		Name:         "Mindswipe",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				x := strconv.Itoa(ctx.X())
				return effects.CounterUnlessPaid{
					StackID:  targets[0].ID,
					Cost:     "{" + x + "}",
					Question: "Mindswipe — pay {" + x + "} or your spell is countered",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.X() <= 0 || len(ctx.Item.Targets) == 0 {
					return nil
				}
				controller, ok := ctx.Game.LastKnownSpellControllerForEffect(ctx.Item.Targets[0].ID)
				if !ok || controller == uuid.Nil {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: ctx.X()}.Apply(ctx)
			}),
		),
	})
}
