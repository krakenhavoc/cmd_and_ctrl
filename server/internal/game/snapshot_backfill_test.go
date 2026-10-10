package game

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
)

// snapshot_backfill_test.go — #683: a restore point written before
// Card.VariableToughness existed gets the flag recomputed, so a `*`
// creature restored across the deploy keeps the placeholder skip.
//
// testdata/snapshot_pre683.json is a real pre-change file. It was
// written by origin/develop at 7c1ae9b, the commit #683 branches from,
// using that binary's CaptureSnapshot and json.MarshalIndent. The
// hands, libraries, command zones and event log were emptied first to
// keep the file small. Its battlefield holds five creatures owned and
// controlled by seat 0, each with printed toughness 0 and one +1/+1
// counter:
//
//	…0001 "Star Creature"    printing …000a
//	…0002 "Zero Zero"        printing …000b
//	…0003 "Construct Token"  no printing (a token)
//	…0004 a Clone copying Star Creature: printing …000a,
//	      PrintedSelf printing …000c (Clone itself, a printed 0/0)
//	…0005 "Star Front"       printing …000d, a transform card whose
//	      front face has toughness 0 and whose back face is a 3/3
//
// Do not regenerate it with this branch's encoder: the point is that
// the file has no variableToughness key.

var (
	pre683Star   = uuid.MustParse("00000000-0000-4000-8000-000000000001")
	pre683Zero   = uuid.MustParse("00000000-0000-4000-8000-000000000002")
	pre683Token  = uuid.MustParse("00000000-0000-4000-8000-000000000003")
	pre683Clone  = uuid.MustParse("00000000-0000-4000-8000-000000000004")
	pre683Faced  = uuid.MustParse("00000000-0000-4000-8000-000000000005")
	pre683Lookup = map[string]struct {
		top   bool
		faces []bool
	}{
		"00000000-0000-4000-9000-00000000000a": {top: true},                   // toughness "*"
		"00000000-0000-4000-9000-00000000000b": {top: false},                  // toughness "0"
		"00000000-0000-4000-9000-00000000000c": {top: false},                  // Clone, "0"
		"00000000-0000-4000-9000-00000000000d": {faces: []bool{true, false}},  // "*" // "3"
		"00000000-0000-4000-9000-00000000000e": {faces: []bool{false, false}}, // "0" // "3"
	}
)

// usePrintedLookup installs a PrintedVariableToughness that knows the
// fixture's printings, the way cmd/server installs the Scryfall one.
func usePrintedLookup(t *testing.T) {
	t.Helper()
	prev := PrintedVariableToughness
	PrintedVariableToughness = func(id string) (bool, []bool, bool) {
		p, ok := pre683Lookup[id]
		return p.top, p.faces, ok
	}
	t.Cleanup(func() { PrintedVariableToughness = prev })
}

// usePrintedLookupNone removes the lookup: a server with no Scryfall
// dump, where restore falls back to the pre-#683 rule.
func usePrintedLookupNone(t *testing.T) {
	t.Helper()
	prev := PrintedVariableToughness
	PrintedVariableToughness = nil
	t.Cleanup(func() { PrintedVariableToughness = prev })
}

func restorePre683Fixture(t *testing.T) *Game {
	t.Helper()
	return restorePre683FixtureWith(t, nil)
}

// restorePre683FixtureWith lets edit change the decoded file before it
// is restored. An edit must leave VariableToughness nil, or the card is
// no longer a pre-#683 one.
func restorePre683FixtureWith(t *testing.T, edit func(*GameSnapshot)) *Game {
	t.Helper()
	raw, err := os.ReadFile("testdata/snapshot_pre683.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if bytes.Contains(bytes.ToLower(raw), []byte("variabletoughness")) {
		t.Fatal("the pre-#683 fixture carries a VariableToughness key; it is no longer a pre-change file")
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if edit != nil {
		edit(&snap)
	}
	g, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("restore fixture: %v", err)
	}
	return g
}

