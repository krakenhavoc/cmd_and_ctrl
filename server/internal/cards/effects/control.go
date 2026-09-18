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
	if c.Target == uuid.Nil {
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
