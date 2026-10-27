package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestRandomEffectsEmitBatchedEventsAndRewind(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	var rolls, picks []int
	g.WithWriteLock(func() {
		var err error
		rolls, err = g.RollDiceForEffect(RandomDraw{Player: p.ID}, 20, 3)
		if err != nil {
			t.Fatalf("RollDiceForEffect: %v", err)
		}
		if _, err := g.FlipCoinsForEffect(RandomDraw{Player: p.ID}, 2); err != nil {
			t.Fatalf("FlipCoinsForEffect: %v", err)
		}
		picks = nil
		for _, id := range g.ChooseAtRandomForEffect(RandomDraw{Player: p.ID}, []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}, 2) {
			picks = append(picks, int(id[0]))
		}
	})
	if len(rolls) != 3 || len(picks) != 2 {
		t.Fatalf("results = %v picks=%v", rolls, picks)
	}
	var randomEvents []Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventRollDie || ev.Kind == EventFlipCoin {
				randomEvents = append(randomEvents, ev)
			}
		}
	})
	if len(randomEvents) != 5 || randomEvents[0].BatchSeq != randomEvents[0].Seq || randomEvents[1].BatchSeq != randomEvents[0].Seq || randomEvents[3].BatchSeq != randomEvents[3].Seq {
		t.Fatalf("bad batches: %#v", randomEvents)
	}
	if randomEvents[0].Sides != 20 {
		t.Fatalf("roll sides = %d", randomEvents[0].Sides)
	}

	before := g.Clone()
	var first, again []int
	g.WithWriteLock(func() { first, _ = g.RollDiceForEffect(RandomDraw{Player: p.ID}, 6, 2) })
	g.RestoreFrom(before)
	g.WithWriteLock(func() { again, _ = g.RollDiceForEffect(RandomDraw{Player: p.ID}, 6, 2) })
	if !reflect.DeepEqual(first, again) {
		t.Fatalf("undo roll = %v, redo = %v", first, again)
	}
}

func TestCoinCallUsesWonLossDrawAndValidatesAnswer(t *testing.T) {
	g := newActiveGame(t)
	p, other := g.Seats[0], g.Seats[1]
	g.SetRNGKeyForTest(testRNGKey())
	var got CoinFlipResult
	var choice uuid.UUID
	g.WithWriteLock(func() {
		choice = g.FlipCoinForEffect(coinFlipHooked(CoinFlipSpec{Flipper: p.ID, Coins: 2}, func(_ *Game, r CoinFlipResult) error { got = r; return nil }))
	})
	if err := g.ResolveCoinCall(choice, other.ID, "heads"); err != ErrNotTheChooser {
		t.Fatalf("wrong chooser = %v", err)
	}
	if err := g.ResolveCoinCall(choice, p.ID, "stop"); err != ErrInvalidParam {
		t.Fatalf("stop without permission = %v", err)
	}
	before := g.Clone()
	if err := g.ResolveCoinCall(choice, p.ID, "heads"); err != nil {
		t.Fatalf("heads: %v", err)
	}
	first := got
	g.RestoreFrom(before)
	var restored uuid.UUID
	g.ReadSnapshot(func() { restored = g.PendingChoices[0].ID })
	if err := g.ResolveCoinCall(restored, p.ID, "tails"); err != nil {
		t.Fatalf("tails: %v", err)
	}
	if !reflect.DeepEqual(first.Won, got.Won) {
		t.Fatalf("changed call changed win/loss: %v vs %v", first.Won, got.Won)
	}
	if reflect.DeepEqual(first.Faces, got.Faces) {
		t.Fatalf("changed call did not invert faces: %v", got.Faces)
	}
}

func TestCoinCallStopDoesNotDrawOrRunFaceResult(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	called := false
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.FlipCoinForEffect(coinFlipHooked(CoinFlipSpec{Flipper: p.ID, AllowStop: true}, func(_ *Game, r CoinFlipResult) error { called = r.Stopped; return nil }))
	})
	if err := g.ResolveCoinCall(id, p.ID, "stop"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("stop continuation did not receive stopped result")
	}
}

