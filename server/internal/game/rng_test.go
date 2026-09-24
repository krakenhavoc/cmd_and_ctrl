package game

import (
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// rng_test.go covers ADR 0054 sub-PR 1: keyed streams, their rewind
// through Clone / RestoreFrom, and every draw site moved onto them.
// Snapshot persistence is in snapshot_test.go.

// testRNGKey is a fixed, non-zero key for known-answer tests.
func testRNGKey() [32]byte {
	var k [32]byte
	for i := range k {
		k[i] = byte(i + 1)
	}
	return k
}

var (
	rngTestPlayer = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	rngTestOgre   = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	rngTestPuzzle = uuid.MustParse("33333333-3333-4333-8333-333333333333")
)

// libraryNames lists a player's library top to bottom by name. Names
// rather than instance IDs so two separate games with the same deck
// can be compared.
func libraryNames(p *Player) []string {
	out := make([]string, len(p.Library.Cards))
	for i, c := range p.Library.Cards {
		out[i] = c.Name
	}
	return out
}

func handNames(p *Player) []string {
	out := make([]string, len(p.Hand.Cards))
	for i, c := range p.Hand.Cards {
		out[i] = c.Name
	}
	return out
}

func graveyardNames(p *Player) []string {
	out := make([]string, len(p.Graveyard.Cards))
	for i, c := range p.Graveyard.Cards {
		out[i] = c.Name
	}
	return out
}

// TestRNGKnownAnswer pins the derivation. A silent change to rngSeed
// (the domain string, the field order, the encoding) or to the ChaCha8
// wrapping would change what every restored game draws next, so it
// must be a deliberate, visible edit of this test.
func TestRNGKnownAnswer(t *testing.T) {
	seed := rngSeed(testRNGKey(), "roll", [16]byte(rngTestPlayer), rngTestOgre, 3, 0)
	if got, want := hex.EncodeToString(seed[:]), rngKnownSeedHex; got != want {
		t.Errorf("rngSeed = %s, want %s", got, want)
	}

	g := NewGame()
	g.SetRNGKeyForTest(testRNGKey())
	g.Turn.Seq = 3
	g.Turn.ActiveSeat = 3
	var got []uint64
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			got = append(got, g.randForLocked(rngStream{kind: "roll", player: rngTestPlayer, source: rngTestOgre}).Uint64())
		}
	})
	if !reflect.DeepEqual(got, rngKnownDraws) {
		t.Errorf("first Uint64 of draws 0..2 = %#v, want %#v", got, rngKnownDraws)
	}
}

// Pinned by TestRNGKnownAnswer. Change only on purpose.
var (
	rngKnownSeedHex = "1f7599ae610947e8bbc25c66e09ad6e5a5ec987ef58ca878ebacc4656fe68f91"
	rngKnownDraws   = []uint64{0xc649bbf894e6761a, 0x314f8bad34e1dca6, 0x30c23d5f49ded4a7}
)

// TestRNGSameKeySameResults: two games with the same key, driven
// through the same actions, shuffle and discard identically. The
// players are different UUIDs in each game; the seat is what keys the
// stream (rng.go), which is what keeps a seeded test deterministic
// from run to run.
func TestRNGSameKeySameResults(t *testing.T) {
	run := func() (opening, reshuffled, discarded []string) {
		g := newActiveGame(t) // Start(rand.New(rand.NewPCG(1, 2)))
		p := g.Seats[0]
		opening = libraryNames(p)
		if err := g.ShuffleLibrary(p.ID); err != nil {
			t.Fatalf("ShuffleLibrary: %v", err)
		}
		reshuffled = libraryNames(p)
		g.WithWriteLock(func() {
			if err := g.DiscardRandomForEffect(p.ID, 2); err != nil {
				t.Fatalf("DiscardRandomForEffect: %v", err)
			}
		})
		return opening, reshuffled, graveyardNames(p)
	}
	o1, r1, d1 := run()
	o2, r2, d2 := run()
	if !reflect.DeepEqual(o1, o2) {
		t.Error("same seed, different opening library order")
	}
	if !reflect.DeepEqual(r1, r2) {
		t.Error("same seed, different reshuffle")
	}
	if !reflect.DeepEqual(d1, d2) {
		t.Errorf("same seed, different random discard: %v vs %v", d1, d2)
	}
	if reflect.DeepEqual(o1, r1) {
		t.Error("a reshuffle left the library in the opening order")
	}
}

