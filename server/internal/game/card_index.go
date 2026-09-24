package game

import (
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
)

// card_index.go answers "which zone holds this card, and where" in
// O(1) instead of by walking every zone (#1479, ADR 0094).
//
// Before this, findCardZoneLocked ran Zone.Contains over the
// battlefield, the stack, exile and each seat's library, hand,
// graveyard and command zone in turn, and the view asks once per card
// it projects. That made every view O(cards²): about 7% of a bot
// table's CPU once #1401 had taken the public log out of the way.
//
// # The index is a hint, and every answer is checked
//
// The index maps an instance ID to the zone and position the card had
// the last time the index looked. It is NEVER trusted: a lookup
// resolves the zone against the live game and reads the card at that
// position, and the answer counts only if that card has the ID asked
// for. When it does not, the lookup searches outward from the old
// position (a removal lower in the zone shifts a card by one or two
// places), and failing that it falls back to the reference walk
// below, which is the pre-#1479 scan, unchanged.
//
// That is why the index needs no invalidation — not on a zone move,
// not on token creation or removal, not on undo (RestoreFrom), Clone
// or snapshot load. A stale entry costs one check and a fallback; it
// cannot produce a wrong answer. Three things make that true, and the
// tests hold each of them:
//
//  1. A hint names its zone by SLOT and SEAT INDEX, never by pointer,
//     and is resolved through the live Game on every read. So a hint
//     that outlives an undo — RestoreFrom swaps every zone pointer —
//     is read against the restored zones, never the discarded ones.
//     TestCardIndexHintsHoldNoPointers pins the no-pointer shape.
//  2. The index and the reference walk cover the SAME zones in the
//     same order, because both go through eachIndexedZoneLocked.
//     PhasedOut is deliberately in neither (ADR 0084).
//  3. An instance ID is in at most one of those zones at a time. That
//     is an engine invariant rather than something this file can
//     check cheaply, which is what the cross-check mode is for.
//
// # Concurrency
//
// Views run under the READ lock, and several of them — the room's and
// each bot seat's — run at once. So the table is immutable once
// published, and replaced whole through an atomic pointer. A read on
// the fast path is an atomic load, a map read and one comparison: no
// lock, no write, nothing shared that is written.
//
// The table is rebuilt when the work spent off the fast path since
// it was built passes a multiple of its size, which keeps a rebuild's
// O(cards) cost amortised against the fallbacks it saves. Two readers
// that cross the threshold together may both rebuild; the tables they
// build are identical (nobody can move a card while they hold the read
// lock), and whichever is published last wins.
//
// # Cross-check mode
//
// SetCardIndexCrossCheck makes every lookup also run the reference
// walk and report any disagreement. The game, protocol, effects,
// legal and actions test packages turn it on for their whole run, so
// every lookup any of those tests makes is checked against the scan
// it replaced.

// cardSlot names one kind of indexed zone. Per-seat slots are resolved
// together with a seat index.
type cardSlot uint8

const (
	slotBattlefield cardSlot = iota
	slotStack
	slotExile
	slotLibrary
	slotHand
	slotGraveyard
	slotCommand
)

// cardLoc is where the index last saw a card. It holds no pointer on
// purpose — see point 1 above.
type cardLoc struct {
	slot cardSlot
	seat int32
	pos  int32
}

// cardIndexTable is one published, immutable snapshot of the index.
type cardIndexTable struct {
	locs map[uuid.UUID]cardLoc
	// size is the number of cards indexed, which sets how much
	// off-fast-path work buys a rebuild.
	size int
}

// cardLocationIndex is the Game's card-location hint table (#1479). The zero
// value is ready to use: the first lookup builds the table.
type cardLocationIndex struct {
	table atomic.Pointer[cardIndexTable]
	// waste counts the cards examined off the fast path since the
	// current table was built, where a rebuild would have saved them:
	// outward searches, and fallback walks that found their card.
	waste atomic.Int64
	// building keeps concurrent readers that cross the threshold
	// together from all rebuilding at once. It is advisory; see the
	// file comment.
	building atomic.Bool

	// Diagnostics, all counted off the fast path only: a fast-path
	// counter would be one more cache line every concurrent view
	// writes to.
	nearHits  atomic.Int64
	fallbacks atomic.Int64
	rebuilds  atomic.Int64
}