func TestSourcedRandomDrawUsesDeckOrdinalAcrossRuns(t *testing.T) {
	run := func() []int {
		g := NewGame()
		deck := []Card{NewCard("first", uuid.Nil), NewCard("source", uuid.Nil)}
		p, err := g.AddPlayer("one", deck)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := g.AddPlayer("two", []Card{NewCard("other", uuid.Nil)}); err != nil {
			t.Fatal(err)
		}
		g.SetRNGKeyForTest(testRNGKey())
		var got []int
		g.WithWriteLock(func() {
			got, err = g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: p.Library.Cards[0].InstanceID}, 20, 4)
		})
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	if a, b := run(), run(); !reflect.DeepEqual(a, b) {
		t.Fatalf("sourced rolls differ across runs: %v / %v", a, b)
	}
}

// TestCoinCallWithAContinuationIsARestorePoint (#4361): the closure form is
// gone, so a coin call with something hanging off it holds no frame. A
// strict restore mid-prompt keeps the call and its continuation, and the
// restored table answers into it.
func TestCoinCallWithAContinuationIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	var got []CoinFlipResult
	g.WithWriteLock(func() {
		g.FlipCoinForEffect(coinFlipHooked(CoinFlipSpec{Flipper: p.ID, Coins: 2}, func(_ *Game, r CoinFlipResult) error {
			got = append(got, r)
			return nil
		}))
	})
	snap := g.CaptureSnapshot()
	if snap.Continuations.ChoiceResumeFrames != 0 || !snap.Restorable() {
		t.Fatalf("coin continuation census = %#v, restorable=%v", snap.Continuations, snap.Restorable())
	}
	if got := snap.PendingChoices[0].CoinCount; got != 2 {
		t.Fatalf("snapshot coin count = %d, want 2", got)
	}
	restored := strictRestore(t, g)
	c := restored.PendingChoices
	if len(c) != 1 || c[0].Kind != PendingChoiceCoinCall || c[0].PickThen != testCoinHookThen.Key() {
		t.Fatalf("restored prompts = %+v, want the keyed coin call", c)
	}
	if err := restored.ResolveCoinCall(c[0].ID, p.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall on the restored table: %v", err)
	}
	if len(got) != 1 || len(got[0].Won) != 2 || got[0].Call != "heads" {
		t.Fatalf("the restored call's continuation ran with %+v, want once with two coins", got)
	}
}

func TestRandomEffectSnapshotRestoresSourcedStreams(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	draw := RandomDraw{Player: p.ID, Source: p.Hand.Cards[0].InstanceID}
	g.WithWriteLock(func() { _, _ = g.RollDiceForEffect(draw, 20, 2) })
	encoded, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err = json.Unmarshal(encoded, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.sourceOrdinals) == 0 {
		t.Fatal("source ordinals were not restored")
	}
	var original, again []int
	g.WithWriteLock(func() { original, _ = g.RollDiceForEffect(draw, 20, 20) })
	restored.WithWriteLock(func() { again, _ = restored.RollDiceForEffect(draw, 20, 20) })
	if !reflect.DeepEqual(original, again) {
		t.Fatalf("snapshot changed next draws: %v / %v", original, again)
	}
}

func TestRandomEffectsBoundsAndIndependentStreams(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	a, b := p.Hand.Cards[0].InstanceID, p.Hand.Cards[1].InstanceID
	before := g.Clone()
	g.WithWriteLock(func() {
		count := len(g.Events)
		for _, args := range [][2]int{{0, 1}, {-1, 1}, {6, 0}, {6, -1}} {
			if _, err := g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: a}, args[0], args[1]); err != ErrInvalidParam {
				t.Fatal("invalid dice accepted")
			}
		}
		if _, err := g.FlipCoinsForEffect(RandomDraw{Player: p.ID}, 0); err != ErrInvalidParam {
			t.Fatal("invalid coin count accepted")
		}
		if len(g.Events) != count || !reflect.DeepEqual(g.rngCounters, before.rngCounters) {
			t.Fatal("invalid draw changed game")
		}
		ids := []uuid.UUID{a, a, b}
		original := append([]uuid.UUID(nil), ids...)
		got := g.ChooseAtRandomForEffect(RandomDraw{Player: p.ID, Source: a}, ids, 5)
		if len(got) != 2 || got[0] == got[1] || !reflect.DeepEqual(ids, original) {
			t.Fatalf("pick must be distinct without changing input: %v", got)
		}
		_, _ = g.FlipCoinsForEffect(RandomDraw{Player: p.ID, Source: a}, 32)
		_, _ = g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: b}, 20, 32)
	})
	var rolls, isolated []int
	g.WithWriteLock(func() { rolls, _ = g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: a}, 20, 32) })
	before.WithWriteLock(func() { isolated, _ = before.RollDiceForEffect(RandomDraw{Player: p.ID, Source: a}, 20, 32) })
	if !reflect.DeepEqual(rolls, isolated) {
		t.Fatal("another kind or source perturbed a roll stream")
	}
}