func TestPre683RestorePointBackfillsTurnIdentity(t *testing.T) {
	g := restorePre683Fixture(t)
	wantSeq := g.Turn.Round*MaxPlayers + g.Turn.ActiveSeat
	if g.Turn.Seq != wantSeq {
		t.Fatalf("restored Turn.Seq = %d, want legacy RNG index %d", g.Turn.Seq, wantSeq)
	}
	if g.Turn.OrderSeat != g.Turn.ActiveSeat || g.Turn.Extra || g.Turn.ExtraRef != 0 {
		t.Errorf("legacy turn provenance = %+v, want a normal turn ordered at the active seat", g.Turn)
	}
	for seat, p := range g.Seats {
		want := g.Turn.Round
		if seat > g.Turn.ActiveSeat {
			want--
		}
		if want < 0 {
			want = 0
		}
		if p.TurnsBegun != want {
			t.Errorf("seat %d TurnsBegun = %d, want %d", seat, p.TurnsBegun, want)
		}
	}

	before := g.Turn.Seq
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn after legacy restore: %v", err)
	}
	if g.Turn.Seq != before+1 {
		t.Errorf("next turn Seq = %d, want %d", g.Turn.Seq, before+1)
	}
}

func TestPre683RestorePointConvertsLegacyTurnStamps(t *testing.T) {
	var oldRound int
	g := restorePre683FixtureWith(t, func(s *GameSnapshot) {
		oldRound = s.Turn.Round
		s.Seats[0].CastPermissions = append(s.Seats[0].CastPermissions, CastPermission{
			Player:              s.Seats[0].ID,
			Duration:            WhileInZoneDuration(),
			LegacyNotBeforeTurn: oldRound + 1,
		})
		s.DelayedTriggers = append(s.DelayedTriggers, delayedTriggerSnapshot{
			ID:          uuid.New(),
			Controller:  s.Seats[0].ID,
			At:          StepEnd,
			CreatedTurn: oldRound,
		})
	})

	perm := g.Seats[0].CastPermissions[len(g.Seats[0].CastPermissions)-1]
	if want := (oldRound + 1) * MaxPlayers; perm.NotBeforeSeq != want {
		t.Errorf("legacy not-before floor = %d, want Seq %d", perm.NotBeforeSeq, want)
	}
	if perm.LegacyNotBeforeTurn != 0 {
		t.Errorf("legacy not-before stamp survived migration: %+v", perm)
	}
	delayed := g.DelayedTriggers[len(g.DelayedTriggers)-1]
	if want := oldRound * MaxPlayers; delayed.CreatedSeq != want {
		t.Errorf("legacy delayed-trigger stamp = %d, want Seq %d", delayed.CreatedSeq, want)
	}
}

// With the printings known, the backfill gives each card the flag a
// fresh import would have: the `*` creature, the Clone copying it and
// the transform card's `*` front are variable; the printed 0/0 and the
// Clone's own printed values are not; the token, which has no
// printing, keeps the pre-#683 skip.
func TestPre683RestorePointBackfillsVariableToughness(t *testing.T) {
	usePrintedLookup(t)
	g := restorePre683Fixture(t)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"`*` creature", pre683Star, true},
		{"printed 0/0", pre683Zero, false},
		{"token with no printing", pre683Token, true},
		{"Clone of the `*` creature", pre683Clone, true},
		{"transform card, `*` front face up", pre683Faced, true},
	} {
		if got := battlefieldCard(t, g, tc.id).VariableToughness; got != tc.want {
			t.Errorf("%s: VariableToughness = %v, want %v", tc.name, got, tc.want)
		}
	}
	clone := battlefieldCard(t, g, pre683Clone)
	if clone.PrintedSelf == nil || clone.PrintedSelf.VariableToughness {
		t.Errorf("the Clone's own printed values (a printed 0/0) came back variable: %+v", clone.PrintedSelf)
	}
	faced := battlefieldCard(t, g, pre683Faced)
	if len(faced.Faces) != 2 || !faced.Faces[0].VariableToughness || faced.Faces[1].VariableToughness {
		t.Errorf("per-face backfill wrong, want [true false]: %+v", faced.Faces)
	}
}

