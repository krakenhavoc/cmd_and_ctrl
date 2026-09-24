package game

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// card_index_test.go is #1479's property test and the behaviour tests
// around it (ADR 0094). The source checks are in
// card_index_source_test.go; the cost test is in
// protocol/view_cost_test.go, where the view that pays for lookups is.

// indexProbe checks, for every card the game has ever held and for a
// few IDs it never held, that the index gives the zone and position
// the zone walk gives. It calls locateIndexedLocked directly, so the
// comparison does not depend on the cross-check mode being on.
//
// The walk it compares against is written out here, zone by zone in
// the order findCardZoneLocked walked them before #1479, rather than
// borrowed from card_index.go — so the index cannot agree with the
// reference by sharing a mistake with it. It is computed once per
// check, for every ID at once, which is what keeps the test fast
// enough to run under -race.
type indexProbe struct {
	seen    map[uuid.UUID]bool
	strays  []uuid.UUID
	lookups int
}

func newIndexProbe() *indexProbe {
	p := &indexProbe{seen: map[uuid.UUID]bool{}}
	for i := 0; i < 4; i++ {
		p.strays = append(p.strays, uuid.New())
	}
	p.strays = append(p.strays, uuid.Nil)
	return p
}

type zonePos struct {
	z   *Zone
	pos int
}

// walkReference is the pre-#1479 findCardZoneLocked's answer for every
// card at once: the first zone in walk order holding the ID, and the
// first position in that zone.
func walkReference(g *Game) map[uuid.UUID]zonePos {
	out := map[uuid.UUID]zonePos{}
	add := func(z *Zone) {
		for i := range z.Cards {
			if _, ok := out[z.Cards[i].InstanceID]; !ok {
				out[z.Cards[i].InstanceID] = zonePos{z, i}
			}
		}
	}
	add(g.Battlefield)
	add(g.Stack)
	add(g.Exile)
	for _, p := range g.Seats {
		add(p.Library)
		add(p.Hand)
		add(p.Graveyard)
		add(p.Command)
	}
	return out
}

// check must run with g.mu held.
func (p *indexProbe) check(t *testing.T, g *Game, where string) {
	t.Helper()
	ref := walkReference(g)
	for id := range ref {
		p.seen[id] = true
	}
	for i := range g.PhasedOut.Cards {
		p.seen[g.PhasedOut.Cards[i].InstanceID] = true
	}
	// Every card in a zone, every phased-out card, the strays, and up
	// to 16 cards that have left the game (tokens that ceased to exist,
	// the old IDs of flickered cards). Each of those costs a full walk
	// in both the index and the reference, so all of them every step
	// would make this test the slowest in the package under -race; map
	// order makes the 16 a different sample each step.
	ids := append([]uuid.UUID(nil), p.strays...)
	gone := 0
	for id := range p.seen {
		if _, live := ref[id]; !live {
			if gone >= 16 {
				continue
			}
			gone++
		}
		ids = append(ids, id)
	}
	for _, id := range ids {
		p.lookups++
		gotZ, gotPos := g.locateIndexedLocked(id)
		want, ok := ref[id]
		if !ok {
			want = zonePos{nil, -1}
		}
		if gotZ != want.z || gotPos != want.pos {
			t.Fatalf("%s: lookup of %s: index says %s, walk says %s",
				where, id, describeZonePos(gotZ, gotPos), describeZonePos(want.z, want.pos))
		}
	}
}

// randomCardIn picks a card from one of the game's indexed zones, or
// uuid.Nil when there is none.
func randomCardIn(g *Game, rng *rand.Rand, kinds ...ZoneKind) (uuid.UUID, *Zone) {
	var pool []*Zone
	g.eachIndexedZoneLocked(func(_ cardSlot, _ int, z *Zone) bool {
		for _, k := range kinds {
			if z.Kind == k && len(z.Cards) > 0 {
				pool = append(pool, z)
			}
		}
		return false
	})
	if len(pool) == 0 {
		return uuid.Nil, nil
	}
	z := pool[rng.IntN(len(pool))]
	return z.Cards[rng.IntN(len(z.Cards))].InstanceID, z
}

var sandboxZoneKinds = []ZoneKind{ZoneBattlefield, ZoneExile, ZoneLibrary, ZoneHand, ZoneGraveyard}

func randomZoneRef(g *Game, rng *rand.Rand) ZoneRef {
	k := sandboxZoneKinds[rng.IntN(len(sandboxZoneKinds))]
	switch k {
	case ZoneBattlefield, ZoneExile:
		return ZoneRef{Kind: k}
	}
	return ZoneRef{Kind: k, Owner: g.Seats[rng.IntN(len(g.Seats))].ID}
}