func TestRuntimeRandomSourcesAreAssignedAtCreationAndPersist(t *testing.T) {
	run := func(reverse bool) ([]int, []int) {
		g := newActiveGame(t)
		p := g.Seats[0]
		g.SetRNGKeyForTest(testRNGKey())
		ids, err := g.SpawnCards(uuid.Nil, p.ID, ZoneHand, devTemplate(), 2)
		if err != nil {
			t.Fatal(err)
		}
		if g.sourceOrdinals[ids[0]] == g.sourceOrdinals[ids[1]] || g.sourceOrdinalNext != 2 {
			t.Fatal("runtime ordinals not assigned at creation")
		}
		var a, b []int
		draw := func(id uuid.UUID) []int {
			v, e := g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: id}, 6, 16)
			if e != nil {
				t.Fatal(e)
			}
			return v
		}
		g.WithWriteLock(func() {
			if reverse {
				b = draw(ids[1])
				a = draw(ids[0])
			} else {
				a = draw(ids[0])
				b = draw(ids[1])
			}
		})
		snapshot := g.CaptureSnapshot()
		restored, err := snapshot.RestoreStrict()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(g.sourceOrdinals, restored.sourceOrdinals) || restored.sourceOrdinalNext != 2 {
			t.Fatal("runtime ordinal snapshot mismatch")
		}
		return a, b
	}
	a, b := run(false)
	c, d := run(true)
	if !reflect.DeepEqual(a, c) || !reflect.DeepEqual(b, d) {
		t.Fatal("first-use order changed runtime source identity")
	}
}

func TestCoinStopPreservesNextDrawAndClonedChain(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	var chain func(*Game, CoinFlipResult) error
	chain = func(g *Game, r CoinFlipResult) error {
		if r.Stopped {
			if r.Call != "" || len(r.Faces) != 0 || len(r.Won) != 0 {
				t.Fatal("stop carried an outcome")
			}
			return nil
		}
		g.FlipCoinForEffect(coinFlipHooked(CoinFlipSpec{Flipper: p.ID, AllowStop: true, Wins: 1}, chain))
		return nil
	}
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.FlipCoinForEffect(coinFlipHooked(CoinFlipSpec{Flipper: p.ID}, chain)) })
	if err := g.ResolveCoinCall(id, p.ID, "heads"); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatal("chain did not resume")
	}
	before := g.Clone()
	counters := cloneRNGCounters(g.rngCounters)
	if err := g.ResolveCoinCall(g.PendingChoices[0].ID, p.ID, "stop"); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 0 || !reflect.DeepEqual(counters, g.rngCounters) {
		t.Fatal("stop consumed randomness")
	}
	g.RestoreFrom(before)
	if err := g.ResolveCoinCall(g.PendingChoices[0].ID, p.ID, "tails"); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatal("cloned continuation was lost")
	}
}

func TestBlinkPreservesRandomSourceStream(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	ids, err := g.SpawnCards(uuid.Nil, p.ID, ZoneBattlefield, devTemplate(), 1)
	if err != nil {
		t.Fatal(err)
	}
	original := ids[0]
	g.WithWriteLock(func() { _, _ = g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: original}, 20, 1) })
	before := g.Clone()
	var next uuid.UUID
	var blinked, unblinked []int
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(original); err != nil {
			t.Fatal(err)
		}
		var err error
		next, err = g.ReturnFromExileToBattlefieldForEffect(original, p.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		blinked, _ = g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: next}, 20, 16)
	})
	before.WithWriteLock(func() { unblinked, _ = before.RollDiceForEffect(RandomDraw{Player: p.ID, Source: original}, 20, 16) })
	if next == original || !reflect.DeepEqual(blinked, unblinked) {
		t.Fatal("blink changed random stream identity")
	}
}

