package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	prince := effects.WhenThisEnters("Charming Prince — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	prince.Modes = effects.ChooseOne(
		effects.ModeWithPurpose(effects.ModeDoing("Scry 2.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.Scry{Player: item.Controller, N: 2}.Apply(ctx)
		}), game.Purpose{Scry: 2}),
		effects.ModeDoing("You gain 3 life.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
		}),
		effects.ModeDoing("Exile another target creature you own. Return it to the battlefield "+
			"under your control at the beginning of the next end step.",
			effects.Another(effects.TargetCreature("another target creature you own", effects.YouOwn())),
			charmingPrinceExileAndReturn),
	)
	effects.Register(effects.Spec{
		OracleID:     "c48d844c-3976-4fa5-8e0d-3f0e535e7619",
		Name:         "Charming Prince",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{prince},
	})
}

// charmingPrinceExileAndReturn is the third bullet's body: exile the
// chosen creature, then schedule its return at the next end step.
func charmingPrinceExileAndReturn(item *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	exiled := t.ID
	return effects.ExileTarget{Target: exiled, ThenTail: effects.WithCarry(ctx.ItemTail(charmingPrinceExiled), exiled)}.Apply(ctx)
}

// charmingPrinceExiled is the clause after the exile (#4253 slice 9).
var charmingPrinceExiled = effects.CardTail("charming-prince/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	exiled, controller := effects.TailCarry0(o), ctx.Controller()
	ok := len(o.IDs) == 1

	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:      "Charming Prince — return the exiled creature",
		Controller: controller,
		Cards:      []uuid.UUID{exiled},
		Body:       effects.ReturnExiledToOwnersBody,
	}.Apply(ctx)
})
