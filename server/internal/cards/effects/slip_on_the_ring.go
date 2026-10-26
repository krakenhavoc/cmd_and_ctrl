package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slip On the Ring — Instant {1}{W}:
//
//	"Exile target creature you own, then return it to the battlefield
//	 under your control. The Ring tempts you."
//
// A creature you own but don't control may be the target (2023-06-16
// ruling). It returns as a new object (CR 400.7), and the tempt runs
// once it has: it may be chosen as the Ring-bearer. A token that is
// exiled does not return (CR 111.8), and the Ring still tempts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6b28298-9624-49d7-ad73-fcd1fd3556d2",
		Name:         "Slip On the Ring",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you own", YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := legalTargetIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			target := ids[0]
			return ExileTarget{
				Target:   target,
				ThenTail: withCarry(ctx.ItemTail(slipOnTheRingExiled), target),
			}.Apply(ctx)
		},
	})
}

// slipOnTheRingExiled is the clause after the exile (#4253 slice 9).
var slipOnTheRingExiled = CardTail("slip-on-the-ring/after-exile", func(ctx *Context, o game.TailOutcome) error {
	target, tempt := tailCarry0(o), ringTemptsYouNext(ctx.Item)
	exiled := len(o.IDs) == 1

	if !exiled {
		return tempt(ctx.Game)
	}
	return ReturnFromExile{
		Target:     target,
		Controller: ctx.Controller(),
		ThenTail:   ctx.ItemTail(ringTemptsYouTail),
	}.Apply(ctx)
})
