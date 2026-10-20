package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Repeated Reverberation — Instant {2}{R}{R}:
//
//	"When you next cast an instant spell, cast a sorcery spell, or
//	 activate a loyalty ability this turn, copy that spell or ability
//	 twice. You may choose new targets for the copies."
//
// One CR 603.7 delayed trigger (Doublecast's shape, #663) watching two
// event kinds: a cast of an instant or sorcery by you, or your
// activation of a loyalty ability (an ability with a loyalty symbol in
// its cost, CR 606.2, printed or granted). Whichever happens first fires
// it and uses it up (CR 603.7b); a cast or activation that does not
// match leaves it waiting, and it ends with the turn (CR 514.2).
//
// The body reads the triggering event off the item. A spell is copied
// twice from last-known information if it has already left the stack
// ("copy that spell" does not target, CR 608.2h); an ability is copied
// twice while it is still on the stack (CR 707.10) and not at all if it
// has left. Copies are created, not cast or activated, so they neither
// fire this again nor count toward CR 606.3's loyalty limit.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "9a664735-6e38-407f-9d2f-a6d279cd6309",
		Name:         "Repeated Reverberation",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DelayedOnEvent{
				Label:     "Repeated Reverberation — copy that spell or ability twice",
				On:        []game.EventKind{game.EventCast, game.EventActivateAbility},
				Condition: repeatedReverberationCondition,
				Body:      repeatedReverberationBody,
			}.Apply(ctx)
		},
	})
}

var (
	// "When you next cast an instant spell, cast a sorcery spell, or
	// activate a loyalty ability this turn". On-disk identity.
	repeatedReverberationCondition = game.DelayedCondition("repeated-reverberation/next-instant-sorcery-or-loyalty",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			if ev.Actor != dt.Controller {
				return false
			}
			switch ev.Kind {
			case game.EventCast:
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery())
			case game.EventActivateAbility:
				return ev.Loyalty
			}
			return false
		})

	// "Copy that spell or ability twice. You may choose new targets for
	// the copies." On-disk identity.
	repeatedReverberationBody = game.SimpleDelayedBody("repeated-reverberation/copy-twice", repeatedReverberationCopy)
)

func repeatedReverberationCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ev := ctx.Trigger().Event
	switch ev.Kind {
	case game.EventCast:
		if ev.CardID == uuid.Nil {
			return nil
		}
		return CopySpell{
			StackID:          ev.CardID,
			Controller:       item.Controller,
			Count:            2,
			ChooseNewTargets: true,
			FromLastKnown:    true,
		}.Apply(ctx)
	case game.EventActivateAbility:
		if ev.StackItemID == uuid.Nil {
			return nil
		}
		return CopyAbility{
			ItemID:           ev.StackItemID,
			Controller:       item.Controller,
			Count:            2,
			ChooseNewTargets: true,
		}.Apply(ctx)
	}
	return nil
}