// The headline: a `*` creature restored from a pre-#683 file gains a
// counter, loses it, and is still on the battlefield. Without the
// backfill it decodes with VariableToughness false and dies. The
// printed 0/0 beside it, whose printing is numeric, dies as #683
// intends.
func TestPre683StarCreatureSurvivesGainingAndLosingACounter(t *testing.T) {
	usePrintedLookup(t)
	g := restorePre683Fixture(t)
	me := g.Seats[0]

	for _, id := range []uuid.UUID{pre683Star, pre683Zero, pre683Token, pre683Clone, pre683Faced} {
		// Each already holds one counter from the old game; give one
		// more after the restore, then take both off.
		if err := g.AddCounter(id, CounterPlusOne, 1); err != nil {
			t.Fatalf("add counter: %v", err)
		}
		runSBAsForTest(g)
		for i := 0; i < 2; i++ {
			if err := g.AddCounter(id, CounterPlusOne, -1); err != nil {
				t.Fatalf("remove counter: %v", err)
			}
		}
	}
	runSBAsForTest(g)

	for _, tc := range []struct {
		name  string
		id    uuid.UUID
		alive bool
	}{
		{"`*` creature", pre683Star, true},
		{"printed 0/0", pre683Zero, false},
		{"token with no printing (pre-#683 skip)", pre683Token, true},
		{"Clone of the `*` creature", pre683Clone, true},
		{"transform card, `*` front face up", pre683Faced, true},
	} {
		if got := g.Battlefield.Contains(tc.id); got != tc.alive {
			t.Errorf("%s: on the battlefield = %v, want %v", tc.name, got, tc.alive)
		}
	}
	if !me.Graveyard.Contains(pre683Zero) {
		t.Error("the printed 0/0 left the battlefield but is not in its owner's graveyard")
	}
}

// With no Scryfall index the backfill cannot tell a `*` from a printed
// 0, so it keeps the pre-#683 skip for every 0-toughness card. The `*`
// creature survives, which is the point; the printed 0/0 survives
// too, which is the documented limit.
func TestPre683RestoreWithoutAnIndexKeepsTheOldSkip(t *testing.T) {
	usePrintedLookupNone(t)
	g := restorePre683Fixture(t)

	for _, id := range []uuid.UUID{pre683Star, pre683Zero, pre683Token, pre683Clone, pre683Faced} {
		if !battlefieldCard(t, g, id).VariableToughness {
			t.Errorf("card %s: toughness 0 with no index, want VariableToughness true", id)
		}
	}
	faced := battlefieldCard(t, g, pre683Faced)
	if len(faced.Faces) != 2 || !faced.Faces[0].VariableToughness || faced.Faces[1].VariableToughness {
		t.Errorf("per-face fallback wrong, want [true false] (only the 0-toughness face): %+v", faced.Faces)
	}
	if ps := battlefieldCard(t, g, pre683Clone).PrintedSelf; ps == nil || !ps.VariableToughness {
		t.Errorf("the Clone's printed values: want the fallback true, got %+v", ps)
	}

	for _, id := range []uuid.UUID{pre683Star, pre683Zero} {
		if err := g.AddCounter(id, CounterPlusOne, -1); err != nil {
			t.Fatalf("remove counter: %v", err)
		}
	}
	runSBAsForTest(g)
	if !g.Battlefield.Contains(pre683Star) {
		t.Error("the `*` creature died losing its last counter after a restore with no index")
	}
	if !g.Battlefield.Contains(pre683Zero) {
		t.Error("the printed 0/0 died with no index; the fallback is meant to keep the pre-#683 skip")
	}
}

