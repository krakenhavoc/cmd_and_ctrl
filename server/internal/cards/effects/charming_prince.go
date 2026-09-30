package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charming Prince — Creature — Human Noble {1}{W}, 2/2 (EDHREC rank
// 3005):
//
//	"When this creature enters, choose one —
//	 • Scry 2.
//	 • You gain 3 life.
//	 • Exile another target creature you own. Return it to the
//	   battlefield under your control at the beginning of the next
//	   end step."
//
// A modal ETB (CR 603.3c): the third bullet's target clause rides its
// own ModeOption, and its body exiles the chosen creature, then
// schedules a CR 603.7 delayed trigger — a registered body
// (returnExiledToOwnersBody), not a captured closure, so a table
// holding the wait is still a restore point. Because the clause
// requires a creature the caster OWNS, "under its owner's control"
// and "under your control" name the same player, so the shared flicker
// body is exactly the printed effect.
//
// "Another" on the third bullet is object identity (effects.Another, CR 109.1):
// a second Charming Prince, or a token copy of this one, is a legal target.
func init() {
	prince := WhenThisEnters("Charming Prince — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	prince.Modes = ChooseOne(
		ModeDoing("Scry 2.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return Scry{Player: item.Controller, N: 2}.Apply(ctx)
		}),
		ModeDoing("You gain 3 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
		}),
		ModeDoing("Exile another target creature you own. Return it to the battlefield "+
			"under your control at the beginning of the next end step.",
			Another(TargetCreature("another target creature you own", YouOwn())),
			charmingPrinceExileAndReturn),
	)
	Register(Spec{
		OracleID:     "c48d844c-3976-4fa5-8e0d-3f0e535e7619",
		Name:         "Charming Prince",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{prince},
	})
}

// charmingPrinceExileAndReturn is the third bullet's body: exile the
// chosen creature, then schedule its return at the next end step.
func charmingPrinceExileAndReturn(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	exiled := t.ID
	controller := item.Controller
	return ExileTarget{Target: exiled, Then: func(ctx *Context, ok bool) error {
		if !ok {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label:      "Charming Prince — return the exiled creature",
			Controller: controller,
			Cards:      []uuid.UUID{exiled},
			Body:       returnExiledToOwnersBody,
		}.Apply(ctx)
	}}.Apply(ctx)
}
