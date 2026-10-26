package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_dread.go — the catalog's words for manifest dread
// (CR 701.62a, ADR 0082's 2026-10-07 amendment, #2570).
//
//	ManifestDread{}                                       "Manifest dread."
//	ManifestDread{ThenTail: PutCountersOnManifested(…)}   "…, then put two
//	                                                       +1/+1 counters on that creature."
//	ManifestDreadTimes{N: 2}                              "Manifest dread twice."
//
// The look, the controller-only prompt and the two moves are
// game.ManifestDreadKeyedForEffect; this file is the vocabulary a card
// file writes in and the continuations the printed "then" clauses
// share. Since #4253 slice 9 the "then" is a registered tail
// (ManifestDreadTail), so a table waiting on the pick, the face-down
// entry or the graveyard card is a restore point.

// ManifestDread is one manifest dread by `Player`, or the resolving
// item's controller when that is zero.
//
// `ThenTail` is the printed text after "then", a registered tail
// (ManifestDreadTail) told what happened. It runs exactly once, after the
// last move settles. An empty library still runs it, with a zero
// Manifested; every ready-made continuation below checks for that.
type ManifestDread struct {
	Player   uuid.UUID
	ThenTail game.KeyedTail
}

func (m ManifestDread) Apply(ctx *Context) error {
	player := m.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.ManifestDreadKeyedForEffect(player, ctx.Source(), m.ThenTail)
}

// ManifestDreadTimes is "manifest dread N times" and "manifest dread
// twice" (They Came from the Pipes): the action taken N times in a row,
// each with its own look, its own prompt and its own entry window. A
// later look sees the library the earlier one left behind, as the rules
// say. Each repetition starts the next from its tail (manifest-dread/again).
type ManifestDreadTimes struct {
	Player uuid.UUID
	N      int
}

func (m ManifestDreadTimes) Apply(ctx *Context) error {
	if m.N <= 0 {
		return nil
	}
	player := m.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source := ctx.Source()
	return ctx.Game.ManifestDreadKeyedForEffect(player, source,
		game.KeyedTail{Ref: manifestDreadAgain, Player: player, Source: source, N: m.N - 1})
}

// PutCountersOnManifested is "…, then put <counters> on that creature"
// (Moldering Gym, Experimental Lab): the counters ride the tail as its
// word. Nothing happens when nothing was manifested or the creature has
// already left.
func PutCountersOnManifested(counters ...CounterAmount) game.KeyedTail {
	return game.KeyedTail{Ref: manifestDreadPutCounters, Word: encodeCounterAmounts(counters)}
}

// ManifestDreadWhenItDiesThisTurn schedules "When <id> dies this turn,
// manifest dread" (Turn Inside Out): an event-delayed trigger pinned to
// the object until cleanup, so a flicker ends it (CR 400.7), controlled
// by the resolving object's controller. The same trigger shape as
// ExileWhenItDiesThisTurn.
func ManifestDreadWhenItDiesThisTurn(ctx *Context, id uuid.UUID, label string) {
	d := ctx.Game.PinnedTo(ctx.Game.UntilEndOfTurnDuration(), id)
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        label,
		On:           []game.EventKind{game.EventLTB},
		Condition:    theListedObjectDiedCondition,
		Cards:        []uuid.UUID{id},
		Duration:     &d,
		Body:         manifestDreadAfterItDiedBody,
	})
}

// manifestDreadBody is the delayed trigger's body: the trigger's
// controller manifests dread.
func manifestDreadBody(g *game.Game, item *game.StackItem) error {
	return g.ManifestDreadKeyedForEffect(item.Controller, item.SourceCardID, game.KeyedTail{})
}

// CounterAmount is one kind of counter and how many.
type CounterAmount struct {
	Kind string
	N    int
}

// AttachSourceToManifested is "…, then attach this Equipment to that
// creature" for a resolving ability whose source is `source`. The
// source is read once, when the ability resolves — a prompt cannot let
// it change, because nobody gets priority in between — and the attach
// is skipped quietly if the Equipment is no longer the permanent that
// asked (CR 400.7, CR 701.3b), the way every attach-this-permanent
// ability is (AttachSourceForEffect): the tail is then empty.
func AttachSourceToManifested(ctx *Context) game.KeyedTail {
	item := ctx.Item
	if ctx.Game.AbilitySourceGoneForEffect(item) {
		return game.KeyedTail{}
	}
	return game.KeyedTail{Ref: manifestDreadAttachSource, Player: ctx.Controller(), Source: item.SourceCardID}
}

// WheneverYouManifestDread matches EventManifestDread for the source's
// controller. `ev.CardID` is the manifested permanent and `ev.Target`
// the card put into the graveyard "this way".
func WheneverYouManifestDread(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventManifestDread && ev.Actor == source.Controller
}