// cardIndexRebuildFactor: rebuild once the off-fast-path work since
// the last build reaches this many times the table's size. Building a
// table touches every card and inserts it into a map, several times
// the cost of the comparison a fallback walk makes per card; a factor
// in that range keeps the rebuilds from costing more than the
// fallbacks they end.
const cardIndexRebuildFactor = 8

// cardIndexMinWaste keeps a tiny game (a unit test's handful of cards)
// from rebuilding on every fallback.
const cardIndexMinWaste = 64

// eachIndexedZoneLocked calls fn for every zone a card lookup covers,
// in the order the reference walk has always used: the battlefield,
// the stack, exile, then each seat's library, hand, graveyard and
// command zone. It stops when fn returns true. seat is -1 for a shared
// zone.
//
// The index builder and the reference walk both go through here, so
// they cannot disagree about which zones exist. PhasedOut is not
// listed: a phased-out permanent is treated as though it does not
// exist (CR 702.26b, ADR 0084).
func (g *Game) eachIndexedZoneLocked(fn func(slot cardSlot, seat int, z *Zone) bool) {
	if g.Battlefield != nil && fn(slotBattlefield, -1, g.Battlefield) {
		return
	}
	if g.Stack != nil && fn(slotStack, -1, g.Stack) {
		return
	}
	if g.Exile != nil && fn(slotExile, -1, g.Exile) {
		return
	}
	for i, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, sz := range [...]struct {
			slot cardSlot
			z    *Zone
		}{
			{slotLibrary, p.Library},
			{slotHand, p.Hand},
			{slotGraveyard, p.Graveyard},
			{slotCommand, p.Command},
		} {
			if sz.z != nil && fn(sz.slot, i, sz.z) {
				return
			}
		}
	}
}

// indexedZoneLocked resolves a slot and seat index against the live
// game. Nil when the seat no longer exists.
func (g *Game) indexedZoneLocked(slot cardSlot, seat int32) *Zone {
	switch slot {
	case slotBattlefield:
		return g.Battlefield
	case slotStack:
		return g.Stack
	case slotExile:
		return g.Exile
	}
	if seat < 0 || int(seat) >= len(g.Seats) {
		return nil
	}
	p := g.Seats[seat]
	if p == nil {
		return nil
	}
	switch slot {
	case slotLibrary:
		return p.Library
	case slotHand:
		return p.Hand
	case slotGraveyard:
		return p.Graveyard
	case slotCommand:
		return p.Command
	}
	return nil
}

// findCardLinearLocked is the reference walk: the pre-#1479
// findCardZoneLocked, returning the position as well. It is the
// fallback when a hint does not hold, and the oracle the cross-check
// mode and the tests compare the index against. The second result is
// -1 when no zone holds the card; the third is the number of cards it
// examined.
func (g *Game) findCardLinearLocked(cardID uuid.UUID) (*Zone, int, int) {
	var (
		found   *Zone
		at      = -1
		scanned int
	)
	g.eachIndexedZoneLocked(func(_ cardSlot, _ int, z *Zone) bool {
		for i := range z.Cards {
			scanned++
			if z.Cards[i].InstanceID == cardID {
				found, at = z, i
				return true
			}
		}
		return false
	})
	return found, at, scanned
}

// locateCardLocked returns the zone holding the card and its position
// in that zone's Cards, or (nil, -1) when no indexed zone holds it.
// Every "which zone is this card in" question in the package comes
// through here. Caller must hold g.mu (read or write).
func (g *Game) locateCardLocked(cardID uuid.UUID) (*Zone, int) {
	if lookupCounting.Load() {
		lookupCalls.Add(1)
	}
	z, pos := g.locateIndexedLocked(cardID)
	if cardIndexCrossCheck.Load() {
		g.crossCheckCardLocked(cardID, z, pos)
	}
	return z, pos
}

