package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_dread.go — the catalog's words for manifest dread
// (CR 701.62a, ADR 0082's 2026-10-07 amendment, #2570).
//
//	ManifestDread{}                                  "Manifest dread."
//	ManifestDread{Then: PutCountersOnManifested(…)}  "…, then put two
//	                                                  +1/+1 counters on that creature."
//	ManifestDreadTimes{N: 2}                         "Manifest dread twice."
//	ManifestDreadTimes{N: x, Then: …}                "Manifest dread X times, then …"
//
// The look, the controller-only prompt and the two moves are
// game.ManifestDreadThenForEffect; this file is the vocabulary a card
// file writes in and the continuations the printed "then" clauses
// share.

// ManifestDread is one manifest dread by `Player`, or the resolving
// item's controller when that is zero.
//
// `Then` is the printed text after "then", told what happened. It runs
// exactly once, after the last move settles — which can be after the
// controller's answer, so it must capture only scalars and use the
// *game.Game it is handed. An empty library still runs it, with a zero
// Manifested; every ready-made continuation below checks for that.
type ManifestDread struct {
	Player uuid.UUID
	Then   func(g *game.Game, res game.ManifestDreadResult) error
}

func (m ManifestDread) Apply(ctx *Context) error {
	player := m.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.ManifestDreadThenForEffect(player, ctx.Source(), m.Then)
}

// ManifestDreadTimes is "manifest dread N times" (Valgavoth's Onslaught)
// and "manifest dread twice" (They Came from the Pipes): the action
// taken N times in a row, each with its own look, its own prompt and
// its own entry window. A later look sees the library the earlier one
// left behind, as the rules say.
//
// `Then` is told every result, one per repetition, in order. It runs
// once, after the last repetition — N of zero or less runs it at once
// with an empty list.
type ManifestDreadTimes struct {
	Player uuid.UUID
	N      int
	Then   func(g *game.Game, results []game.ManifestDreadResult) error
}

func (m ManifestDreadTimes) Apply(ctx *Context) error {
	player := m.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return manifestDreadRepeat(ctx.Game, player, ctx.Source(), m.N, nil, m.Then)
}

// manifestDreadRepeat takes the action `left` more times, carrying the
// results so far by value (the property that makes an undo across a
// prompt replay identically).
func manifestDreadRepeat(g *game.Game, player, source uuid.UUID, left int,
	done []game.ManifestDreadResult, then func(*game.Game, []game.ManifestDreadResult) error) error {
	if left <= 0 {
		if then == nil {
			return nil
		}
		return then(g, done)
	}
	return g.ManifestDreadThenForEffect(player, source, func(g *game.Game, res game.ManifestDreadResult) error {
		next := append(append([]game.ManifestDreadResult(nil), done...), res)
		return manifestDreadRepeat(g, player, source, left-1, next, then)
	})
}

// PutCountersOnManifested is "…, then put <n> <kind> counters on that
// creature", one entry per kind. Nothing happens when nothing was
// manifested (an empty library, or a replacement that sent the card
// elsewhere), and a manifested card that has left the battlefield by
// the time the counters would go on is skipped.
func PutCountersOnManifested(counters ...CounterAmount) func(*game.Game, game.ManifestDreadResult) error {
	return func(g *game.Game, res game.ManifestDreadResult) error {
		if res.Manifested == uuid.Nil || !b15OnBattlefield(g, res.Manifested) {
			return nil
		}
		for _, c := range counters {
			if err := g.AddCounterForEffect(res.Manifested, c.Kind, c.N); err != nil {
				return err
			}
		}
		return nil
	}
}

// putCountersOnEachManifested is "…, then put X <kind> counters on each
// of those creatures" for ManifestDreadTimes.
func putCountersOnEachManifested(kind string, n int) func(*game.Game, []game.ManifestDreadResult) error {
	return func(g *game.Game, results []game.ManifestDreadResult) error {
		if n <= 0 {
			return nil
		}
		for _, res := range results {
			if res.Manifested == uuid.Nil || !b15OnBattlefield(g, res.Manifested) {
				continue
			}
			if err := g.AddCounterForEffect(res.Manifested, kind, n); err != nil {
				return err
			}
		}
		return nil
	}
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
	return g.ManifestDreadThenForEffect(item.Controller, item.SourceCardID, nil)
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
// ability is (AttachSourceForEffect).
func AttachSourceToManifested(ctx *Context) func(*game.Game, game.ManifestDreadResult) error {
	item := ctx.Item
	if ctx.Game.AbilitySourceGoneForEffect(item) {
		return func(*game.Game, game.ManifestDreadResult) error { return nil }
	}
	source := item.SourceCardID
	return func(g *game.Game, res game.ManifestDreadResult) error {
		if res.Manifested == uuid.Nil {
			return nil
		}
		return g.AttachForEffect(source, game.TargetRef{Kind: game.TargetCard, ID: res.Manifested})
	}
}

// WheneverYouManifestDread matches EventManifestDread for the source's
// controller. `ev.CardID` is the manifested permanent and `ev.Target`
// the card put into the graveyard "this way".
func WheneverYouManifestDread(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventManifestDread && ev.Actor == source.Controller
}
