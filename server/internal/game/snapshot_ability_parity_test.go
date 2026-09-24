package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// snapshot_ability_parity_test.go is #522's "rebuilt is never verified
// to have been rebuilt" half.
//
// A restore re-derives a card's abilities from the running binary's
// catalog. Nothing compared what came back against what the restore
// point recorded, so a card whose entry was removed, renamed or cut
// down between two deploys came back silently stripped. The owner's
// decision on #515: restore the table, report the card loudly, and
// flag it as not automated for the rest of the game. More abilities
// than captured is not a mismatch.
//
// Every test here swaps CatalogLookup between the capture and the
// restore, which is exactly what a deploy does.

const parityOracle = "00000000-0000-4000-8000-000000000522"

// parityDef is a catalog entry with one of each measured slot.
func parityDef(triggers, statics int) *CardDef {
	d := &CardDef{
		Activated:    []ActivatedAbilityShape{{Label: "{T}: do a thing"}},
		Replacements: []ReplacementEffect{{Label: "a replacement"}},
	}
	for range triggers {
		d.Triggered = append(d.Triggered, TriggeredAbility{Watches: []EventKind{EventETB}})
	}
	for range statics {
		d.Static = append(d.Static, StaticAbility{Layer: Layer6Ability})
	}
	return d
}

// withCatalog installs a one-entry catalog (nil def = no entry) for
// the rest of the test.
func withCatalog(t *testing.T, key string, def *CardDef) {
	t.Helper()
	prev := CatalogLookup
	t.Cleanup(func() { CatalogLookup = prev })
	CatalogLookup = func(k string) *CardDef {
		if k == key && def != nil {
			return def
		}
		return nil
	}
}

// pushParityCard puts the probe card on seat 0's battlefield.
func pushParityCard(g *Game) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{
			InstanceID: id,
			Name:       "Parity Probe",
			TypeLine:   "Creature — Test",
			OracleID:   parityOracle,
			Power:      2,
			Toughness:  2,
			Owner:      g.Seats[0].ID,
			Controller: g.Seats[0].ID,
		})
	})
	return id
}

// throughJSON is the deploy: encode with one binary, decode with the
// next.
func throughJSON(t *testing.T, s *GameSnapshot) *GameSnapshot {
	t.Helper()
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out GameSnapshot
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return &out
}

func TestCaptureMeasuresTheCatalogEntry(t *testing.T) {
	withCatalog(t, parityOracle, parityDef(2, 1))
	g := newRestorableGame(t)
	id := pushParityCard(g)

	snap := throughJSON(t, g.CaptureSnapshot())
	var got *cardSnapshot
	for i := range snap.Battlefield.Cards {
		if snap.Battlefield.Cards[i].InstanceID == id {
			got = &snap.Battlefield.Cards[i]
		}
	}
	if got == nil || got.CatalogAbilities == nil {
		t.Fatalf("the capture recorded no catalog measurement for a catalogued card: %+v", got)
	}
	want := AbilityCounts{Activated: 1, Triggered: 2, Static: 1, Replacements: 1}
	if *got.CatalogAbilities != want {
		t.Errorf("recorded %+v, want %+v", *got.CatalogAbilities, want)
	}
	// An uncatalogued card carries no record at all, so a library of
	// a hundred basic lands costs the file nothing.
	for _, c := range snap.Seats[0].Library.Cards {
		if c.CatalogAbilities != nil {
			t.Fatalf("uncatalogued %q carries a catalog record", c.Name)
		}
	}
}

// TestRemovedCatalogEntryIsReportedAndFlagged is the acceptance
// criterion: remove the entry, and the card is reported rather than
// silently stripped.
func TestRemovedCatalogEntryIsReportedAndFlagged(t *testing.T) {
	withCatalog(t, parityOracle, parityDef(1, 1))
	g := newRestorableGame(t)
	id := pushParityCard(g)
	snap := throughJSON(t, g.CaptureSnapshot())
	if !snap.Restorable() {
		t.Fatalf("setup: not a restore point: %+v", snap.Continuations)
	}

	// The next binary has no entry for the card.
	withCatalog(t, parityOracle, nil)

	sf := snap.AbilityShortfalls()
	if len(sf) != 1 {
		t.Fatalf("shortfalls = %+v, want exactly the probe", sf)
	}
	if sf[0].CardID != id || sf[0].Name != "Parity Probe" || sf[0].OracleID != parityOracle {
		t.Errorf("shortfall names %+v, want the probe", sf[0])
	}
	if !sf[0].EntryMissing || sf[0].Zone != ZoneBattlefield {
		t.Errorf("shortfall = %+v, want EntryMissing on the battlefield", sf[0])
	}
	if sf[0].Captured.Total() != 4 || sf[0].Restored.Total() != 0 {
		t.Errorf("counts captured %+v restored %+v, want 4 and 0", sf[0].Captured, sf[0].Restored)
	}

	// Restored anyway — the owner's call is not to abandon the table.
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict refused a table with a stripped card: %v", err)
	}
	c := findCardInZone(restored.Battlefield, id)
	if c == nil {
		t.Fatal("the stripped card did not come back")
	}
	if !c.AbilitiesLostOnRestore {
		t.Fatal("the stripped card is not flagged")
	}
	if !Unimplemented(*c) {
		t.Error("a flagged card must read as unimplemented, so the client shows the manual chip")
	}

	// The flag survives the next restore point, even under a binary
	// that has the entry back — only the card leaving the game clears
	// it.
	withCatalog(t, parityOracle, parityDef(1, 1))
	again, err := throughJSON(t, restored.CaptureSnapshot()).RestoreStrict()
	if err != nil {
		t.Fatalf("second RestoreStrict: %v", err)
	}
	c2 := findCardInZone(again.Battlefield, id)
	if c2 == nil || !c2.AbilitiesLostOnRestore {
		t.Fatalf("the flag did not survive a later restore point: %+v", c2)
	}

	// And a zone change is not the card leaving the game.
	var moved bool
	again.WithWriteLock(func() {
		owner := again.PlayerByIDForEffect(c2.Owner)
		if _, err := MoveCard(again.Battlefield, owner.Graveyard, id); err == nil {
			moved = true
		}
	})
	if !moved {
		t.Fatal("setup: could not move the card to its graveyard")
	}
	if gy := findCardInZone(again.PlayerByID(c2.Owner).Graveyard, id); gy == nil || !gy.AbilitiesLostOnRestore {
		t.Errorf("a zone change cleared the flag: %+v", gy)
	}
	if u := again.Clone(); findCardInZone(u.PlayerByID(c2.Owner).Graveyard, id) == nil ||
		!findCardInZone(u.PlayerByID(c2.Owner).Graveyard, id).AbilitiesLostOnRestore {
		t.Error("Clone (undo) dropped the flag")
	}
}