// TestRNGDifferentSeedsDiffer is the negative case: a different seed
// gives a different opening order.
func TestRNGDifferentSeedsDiffer(t *testing.T) {
	mk := func(seed uint64) []string {
		g := NewGame()
		for i := 0; i < 2; i++ {
			if _, err := g.AddPlayer("P", buildTestDeck("C")); err != nil {
				t.Fatalf("AddPlayer: %v", err)
			}
		}
		if err := g.Start(rand.New(rand.NewPCG(seed, seed))); err != nil {
			t.Fatalf("Start: %v", err)
		}
		return libraryNames(g.Seats[0])
	}
	if reflect.DeepEqual(mk(1), mk(2)) {
		t.Error("two seeds dealt the same opening library")
	}
}

// TestRNGSeatsHaveTheirOwnStreams: the same key does not give every
// seat the same shuffle.
func TestRNGSeatsHaveTheirOwnStreams(t *testing.T) {
	g := newActiveGame(t)
	if reflect.DeepEqual(libraryNames(g.Seats[0]), libraryNames(g.Seats[1])) {
		t.Error("both seats' identical decks were shuffled into the same order")
	}
}

// TestRNGStreamIndependence is the fishing scenario from ADR 0054's
// Context: draws on other streams — another source's rolls, a shuffle
// — do not move a stream's next value.
func TestRNGStreamIndependence(t *testing.T) {
	ogre := rngStream{kind: "roll", player: rngTestPlayer, source: rngTestOgre}
	fresh := func() *Game {
		g := NewGame()
		g.SetRNGKeyForTest(testRNGKey())
		return g
	}

	a := fresh()
	var want uint64
	a.WithWriteLock(func() { want = a.randForLocked(ogre).Uint64() })

	b := fresh()
	var got uint64
	b.WithWriteLock(func() {
		for i := 0; i < 5; i++ {
			b.randForLocked(rngStream{kind: "roll", player: rngTestPlayer, source: rngTestPuzzle}).Uint64()
			b.randForLocked(rngStream{kind: rngStreamShuffle, player: rngTestPlayer}).Uint64()
		}
		got = b.randForLocked(ogre).Uint64()
	})
	if got != want {
		t.Errorf("draws on other streams moved the Ogre's roll: %d, want %d", got, want)
	}

	// Negative: the SAME stream does move.
	b.WithWriteLock(func() { got = b.randForLocked(ogre).Uint64() })
	if got == want {
		t.Error("a stream's second draw repeated its first")
	}
}

// TestRNGFishingIsRewound is test plan item 6: roll for source A,
// restore, draw on source B, roll for A again — same value.
func TestRNGFishingIsRewound(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0].ID
	ogre := rngStream{kind: "roll", player: p, source: rngTestOgre}
	pre := g.Clone()
	var first, second uint64
	g.WithWriteLock(func() {
		first = g.randForLocked(ogre).Uint64()
		g.RestoreFrom(pre)
		g.randForLocked(rngStream{kind: "roll", player: p, source: rngTestPuzzle}).Uint64()
		if err := g.ShuffleLibraryForEffect(p); err != nil {
			t.Fatalf("ShuffleLibraryForEffect: %v", err)
		}
		second = g.randForLocked(ogre).Uint64()
	})
	if first != second {
		t.Errorf("undo + a draw elsewhere changed the roll: %d then %d", first, second)
	}
}

