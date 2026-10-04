package game

import (
	"testing"

	"github.com/google/uuid"
)

// ADR 0113 §3 (#2074): the maximum hand size as a CR 613.11
// timestamp-order fold. These tests stub CatalogHandSize, the slot that
// exists so a game-package test needs no catalog; the card-level tests
// with the printed cards are in cards/effects/max_hand_size_test.go.

// withHandSizeStatics answers CatalogHandSize from a map of stub keys.
func withHandSizeStatics(t *testing.T, defs map[string][]HandSizeStatic) {
	t.Helper()
	prev := CatalogHandSize
	CatalogHandSize = func(key string) []HandSizeStatic {
		if hs, ok := defs[key]; ok {
			return hs
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogHandSize = prev })
}

// pushHandSizePermanent puts a permanent with the stub key `key` under
// `controller`, stamped as having entered at `at`.
func pushHandSizePermanent(g *Game, controller uuid.UUID, key string, at int64) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{
			InstanceID: id, Name: key, TypeLine: "Artifact", OracleID: key,
			Owner: controller, Controller: controller, EnteredBattlefieldAt: at,
		})
	})
	return id
}

func maxHandOf(g *Game, p *Player) int {
	var n int
	g.WithWriteLock(func() { n = g.EffectiveMaxHandSizeLocked(p) })
	return n
}

const (
	hsNoMax     = "test-you-have-no-maximum"
	hsSetTwo    = "test-your-maximum-is-two"
	hsOppMinus7 = "test-each-opponent-minus-seven"
	hsEveryNoMx = "test-players-have-no-maximum"
	hsYouPlus1  = "test-your-maximum-plus-one"
)

func hsDefs() map[string][]HandSizeStatic {
	return map[string][]HandSizeStatic{
		hsNoMax:     {{Players: HandSizeYou, Kind: HandSizeNoMaximum}},
		hsSetTwo:    {{Players: HandSizeYou, Kind: HandSizeSet, N: 2}},
		hsOppMinus7: {{Players: HandSizeEachOpponent, Kind: HandSizeModify, N: -7}},
		hsEveryNoMx: {{Players: HandSizeEachPlayer, Kind: HandSizeNoMaximum}},
		hsYouPlus1:  {{Players: HandSizeYou, Kind: HandSizeModify, N: 1}},
	}
}

// CR 613.11, the Null Profusion ruling (2009-10-01): "no maximum" then
// "is two" is two; "is two" then "no maximum" is no maximum.
func TestMaxHandSizeFoldsInTimestampOrder(t *testing.T) {
	withHandSizeStatics(t, hsDefs())

	g := newActiveGame(t)
	me := g.Seats[0]
	pushHandSizePermanent(g, me.ID, hsNoMax, 100)
	pushHandSizePermanent(g, me.ID, hsSetTwo, 200)
	if got := maxHandOf(g, me); got != 2 {
		t.Errorf("no maximum, then is two: %d, want 2", got)
	}

	g = newActiveGame(t)
	me = g.Seats[0]
	pushHandSizePermanent(g, me.ID, hsSetTwo, 100)
	pushHandSizePermanent(g, me.ID, hsNoMax, 200)
	if got := maxHandOf(g, me); got != NoMaxHandSize {
		t.Errorf("is two, then no maximum: %d, want no maximum", got)
	}

	// A modification after "no maximum" leaves it unbounded; one after
	// a set number adds to it.
	g = newActiveGame(t)
	me = g.Seats[0]
	pushHandSizePermanent(g, me.ID, hsSetTwo, 100)
	pushHandSizePermanent(g, me.ID, hsYouPlus1, 200)
	if got := maxHandOf(g, me); got != 3 {
		t.Errorf("is two, then +1: %d, want 3", got)
	}
	pushHandSizePermanent(g, me.ID, hsNoMax, 300)
	pushHandSizePermanent(g, me.ID, hsYouPlus1, 400)
	if got := maxHandOf(g, me); got != NoMaxHandSize {
		t.Errorf("no maximum, then +1: %d, want no maximum", got)
	}
}

// The battlefield order is not the timestamp order: a permanent pushed
// later but stamped earlier still applies first.
func TestMaxHandSizeSortsByTimestampNotBattlefieldOrder(t *testing.T) {
	withHandSizeStatics(t, hsDefs())
	g := newActiveGame(t)
	me := g.Seats[0]
	pushHandSizePermanent(g, me.ID, hsNoMax, 300)
	pushHandSizePermanent(g, me.ID, hsSetTwo, 100)
	if got := maxHandOf(g, me); got != NoMaxHandSize {
		t.Errorf("is two (earlier), no maximum (later): %d, want no maximum", got)
	}
}