func zoneRefOf(z *Zone) ZoneRef {
	return ZoneRef{Kind: z.Kind, Owner: z.Owner}
}

// TestCardIndexAgreesWithTheZoneWalk is the property: across random
// games — sandbox moves between every pair of zones, moves to the
// bottom, draws, mills, shuffles, tokens created and ceasing to exist,
// flickers that mint a new instance ID, phasing out and in, undo to a
// random depth through Clone / RestoreFrom (the pair ws.Room uses),
// play continued on a clone, and persisted-snapshot restores — every
// lookup through the index gives the zone and position the reference
// walk gives, for every card the game has ever held and for IDs it
// never held.
//
// Nothing in the engine invalidates the index; this is the test that
// it never needs to.
func TestCardIndexAgreesWithTheZoneWalk(t *testing.T) {
	const steps = 400
	for seed := uint64(1); seed <= 3; seed++ {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			g := newActiveGameWithSeats(t, 4)
			rng := rand.New(rand.NewPCG(seed, 1479))
			probe := newIndexProbe()
			var undo []*Game
			counts := map[string]int{}
			token := Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}

			for step := 0; step < steps; step++ {
				op := ""
				// Undo, clone and restore replace g itself, so they run
				// outside the lock.
				switch r := rng.IntN(100); {
				case r < 6 && len(undo) > 0:
					at := len(undo) - 1 - rng.IntN(min(len(undo), 5))
					snap := undo[at]
					undo = undo[:at]
					g.WithWriteLock(func() { g.RestoreFrom(snap) })
					op = "undo"
				case r < 8:
					g = g.Clone()
					undo = nil
					op = "continue on a clone"
				case r < 9:
					restored, err := g.CaptureSnapshot().Restore()
					if err != nil {
						t.Fatalf("step %d: snapshot restore: %v", step, err)
					}
					g = restored
					undo = nil
					op = "snapshot restore"
				default:
					undo = append(undo, g.Clone())
					g.WithWriteLock(func() { op = randomIndexOp(g, rng, token) })
				}
				counts[op]++
				g.WithWriteLock(func() { probe.check(t, g, fmt.Sprintf("step %d after %s", step, op)) })
			}

			near, fallbacks, rebuilds := g.CardIndexStats()
			t.Logf("seed %d: %d lookups checked, %d ids seen; ops %v; last game's index: %d near hits, %d fallbacks, %d rebuilds",
				seed, probe.lookups, len(probe.seen), counts, near, fallbacks, rebuilds)
			for _, must := range []string{"undo", "continue on a clone", "snapshot restore", "move", "move to bottom",
				"token", "token dies", "flicker", "phase out", "phase in", "draw", "shuffle"} {
				if counts[must] == 0 {
					t.Errorf("seed %d never exercised %q; widen the run", seed, must)
				}
			}
		})
	}
}

// randomIndexOp applies one random zone mutation and names it. Caller
// holds g.mu. Errors are ignored: a refused move is a move that did not
// happen, and the probe checks the state either way.
func randomIndexOp(g *Game, rng *rand.Rand, token Card) string {
	seat := g.Seats[rng.IntN(len(g.Seats))].ID
	switch rng.IntN(12) {
	case 0, 1, 2:
		id, z := randomCardIn(g, rng, sandboxZoneKinds...)
		if z == nil {
			return "nothing to move"
		}
		_ = g.moveCardByRefLocked(zoneRefOf(z), randomZoneRef(g, rng), id, false, false)
		return "move"
	case 3:
		id, z := randomCardIn(g, rng, sandboxZoneKinds...)
		if z == nil {
			return "nothing to move"
		}
		_ = g.moveCardByRefLocked(zoneRefOf(z), randomZoneRef(g, rng), id, false, true)
		return "move to bottom"
	case 4:
		_, _ = g.CreateTokensForEffect(seat, token, 1+rng.IntN(3), TokenEntryOptions{})
		return "token"
	case 5:
		// A token that leaves the battlefield ceases to exist at the
		// next state check (CR 704.5d).
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() {
				_ = g.DestroyPermanentForEffect(c.InstanceID)
				g.runStateChecksLocked()
				return "token dies"
			}
		}
		return "no token to kill"
	case 6:
		// Exile and return: the card comes back as a new object with a
		// new instance ID (CR 400.7).
		id, _ := randomCardIn(g, rng, ZoneBattlefield)
		if id == uuid.Nil {
			return "nothing to flicker"
		}
		if err := g.ExileCardForEffect(id); err != nil {
			return "flicker refused"
		}
		_, _ = g.ReturnFromExileToBattlefieldForEffect(id, seat, false)
		return "flicker"
	case 7:
		id, _ := randomCardIn(g, rng, ZoneBattlefield)
		if id == uuid.Nil {
			return "nothing to phase"
		}
		_ = g.PhaseOutForEffect(uuid.Nil, id)
		return "phase out"
	case 8:
		if len(g.PhasedOut.Cards) == 0 {
			return "nothing phased out"
		}
		g.phaseInLocked([]uuid.UUID{g.PhasedOut.Cards[rng.IntN(len(g.PhasedOut.Cards))].InstanceID})
		return "phase in"
	case 9:
		_ = g.DrawNForEffect(seat, 1+rng.IntN(2))
		return "draw"
	case 10:
		_ = g.MillNForEffect(seat, 1+rng.IntN(3))
		return "mill"
	default:
		_ = g.ShuffleLibraryForEffect(seat)
		return "shuffle"
	}
}

