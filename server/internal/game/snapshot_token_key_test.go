package game

import (
	"testing"

	"github.com/google/uuid"
)

// snapshot_token_key_test.go covers #521 at the engine seam: a token
// carries a catalog key instead of an oracle ID, the census asks the
// catalog rather than asking after an oracle ID, and CR 707.2's token
// COPY is not disturbed by either.
//
// The end-to-end half — a real Treasure, Food, Clue and Blood on a
// real board, captured, restored and then cracked — lives in
// cards/effects/token_restore_test.go, where the actual catalog is
// wired up. This file is the seam, with the catalog stubbed so each
// property can be stated on its own.

// stubTokenCatalog wires CatalogManaAbilities / CatalogActivatedAbilities
// to answer for one key and nothing else, the way the real catalog
// answers for "token:treasure".
func stubTokenCatalog(t *testing.T, key string, mana []ManaAbilityShape, activated []ActivatedAbilityShape) {
	t.Helper()
	prevMana, prevActivated := CatalogManaAbilities, CatalogActivatedAbilities
	CatalogManaAbilities = func(k string) []ManaAbilityShape {
		if k != key {
			return nil
		}
		return mana
	}
	CatalogActivatedAbilities = func(k string) []ActivatedAbilityShape {
		if k != key {
			return nil
		}
		return activated
	}
	t.Cleanup(func() {
		CatalogManaAbilities, CatalogActivatedAbilities = prevMana, prevActivated
	})
}

// TestTokenKeyNamespaceCannotCollideWithAnOracleID is the whole
// safety argument for the prefix in one assertion: a Scryfall oracle
// ID is a UUID, and a UUID has no colon in it.
func TestTokenKeyNamespaceCannotCollideWithAnOracleID(t *testing.T) {
	key := TokenKey("treasure")
	if key != "token:treasure" {
		t.Errorf("TokenKey(treasure) = %q, want token:treasure", key)
	}
	if !IsTokenKey(key) || IsTokenKey(uuid.New().String()) {
		t.Errorf("IsTokenKey does not separate the namespaces: %q", key)
	}
	if got := TokenKey(key); got != key {
		t.Errorf("TokenKey is not idempotent: %q -> %q", key, got)
	}
	if got := TokenKey(""); got != "" {
		t.Errorf("TokenKey(\"\") = %q, want the empty key rather than a bare prefix", got)
	}
}

