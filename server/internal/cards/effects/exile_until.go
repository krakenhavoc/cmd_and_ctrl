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

	// On and Condition are "until <an event>": the events watched and a
	// registered condition over them (one that reads no parameters).
	On        []game.EventKind
	Condition game.ConditionRef

	// Label names the held card on the table ("Ossification — the
	// exiled card returns when it leaves the battlefield").
	Label string

	// Grant, when set, is a cast permission over the exiled card,
	// stamped once it has landed in exile: Hostage Taker's "you may
	// cast that card for as long as it remains exiled". Its Zone is
	// filled in.
	Grant *game.CastPermission

	// ThenTail is the rest of the effect (#4253): a registered tail
	// (CardTail), told the card as TailOutcome.IDs when it was exiled and
	// nothing when it was not — "for each permanent exiled this way"
	// (Abdel Adrian, Gorion's Ward) counts them. It runs after the return
	// is recorded, and not at all when the "until" event had already
	// happened (nothing was exiled this way). On the "until this leaves
	// the battlefield" form it rides the keyed exile as the record's
	// Next, so a prompt the exile pauses on is a restore point.
	ThenTail game.KeyedTail
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
	if e.ThisLeaves && len(e.On) == 0 && e.Condition.Key() == "" {
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
		if !e.ThenTail.IsZero() {
			next := e.ThenTail
			t.Next = &next
		}
		return ExileTarget{Target: target, ThenTail: t}.Apply(ctx)
	}
	// An event-conditioned "until" (Palace Jailer's monarch, #4253): the
	// record's condition key and the events it watches ride the tail's
	// Words, the "leaves" object (when there is one) its Carry and N, so a
	// CR 616 or CR 903.9 prompt the exile pauses on is a restore point.
	t := game.KeyedTail{Ref: exileUntilEventRecorded, Player: u.Controller, Source: u.SourceCardID, Word: e.Label}
	if e.Grant != nil {
		grant := *e.Grant
		t.Permission = &grant
	}
	if record {
		t.Words = append([]string{e.Condition.Key()}, eventKindWords(e.On)...)
		if u.Leaves.ID != uuid.Nil {
			t.Carry, t.N = []uuid.UUID{u.Leaves.ID}, u.Leaves.Epoch
		}
	}
	if !e.ThenTail.IsZero() {
		next := e.ThenTail
		t.Next = &next
	}
	return ExileTarget{Target: target, ThenTail: t}.Apply(ctx)
}

// eventKindWords spells event kinds as a tail's words.
func eventKindWords(kinds []game.EventKind) []string {
	out := make([]string, len(kinds))
	for i, k := range kinds {
		out[i] = string(k)
	}
	return out
}

// exileUntilEventRecorded is the event-conditioned ExileUntil's keyed tail
// (#4253), told the card when it reached exile: the grant (Permission), and
// the return's record "until <Words[1:]> satisfies the condition Words[0]"
// (and, with a Carry, "until <Carry[0] at epoch N> leaves"), labelled with
// Word, for the effect of Player and Source. No Words is no record. Next,
// when set, is the caller's ExileUntil.ThenTail, told the same IDs. An
// on-disk key.
var exileUntilEventRecorded = game.RegisterTail("exile-until/record-event-return", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 1 {
		exileUntilEventRecord(g, o)
	}
	if next := o.Tail.Next; next != nil {
		return next.Run(g, o.IDs)
	}
	return nil
})

// exileUntilEventRecord is exileUntilEventRecorded's grant and record.
func exileUntilEventRecord(g *game.Game, o game.TailOutcome) {
	target := o.IDs[0]
	if p := o.Tail.Permission; p != nil {
		perm := *p
		perm.Zone = game.ZoneExile
		g.GrantCastPermissionOverCardForEffect(target, perm)
	}
	if len(o.Tail.Words) == 0 {
		return
	}
	u := game.UntilReturn{Controller: o.Tail.Player, SourceCardID: o.Tail.Source, Label: o.Tail.Word}
	if key := o.Tail.Words[0]; key != "" {
		cond, ok := game.ConditionRefByKey(key)
		if !ok {
			return
		}
		u.Condition = cond
	}
	for _, k := range o.Tail.Words[1:] {
		u.On = append(u.On, game.EventKind(k))
	}
	if len(o.Tail.Carry) > 0 {
		u.Leaves = game.ObjectRef{ID: o.Tail.Carry[0], Epoch: o.Tail.N}
	}
	g.ScheduleUntilReturnForEffect(target, u)
}

// exileUntilRecorded is ExileUntil's keyed tail (#4253 slice 10a), told
// the card when it reached exile: the grant (the tail's Permission) and the
// return's record, "until <Carry[0] at epoch N> leaves the battlefield",
// labelled with the tail's Word, for the effect of Player and Source. A
// tail with no Carry exiles with no record (an item that names no source
// object). The tail's Next, when set, is the caller's ExileUntil.ThenTail,
// told the same IDs once the record is written. An on-disk key.
var exileUntilRecorded = game.RegisterTail("exile-until/record-return", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 1 {
		exileUntilRecord(g, o)
	}
	if next := o.Tail.Next; next != nil {
		return next.Run(g, o.IDs)
	}
	return nil
})

// exileUntilRecord is exileUntilRecorded's grant and record, over the card
// that reached exile.
func exileUntilRecord(g *game.Game, o game.TailOutcome) {
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
}

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
