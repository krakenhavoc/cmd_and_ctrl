package game

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/uuid"
)

// table_roll_test.go — ADR 0121 §5 and §9: the "Roll a die" table
// action. Its own stream outside the turn counters, a counter an undo
// does not rewind, no game roll event, and a log line an undo does not
// erase.

func tableRollEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventTableRoll {
			out = append(out, ev)
		}
	}
	return out
}

func mustTableRoll(t *testing.T, g *Game, p uuid.UUID, die string) Event {
	t.Helper()
	ev, err := g.RollTableDie(p, die)
	if err != nil {
		t.Fatalf("RollTableDie(%s): %v", die, err)
	}
	return ev
}

// TestTableRollKnownAnswer pins the table stream's derivation:
// rngSeed(key, "table", seat, nil, 0, tableRollNext). A change to it
// changes what a restored game's next table roll is, so it must be a
// deliberate, visible edit of this test.
func TestTableRollKnownAnswer(t *testing.T) {
	g := newActiveGame(t)
	g.SetRNGKeyForTest(testRNGKey())
	var seat0 [16]byte
	g.WithWriteLock(func() { seat0, _ = g.rngPlayerIdentityLocked(g.Seats[0].ID) })
	seed := rngSeed(testRNGKey(), "table", seat0, uuid.Nil, 0, 0)
	if got := hex.EncodeToString(seed[:]); got != tableRollKnownSeedHex {
		t.Errorf("table seed = %s, want %s", got, tableRollKnownSeedHex)
	}

	a, b := g.Seats[0].ID, g.Seats[1].ID
	got := []string{}
	for _, roll := range []struct {
		p   uuid.UUID
		die string
	}{{a, TableDieD20}, {a, TableDieD6}, {b, TableDieCoin}, {b, TableDieD20}, {a, TableDieCoin}} {
		ev := mustTableRoll(t, g, roll.p, roll.die)
		if ev.Sides > 0 {
			got = append(got, roll.die+":"+strconv.Itoa(ev.Amount))
		} else {
			got = append(got, roll.die+":"+ev.Label)
		}
	}
	if !reflect.DeepEqual(got, tableRollKnownResults) {
		t.Errorf("table rolls = %#v, want %#v", got, tableRollKnownResults)
	}
}

// Pinned by TestTableRollKnownAnswer. Change only on purpose.
var (
	tableRollKnownSeedHex = "400ce5cf580f10e5ec007dc3a370a56338b64b51c607e812c91a37cf96a85d3f"
	tableRollKnownResults = []string{"d20:14", "d6:2", "coin:heads", "d20:12", "coin:heads"}
)

// TestTableRollLeavesTheGameStreamsAlone: a table roll never touches
// rngCounters, so a card roll, a shuffle and a pick draw exactly what
// they would have drawn without it, in the same turn.
func TestTableRollLeavesTheGameStreamsAlone(t *testing.T) {
	run := func(tableRolls int) (counters map[string]uint64, rolls []int, library []string, picked []int) {
		g := newActiveGame(t)
		g.SetRNGKeyForTest(testRNGKey())
		p := g.Seats[0]
		// One draw on each stream before, so the counters are not empty.
		g.WithWriteLock(func() {
			if _, err := g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: rngTestOgre}, 20, 1); err != nil {
				t.Fatal(err)
			}
		})
		for i := 0; i < tableRolls; i++ {
			mustTableRoll(t, g, g.Seats[i%2].ID, []string{TableDieD6, TableDieD20, TableDieCoin}[i%3])
		}
		g.WithWriteLock(func() {
			counters = cloneRNGCounters(g.rngCounters)
			var err error
			rolls, err = g.RollDiceForEffect(RandomDraw{Player: p.ID, Source: rngTestOgre}, 20, 3)
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]uuid.UUID, 0, len(p.Hand.Cards))
			at := map[uuid.UUID]int{}
			for i, c := range p.Hand.Cards {
				ids = append(ids, c.InstanceID)
				at[c.InstanceID] = i
			}
			// Positions, not IDs: each run's cards have their own IDs.
			for _, id := range g.pickAtRandomLocked(rngStream{kind: rngStreamPick, player: p.ID}, ids, 3) {
				picked = append(picked, at[id])
			}
		})
		if err := g.ShuffleLibrary(p.ID); err != nil {
			t.Fatal(err)
		}
		return counters, rolls, libraryNames(p), picked
	}
	c0, r0, l0, p0 := run(0)
	c5, r5, l5, p5 := run(5)
	if !reflect.DeepEqual(c0, c5) {
		t.Errorf("table rolls moved rngCounters: %v, want %v", c5, c0)
	}
	if !reflect.DeepEqual(r0, r5) {
		t.Errorf("a card roll after table rolls = %v, want %v", r5, r0)
	}
	if !reflect.DeepEqual(l0, l5) {
		t.Error("a shuffle after table rolls came out differently")
	}
	if len(p0) != 3 || !reflect.DeepEqual(p0, p5) {
		t.Errorf("a pick after table rolls = %v, want %v", p5, p0)
	}
}

