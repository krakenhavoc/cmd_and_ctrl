package game

import (
	"sync"

	"github.com/google/uuid"
)

// discard_keyed.go — #4253 slice 9: a discard batch with the rest of the
// batch as data.
//
// discardBatchLocked discards one card at a time, each from the previous
// leg's route continuation, so a leg paused on a CR 616 prompt (madness's
// "you may exile it instead", Library of Leng's "you may put it on top",
// an ordering prompt between them) holds the rest of the discard behind
// it. The keyed form carries the same state the closure captured as data
// on the leg's route (zoneRoute.discardLegs): the cards still to discard,
// the ones discarded this way, the legs a replacement put on top of the
// library with the object epoch each landed with, the event-log mark the
// leg began at, and the caller's continuation as a registered tail.
//
// A discard is its own replacement event kind (RepEventDiscard, #650), so
// its frame is keyed under its own resume key
// (ReplacementResumeDiscard); before this slice no discard frame was.

// discardBatchRun is the rest of a keyed discard batch, riding the route
// of the leg now moving (Leg). Never written after it is built, so an undo
// snapshot may share it.
type discardBatchRun struct {
	Player uuid.UUID `json:"player"`
	// Source is the card that caused the discard (discardOptions.source),
	// uuid.Nil for none.
	Source uuid.UUID    `json:"source"`
	Cause  DiscardCause `json:"cause,omitempty"`
	Leg    uuid.UUID    `json:"leg"`
	// Mark is the event-log position before Leg's move
	// (discardLegPutOnTopLocked).
	Mark   uint64               `json:"mark,omitempty"`
	Rest   []uuid.UUID          `json:"rest,omitempty"`
	Landed []uuid.UUID          `json:"landed,omitempty"`
	Tops   []discardTopSnapshot `json:"tops,omitempty"`
	// Then is the caller's continuation, told the cards discarded this
	// way (discardedThisWayLocked), in batch order, as TailOutcome.IDs.
	Then *keyedTailSnapshot `json:"then,omitempty"`
}

// discardTopSnapshot is a discardTopPlacement on disk.
type discardTopSnapshot struct {
	ID    uuid.UUID `json:"id"`
	Epoch int       `json:"epoch,omitempty"`
}

func cloneDiscardBatchRun(r *discardBatchRun) *discardBatchRun {
	if r == nil {
		return nil
	}
	out := *r
	out.Rest = copyUUIDs(r.Rest)
	out.Landed = copyUUIDs(r.Landed)
	out.Tops = append([]discardTopSnapshot(nil), r.Tops...)
	if r.Then != nil {
		out.Then = snapshotKeyedTail(restoreKeyedTail(r.Then))
	}
	return &out
}

// discardCardsKeyedLocked is discardCardsLocked with the continuation as
// data: `then` is told the cards discarded this way. A zero tail is "no
// continuation".
//
// Caller must hold g.mu.
func (g *Game) discardCardsKeyedLocked(playerID uuid.UUID, cards []uuid.UUID, cause DiscardCause, source uuid.UUID, then KeyedTail) error {
	return g.discardBatchKeyedLocked(&discardBatchRun{
		Player: playerID, Source: source, Cause: cause, Rest: copyUUIDs(cards), Then: snapshotKeyedTail(then),
	})
}

// discardBatchKeyedLocked discards the head of run.Rest with the rest of
// the batch riding its route, or finishes the batch: discardBatchLocked's
// keyed twin, which it must not drift from.
//
// Caller must hold g.mu.
func (g *Game) discardBatchKeyedLocked(run *discardBatchRun) error {
	p := g.playerByIDLocked(run.Player)
	cards := run.Rest
	for len(cards) > 0 {
		next, rest := cards[0], cards[1:]
		// #2178: an opponent's spell or ability can't make this player
		// discard.
		if (run.Cause == DiscardCauseEffect || run.Cause == "") && g.effectDiscardBlockedLocked(run.Player) {
			break
		}
		if p == nil || !p.Hand.Contains(next) {
			cards = rest
			continue
		}
		leg := cloneDiscardBatchRun(run)
		leg.Leg, leg.Rest, leg.Mark = next, copyUUIDs(rest), g.eventSeq
		_, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:        next,
			Dst:           ZoneGraveyard,
			DstOwner:      run.Player,
			Actor:         run.Player,
			Discard:       true,
			DiscardCause:  run.Cause,
			Source:        run.Source,
			Cause:         discardMoveCause(run.Cause, run.Player),
			MustSettleNow: run.Cause == DiscardCauseCost,
			discardLegs:   leg,
		})
		return err
	}
	then := restoreKeyedTail(run.Then)
	// ADR 0088 amendment 2026-10-14 (#2949), #3541: the cards a
	// replacement put on top of the library are ordered by their owner,
	// as discardBatchLocked does; the rest of the effect waits for it.
	if top, displaced := g.discardBatchTopPlacementsLocked(run.Player, discardTopsOf(run.Tops)); len(top) >= 2 || (len(top) == 1 && displaced) {
		var after KeyedTail
		if !then.IsZero() {
			after = KeyedTail{Ref: discardOrderPlaced, Carry: copyUUIDs(run.Landed), Next: &then}
		}
		return g.PutInLibraryInChosenOrderThenForEffect(PutInLibrarySpec{
			Chooser:   run.Player,
			Source:    run.Source,
			Cards:     top,
			From:      ZoneLibrary,
			Placement: LibraryPlaceTop,
			Reason:    "Put the discarded cards on top of your library in any order",
			ThenTail:  after,
		})
	}
	return g.runKeyedTailLocked(then, run.Landed, 0)
}