func TestDeckRandomOrdinalsFollowLobbyReseating(t *testing.T) {
	g := NewGame()
	first, err := g.AddPlayer("first", []Card{NewCard("first", uuid.Nil)})
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.AddPlayer("second", []Card{NewCard("second", uuid.Nil)})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.RemovePlayer(first.ID); err != nil {
		t.Fatal(err)
	}
	third, err := g.AddPlayer("third", []Card{NewCard("third", uuid.Nil)})
	if err != nil {
		t.Fatal(err)
	}
	a := g.sourceOrdinals[second.Library.Cards[0].InstanceID]
	b := g.sourceOrdinals[third.Library.Cards[0].InstanceID]
	if a != 0 || b != 1<<16 {
		t.Fatalf("reused deck source ordinal after reseating: %d / %d", a, b)
	}
}

// CR 706.6 (#4174): an ignored roll "is considered to have never
// happened". The lowest of two d20 is dropped before any event exists,
// so exactly one EventRollDie is emitted (the kept, highest one), and the
// result is the larger of the two draws.
func TestRollDiceIgnoringLowestEmitsOnlyTheKeptRoll(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())

	// The control: the same stream and key, both dice announced.
	control := g.Clone()
	var both []int
	control.WithWriteLock(func() { both, _ = control.RollDiceForEffect(RandomDraw{Player: p.ID}, 20, 2) })
	if len(both) != 2 {
		t.Fatalf("control rolled %v", both)
	}
	want := both[0]
	if both[1] > want {
		want = both[1]
	}

	var kept []int
	g.WithWriteLock(func() {
		var err error
		kept, err = g.RollDiceIgnoringLowestForEffect(RandomDraw{Player: p.ID}, 20, 2, 1)
		if err != nil {
			t.Fatalf("RollDiceIgnoringLowestForEffect: %v", err)
		}
	})
	if len(kept) != 1 || kept[0] != want {
		t.Fatalf("kept %v, want [%d] (the higher of %v)", kept, want, both)
	}
	var events []Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventRollDie {
				events = append(events, ev)
			}
		}
	})
	if len(events) != 1 || events[0].Amount != want || events[0].Sides != 20 || events[0].BatchSeq != events[0].Seq {
		t.Fatalf("events = %#v, want one d20 event of %d", events, want)
	}
}

// The ignore count validates, and the dice kept keep their rolled order.
func TestRollDiceIgnoringLowestValidatesAndKeepsOrder(t *testing.T) {
	g := newActiveGame(t)
	g.SetRNGKeyForTest(testRNGKey())
	g.WithWriteLock(func() {
		for _, bad := range []struct{ sides, n, ignore int }{{20, 2, 2}, {20, 2, -1}, {20, 0, 0}, {0, 2, 1}} {
			if _, err := g.RollDiceIgnoringLowestForEffect(RandomDraw{Player: g.Seats[0].ID}, bad.sides, bad.n, bad.ignore); err != ErrInvalidParam {
				t.Errorf("%+v: err = %v, want ErrInvalidParam", bad, err)
			}
		}
	})
	if got := keepAllButLowest([]int{5, 2, 9, 2}, 1); !reflect.DeepEqual(got, []int{5, 9, 2}) {
		t.Errorf("drop one of a tie for lowest: %v", got)
	}
	if got := keepAllButLowest([]int{5, 2, 9, 7}, 2); !reflect.DeepEqual(got, []int{9, 7}) {
		t.Errorf("drop two: %v", got)
	}
}

// Ignoring a roll draws the same number of dice, so it replays exactly
// after an undo and leaves the stream where an unignored pair would.
func TestRollDiceIgnoringLowestIsReproducibleOnRestore(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	before := g.Clone()
	var first, again, next1, next2 []int
	g.WithWriteLock(func() {
		first, _ = g.RollDiceIgnoringLowestForEffect(RandomDraw{Player: p.ID}, 20, 2, 1)
		next1, _ = g.RollDiceForEffect(RandomDraw{Player: p.ID}, 20, 1)
	})
	g.RestoreFrom(before)
	g.WithWriteLock(func() {
		again, _ = g.RollDiceIgnoringLowestForEffect(RandomDraw{Player: p.ID}, 20, 2, 1)
		next2, _ = g.RollDiceForEffect(RandomDraw{Player: p.ID}, 20, 1)
	})
	if !reflect.DeepEqual(first, again) || !reflect.DeepEqual(next1, next2) {
		t.Fatalf("undo roll = %v then %v, redo = %v then %v", first, next1, again, next2)
	}
}