func (g *Game) locateIndexedLocked(cardID uuid.UUID) (*Zone, int) {
	idx := &g.cardIndex
	t := idx.table.Load()
	if t == nil {
		t = g.rebuildCardIndexLocked()
	}
	if loc, ok := t.locs[cardID]; ok {
		if z := g.indexedZoneLocked(loc.slot, loc.seat); z != nil {
			p := int(loc.pos)
			if p < len(z.Cards) && z.Cards[p].InstanceID == cardID {
				noteLookupWork(1)
				return z, p
			}
			// The zone is probably right and the position has moved:
			// a card lower down left, or one was put on the bottom.
			// Search outward from where the card was.
			found, scanned := nearIndexOf(z, cardID, p)
			noteLookupWork(scanned)
			if found >= 0 {
				idx.nearHits.Add(1)
				g.chargeCardIndexLocked(t, scanned)
				return z, found
			}
			g.chargeCardIndexLocked(t, scanned)
		}
	}
	idx.fallbacks.Add(1)
	z, pos, scanned := g.findCardLinearLocked(cardID)
	noteLookupWork(scanned)
	if z != nil {
		// Only a walk that FOUND the card counts towards a rebuild:
		// the new table would answer it on the fast path. A card no
		// zone holds — a token that has ceased to exist, the old ID
		// of a card that changed zones — costs the same walk with or
		// without a rebuild, so rebuilding for it would be pure cost.
		g.chargeCardIndexLocked(t, scanned)
	}
	return z, pos
}

// nearIndexOf searches z for id outward from `from`: from-1, from+1,
// from-2, … It returns the position (or -1) and how many cards it
// examined. The card at `from` itself has already been checked.
func nearIndexOf(z *Zone, id uuid.UUID, from int) (int, int) {
	n := len(z.Cards)
	scanned := 0
	for d := 1; from-d >= 0 || from+d < n; d++ {
		if i := from - d; i >= 0 && i < n {
			scanned++
			if z.Cards[i].InstanceID == id {
				return i, scanned
			}
		}
		if i := from + d; i >= 0 && i < n {
			scanned++
			if z.Cards[i].InstanceID == id {
				return i, scanned
			}
		}
	}
	return -1, scanned
}

// chargeCardIndexLocked records off-fast-path work against table t
// and rebuilds once the work reaches cardIndexRebuildFactor × its
// size. Work charged against a table that has already been replaced
// is dropped.
func (g *Game) chargeCardIndexLocked(t *cardIndexTable, work int) {
	idx := &g.cardIndex
	if idx.table.Load() != t {
		return
	}
	limit := int64(cardIndexRebuildFactor * t.size)
	if limit < cardIndexMinWaste {
		limit = cardIndexMinWaste
	}
	if idx.waste.Add(int64(work)) < limit {
		return
	}
	if !idx.building.CompareAndSwap(false, true) {
		return
	}
	defer idx.building.Store(false)
	g.rebuildCardIndexLocked()
}

// rebuildCardIndexLocked indexes every card in every indexed zone and
// publishes the table. If an ID appears twice, the first zone in walk
// order wins, which is what the reference walk would answer.
func (g *Game) rebuildCardIndexLocked() *cardIndexTable {
	n := 0
	g.eachIndexedZoneLocked(func(_ cardSlot, _ int, z *Zone) bool {
		n += len(z.Cards)
		return false
	})
	t := &cardIndexTable{locs: make(map[uuid.UUID]cardLoc, n), size: n}
	g.eachIndexedZoneLocked(func(slot cardSlot, seat int, z *Zone) bool {
		for i := range z.Cards {
			id := z.Cards[i].InstanceID
			if _, dup := t.locs[id]; dup {
				continue
			}
			t.locs[id] = cardLoc{slot: slot, seat: int32(seat), pos: int32(i)}
		}
		return false
	})
	noteLookupWork(n)
	g.cardIndex.rebuilds.Add(1)
	g.cardIndex.waste.Store(0)
	g.cardIndex.table.Store(t)
	return t
}

