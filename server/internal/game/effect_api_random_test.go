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
		choice = g.FlipCoinForEffect(CoinFlipSpec{Flipper: p.ID, Coins: 2, Then: func(_ *Game, r CoinFlipResult) error { got = r; return nil }})
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
		id = g.FlipCoinForEffect(CoinFlipSpec{Flipper: p.ID, AllowStop: true, Then: func(_ *Game, r CoinFlipResult) error { called = r.Stopped; return nil }})
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

func TestCoinCallSnapshotCensusesContinuation(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		g.FlipCoinForEffect(CoinFlipSpec{Flipper: p.ID, Then: func(*Game, CoinFlipResult) error { return nil }})
	})
	snap := g.CaptureSnapshot()
	if snap.Continuations.ChoiceResumeFrames != 1 || snap.Restorable() {
		t.Fatalf("coin continuation census = %#v, restorable=%v", snap.Continuations, snap.Restorable())
	}
	if got := snap.PendingChoices[0].CoinCount; got != 1 {
		t.Fatalf("snapshot coin count = %d, want 1", got)
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
		ids, err := g.SpawnCardsForDev(p.ID, ZoneHand, devTemplate(), 2)
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
		g.FlipCoinForEffect(CoinFlipSpec{Flipper: p.ID, AllowStop: true, Wins: 1, Then: chain})
		return nil
	}
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.FlipCoinForEffect(CoinFlipSpec{Flipper: p.ID, Then: chain}) })
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
	ids, err := g.SpawnCardsForDev(p.ID, ZoneBattlefield, devTemplate(), 1)
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