// TestRNGTurnScoping: the counters belong to one turn. On the next turn
// they are cleared, and the same stream gives a different value even
// at counter 0. "Next turn" includes the next SEAT's turn in the same
// round: Turn.Round only counts rounds.
func TestRNGTurnScoping(t *testing.T) {
	g := NewGame()
	g.SetRNGKeyForTest(testRNGKey())
	s := rngStream{kind: "roll", player: rngTestPlayer, source: rngTestOgre}
	var t1, t2 uint64
	g.WithWriteLock(func() {
		g.Turn.Seq = 1
		t1 = g.randForLocked(s).Uint64()
		g.randForLocked(rngStream{kind: rngStreamPick, player: rngTestPlayer}).Uint64()
		if len(g.rngCounters) != 2 {
			t.Errorf("turn 1 counters = %v, want two streams", g.rngCounters)
		}
		g.Turn.Seq = 2
		t2 = g.randForLocked(s).Uint64()
		if want := 2; g.rngTurn != want {
			t.Errorf("rngTurn = %d, want %d", g.rngTurn, want)
		}
		if len(g.rngCounters) != 1 {
			t.Errorf("turn 2 counters = %v, want only the stream drawn this turn", g.rngCounters)
		}
	})
	if t1 == t2 {
		t.Error("the same stream gave the same first value on two turns")
	}

	// Same round, next seat's turn: also a new turn.
	var t3 uint64
	g.WithWriteLock(func() {
		g.Turn.ActiveSeat = 1
		g.Turn.Seq++
		t3 = g.randForLocked(s).Uint64()
		if len(g.rngCounters) != 1 {
			t.Errorf("seat 1's turn counters = %v, want only the stream drawn this turn", g.rngCounters)
		}
	})
	if t3 == t2 {
		t.Error("the next seat's turn in the same round reused the stream's value")
	}

	// Negative: a second draw in the SAME turn does not reset.
	g.WithWriteLock(func() {
		g.randForLocked(s).Uint64()
		if n := g.rngCounters["roll/"+rngTestPlayer.String()+"/"+rngTestOgre.String()]; n != 2 {
			t.Errorf("counter after two draws in one turn = %d, want 2", n)
		}
	})
}

// TestRNGLazyKey: a game that never called Start still draws from a
// real, persistable key — minted on the first draw — and two such
// games do not share it.
func TestRNGLazyKey(t *testing.T) {
	a, b := NewGame(), NewGame()
	var x, y uint64
	a.WithWriteLock(func() { x = a.randForLocked(rngStream{kind: rngStreamPick}).Uint64() })
	b.WithWriteLock(func() { y = b.randForLocked(rngStream{kind: rngStreamPick}).Uint64() })
	if a.rngKey == ([32]byte{}) {
		t.Fatal("first draw did not mint a key")
	}
	if a.rngKey == b.rngKey || x == y {
		t.Error("two unstarted games minted the same key")
	}
	if got := a.CaptureSnapshot().RNG.Kind; got != rngKindKeyed {
		t.Errorf("rng kind after a lazy mint = %q, want %q", got, rngKindKeyed)
	}
}

func TestSetRNGKeyForTestRejectsZeroKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("SetRNGKeyForTest accepted the all-zero sentinel key")
		}
	}()
	NewGame().SetRNGKeyForTest([32]byte{})
}