// CR 107.1b: the fold is not clamped on the way, the result is. Two
// "each opponent -7" make -7, which is reported as zero, never as the
// -1 sentinel.
func TestMaxHandSizeNeverGoesBelowZero(t *testing.T) {
	withHandSizeStatics(t, hsDefs())
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushHandSizePermanent(g, me.ID, hsOppMinus7, 100)
	if got := maxHandOf(g, opp); got != 0 {
		t.Errorf("one -7: %d, want 0", got)
	}
	pushHandSizePermanent(g, me.ID, hsOppMinus7, 200)
	if got := maxHandOf(g, opp); got != 0 {
		t.Errorf("two -7 (=-7): %d, want 0", got)
	}
	// Not clamped on the way: 7 - 7 - 7 + 1 is -6, so zero. Clamping
	// each step would have made it 1.
	pushHandSizePermanent(g, opp.ID, hsYouPlus1, 300)
	if got := maxHandOf(g, opp); got != 0 {
		t.Errorf("-7, -7, +1 (=-6): %d, want 0", got)
	}
	if got := maxHandOf(g, me); got != DefaultMaxHandSize {
		t.Errorf("the controller: %d, want 7", got)
	}
}

// CR 514.1: the cleanup discard is the ACTIVE player's, against their
// own maximum. The active player controls "each opponent -7"; their
// own maximum stays seven, so of nine cards they discard two.
func TestCleanupDiscardUsesTheActivePlayersOwnMaximum(t *testing.T) {
	withHandSizeStatics(t, hsDefs())
	g := newActiveGame(t)
	active, opp := g.Seats[0], g.Seats[1]
	pushHandSizePermanent(g, active.ID, hsOppMinus7, 100)
	for active.Hand.Size() < 9 {
		_ = g.DrawCard(active.ID)
	}
	advanceTo(t, g, StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past End: %v", err)
	}
	if g.Turn.Step != StepCleanup {
		t.Fatalf("expected paused at Cleanup, got %q", g.Turn.Step)
	}
	if n := g.DiscardPending[active.ID]; n != 2 {
		t.Errorf("active player discards %d, want 2 (nine against their own seven)", n)
	}
	if _, ok := g.DiscardPending[opp.ID]; ok {
		t.Errorf("the non-active opponent was asked to discard: %+v", g.DiscardPending)
	}
}

// The opponent under two "-7"s discards their whole hand in their own
// cleanup step (the Jin-Gitaxias ruling, CR 107.1b).
func TestCleanupDiscardAtZeroIsTheWholeHand(t *testing.T) {
	withHandSizeStatics(t, hsDefs())
	g := newActiveGame(t)
	me := g.Seats[0]
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID != me.ID {
		t.Fatalf("setup: seat 0 is not active")
	}
	other := g.Seats[1]
	pushHandSizePermanent(g, other.ID, hsOppMinus7, 100)
	pushHandSizePermanent(g, other.ID, hsOppMinus7, 200)
	for active.Hand.Size() < 4 {
		_ = g.DrawCard(active.ID)
	}
	advanceTo(t, g, StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past End: %v", err)
	}
	if n, hand := g.DiscardPending[active.ID], active.Hand.Size(); n != hand {
		t.Errorf("discard %d of %d, want the whole hand", n, hand)
	}
}