// TestFewerAbilitiesInOneSlotIsAShortfall is the refactor case: the
// entry still exists, with a trigger gone.
func TestFewerAbilitiesInOneSlotIsAShortfall(t *testing.T) {
	withCatalog(t, parityOracle, parityDef(2, 1))
	g := newRestorableGame(t)
	id := pushParityCard(g)
	snap := throughJSON(t, g.CaptureSnapshot())

	withCatalog(t, parityOracle, parityDef(1, 1))
	sf := snap.AbilityShortfalls()
	if len(sf) != 1 || sf[0].EntryMissing {
		t.Fatalf("shortfalls = %+v, want one, entry present", sf)
	}
	if sf[0].Captured.Triggered != 2 || sf[0].Restored.Triggered != 1 {
		t.Errorf("trigger counts %d → %d, want 2 → 1", sf[0].Captured.Triggered, sf[0].Restored.Triggered)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if c := findCardInZone(restored.Battlefield, id); c == nil || !c.AbilitiesLostOnRestore {
		t.Error("a card that lost one trigger is not flagged")
	}
}

// TestMoreAbilitiesIsNotAShortfall: a new build adding abilities is
// normal.
func TestMoreAbilitiesIsNotAShortfall(t *testing.T) {
	withCatalog(t, parityOracle, parityDef(1, 0))
	g := newRestorableGame(t)
	id := pushParityCard(g)
	snap := throughJSON(t, g.CaptureSnapshot())

	withCatalog(t, parityOracle, parityDef(3, 2))
	if sf := snap.AbilityShortfalls(); len(sf) != 0 {
		t.Fatalf("more abilities reported as a shortfall: %+v", sf)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if c := findCardInZone(restored.Battlefield, id); c == nil || c.AbilitiesLostOnRestore {
		t.Error("a card that gained abilities was flagged")
	}
}

// TestTokenInstanceAbilitiesAreCompared is the half the counts on disk
// were always for: a token's instance closures, re-derived by key.
func TestTokenInstanceAbilitiesAreCompared(t *testing.T) {
	key := TokenKey("treasure")
	stubTokenCatalog(t, key,
		[]ManaAbilityShape{{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}}, nil)
	g := newRestorableGame(t)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{
			InstanceID:    id,
			Name:          "Treasure",
			TypeLine:      "Token Artifact — Treasure",
			TokenKey:      key,
			Owner:         g.Seats[0].ID,
			Controller:    g.Seats[0].ID,
			ManaAbilities: []ManaAbilityShape{{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}},
		})
	})
	snap := throughJSON(t, g.CaptureSnapshot())
	if !snap.Restorable() {
		t.Fatalf("setup: %+v", snap.Continuations)
	}
	if sf := snap.AbilityShortfalls(); len(sf) != 0 {
		t.Fatalf("an intact Treasure reported short: %+v", sf)
	}

	// The next binary's token template lost its ability.
	stubTokenCatalog(t, key, nil, nil)
	sf := snap.AbilityShortfalls()
	if len(sf) != 1 || sf[0].TokenKey != key || sf[0].Captured.Mana != 1 || sf[0].Restored.Mana != 0 {
		t.Fatalf("shortfalls = %+v, want the Treasure's mana ability", sf)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if c := findCardInZone(restored.Battlefield, id); c == nil || !c.AbilitiesLostOnRestore || !Unimplemented(*c) {
		t.Errorf("the stripped Treasure is not flagged manual: %+v", c)
	}
}

// TestAFaceDownFlaggedCardDoesNotLeakTheFlag keeps CR 708.2a ahead of
// the flag: a face-down permanent has no rules to be missing, and
// saying otherwise would tell the table something about a card it
// cannot read.
func TestAFaceDownFlaggedCardDoesNotLeakTheFlag(t *testing.T) {
	c := Card{AbilitiesLostOnRestore: true}
	c.SetFaceDown(FaceDownManifested)
	if Unimplemented(c) {
		t.Error("a face-down permanent read as unimplemented")
	}
}