// TestCatalogKeyPrefersTheOracleID is CR 707.2 stated as a
// precedence rule: a token that is a COPY of a printed card carries
// that card's oracle ID, and the oracle ID must win — otherwise a
// Clone of Llanowar Elves would resolve to a token's entry.
func TestCatalogKeyPrefersTheOracleID(t *testing.T) {
	cases := []struct {
		name string
		card Card
		want string
	}{
		{"a true token", Card{TokenKey: TokenKey("treasure")}, "token:treasure"},
		{"a printed card", Card{OracleID: "oracle-1"}, "oracle-1"},
		{
			"a token copy carrying both",
			Card{OracleID: "oracle-1", TokenKey: TokenKey("treasure")},
			"oracle-1",
		},
		{"neither", Card{}, ""},
	}
	for _, tc := range cases {
		if got := CatalogKey(tc.card); got != tc.want {
			t.Errorf("%s: CatalogKey = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestCopyingATokenCarriesItsKeyAndCopyingACardClearsIt is the same
// rule one layer down, through the real copy machinery: the token key
// is a COPIABLE VALUE, so a copy of a Treasure is a Treasure, and a
// Treasure that becomes a copy of a printed card stops being one.
func TestCopyingATokenCarriesItsKeyAndCopyingACardClearsIt(t *testing.T) {
	treasure := Card{
		InstanceID: uuid.New(),
		Name:       "Treasure",
		TypeLine:   "Token Artifact — Treasure",
		TokenKey:   TokenKey("treasure"),
	}
	printed := Card{
		InstanceID: uuid.New(),
		Name:       "Llanowar Elves",
		TypeLine:   "Creature — Elf Druid",
		OracleID:   "oracle-llanowar",
	}

	clone := Card{InstanceID: uuid.New(), Name: "Clone"}
	clone.applyCopy(CopiableValuesOf(treasure), treasure)
	if got := CatalogKey(clone); got != TokenKey("treasure") {
		t.Errorf("a copy of a Treasure resolves to %q, want the token key", got)
	}

	// The other direction, and the one CR 707.2 is usually quoted
	// for: a TOKEN that copies a printed card resolves to that card.
	tok := treasure
	tok.InstanceID = uuid.New()
	tok.applyCopy(CopiableValuesOf(printed), printed)
	if got := CatalogKey(tok); got != "oracle-llanowar" {
		t.Errorf("a token copy of Llanowar Elves resolves to %q, want the printed oracle ID", got)
	}
	// And leaving the battlefield puts its own identity back.
	tok.restorePrintedSelf()
	if got := CatalogKey(tok); got != TokenKey("treasure") {
		t.Errorf("after the copy ended the token resolves to %q, want its own key", got)
	}
}

// TestATokenWithACatalogKeyIsNotCensused is the issue in one test:
// the board holds a token, the catalog can answer for it, and the
// snapshot is still a restore point.
func TestATokenWithACatalogKeyIsNotCensused(t *testing.T) {
	stubTokenCatalog(t, TokenKey("treasure"),
		[]ManaAbilityShape{{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}}, nil)

	g := newRestorableGame(t)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{
			InstanceID: id,
			Name:       "Treasure",
			TypeLine:   "Token Artifact — Treasure",
			TokenKey:   TokenKey("treasure"),
			Owner:      g.Seats[0].ID,
			Controller: g.Seats[0].ID,
		})
	})

	snap := g.CaptureSnapshot()
	if snap.Continuations.IntrinsicAbilityCards != 0 {
		t.Errorf("a Treasure was censused: %+v", snap.Continuations)
	}
	if !snap.Restorable() {
		t.Fatalf("a board with a Treasure on it is not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	c := findCardInZone(restored.Battlefield, id)
	if c == nil {
		t.Fatal("restored battlefield lost the Treasure")
	}
	if c.TokenKey != TokenKey("treasure") {
		t.Fatalf("restored token key = %q, want the key it was captured with", c.TokenKey)
	}
	// The abilities come back through the ordinary accessor, which is
	// the point: nothing downstream has a token arm.
	if ma := ManaAbilitiesForCard(*c); len(ma) != 1 || !ma[0].SacrificeCost {
		t.Errorf("restored Treasure mana abilities = %+v, want the sac-for-mana ability", ma)
	}
}

// TestInstanceAbilityBeyondTheCatalogEntryIsStillCensused keeps the
// counter reachable. #521 is allowed to make it accurate; it is not
// allowed to make it unreachable, because the moment nothing can
// increment it a genuinely lossy snapshot would pass as a restore
// point.
func TestInstanceAbilityBeyondTheCatalogEntryIsStillCensused(t *testing.T) {
	// The catalog knows this key and declares ONE ability.
	stubTokenCatalog(t, TokenKey("treasure"),
		[]ManaAbilityShape{{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}}, nil)

	g := newRestorableGame(t)
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{
			InstanceID: uuid.New(),
			Name:       "Treasure",
			TypeLine:   "Token Artifact — Treasure",
			TokenKey:   TokenKey("treasure"),
			Owner:      g.Seats[0].ID,
			Controller: g.Seats[0].ID,
			// TWO abilities on the instance where the entry declares
			// one: something stamped a second on at runtime, and
			// restore would rebuild the entry's list and lose it.
			ManaAbilities: []ManaAbilityShape{
				{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"},
				{SacrificeCost: true, Produced: "{C}", Condition: func(*Game, uuid.UUID, uuid.UUID) bool { return true }},
			},
		})
	})

	snap := g.CaptureSnapshot()
	if snap.Continuations.IntrinsicAbilityCards != 1 {
		t.Errorf("census counted %d, want 1; census = %+v",
			snap.Continuations.IntrinsicAbilityCards, snap.Continuations)
	}
	if snap.Restorable() {
		t.Error("a snapshot that would lose an ability must not be a restore point")
	}
}
