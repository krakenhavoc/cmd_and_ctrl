package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// control.go — gain and exchange of control from a spell or an
// ability (CR 613.1b, CR 701.12; ADR 0063, #756). The durations these
// take are built in durations.go.

// GainControl is "gain control of target permanent", for any
// duration (CR 613.1b).
//
// It is a layer-2 continuous effect, not a write to Card.Controller,
// and that distinction is the whole primitive:
//
//   - Control REVERTS by itself when the effect ends, because the
//     layer engine reseeds every permanent's controller from
//     Card.BaseController on each pass. Nothing remembers who had it.
//   - Two control-changers on one permanent sort by timestamp and the
//     later one wins (CR 613.7) — an Act of Treason cast on a
//     Mind Controlled creature takes it, and at cleanup hands it back
//     to the Mind Control's controller, not to its owner.
//   - The change is visible to combat, targeting, activated
//     abilities and the wire, because the recompute materialises
//     layer 2's answer onto Card.Controller before anything reads it.
//
// The two CR consequences ride with that materialisation: the
// permanent leaves combat (CR 506.4) and is summoning-sick under its
// new controller (CR 302.6). A "gain control and attack with it" card
// therefore has to grant haste, exactly as it prints.
type GainControl struct {
	// Target is the permanent to take. The effect is pinned to it as
	// the object it is now, so one flickered in response is not
	// taken (CR 400.7).
	Target uuid.UUID

	// Controller is the player who gains control. Defaults to the
	// effect's own controller; set it for "target opponent gains
	// control of ~".
	Controller uuid.UUID

	// Duration is how long the theft lasts. The zero value is until
	// end of turn.
	Duration game.Duration

	Label string
}

func (c GainControl) Apply(ctx *Context) error {
	if c.Target == uuid.Nil || ctx.isNewSourceObject(c.Target) { // #1432
		return nil
	}
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	ctx.Game.GainControlForEffect(ctx.Source(), c.Target, controller, c.Duration,
		eotLabel(c.Label, "gain control"))
	return nil
}

// ExchangeControl is "exchange control of two target permanents"
// (CR 701.12) — Switcheroo.
//
// All or nothing (CR 701.12b): if either permanent has left the
// battlefield by the time the spell resolves, no control is
// exchanged at all. The two halves share one CR 613.7 timestamp, so
// a later control-changer beats both or neither.
//
// Neither half has a stated duration, so both are CR 611.2a
// indefinite.
type ExchangeControl struct {
	A, B  uuid.UUID
	Label string
}

func (e ExchangeControl) Apply(ctx *Context) error {
	ctx.Game.ExchangeControlForEffect(ctx.Source(), e.A, e.B,
		eotLabel(e.Label, "exchange control"))
	return nil
}

// GainControlOfSpell is "gain control of target spell" (ADR 0104,
// #1745; CR 611.1, CR 613.1b) — Aethersnatch, Commandeer, Invert
// Polarity's won flip.
//
// It is the same layer-2 effect GainControl is, pinned to a spell on
// the stack instead of a permanent: the spell's "you" becomes the new
// controller at once (the stack step of the layer pass materialises it
// before Apply returns), and a permanent spell becomes a permanent
// under them whose default controller is still the caster (CR 110.2b,
// CR 400.7a), so it goes home if the thief leaves the game.
//
// With ChooseNewTargets the new controller is then offered "you may
// choose new targets for it" (CR 115.7d) — AFTER the control change,
// because which targets are legal is judged for the spell's controller
// and it has to be theirs by then. A spell that is no longer on the
// stack, or one the player already controls, leaves the card doing
// nothing.
type GainControlOfSpell struct {
	// Spell is the spell to take — the stack item's ID, which is the
	// spell card's instance ID.
	Spell uuid.UUID

	// Controller is who gains control. Defaults to the effect's own
	// controller.
	Controller uuid.UUID

	// ChooseNewTargets offers the new controller "you may choose new
	// targets for it".
	ChooseNewTargets bool

	Label string
}

func (c GainControlOfSpell) Apply(ctx *Context) error {
	if c.Spell == uuid.Nil {
		return nil
	}
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	if !ctx.Game.GainControlOfSpellForEffect(ctx.Source(), c.Spell, controller,
		eotLabel(c.Label, "gain control of a spell")) {
		return nil
	}
	if !c.ChooseNewTargets {
		return nil
	}
	return ChangeTargets{
		StackID:  c.Spell,
		Chooser:  controller,
		Policy:   game.RetargetChooseNew,
		Optional: true,
	}.Apply(ctx)
}

// ExchangeControlOfSpellAnd is "exchange control of <spell> and
// <permanent>" (ADR 0104; CR 701.12) — Perplexing Chimera's "exchange
// control of this creature and that spell", Sudden Substitution's
// "exchange control of target noncreature spell and target creature".
//
// All or nothing (CR 701.12a): if either object is gone, nothing is
// exchanged. If one player controls both, it does nothing (CR 701.12b).
// Both halves are indefinite and share one CR 613.7 timestamp.
//
// With ChooseNewTargets "the spell's controller may choose new
// targets for it" — the spell's controller AFTER the exchange.
type ExchangeControlOfSpellAnd struct {
	Spell     uuid.UUID
	Permanent uuid.UUID

	ChooseNewTargets bool

	Label string
}

func (e ExchangeControlOfSpellAnd) Apply(ctx *Context) error {
	_, err := e.ApplyAndReport(ctx)
	return err
}

// ApplyAndReport is Apply that also says whether control was
// exchanged — Perplexing Chimera's "if you do".
func (e ExchangeControlOfSpellAnd) ApplyAndReport(ctx *Context) (bool, error) {
	if e.Spell == uuid.Nil || e.Permanent == uuid.Nil {
		return false, nil
	}
	if !ctx.Game.ExchangeControlOfSpellAndPermanentForEffect(ctx.Source(), e.Spell, e.Permanent,
		eotLabel(e.Label, "exchange control")) {
		return false, nil
	}
	if !e.ChooseNewTargets {
		return true, nil
	}
	controller, ok := ctx.Game.SpellControllerForEffect(e.Spell)
	if !ok {
		return true, nil
	}
	return true, ChangeTargets{
		StackID:  e.Spell,
		Chooser:  controller,
		Policy:   game.RetargetChooseNew,
		Optional: true,
	}.Apply(ctx)
}