// A file this binary writes says false outright for a printed 0/0, and
// restore must keep that false rather than backfill it: with no index
// the backfill would say true, and the 0/0 would stop dying.
func TestSnapshotKeepsAnExplicitFalseVariableToughness(t *testing.T) {
	usePrintedLookupNone(t)
	g := newRestorableGame(t)
	me := g.Seats[0]
	zero := pushZeroZero(g, me, map[string]int{CounterPlusOne: 1})

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Contains(raw, []byte(`"variableToughness":false`)) {
		t.Fatal(`the snapshot does not write "variableToughness":false; restore cannot tell it from a pre-#683 file`)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if battlefieldCard(t, restored, zero).VariableToughness {
		t.Fatal("an explicit false was backfilled to true")
	}
	if err := restored.AddCounter(zero, CounterPlusOne, -1); err != nil {
		t.Fatalf("remove counter: %v", err)
	}
	runSBAsForTest(restored)
	if restored.Battlefield.Contains(zero) {
		t.Error("a restored printed 0/0 survived losing its last counter")
	}
}

// snapshotCardByID finds a battlefield card in a decoded snapshot.
func snapshotCardByID(t *testing.T, snap *GameSnapshot, id uuid.UUID) *cardSnapshot {
	t.Helper()
	for i := range snap.Battlefield.Cards {
		if snap.Battlefield.Cards[i].InstanceID == id {
			return &snap.Battlefield.Cards[i]
		}
	}
	t.Fatalf("card %s is not on the fixture's battlefield", id)
	return nil
}

// A token copy of a double-faced card keeps the printing's ScryfallID
// but not its Faces (TokenCopyTemplate). Scryfall gives such a printing
// no top-level toughness, so the lookup's top answer is false even when
// the copied face is `*` (Katilda, Dawnhart Martyr). The backfill must
// not take that false: a fresh token copy of the `*` face carries true,
// and the pre-#683 skip kept it too. Here the fixture's Construct token
// becomes a token copy of printing …000d ("*" // "3"): it gains and
// loses a counter and survives. A token copy of …000e, whose faces are
// both numeric, gets false as a fresh copy would, and dies.
func TestPre683TokenCopyOfADoubleFacedCardUsesItsFaces(t *testing.T) {
	for _, tc := range []struct {
		name     string
		printing string
		variable bool
	}{
		{"`*` front face", "00000000-0000-4000-9000-00000000000d", true},
		{"numeric faces", "00000000-0000-4000-9000-00000000000e", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			usePrintedLookup(t)
			g := restorePre683FixtureWith(t, func(snap *GameSnapshot) {
				tok := snapshotCardByID(t, snap, pre683Token)
				tok.ScryfallID = tc.printing
				if len(tok.Faces) != 0 {
					t.Fatalf("the fixture token has faces: %+v", tok.Faces)
				}
				// The Clone's own printed values, faceless on the same
				// double-faced printing, take the same road.
				clone := snapshotCardByID(t, snap, pre683Clone)
				if clone.PrintedSelf == nil {
					t.Fatal("the fixture Clone has no printedSelf")
				}
				clone.PrintedSelf.ScryfallID = tc.printing
				clone.PrintedSelf.Faces = nil
			})

			tok := battlefieldCard(t, g, pre683Token)
			if tok.VariableToughness != tc.variable {
				t.Errorf("token copy: VariableToughness = %v, want %v", tok.VariableToughness, tc.variable)
			}
			if ps := battlefieldCard(t, g, pre683Clone).PrintedSelf; ps == nil || ps.VariableToughness != tc.variable {
				t.Errorf("faceless PrintedSelf: want VariableToughness %v, got %+v", tc.variable, ps)
			}

			if err := g.AddCounter(pre683Token, CounterPlusOne, 1); err != nil {
				t.Fatalf("add counter: %v", err)
			}
			runSBAsForTest(g)
			for i := 0; i < 2; i++ {
				if err := g.AddCounter(pre683Token, CounterPlusOne, -1); err != nil {
					t.Fatalf("remove counter: %v", err)
				}
			}
			runSBAsForTest(g)
			if got := g.Battlefield.Contains(pre683Token); got != tc.variable {
				t.Errorf("token copy on the battlefield after losing its last counter = %v, want %v", got, tc.variable)
			}
		})
	}
}