// --- cross-check mode ---------------------------------------------------

var (
	cardIndexCrossCheck atomic.Bool
	// cardIndexMismatch receives a cross-check failure. Guarded by
	// SetCardIndexCrossCheck's contract: set once, before the lookups
	// it is to see.
	cardIndexMismatch func(msg string)
)

// SetCardIndexCrossCheck turns the card-index debug mode on (report
// non-nil) or off (nil). While it is on, every card lookup also runs
// the reference walk, and a disagreement is passed to report — a test
// package's TestMain passes a panic, so the failing lookup's stack is
// the one printed. It returns a function that restores the previous
// setting.
//
// Test and debugging use only: it doubles the cost of every lookup.
// Not safe to call while lookups are running.
func SetCardIndexCrossCheck(report func(msg string)) (restore func()) {
	prevOn, prevReport := cardIndexCrossCheck.Load(), cardIndexMismatch
	cardIndexMismatch = report
	cardIndexCrossCheck.Store(report != nil)
	return func() {
		cardIndexMismatch = prevReport
		cardIndexCrossCheck.Store(prevOn)
	}
}

func (g *Game) crossCheckCardLocked(cardID uuid.UUID, z *Zone, pos int) {
	wantZ, wantPos, _ := g.findCardLinearLocked(cardID)
	if z == wantZ && pos == wantPos {
		return
	}
	if report := cardIndexMismatch; report != nil {
		report(fmt.Sprintf("card index disagrees with the zone walk for %s: index says %s, walk says %s",
			cardID, describeZonePos(z, pos), describeZonePos(wantZ, wantPos)))
	}
}

func describeZonePos(z *Zone, pos int) string {
	if z == nil {
		return fmt.Sprintf("no zone (pos %d)", pos)
	}
	// The pointer is there because two zones can share a kind and an
	// owner: the live one and one an undo discarded.
	return fmt.Sprintf("%s of %s at %d/%d (zone %p)", z.Kind, z.Owner, pos, len(z.Cards), z)
}

// --- measurement ----------------------------------------------------------

var (
	lookupCounting atomic.Bool
	lookupCalls    atomic.Int64
	lookupWork     atomic.Int64
)

// noteLookupWork adds n cards examined to the lookup-work counter, if
// a test has turned counting on. The check is one read of a flag
// nobody writes during a game, so it costs nothing measurable when
// counting is off.
func noteLookupWork(n int) {
	if lookupCounting.Load() {
		lookupWork.Add(int64(n))
	}
}

// CardLookupWork is what CountCardLookupWork has counted so far.
type CardLookupWork struct {
	// Lookups is how many card lookups ran.
	Lookups int64
	// Examined is how many cards they examined: one per fast-path
	// hit, one per card an outward search or a fallback walk compared,
	// and one per card a table rebuild indexed.
	Examined int64
}

// CountCardLookupWork starts counting card lookups and the cards they
// examine, and returns a reader for the running totals and a stop
// function. The cost tests use it to show that a lookup's cost does
// not grow with the board. Counting is global, not per game; tests
// that use it must not run in parallel.
func CountCardLookupWork() (read func() CardLookupWork, stop func()) {
	lookupCalls.Store(0)
	lookupWork.Store(0)
	lookupCounting.Store(true)
	read = func() CardLookupWork {
		return CardLookupWork{Lookups: lookupCalls.Load(), Examined: lookupWork.Load()}
	}
	return read, func() { lookupCounting.Store(false) }
}

// CardIndexStats reports the index's off-fast-path activity since this
// *Game was created: lookups that found their card near its old
// position, lookups that fell back to the full walk, and table
// rebuilds. For tests and profiling.
func (g *Game) CardIndexStats() (nearHits, fallbacks, rebuilds int64) {
	return g.cardIndex.nearHits.Load(), g.cardIndex.fallbacks.Load(), g.cardIndex.rebuilds.Load()
}