// TestTableRollIsNotAGameRoll: one EventTableRoll, and no EventRollDie
// or EventFlipCoin, no pending choice, nothing else on the log.
func TestTableRollIsNotAGameRoll(t *testing.T) {
	g := newActiveGame(t)
	before := len(g.Events)
	p := g.Seats[1].ID
	d20 := mustTableRoll(t, g, p, TableDieD20)
	d6 := mustTableRoll(t, g, p, TableDieD6)
	coin := mustTableRoll(t, g, p, TableDieCoin)
	added := g.Events[before:]
	if len(added) != 3 {
		t.Fatalf("three table rolls emitted %d events: %+v", len(added), added)
	}
	for _, ev := range added {
		if ev.Kind != EventTableRoll || ev.Actor != p {
			t.Errorf("event %+v, want an EventTableRoll by the roller", ev)
		}
	}
	if d20.Sides != 20 || d20.Amount < 1 || d20.Amount > 20 || d20.RollID != 1 {
		t.Errorf("d20 = %+v", d20)
	}
	if d6.Sides != 6 || d6.Amount < 1 || d6.Amount > 6 || d6.RollID != 2 {
		t.Errorf("d6 = %+v", d6)
	}
	if coin.Sides != 0 || (coin.Label != "heads" && coin.Label != "tails") || coin.RollID != 3 {
		t.Errorf("coin = %+v", coin)
	}
	if len(g.PendingChoices) != 0 || len(g.PendingTriggers) != 0 {
		t.Errorf("a table roll left %d choices and %d triggers", len(g.PendingChoices), len(g.PendingTriggers))
	}
}

// TestTableRollIsLegalWhenever: during the opening roll and the
// mulligan, and mid-turn; refused for an eliminated seat, an unknown
// die and an ended game.
func TestTableRollIsLegalWhenever(t *testing.T) {
	open := openOpeningRollForTest(t, 4, 1, 2)
	mustTableRoll(t, open, open.Seats[2].ID, TableDieD20)
	if open.OpeningRoll == nil || len(open.OpeningRoll.Rounds[0].Rolls) != 0 {
		t.Fatalf("a table roll moved the opening roll: %+v", open.OpeningRoll)
	}

	mull := newActiveGameMulligansOpen(t, 2)
	mustTableRoll(t, mull, mull.Seats[0].ID, TableDieCoin)
	if !mull.MulligansOpen {
		t.Fatal("a table roll closed the mulligan")
	}

	g := newActiveGame(t)
	mustTableRoll(t, g, g.Seats[1].ID, TableDieD6)
	if _, err := g.RollTableDie(g.Seats[0].ID, "d12"); !errors.Is(err, ErrUnknownTableDie) {
		t.Errorf("a d12: %v, want ErrUnknownTableDie", err)
	}
	if _, err := g.RollTableDie(uuid.New(), TableDieD6); !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("a stranger: %v, want ErrPlayerNotFound", err)
	}
	g.Seats[1].Eliminated = true
	if _, err := g.RollTableDie(g.Seats[1].ID, TableDieD6); !errors.Is(err, ErrPlayerEliminated) {
		t.Errorf("an eliminated seat: %v, want ErrPlayerEliminated", err)
	}
	g.End()
	if _, err := g.RollTableDie(g.Seats[0].ID, TableDieD6); !errors.Is(err, ErrGameNotActive) {
		t.Errorf("an ended game: %v, want ErrGameNotActive", err)
	}
}