// TestCardIndexKeepsItsTableThroughAnUndo: RestoreFrom does not touch
// the index. The table built before the undo is still the one in use
// after it, and it answers for the restored zones — including for the
// card the undone action moved, whose hint now points at the zone it
// was moved TO.
func TestCardIndexKeepsItsTableThroughAnUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var moved uuid.UUID
	g.WithWriteLock(func() {
		moved = me.Hand.Cards[0].InstanceID
		g.findCardZoneLocked(moved) // build the table
	})
	pre := g.Clone()
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me.ID}, ZoneRef{Kind: ZoneBattlefield}, moved); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		g.rebuildCardIndexLocked() // the hint now says "battlefield"
		if z := g.findCardZoneLocked(moved); z != g.Battlefield {
			t.Fatalf("after the move: %v", z)
		}
	})
	table := g.cardIndex.table.Load()
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	g.WithWriteLock(func() {
		if g.cardIndex.table.Load() != table {
			t.Fatal("RestoreFrom replaced the index table; it is meant to keep it and let every read check it")
		}
		if z := g.findCardZoneLocked(moved); z != g.Seats[0].Hand {
			t.Fatalf("after the undo the card is in %v, want the restored hand", z)
		}
		if g.Battlefield.Contains(moved) {
			t.Fatal("the restored battlefield holds the moved card")
		}
	})
}

// TestCardIndexStartsEmptyOnACloneAndARestoredSnapshot: neither copies
// the table. A new *Game builds its own on its first lookup, against
// its own zones.
func TestCardIndexStartsEmptyOnACloneAndARestoredSnapshot(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() { g.findCardZoneLocked(g.Seats[0].Hand.Cards[0].InstanceID) })
	if g.cardIndex.table.Load() == nil {
		t.Fatal("a lookup did not build the table")
	}
	c := g.Clone()
	if c.cardIndex.table.Load() != nil {
		t.Fatal("the clone carried the source's index table")
	}
	r, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatal(err)
	}
	if r.cardIndex.table.Load() != nil {
		t.Fatal("the restored game carried an index table")
	}
	for _, x := range []*Game{c, r} {
		x.WithWriteLock(func() {
			id := x.Seats[0].Hand.Cards[0].InstanceID
			if z := x.findCardZoneLocked(id); z != x.Seats[0].Hand {
				t.Fatalf("lookup on the new game answered %v, want its own hand zone", z)
			}
		})
	}
}

// TestCardIndexFindsAShiftedCardWithoutAFallback: a removal lower in a
// zone moves every card above it down one place. The hint's zone is
// still right, and the lookup finds the card next to where it was
// rather than walking the game.
func TestCardIndexFindsAShiftedCardWithoutAFallback(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		hand := me.Hand
		last := hand.Cards[len(hand.Cards)-1].InstanceID
		g.rebuildCardIndexLocked()
		if _, err := hand.Remove(hand.Cards[0].InstanceID); err != nil {
			t.Fatal(err)
		}
		near0, fall0, _ := g.CardIndexStats()
		z, pos := g.locateCardLocked(last)
		near1, fall1, _ := g.CardIndexStats()
		if z != hand || pos != len(hand.Cards)-1 {
			t.Fatalf("shifted card: got %s", describeZonePos(z, pos))
		}
		if near1-near0 != 1 || fall1 != fall0 {
			t.Fatalf("shifted card: %d near hits and %d fallbacks, want 1 and 0", near1-near0, fall1-fall0)
		}
	})
}