// TestStartKeepsATestKey: Start(nil) keeps a key set before it, and
// Start with a source replaces it.
func TestStartKeepsATestKey(t *testing.T) {
	mk := func() *Game {
		g := NewGame()
		for i := 0; i < 2; i++ {
			if _, err := g.AddPlayer("P", buildTestDeck("C")); err != nil {
				t.Fatalf("AddPlayer: %v", err)
			}
		}
		g.SetRNGKeyForTest(testRNGKey())
		return g
	}
	g := mk()
	if err := g.Start(nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.rngKey != testRNGKey() {
		t.Error("Start(nil) replaced a key set by SetRNGKeyForTest")
	}
	h := mk()
	if err := h.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if h.rngKey == testRNGKey() {
		t.Error("Start with a source kept the test key instead of deriving one")
	}
}

// TestPickAtRandomIsUniform: under a fixed key, each of five IDs is the
// single pick about a fifth of the time. Deterministic, so it cannot
// flake; it catches an off-by-one that never picks the last element.
func TestPickAtRandomIsUniform(t *testing.T) {
	g := NewGame()
	g.SetRNGKeyForTest(testRNGKey())
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	counts := map[uuid.UUID]int{}
	const trials = 20000
	g.WithWriteLock(func() {
		for i := 0; i < trials; i++ {
			got := g.pickAtRandomLocked(rngStream{kind: rngStreamPick, player: rngTestPlayer}, ids, 1)
			if len(got) != 1 {
				t.Fatalf("pick of 1 returned %d", len(got))
			}
			counts[got[0]]++
		}
		if got := g.pickAtRandomLocked(rngStream{kind: rngStreamPick}, ids, 9); len(got) != len(ids) {
			t.Errorf("pick of 9 from 5 returned %d, want all 5", len(got))
		}
		if got := g.pickAtRandomLocked(rngStream{kind: rngStreamPick}, ids, 0); got != nil {
			t.Errorf("pick of 0 returned %v, want nil", got)
		}
	})
	for _, id := range ids {
		if c := counts[id]; c < trials/5-trials/50 || c > trials/5+trials/50 {
			t.Errorf("id picked %d of %d times, want %d ± %d", c, trials, trials/5, trials/50)
		}
	}
}

// ---- Rewind at every draw site (test plan item 5) ----

// rewinds runs act, restores the pre-act clone, runs act again, and
// requires the same result. act runs without the lock held, and must
// look its player up again each time: RestoreFrom adopts the clone's
// Player objects, so a *Player captured before the restore is stale.
// (The negative — a new draw without an undo differs — is
// TestRNGStreamIndependence's last check and TestUndoRewindsAShuffle
// in internal/ws.)
func rewinds[T any](t *testing.T, g *Game, act func() T) {
	t.Helper()
	pre := g.Clone()
	first := act()
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	second := act()
	if !reflect.DeepEqual(first, second) {
		t.Errorf("undo then redo drew differently:\nfirst:  %v\nsecond: %v", first, second)
	}
}

func TestRewindOpeningShuffle(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer("P", buildTestDeck("C")); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	g.SetRNGKeyForTest(testRNGKey())
	start := func() []string {
		if err := g.Start(nil); err != nil {
			t.Fatalf("Start: %v", err)
		}
		return append(libraryNames(g.Seats[0]), handNames(g.Seats[1])...)
	}
	pre := g.Clone()
	first := start()
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if second := start(); !reflect.DeepEqual(first, second) {
		t.Errorf("undo of Start then Start again dealt differently:\nfirst:  %v\nsecond: %v", first, second)
	}
}

func TestRewindShuffleLibrary(t *testing.T) {
	g := newActiveGame(t)
	rewinds(t, g, func() []string {
		p := g.Seats[0]
		if err := g.ShuffleLibrary(p.ID); err != nil {
			t.Fatalf("ShuffleLibrary: %v", err)
		}
		return libraryNames(p)
	})
}

func TestRewindShuffleLibraryForEffect(t *testing.T) {
	g := newActiveGame(t)
	rewinds(t, g, func() []string {
		p := g.Seats[1]
		g.WithWriteLock(func() {
			if err := g.ShuffleLibraryForEffect(p.ID); err != nil {
				t.Fatalf("ShuffleLibraryForEffect: %v", err)
			}
		})
		return libraryNames(p)
	})
}

func TestRewindSearchShuffle(t *testing.T) {
	g := newActiveGame(t)
	rewinds(t, g, func() []string {
		p := g.Seats[0]
		g.WithWriteLock(func() {
			spec := SearchLibrarySpec{Player: p.ID, Shuffle: true}
			if err := g.finishSearchLocked(spec, p, nil); err != nil {
				t.Fatalf("finishSearchLocked: %v", err)
			}
		})
		return libraryNames(p)
	})
}

func TestRewindMulligan(t *testing.T) {
	g := newActiveGameMulligansOpen(t, 2)
	rewinds(t, g, func() []string {
		p := g.Seats[0]
		if err := g.Mulligan(p.ID, 7); err != nil {
			t.Fatalf("Mulligan: %v", err)
		}
		return handNames(p)
	})
}

func TestRewindDiscardRandom(t *testing.T) {
	g := newActiveGame(t)
	rewinds(t, g, func() []string {
		p := g.Seats[0]
		g.WithWriteLock(func() {
			if err := g.DiscardRandomForEffect(p.ID, 3); err != nil {
				t.Fatalf("DiscardRandomForEffect: %v", err)
			}
		})
		return graveyardNames(p)
	})
}

func TestRewindCascadeRandomBottom(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; i < 6; i++ {
			if _, err := MoveCard(p.Library, g.Exile, p.Library.Cards[0].InstanceID); err != nil {
				t.Fatalf("MoveCard: %v", err)
			}
		}
	})
	ids := make([]uuid.UUID, len(g.Exile.Cards))
	for i, c := range g.Exile.Cards {
		ids[i] = c.InstanceID
	}
	rewinds(t, g, func() []string {
		p := g.Seats[0]
		g.WithWriteLock(func() {
			if err := g.PutOnBottomInRandomOrderForEffect(p.ID, ZoneExile, ids); err != nil {
				t.Fatalf("PutOnBottomInRandomOrderForEffect: %v", err)
			}
		})
		// The bottom of a zone is Cards[0] (Zone.PushBottom).
		return libraryNames(p)[:len(ids)]
	})
}

