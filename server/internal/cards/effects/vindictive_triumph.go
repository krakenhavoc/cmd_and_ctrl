package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vindictive Triumph — Instant {W}{B}{B} (Reality Fracture, tracker #2795):
//
//	"Exile target creature or planeswalker. If that permanent's mana value
//	 was 3 or less, return it to the battlefield tapped under your control.
//	 Exile it at the beginning of the next end step."
//
// The mana value is read BEFORE the exile (CR 608.2h last known
// information), the return runs as the exile's continuation (so a commander
// that stops to ask about the command zone is not returned early), and it
// comes back as a new object under the caster's control, tapped. The
// delayed exile is scheduled against that NEW object. A token is exiled
// and ceases to exist, so there is nothing to return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6b478d1-5015-49ba-b1aa-cbc2b08107a6",
		Name:         "Vindictive Triumph",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			target := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			mv, known := ctx.Game.ManaValueForEffect(card)
			small := known && mv <= 3
			return ExileTarget{
				Target:   target,
				ThenTail: withN(withCarry(ctx.ItemTail(vindictiveTriumphExiled), target), boolToInt(small)),
			}.Apply(ctx)
		},
	})
}

// vindictiveTriumphExile is "exile it at the beginning of the next end
// step" as a keyed tail (#4253 slice 6): the caster is the tail's Player
// and the spell its Source. An on-disk key.
var vindictiveTriumphExile = EnteredTail("vindictive-triumph/exile-at-end-step", func(ctx *Context, entered uuid.UUID, o game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   o.Tail.Player,
		SourceCardID: o.Tail.Source,
		Label:        "Vindictive Triumph — exile the returned permanent",
		At:           game.StepEnd,
		Cards:        []uuid.UUID{entered},
		Body:         exileListedCardsBody,
	})
	return nil
})

// vindictiveTriumphExiled is the clause after the exile (#4253 slice 9).
var vindictiveTriumphExiled = CardTail("vindictive-triumph/after-exile", func(ctx *Context, o game.TailOutcome) error {
	target, small, caster := tailCarry0(o), o.Tail.N == 1, ctx.Controller()
	exiled := len(o.IDs) == 1

	if !exiled || !small {
		return nil
	}
	return ReturnFromExile{
		Target:     target,
		Controller: caster,
		Tapped:     true,
		ThenTail:   ctx.Tail(vindictiveTriumphExile),
	}.Apply(ctx)
})
