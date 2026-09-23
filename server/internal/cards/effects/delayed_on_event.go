package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delayed_on_event.go — the catalog side of #663: a CR 603.7 delayed
// triggered ability whose condition is an EVENT rather than a step.
//
//	"When you next cast an instant or sorcery spell this turn, copy
//	 that spell. You may choose new targets for the copy."
//	 — Doublecast, Galvanic Iteration, Teach by Example
//
// ScheduleDelayedTrigger (primitives.go) is the step-conditioned
// sibling and stays exactly as it was: "at the beginning of the next
// end step, return that card". This one waits for a thing to happen
// instead of for a step to begin, fires on the FIRST match and then
// ceases to exist (CR 603.7b), and ends with the turn whether or not
// the thing ever happened (CR 514.2). See the 2026-09-18 amendment to
// ADR 0026, which reverses §1-2 for this case and says why.

// DelayedOnEvent is the general constructor: a delayed trigger that
// watches `On`, fires when `Matches` says the event is the one the
// card names, and goes on the stack as `Label` with `Effect`.
//
// It is applied from inside a resolving effect, exactly as
// ScheduleDelayedTrigger is — the trigger is created BY the
// resolution, which is also what makes "next" free: the EventCast of
// the spell that created it was emitted before that resolution began,
// so a Doublecast can never copy itself.
type DelayedOnEvent struct {
	// Label is the stack-overlay copy for the item the trigger puts
	// on the stack, in the shape every trigger uses: "<card> — <what
	// happens>".
	Label string

	// On is the event kinds the trigger watches. Required — with none
	// the trigger is malformed and is dropped rather than queued.
	On []game.EventKind

	// Matches narrows On to the event the card names. Nil matches
	// every event of a watched kind. Read the controller off the
	// DelayedTrigger it is handed; it must capture no *Game and no
	// pointer into a zone slice, for the reason Effect must not.
	Matches func(ev game.Event, dt *game.DelayedTrigger, g *game.Game) bool

	// Controller is who controls the delayed ability. Zero means the
	// controller of the effect scheduling it (CR 603.7d).
	Controller uuid.UUID

	// Cards is the instance IDs the effect acts on, captured now.
	// The TRIGGERING event's object rides separately, as the item's
	// payload — see ctx.PayloadCards().
	Cards []uuid.UUID

	// Optional is the CR 603.5 "you may" on the fired trigger. Nil
	// for the mandatory case, which is every card on this seam.
	Optional *game.TriggerOptionalPrompt

	// Effect runs when the trigger's stack item resolves. Same
	// contract as a triggered ability's: read everything off `item`
	// and the `g` handed in, and declare it as a package-level func
	// so it captures nothing.
	Effect func(g *game.Game, item *game.StackItem) error
}

func (d DelayedOnEvent) Apply(ctx *Context) error {
	controller := d.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: ctx.Source(),
		Label:        d.Label,
		On:           d.On,
		AppliesTo:    d.Matches,
		Optional:     d.Optional,
		Cards:        d.Cards,
		Effect:       d.Effect,
	})
	return nil
}

// WhenYouNextCast is the printed shape: "When you next cast a <spell>
// spell this turn, <do something to that spell>". `spell` narrows
// which cast counts — `Or(Instant(), Sorcery())` for the copy family,
// `Creature()` for the "it enters with additional counters" family —
// and nil counts every spell.
//
// The effect reads the spell that was cast off the item's payload
// (`ctx.PayloadCards()[0]`), because the object is known only when
// the trigger fires, not when it was scheduled.
func WhenYouNextCast(label string, spell CardPredicate, effect Effect) DelayedOnEvent {
	return DelayedOnEvent{
		Label:   label,
		On:      []game.EventKind{game.EventCast},
		Matches: YouNextCast(spell),
		Effect:  effect,
	}
}

// YouNextCast is the condition WhenYouNextCast is built from: an
// EventCast whose actor is the trigger's controller and whose spell
// matches `spell`.
//
// "You" is the delayed trigger's controller (CR 603.7d), read off the
// trigger rather than off a source card — the card that created it is
// in a graveyard by now, and may have changed hands or been exiled.
func YouNextCast(spell CardPredicate) func(ev game.Event, dt *game.DelayedTrigger, g *game.Game) bool {
	return func(ev game.Event, dt *game.DelayedTrigger, g *game.Game) bool {
		if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (spell == nil || spell(g, dt.Controller, c))
	}
}

// copyTheSpellYouJustCast is the whole effect of the "when you next
// cast … copy that spell" family. "That spell" is the one the
// triggering event named, which rides on the item as its payload — so
// this is a package-level func capturing nothing, and Doublecast and
// Galvanic Iteration are one line each.
//
// A spell that has already left the stack by the time the trigger
// resolves — countered in response — is still copied, from last-known
// information: "copy that spell" names the spell and does not target
// it, so CR 608.2h governs rather than CR 608.2b (#1255).
func copyTheSpellYouJustCast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cast := ctx.PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	return CopySpell{
		StackID:          cast[0],
		Controller:       item.Controller,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(ctx)
}