// TestRewindPutOnBottomInRandomOrder: #745's public random-order
// bottom draws on the actor's "random_order" stream (ADR 0054
// Decision 8). Undo-then-redo bottoms the same cards in the same
// order, for a reorder within the library (no zone change) as well as
// a move from exile, and a draw on another stream in between (a
// shuffle of the other seat's library, a random discard) does not move
// it: the fishing the keyed streams exist to close.
func TestRewindPutOnBottomInRandomOrder(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		// Five cards off the top stay in the library (a reorder);
		// five more go to exile first (a zone change).
		for i := 0; i < 5; i++ {
			ids = append(ids, p.Library.Cards[len(p.Library.Cards)-1-i].InstanceID)
		}
		for i := 0; i < 5; i++ {
			id := p.Library.Cards[len(p.Library.Cards)-6].InstanceID
			if _, err := MoveCard(p.Library, g.Exile, id); err != nil {
				t.Fatalf("MoveCard: %v", err)
			}
			ids = append(ids, id)
		}
	})
	bottom := func() []string {
		// Re-read the seat: RestoreFrom replaces the Player values, so
		// a pointer taken before a restore reads a stale library.
		p := g.Seats[0]
		g.WithWriteLock(func() {
			// The bottom takes one source zone per call, so the cards
			// are split by where they are now: the ones still in the
			// library (a reorder), then the ones in exile (a zone
			// change). After the first bottom they are all in the
			// library.
			var inLibrary, inExile []uuid.UUID
			for _, id := range ids {
				switch z := g.findCardZoneLocked(id); {
				case z == nil:
				case z.Kind == ZoneLibrary:
					inLibrary = append(inLibrary, id)
				case z.Kind == ZoneExile:
					inExile = append(inExile, id)
				}
			}
			if err := g.PutOnBottomInRandomOrderForEffect(p.ID, ZoneLibrary, inLibrary); err != nil {
				t.Fatalf("PutOnBottomInRandomOrderForEffect(library): %v", err)
			}
			if err := g.PutOnBottomInRandomOrderForEffect(p.ID, ZoneExile, inExile); err != nil {
				t.Fatalf("PutOnBottomInRandomOrderForEffect(exile): %v", err)
			}
		})
		// The bottom of a zone is Cards[0] (Zone.PushBottom).
		return libraryNames(p)[:len(ids)]
	}
	rewinds(t, g, bottom)

	pre := g.Clone()
	first := bottom()
	// Not vacuous: the next draw on the same stream orders the same
	// ten cards differently.
	if again := bottom(); reflect.DeepEqual(first, again) {
		t.Fatalf("two draws on the random_order stream gave the same order %v; the test cannot see the stream", first)
	}
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if err := g.ShuffleLibrary(g.Seats[1].ID); err != nil {
		t.Fatalf("ShuffleLibrary: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(g.Seats[0].ID, 1); err != nil {
			t.Fatalf("DiscardRandomForEffect: %v", err)
		}
	})
	if second := bottom(); !reflect.DeepEqual(first, second) {
		t.Errorf("a draw on another stream moved the random-order bottom:\nfirst:  %v\nsecond: %v", first, second)
	}
}