// discardOrderPlaced is a keyed discard batch's continuation after its
// owner has ordered the cards a replacement put on top of the library: the
// caller's tail (Next), told the cards discarded this way (Carry).
var discardOrderPlaced = RegisterTail("engine/discard-order-placed", func(g *Game, o TailOutcome) error {
	if o.Tail.Next == nil {
		return nil
	}
	return g.runKeyedTailLocked(*o.Tail.Next, o.Tail.Carry, 0)
})

// runDiscardBatchLegLocked counts the leg that just settled and discards
// the next card: the keyed form of discardBatchLocked's leg closure.
//
// Caller must hold g.mu.
func (g *Game) runDiscardBatchLegLocked(run *discardBatchRun) error {
	next := cloneDiscardBatchRun(run)
	next.Leg, next.Mark = uuid.Nil, 0
	if g.discardedThisWayLocked(run.Player, run.Leg) {
		next.Landed = append(next.Landed, run.Leg)
	}
	if g.discardLegPutOnTopLocked(run.Player, run.Leg, run.Mark) {
		next.Tops = append(next.Tops, discardTopSnapshot{ID: run.Leg, Epoch: g.cardObjectEpochLocked(run.Leg)})
	}
	return g.discardBatchKeyedLocked(next)
}

func discardTopsOf(s []discardTopSnapshot) []discardTopPlacement {
	out := make([]discardTopPlacement, 0, len(s))
	for _, t := range s {
		out = append(out, discardTopPlacement{id: t.ID, epoch: t.Epoch})
	}
	return out
}

// unknownDiscardBatchKeys is checkEffectKeys' half for a keyed discard.
func unknownDiscardBatchKeys(r *discardBatchRun) []string {
	if r == nil {
		return nil
	}
	return unknownKeyedTail(r.Then)
}

// DiscardRandomKeyedForEffect is DiscardRandomThenForEffect with the
// continuation as data (#4253 slice 9): `then` is told the cards really
// discarded, and runs even when nothing was.
//
// Caller must hold g.mu.
func (g *Game) DiscardRandomKeyedForEffect(playerID uuid.UUID, n int, then KeyedTail) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if n <= 0 || p.Hand.Size() == 0 || g.effectDiscardBlockedLocked(playerID) {
		return then.Run(g, nil)
	}
	ids := make([]uuid.UUID, len(p.Hand.Cards))
	for i, c := range p.Hand.Cards {
		ids[i] = c.InstanceID
	}
	picked := g.ChooseAtRandomForEffect(RandomDraw{Player: playerID}, ids, n)
	return g.discardCardsKeyedLocked(playerID, picked, DiscardCauseEffect, uuid.Nil, then)
}

// --- the engine's discard continuations ---------------------------------

var (
	discardThenMu  sync.RWMutex
	discardThenFns = map[string]DiscardThenFunc{}
)

// discardThenDone is a keyed discard-then's continuation once the discard
// has landed (RegisterDiscardThenRule): the registered rest of the card
// named by the tail's Word, told what was discarded.
var discardThenDone = RegisterTail("engine/discard-then-done", func(g *Game, o TailOutcome) error {
	discardThenMu.RLock()
	fn := discardThenFns[o.Tail.Word]
	discardThenMu.RUnlock()
	if fn == nil {
		return nil
	}
	return fn(g, DiscardsDone{
		Player:    o.Tail.Player,
		Source:    o.Tail.Source,
		Discarded: PromptedDiscards{{Seat: o.Tail.Player, Cards: copyUUIDs(o.IDs)}},
		Carry:     copyUUIDs(o.Tail.Carry),
		Item:      o.Tail.Asked.StackItem(),
	})
})

// discardLegSettled is a prompted discard run's leg once its discard has
// landed: the run's per-seat tail, if it is keyed, then the leg's settle.
// The run is Carry[0]; a fan-out nothing waits on has none. N is 1 for a
// closure-form prompt's leg, which settles without the keyed per-seat tail
// (queueDiscardPromptLocked's own leg never ran one).
var discardLegSettled = RegisterTail("engine/discard-leg-settled", func(g *Game, o TailOutcome) error {
	var runID uuid.UUID
	if len(o.Tail.Carry) > 0 {
		runID = o.Tail.Carry[0]
	}
	if run := g.promptRuns[runID]; run != nil && o.Tail.N == 0 {
		g.runLegTailLocked(run.thenKey, o.Tail.Player, o.Tail.Source, o.IDs, run.carry)
	}
	return g.settleRunLegLocked(runID, o.Tail.Player, o.IDs)
})

// discardLegSettledTail is discardLegSettled for one seat's leg of run.
func discardLegSettledTail(run, seat, source uuid.UUID) KeyedTail {
	return KeyedTail{Ref: discardLegSettled, Player: seat, Source: source, Carry: []uuid.UUID{run}}
}

// withTailN sets the number a tail's body reads.
func withTailN(t KeyedTail, n int) KeyedTail {
	t.N = n
	return t
}