// TestCardIndexRebuildsOnceTheFallbacksAddUp: moved cards cost a
// fallback walk each until the table is rebuilt, and the rebuild comes
// once that work passes cardIndexRebuildFactor × the table's size —
// not on every fallback, and not never.
func TestCardIndexRebuildsOnceTheFallbacksAddUp(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	g.WithWriteLock(func() {
		built := g.rebuildCardIndexLocked()
		limit := int64(cardIndexRebuildFactor * built.size)
		_, _, before := g.CardIndexStats()
		// Move a card from the last seat's library to the first seat's
		// graveyard, then look it up until a rebuild happens. Each
		// lookup's fallback walk examines most of the game.
		src := g.Seats[3].Library
		id := src.Cards[0].InstanceID
		if _, err := MoveCard(src, g.Seats[0].Graveyard, id); err != nil {
			t.Fatal(err)
		}
		lookups := 0
		var wasteBefore int64
		for ; lookups < 1000; lookups++ {
			wasteBefore = g.cardIndex.waste.Load()
			g.locateCardLocked(id)
			if _, _, r := g.CardIndexStats(); r > before {
				break
			}
		}
		// Not on the first fallback, and not never: on the lookup that
		// took the work since the build past the limit. One lookup here
		// examines at most the hinted zone plus a walk of the game.
		if lookups == 0 {
			t.Fatal("rebuilt on the first fallback; a rebuild should wait for the work to add up")
		}
		if lookups == 1000 {
			t.Fatalf("never rebuilt in 1000 fallback lookups of a moved card (%d cards of work, limit %d)", g.cardIndex.waste.Load(), limit)
		}
		if wasteBefore >= limit || wasteBefore < limit-int64(2*built.size) {
			t.Fatalf("rebuilt with %d cards of work already charged against a limit of %d (%d × a %d-card table); want the lookup that crossed it",
				wasteBefore, limit, cardIndexRebuildFactor, built.size)
		}
		nearBefore, fallBefore, _ := g.CardIndexStats()
		g.locateCardLocked(id)
		nearAfter, fallAfter, _ := g.CardIndexStats()
		if nearAfter != nearBefore || fallAfter != fallBefore {
			t.Fatal("the lookup after the rebuild was not a fast-path hit")
		}
	})
}

// TestCardIndexIsSafeForConcurrentReaders: views run under the read
// lock, several at once, and a view's lookups may rebuild the table.
// Readers that rebuild together must neither race (go test -race) nor
// publish a table that answers wrongly. Every reader looks up cards
// that have moved since the table was built, so every reader is on the
// rebuild path.
func TestCardIndexIsSafeForConcurrentReaders(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		g.rebuildCardIndexLocked()
		for _, p := range g.Seats {
			for i := 0; i < 10; i++ {
				c := p.Library.Cards[i]
				if _, err := MoveCard(p.Library, g.Seats[0].Graveyard, c.InstanceID); err != nil {
					t.Fatal(err)
				}
				ids = append(ids, c.InstanceID)
			}
		}
	})
	const readers = 8
	errs := make(chan string, readers)
	done := make(chan struct{})
	for r := 0; r < readers; r++ {
		go func(r int) {
			defer func() { done <- struct{}{} }()
			for round := 0; round < 50; round++ {
				g.ReadSnapshot(func() {
					for _, id := range ids {
						z, pos := g.locateCardLocked(id)
						if z != g.Seats[0].Graveyard || z.Cards[pos].InstanceID != id {
							select {
							case errs <- fmt.Sprintf("reader %d: %s answered %s", r, id, describeZonePos(z, pos)):
							default:
							}
							return
						}
					}
				})
			}
		}(r)
	}
	for r := 0; r < readers; r++ {
		<-done
	}
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	if _, _, rebuilds := g.CardIndexStats(); rebuilds < 2 {
		t.Fatalf("%d rebuilds; the readers never reached the rebuild path", rebuilds)
	}
}

// TestCardIndexCrossCheckReportsADuplicate: the debug mode is the net
// under the one assumption the index cannot check for itself — an
// instance ID in one zone at a time. Put one card in two zones and the
// cross-check must say so.
func TestCardIndexCrossCheckReportsADuplicate(t *testing.T) {
	var got []string
	restore := SetCardIndexCrossCheck(func(msg string) { got = append(got, msg) })
	defer restore()

	g := newActiveGame(t)
	g.WithWriteLock(func() {
		c := g.Seats[1].Hand.Cards[0]
		g.rebuildCardIndexLocked() // the hint says: seat 1's hand
		g.Battlefield.PushTop(c)   // … and now the battlefield holds it too
		g.findCardZoneLocked(c.InstanceID)
	})
	if len(got) != 1 || !strings.Contains(got[0], "disagrees") {
		t.Fatalf("cross-check reports %q, want one disagreement", got)
	}
}
