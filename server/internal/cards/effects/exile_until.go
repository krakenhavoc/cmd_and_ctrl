package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_until.go — "exile <it> until <an event>" (CR 610.3, #1729).
//
// The catalog half of game/until_return.go. One primitive for every
// printed "until": the card names the event, and the engine owns the
// rest — when the return happens (immediately after the event, never on
// the stack), what returns (the object it exiled, if it is still that
// object in exile), and under whose control (its owner's, CR 610.3c).
//
// Two shapes of event, and every "until" card in the catalog is one of
// them:
//
//   - THIS LEAVES THE BATTLEFIELD (Ossification, Hostage Taker,
//     Sheltered by Ghosts): ThisLeaves. The object is the resolving
//     ability's own source as it was when the ability triggered, so a
//     source that left and came back is a new object and the exile
//     never starts (CR 610.3b).
//   - AN EVENT (Palace Jailer's "until an opponent becomes the
//     monarch"): On and a registered Condition, the vocabulary an
//     event-conditioned delayed trigger uses.

// ExileUntil exiles Target until the event it names, and records the
// return. Optional fields are noted; the zero value of each does
// nothing.
type ExileUntil struct {
	Target uuid.UUID

	// ThisLeaves is "until this permanent leaves the battlefield".
	ThisLeaves bool

	// On, Condition and CondParams are "until <an event>".
	On         []game.EventKind
	Condition  game.ConditionRef
	CondParams game.EffectParams

	// Label names the held card on the table ("Ossification — the
	// exiled card returns when it leaves the battlefield").
	Label string

	// Grant, when set, is a cast permission over the exiled card,
	// stamped once it has landed in exile: Hostage Taker's "you may
	// cast that card for as long as it remains exiled". Its Zone is
	// filled in.
	Grant *game.CastPermission

	// Then, when set, runs once the exile has settled with whether the
	// card was exiled: "for each permanent exiled this way" (Abdel
	// Adrian, Gorion's Ward) counts the trues. It runs after the return
	// is recorded, and not at all when the "until" event had already
	// happened (nothing was exiled this way).
	Then func(ctx *Context, exiled bool) error
}

// Apply is CR 610.3a/b first — an "until" whose event has already
// happened since the ability triggered moves nothing — and then the
// exile, from whose continuation the return is recorded, because a
// commander can stop on the CR 903.9 prompt and one that took the
// command zone has nothing to come back from.
//
// A leaves-the-battlefield "until" on an item that names no source
// object (one restored from a file written before #1418 stamped it)
// is exiled without a record. The card's legacy leave trigger
// (UntilThisLeavesLegacyReturn) brings it back, exactly as before.
func (e ExileUntil) Apply(ctx *Context) error {
	if e.Target == uuid.Nil {
		return nil
	}
	u := game.UntilReturn{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        e.Label,
		On:           e.On,
		Condition:    e.Condition,
		CondParams:   e.CondParams,
	}
	record := true
	if e.ThisLeaves {
		ref, ok := ctx.SourceRef()
		if ok {
			u.Leaves = ref
		} else {
			record = false
		}
	}
	if record && ctx.Game.UntilEventHappenedForEffect(u, ctx.Item) {
		return nil
	}
	target := e.Target
	if e.ThisLeaves && e.Then == nil && len(e.On) == 0 && e.Condition.Key() == "" {
		// #4253 slice 10a: "until this leaves the battlefield" with no
		// clause of its own (the Oblivion Ring family, Hostage Taker) is
		// a keyed tail: the record and the grant ride it as data.
		t := game.KeyedTail{Ref: exileUntilRecorded, Player: u.Controller, Source: u.SourceCardID, Word: e.Label}
		if e.Grant != nil {
			grant := *e.Grant
			t.Permission = &grant
		}
		if record {
			t.Carry, t.N = []uuid.UUID{u.Leaves.ID}, u.Leaves.Epoch
		}
		return ExileTarget{Target: target, ThenTail: t}.Apply(ctx)
	}
	then := e.Then
	return exileTargetThen(ctx, target, func(ctx *Context, exiled bool) error {
		if !exiled {
			if then != nil {
				return then(ctx, false)
			}
			return nil
		}
		if e.Grant != nil {
			perm := *e.Grant
			perm.Zone = game.ZoneExile
			ctx.Game.GrantCastPermissionOverCardForEffect(target, perm)
		}
		if record {
			ctx.Game.ScheduleUntilReturnForEffect(target, u)
		}
		if then != nil {
			return then(ctx, true)
		}
		return nil
	})
}

// exileUntilRecorded is ExileUntil's keyed tail (#4253 slice 10a), told
// the card when it reached exile: the grant (the tail's Permission) and the
// return's record, "until <Carry[0] at epoch N> leaves the battlefield",
// labelled with the tail's Word, for the effect of Player and Source. A
// tail with no Carry exiles with no record (an item that names no source
// object). An on-disk key.
var exileUntilRecorded = game.RegisterTail("exile-until/record-return", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	target := o.IDs[0]
	if p := o.Tail.Permission; p != nil {
		perm := *p
		perm.Zone = game.ZoneExile
		g.GrantCastPermissionOverCardForEffect(target, perm)
	}
	if len(o.Tail.Carry) > 0 {
		g.ScheduleUntilReturnForEffect(target, game.UntilReturn{
			Controller:   o.Tail.Player,
			SourceCardID: o.Tail.Source,
			Label:        o.Tail.Word,
			Leaves:       game.ObjectRef{ID: o.Tail.Carry[0], Epoch: o.Tail.N},
		})
	}
	return nil
})

// exileChosenTargetUntilThisLeaves is the Oblivion Ring family's entry
// trigger body: exile the card the pick stamped into the item, if it is
// still a legal target (CR 608.2b), until this permanent leaves the
// battlefield. `label` is the held card's line on the table.
func exileChosenTargetUntilThisLeaves(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return ExileUntil{Target: t.ID, ThisLeaves: true, Label: label}.Apply(ctx)
		}
		return nil
	}
}

// UntilThisLeavesLegacyReturn is the leaves-the-battlefield trigger an
// "until this leaves the battlefield" card carried before #1729, kept
// for one job: a card a table exiled with it under an older binary has
// no "until" record, and a restore point holding one must still bring
// it back. It fires only when such a card exists — one exiled with this
// permanent (the exile trigger's label, exileLabel) with no record
// pending — so a card exiled today returns the CR 610.3 way and this
// never reaches the stack. The row keeps its label, so a restored card
// does not count it as an ability lost (AGENTS.md §5).
func UntilThisLeavesLegacyReturn(returnLabel, exileLabel string) game.TriggeredAbility {
	t := On(game.EventLTB, Self, returnLabel, b41ReturnCardsExiledWithToTheBattlefield(exileLabel))
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		if !Self(ev, source, ch, g) {
			return false
		}
		for _, id := range b27ExiledWith(g, source.InstanceID, exileLabel) {
			if !g.UntilReturnPendingForEffect(id) {
				return true
			}
		}
		return false
	}
	return t
}
