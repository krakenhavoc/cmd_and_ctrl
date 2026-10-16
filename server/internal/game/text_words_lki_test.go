package game

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// text_words_lki_test.go — #3872 slice 2: a permanent's changed words are
// last-known information (CR 608.2h, CR 603.10a). The departure record
// carries them as plain data, a snapshot round-trips them, and the
// readers of a departed source (the target re-check, an effect's text
// reader) read them; and a cost's clause reads its source's words on a
// walk that is not targeting.

// entries and textWordsFromEntries are inverses, for a composed change
// and for a net map composing could not build: entries are set side by
// side, never composed, so a swap (black → blue and blue → black at once)
// stays a swap.
func TestTextWordsEntriesRoundTrip(t *testing.T) {
	var w *TextWords
	w = w.with(twChange(TextWordColor, "B", "U"))
	w = w.with(twChange(TextWordColor, "U", "R"))
	w = w.with(twChange(TextWordBasicLandType, "Swamp", "Island"))
	if back := textWordsFromEntries(w.entries()); !back.equal(w) {
		t.Errorf("round trip: %v, want %v", back.Lines(), w.Lines())
	}
	swap := []TextWordChange{
		{Kind: TextWordColor, From: "U", To: "B"},
		{Kind: TextWordColor, From: "B", To: "U"},
	}
	back := textWordsFromEntries(swap)
	if back.color("B") != "U" || back.color("U") != "B" {
		t.Errorf("the rebuilt swap reads black %q, blue %q", back.color("B"), back.color("U"))
	}
	if got := back.entries(); !slices.Equal(got, swap) {
		t.Errorf("entries %+v, want %+v (WUBRG order)", got, swap)
	}
	if textWordsFromEntries(nil) != nil || (*TextWords)(nil).entries() != nil {
		t.Error("no change is nil both ways")
	}
	// A damaged entry is skipped, never read as a wrong word.
	if got := textWordsFromEntries([]TextWordChange{{Kind: TextWordColor, From: "B", To: "B"}}); got != nil {
		t.Errorf("an entry replacing a word with itself reads as %v", got.Lines())
	}
}

// A permanent that leaves takes its changed words into its last-known
// record, the record survives a snapshot, and both the target re-check
// source and the effect's text reader read them.
func TestDepartedSourceKeepsItsChangedWords(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushBattlefieldForTest(g, me.ID, "Swamp Watcher", "Creature — Zombie", "")
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(src)
		c.EnteredBattlefieldAt = timeNowUnixNano()
		if !g.ChangeTextWordForEffect(uuid.Nil, src, twChange(TextWordBasicLandType, "Swamp", "Island"), IndefiniteDuration(), "test") {
			t.Fatal("setup: the change was registered")
		}
	})
	var item *StackItem
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(src)
		item = &StackItem{ID: uuid.New(), Kind: StackItemActivated, Controller: me.ID, SourceCardID: src,
			SourceObject: ObjectRef{ID: src, Epoch: c.ObjectEpoch}}
		if err := g.DestroyPermanentForEffect(src); err != nil {
			t.Fatal(err)
		}
	})
	read := func(g *Game) (reader, target string) {
		g.WithWriteLock(func() {
			reader = g.TextReaderForEffect(item).TextSubtype("Swamp")
			target = g.stackItemSourceLocked(item).lend(Card{}).LentSubtype("Swamp")
		})
		return reader, target
	}
	if reader, target := read(g); reader != "Island" || target != "Island" {
		t.Fatalf("the departed source reads \"Swamp\" as %q (effect) and %q (target re-check), want Island", reader, target)
	}
	var rec PermanentInfo
	g.ReadSnapshot(func() { rec, _ = g.LastKnownPermanentForEffect(src) })
	if !slices.Equal(rec.TextWords, []TextWordChange{{Kind: TextWordBasicLandType, From: "Swamp", To: "Island"}}) {
		t.Errorf("the record carries %+v", rec.TextWords)
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
		t.Fatalf("restore: %v", err)
	}
	if reader, target := read(restored); reader != "Island" || target != "Island" {
		t.Errorf("after a restore the departed source reads \"Swamp\" as %q / %q, want Island", reader, target)
	}
	clone := g.Clone()
	if reader, _ := read(clone); reader != "Island" {
		t.Errorf("a clone's departed source reads \"Swamp\" as %q, want Island", reader)
	}
}

// A cost's clause is judged on a walk that does not target, with no
// source object (it excludes nothing, compares nothing), but its words are
// the source's text: costSourceLocked lends them.
func TestCostWalkReadsTheSourcesWords(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushBattlefieldForTest(g, me.ID, "Altar", "Artifact", "")
	white := pushBattlefieldForTest(g, me.ID, "White Knight", "Creature — Knight", "")
	green := pushBattlefieldForTest(g, me.ID, "Green Bear", "Creature — Bear", "")
	spec := &TargetSpec{
		Mode: "creature", Label: "a white creature", Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature() && c.HasColorWord("W")
		},
		Min: 1, Max: 1,
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case white:
				g.Battlefield.Cards[i].Colors = []string{"W"}
			case green:
				g.Battlefield.Cards[i].Colors = []string{"G"}
			case src:
				g.Battlefield.Cards[i].EnteredBattlefieldAt = timeNowUnixNano()
			}
		}
		g.RecomputeLayersIfStaleLocked()
		if !g.ChangeTextWordForEffect(uuid.Nil, src, twChange(TextWordColor, "W", "G"), IndefiniteDuration(), "test") {
			t.Fatal("setup: the change was registered")
		}
	})
	var cands, printed []uuid.UUID
	var src0 TargetSource
	g.ReadSnapshot(func() {
		cands = g.SacrificeCostCandidatesForEffect(me.ID, src, spec)
		printed = g.SacrificeCostCandidatesForEffect(me.ID, uuid.Nil, spec)
		src0 = g.costSourceLocked(me.ID, src)
	})
	if !slices.Contains(cands, green) || slices.Contains(cands, white) {
		t.Errorf("the cost's candidates %v: want the green creature and not the white one", cands)
	}
	if !slices.Contains(printed, white) || slices.Contains(printed, green) {
		t.Errorf("with no source the printed word stands: %v", printed)
	}
	if src0.Object != nil || src0.Departed != nil {
		t.Error("a cost walk names no source object: it targets nothing and excludes nothing")
	}
}
