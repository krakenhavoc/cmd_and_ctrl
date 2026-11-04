package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_counters_test.go — #4442, ADR 0209 §3: the client's counter cue
// pulses a permanent's power/toughness when +1/+1 or -1/-1 counters are
// put on it, which needs to know a placement from a removal. The
// engine's event carries only the total after the change, so the
// `counters` entry gains `delta`, the signed change, measured by the
// log fold from the card's previous total. These pin that delta, that
// every viewer who may identify the card gets it, and that counters put
// on by one effect share a batch, which is how the cue plays one sound
// for a proliferate.

// countersOf is the card's `counters` entries in log order.
func countersOf(log []LogEvent, card uuid.UUID) []LogEvent {
	var out []LogEvent
	for _, e := range log {
		if e.Kind == LogCounters && e.CardID == card.String() {
			out = append(out, e)
		}
	}
	return out
}

type countLine struct {
	label         string
	amount, delta int
}

func countLines(es []LogEvent) []countLine {
	out := make([]countLine, len(es))
	for i, e := range es {
		out[i] = countLine{e.Label, e.Amount, e.Delta}
	}
	return out
}

func TestACountersEntryCarriesItsSignedChange(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	bear := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", owner.ID)
	add := func(kind string, n int) {
		t.Helper()
		if err := g.AddCounter(bear, kind, n); err != nil {
			t.Fatalf("AddCounter(%s, %d): %v", kind, n, err)
		}
		// A view between each change: the fold resumes, and its counts
		// must carry over.
		_ = ViewOfGame(g)
	}
	add(game.CounterPlusOne, 2)
	add(game.CounterPlusOne, 1)
	add(game.CounterPlusOne, -1)
	add(game.CounterMinusOne, 1)
	// CR 704.5q: the -1/-1 counter and a +1/+1 counter are removed.
	g.RunStateChecksForTest()

	want := []countLine{
		{game.CounterPlusOne, 2, 2},
		{game.CounterPlusOne, 3, 1},
		{game.CounterPlusOne, 2, -1},
		{game.CounterMinusOne, 1, 1},
		{game.CounterPlusOne, 1, -1},
		{game.CounterMinusOne, 0, -1},
	}
	// An opponent reads the same changes the controller does.
	got := countLines(countersOf(FilterViewFor(ViewOfGame(g), other.ID.String()).Log, bear))
	if len(got) != len(want) {
		t.Fatalf("got %d counters entries %v, want %v", len(got), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d: got %+v, want %+v", i, got[i], want[i])
		}
	}

	// A fresh fold over the same events agrees with the resumed one.
	viewAndCheck(t, g, "after the counters")

	// The name on the wire.
	buf, err := json.Marshal(countersOf(ViewOfGame(g).Log, bear)[2])
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(buf, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["delta"] != float64(-1) || raw["amount"] != float64(2) || raw["label"] != game.CounterPlusOne {
		t.Errorf("wire entry %s, want delta -1, amount 2, label +1/+1", buf)
	}
}

// A permanent that leaves the battlefield is a new object when it comes
// back (CR 400.7): its first counter is a placement of one, never a
// change from the count the old object had.
func TestACountersDeltaStartsAgainAfterAZoneChange(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	bear := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", owner.ID)
	if err := g.AddCounter(bear, game.CounterPlusOne, 3); err != nil {
		t.Fatal(err)
	}
	bf := game.ZoneRef{Kind: game.ZoneBattlefield}
	gy := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner.ID}
	if err := g.MoveCardByID(bf, gy, bear); err != nil {
		t.Fatalf("to the graveyard: %v", err)
	}
	if err := g.MoveCardByID(gy, bf, bear); err != nil {
		t.Fatalf("back: %v", err)
	}
	if err := g.AddCounter(bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	es := countersOf(ViewOfGame(g).Log, bear)
	if len(es) == 0 {
		t.Fatal("no counters entry")
	}
	if last := es[len(es)-1]; last.Amount != 1 || last.Delta != 1 {
		t.Errorf("the returned bear's first counter: amount %d delta %d, want 1 and 1", last.Amount, last.Delta)
	}
}

// One proliferate is one batch: every counters entry it writes shares
// it, each a placement of one.
func TestAProliferatesCountersShareABatch(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", me.ID)
	theirs := battlefieldCard(t, g, "Runeclaw Bear", "Creature — Bear", them.ID)
	if err := g.AddCounter(mine, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounter(theirs, game.CounterMinusOne, 1); err != nil {
		t.Fatal(err)
	}
	before := len(ViewOfGame(g).Log)
	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{mine, theirs}, nil); err != nil {
			t.Fatalf("proliferate: %v", err)
		}
	})
	var got []LogEvent
	for _, e := range ViewOfGame(g).Log[before:] {
		if e.Kind == LogCounters {
			got = append(got, e)
		}
	}
	if len(got) != 2 {
		t.Fatalf("got %d counters entries, want 2: %v", len(got), logKinds(ViewOfGame(g).Log[before:]))
	}
	for _, e := range got {
		if e.Delta != 1 || e.Amount != 2 {
			t.Errorf("%s on %s: amount %d delta %d, want 2 and 1", e.Label, e.CardID, e.Amount, e.Delta)
		}
		if e.Batch == 0 || e.Batch != got[0].Batch {
			t.Errorf("batch %d, want one non-zero batch for the proliferate (%d)", e.Batch, got[0].Batch)
		}
	}
}

// The change goes with the count it is half of: a viewer who may not
// identify the card gets neither (log_narrated_silences_test.go).
func TestACountersDeltaIsRedactedWithTheCount(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	hidden := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: hidden, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, FaceDown: true,
			KnownBy: map[uuid.UUID]bool{owner.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCounterPlaced, Target: hidden, Label: game.CounterPlusOne, Amount: 2,
		})
	})
	mine := findLog(t, FilterViewFor(ViewOfGame(g), owner.ID.String()).Log, LogCounters)
	if mine.Delta != 2 {
		t.Errorf("the knower's delta: got %d, want 2", mine.Delta)
	}
	theirs := findLog(t, FilterViewFor(ViewOfGame(g), other.ID.String()).Log, LogCounters)
	if theirs.Delta != 0 || theirs.Amount != 0 || theirs.Label != "" {
		t.Errorf("a non-knower's entry kept %q x%d (delta %d)", theirs.Label, theirs.Amount, theirs.Delta)
	}
}