// TestDiscardRandomDoesNotAlwaysTakeTheFirstCard pins the removal of
// the old "index 0 when the game has no RNG" branch: a game that never
// called Start still discards at random.
func TestDiscardRandomDoesNotAlwaysTakeTheFirstCard(t *testing.T) {
	g := NewGame()
	p, err := g.AddPlayer("P", buildTestDeck("C"))
	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	g.SetRNGKeyForTest(testRNGKey())
	notFirst := 0
	for i := 0; i < 20; i++ {
		g.WithWriteLock(func() {
			g.Turn.Seq = i // a fresh turn: counter 0 on a new stream each time
			for p.Hand.Size() < 5 {
				if _, err := MoveCard(p.Library, p.Hand, p.Library.Cards[0].InstanceID); err != nil {
					t.Fatalf("MoveCard: %v", err)
				}
			}
			first := p.Hand.Cards[0].InstanceID
			if err := g.DiscardRandomForEffect(p.ID, 1); err != nil {
				t.Fatalf("DiscardRandomForEffect: %v", err)
			}
			if p.Hand.Contains(first) {
				notFirst++
			}
		})
	}
	if notFirst == 0 {
		t.Error("20 random discards all took the first card in hand")
	}
}

// TestDiscardRandomEdges: fewer cards than n discards the whole hand,
// and n <= 0 discards nothing.
func TestDiscardRandomEdges(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(p.ID, 0); err != nil {
			t.Fatalf("n=0: %v", err)
		}
		if p.Graveyard.Size() != 0 {
			t.Errorf("n=0 discarded %d cards", p.Graveyard.Size())
		}
		hand := p.Hand.Size()
		if err := g.DiscardRandomForEffect(p.ID, hand+3); err != nil {
			t.Fatalf("n>hand: %v", err)
		}
		if p.Hand.Size() != 0 || p.Graveyard.Size() != hand {
			t.Errorf("n>hand left hand=%d graveyard=%d, want 0 and %d", p.Hand.Size(), p.Graveyard.Size(), hand)
		}
		if err := g.DiscardRandomForEffect(uuid.New(), 1); err != ErrPlayerNotFound {
			t.Errorf("unknown player: %v, want ErrPlayerNotFound", err)
		}
	})
}

// TestNoDirectRandomSource keeps ADR 0054 Decision 1 true: no file in
// internal/game calls into math/rand except rng.go, and zone.go only
// inside Zone.Shuffle (the nil fallback kept for zone-level unit
// tests). Naming the rand.Rand / rand.PCG types in a signature is
// allowed. No game path passes Shuffle a nil source,
// and nothing but rng.go reads crypto/rand. Test files are exempt.
func TestNoDirectRandomSource(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") || name == "rng.go" {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		randNames := map[string]string{}
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			switch path {
			case "math/rand", "math/rand/v2", "crypto/rand":
			default:
				continue
			}
			local := "rand"
			if imp.Name != nil {
				local = imp.Name.Name
			}
			if path == "crypto/rand" {
				t.Errorf("%s imports crypto/rand; mint keys in rng.go", name)
				continue
			}
			randNames[local] = path
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncDecl:
				if name == "zone.go" && n.Name.Name == "Shuffle" && n.Recv != nil {
					return false // the one sanctioned use
				}
			case *ast.SelectorExpr:
				// Naming the TYPES is fine: Start and StartWithSource
				// take a *rand.Rand / *rand.PCG to derive the key from.
				// Calling anything in the package is not.
				switch n.Sel.Name {
				case "Rand", "PCG", "Source":
					return true
				}
				if id, ok := n.X.(*ast.Ident); ok {
					if path, ok := randNames[id.Name]; ok {
						t.Errorf("%s: %s uses %s.%s; draw through randForLocked (rng.go)", fset.Position(n.Pos()), name, path, n.Sel.Name)
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Shuffle" && len(n.Args) == 1 {
					if id, ok := n.Args[0].(*ast.Ident); ok && id.Name == "nil" {
						t.Errorf("%s: Shuffle(nil) in a game path; pass randForLocked's generator", fset.Position(n.Pos()))
					}
				}
			}
			return true
		})
	}
}