// ADR 0113 §3 decision 3: the player's own grant (Finale of Revelation,
// the sandbox action) has a timestamp now, so it no longer always wins.
func TestMaxHandSizeGrantHasATimestamp(t *testing.T) {
	withHandSizeStatics(t, hsDefs())
	now := int64(1000)
	restore := SetClockForTest(func() int64 { now += 10; return now })
	t.Cleanup(restore)

	// Grant, then a later "is two": two.
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.SetMaxHandSizeForEffect(me.ID, NoMaxHandSize); err != nil {
			t.Fatalf("SetMaxHandSizeForEffect: %v", err)
		}
	})
	if me.MaxHandSizeAt == 0 {
		t.Fatalf("the grant was not stamped")
	}
	pushHandSizePermanent(g, me.ID, hsSetTwo, me.MaxHandSizeAt+5)
	if got := maxHandOf(g, me); got != 2 {
		t.Errorf("grant, then is two: %d, want 2", got)
	}

	// "Is two", then the grant: no maximum.
	g = newActiveGame(t)
	me = g.Seats[0]
	pushHandSizePermanent(g, me.ID, hsSetTwo, 1)
	g.WithWriteLock(func() { _ = g.SetMaxHandSizeForEffect(me.ID, NoMaxHandSize) })
	if got := maxHandOf(g, me); got != NoMaxHandSize {
		t.Errorf("is two, then grant: %d, want no maximum", got)
	}

	// The sandbox action is stamped the same way.
	g = newActiveGame(t)
	me = g.Seats[0]
	pushHandSizePermanent(g, me.ID, hsNoMax, 1)
	if err := g.SetMaxHandSize(me.ID, 4); err != nil {
		t.Fatalf("SetMaxHandSize: %v", err)
	}
	if got := maxHandOf(g, me); got != 4 {
		t.Errorf("no maximum, then the sandbox's 4: %d, want 4", got)
	}
}

// A grant restored from a file written before the timestamp existed
// has none, and sorts first: a permanent's "is two" still applies.
func TestUnstampedGrantSortsFirst(t *testing.T) {
	withHandSizeStatics(t, hsDefs())
	g := newActiveGame(t)
	me := g.Seats[0]
	me.MaxHandSize = NoMaxHandSize
	me.MaxHandSizeAt = 0
	pushHandSizePermanent(g, me.ID, hsSetTwo, 50)
	if got := maxHandOf(g, me); got != 2 {
		t.Errorf("unstamped grant, then is two: %d, want 2", got)
	}
}

// "Players have no maximum hand size" reaches every player, and
// "each opponent" never reaches the controller (CR 102.3).
func TestMaxHandSizePlayerScopes(t *testing.T) {
	withHandSizeStatics(t, hsDefs())
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	pushHandSizePermanent(g, me.ID, hsOppMinus7, 100)
	for _, p := range g.Seats[1:] {
		if got := maxHandOf(g, p); got != 0 {
			t.Errorf("opponent %s: %d, want 0", p.Name, got)
		}
	}
	pushHandSizePermanent(g, g.Seats[2].ID, hsEveryNoMx, 200)
	for _, p := range g.Seats {
		if got := maxHandOf(g, p); got != NoMaxHandSize {
			t.Errorf("%s under a later \"players have no maximum\": %d, want no maximum", p.Name, got)
		}
	}
}

// A chosen-player static reaches the chosen player only, and nobody
// before a player is chosen.
func TestMaxHandSizeChosenPlayer(t *testing.T) {
	const key = "test-chosen-player-is-four"
	withHandSizeStatics(t, map[string][]HandSizeStatic{
		key: {{Players: HandSizeChosenPlayer, Kind: HandSizeSet, N: 4}},
	})
	g := newFourPlayerActiveGame(t)
	me, target := g.Seats[0], g.Seats[2]
	id := pushHandSizePermanent(g, me.ID, key, 100)
	for _, p := range g.Seats {
		if got := maxHandOf(g, p); got != DefaultMaxHandSize {
			t.Errorf("no player chosen yet, %s: %d, want 7", p.Name, got)
		}
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].ChosenPlayer = target.ID
			}
		}
	})
	for _, p := range g.Seats {
		want := DefaultMaxHandSize
		if p.ID == target.ID {
			want = 4
		}
		if got := maxHandOf(g, p); got != want {
			t.Errorf("%s: %d, want %d", p.Name, got, want)
		}
	}
}

// ADR 0113 §3 snapshot impact: the grant's timestamp survives a
// snapshot round trip.
func TestMaxHandSizeGrantTimestampRoundTrips(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { _ = g.SetMaxHandSizeForEffect(me.ID, NoMaxHandSize) })
	at := me.MaxHandSizeAt
	_, restored := roundTrip(t, g)
	var got *Player
	for _, p := range restored.Seats {
		if p.ID == me.ID {
			got = p
		}
	}
	if got == nil {
		t.Fatal("player missing after restore")
	}
	if got.MaxHandSize != NoMaxHandSize || got.MaxHandSizeAt != at {
		t.Errorf("restored grant = %d at %d, want %d at %d", got.MaxHandSize, got.MaxHandSizeAt, NoMaxHandSize, at)
	}
}