// TestUndoKeepsTableRolls: an undo of somebody's earlier action does
// not erase a table roll made after it, does not rewind tableRollNext,
// and does not repeat the line when the roll is still in the restored
// log.
func TestUndoKeepsTableRolls(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]

	beforeA := g.Clone()
	if err := g.DrawCard(a.ID); err != nil { // A's action
		t.Fatal(err)
	}
	roll := mustTableRoll(t, g, b.ID, TableDieD20) // B's table roll
	beforeC := g.Clone()
	if err := g.DrawCard(a.ID); err != nil { // another action after the roll
		t.Fatal(err)
	}

	// Undo the action after the roll: the roll is still in the
	// restored log, and is not repeated.
	g.WithWriteLock(func() { g.RestoreFrom(beforeC) })
	if got := tableRollEvents(g); len(got) != 1 || got[0].Seq != roll.Seq {
		t.Fatalf("after undoing a later action: %+v, want the one roll at seq %d", got, roll.Seq)
	}

	// Undo A's draw, from before the roll: the truncation cuts the
	// roll's line, and RestoreFrom puts it back after the restored end.
	g.WithWriteLock(func() { g.RestoreFrom(beforeA) })
	if len(g.Events) != len(beforeA.Events)+1 {
		t.Fatalf("after undoing A the log has %d events, want the restored %d and the roll", len(g.Events), len(beforeA.Events))
	}
	got := tableRollEvents(g)
	if len(got) != 1 {
		t.Fatalf("after undoing A: %d table rolls, want 1", len(got))
	}
	re := got[0]
	if re.RollID != roll.RollID || re.Amount != roll.Amount || re.Actor != b.ID || re.Sides != 20 {
		t.Errorf("re-emitted %+v, want the payload of %+v", re, roll)
	}
	if re.Seq != beforeA.eventSeq+1 || g.Events[len(g.Events)-1].Seq != re.Seq {
		t.Errorf("re-emitted at seq %d, want %d, the first after the restored history", re.Seq, beforeA.eventSeq+1)
	}
	if g.tableRollNext != 1 {
		t.Errorf("tableRollNext = %d after the undo, want 1 (carried forward)", g.tableRollNext)
	}

	// The next table roll is a fresh one, not a replay of roll 1.
	next := mustTableRoll(t, g, b.ID, TableDieD20)
	if next.RollID != 2 {
		t.Errorf("next roll_id = %d, want 2", next.RollID)
	}

	// A second undo, further back, puts both back again, in order.
	g.WithWriteLock(func() { g.RestoreFrom(beforeA) })
	got = tableRollEvents(g)
	if len(got) != 2 || got[0].RollID != 1 || got[1].RollID != 2 {
		t.Fatalf("after a second undo: %+v, want rolls 1 and 2", got)
	}
}

// TestTableRollRingHoldsTheLast32 bounds what an undo carries across.
func TestTableRollRingHoldsTheLast32(t *testing.T) {
	g := newActiveGame(t)
	before := g.Clone()
	for i := 0; i < tableRollRingMax+8; i++ {
		mustTableRoll(t, g, g.Seats[i%2].ID, TableDieD6)
	}
	if len(g.tableRolls) != tableRollRingMax || g.tableRolls[0].RollID != 9 {
		t.Fatalf("ring holds %d rolls from roll_id %d, want %d from 9", len(g.tableRolls), g.tableRolls[0].RollID, tableRollRingMax)
	}
	g.WithWriteLock(func() { g.RestoreFrom(before) })
	got := tableRollEvents(g)
	if len(got) != tableRollRingMax || got[0].RollID != 9 || got[len(got)-1].RollID != tableRollRingMax+8 {
		t.Fatalf("after the undo: %d rolls, want the last %d", len(got), tableRollRingMax)
	}
}

// TestTableRollNextIsPersisted: a restored game's next table roll is
// the live game's next, not a repeat of roll 1.
func TestTableRollNextIsPersisted(t *testing.T) {
	g := newActiveGame(t)
	for i := 0; i < 3; i++ {
		mustTableRoll(t, g, g.Seats[0].ID, TableDieD20)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	want := mustTableRoll(t, g, g.Seats[0].ID, TableDieD20)
	got := mustTableRoll(t, restored, restored.Seats[0].ID, TableDieD20)
	if got.RollID != 4 || got.RollID != want.RollID || got.Amount != want.Amount {
		t.Errorf("restored next roll %+v, live %+v", got, want)
	}
	if len(restored.tableRolls) != 1 {
		t.Errorf("restored ring = %d, want only the roll made since", len(restored.tableRolls))
	}
}

// TestTableRollsAreFair: 20,000 d20s and 10,000 coins under a fixed key
// land on every face within a fixed tolerance (ADR 0054 item 8's
// check, for the table stream).
func TestTableRollsAreFair(t *testing.T) {
	g := newActiveGame(t)
	g.SetRNGKeyForTest(testRNGKey())
	p := g.Seats[0].ID
	var d20 [21]int
	heads := 0
	g.WithWriteLock(func() {
		for i := 0; i < 20000; i++ {
			n, _ := g.tableRollDrawLocked(p, 20)
			d20[n+1]++
		}
		for i := 0; i < 10000; i++ {
			if n, _ := g.tableRollDrawLocked(p, 2); n == 0 {
				heads++
			}
		}
	})
	for face := 1; face <= 20; face++ {
		if d20[face] < 850 || d20[face] > 1150 {
			t.Errorf("d20 face %d came up %d times in 20000, want 1000 ± 150", face, d20[face])
		}
	}
	if heads < 4750 || heads > 5250 {
		t.Errorf("heads came up %d times in 10000, want 5000 ± 250", heads)
	}
}
