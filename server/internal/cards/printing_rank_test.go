package cards

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// #3605 / ADR 0192 §1: with no printing named, the picture for a card
// name is the newest printing from the most ordinary kind of set, not
// whichever record the dump happens to list last.

const raffineOracle = "11111111-1111-1111-1111-111111111111"

func pid(n int) string { return fmt.Sprintf("00000000-0000-0000-0000-%012d", n) }

// raffinePrintings are the printings of one card the issue describes:
// the Streets of New Capenna card, its showcase and promo versions, a
// Secret Lair box with newer art, and a digital copy.
var raffinePrintings = []string{
	`{"id":"` + pid(1) + `","oracle_id":"` + raffineOracle + `","name":"Raffine, Scheming Seer","set":"snc","set_type":"expansion","collector_number":"213","released_at":"2022-04-29","border_color":"black","type_line":"Legendary Creature — Sphinx","lang":"en"}`,
	`{"id":"` + pid(2) + `","oracle_id":"` + raffineOracle + `","name":"Raffine, Scheming Seer","set":"snc","set_type":"expansion","collector_number":"333","released_at":"2022-04-29","border_color":"black","frame_effects":["showcase"],"type_line":"Legendary Creature — Sphinx","lang":"en"}`,
	`{"id":"` + pid(3) + `","oracle_id":"` + raffineOracle + `","name":"Raffine, Scheming Seer","set":"snc","set_type":"expansion","collector_number":"414","released_at":"2022-04-29","border_color":"borderless","type_line":"Legendary Creature — Sphinx","lang":"en"}`,
	`{"id":"` + pid(4) + `","oracle_id":"` + raffineOracle + `","name":"Raffine, Scheming Seer","set":"psnc","set_type":"promo","promo":true,"collector_number":"213s","released_at":"2022-04-22","type_line":"Legendary Creature — Sphinx","lang":"en"}`,
	`{"id":"` + pid(5) + `","oracle_id":"` + raffineOracle + `","name":"Raffine, Scheming Seer","set":"slz","set_type":"box","collector_number":"93","released_at":"2026-09-02","border_color":"black","type_line":"Legendary Creature — Sphinx","lang":"en"}`,
	`{"id":"` + pid(6) + `","oracle_id":"` + raffineOracle + `","name":"Raffine, Scheming Seer","set":"sncd","set_type":"expansion","digital":true,"collector_number":"9","released_at":"2026-10-01","type_line":"Legendary Creature — Sphinx","lang":"en"}`,
	`{"id":"` + pid(7) + `","oracle_id":"` + raffineOracle + `","name":"Raffine, Scheming Seer","set":"snc","set_type":"expansion","collector_number":"213","released_at":"2026-10-02","type_line":"Legendary Creature — Sphinx","lang":"ja"}`,
}

func loadRaffine(t *testing.T, order []int) *Index {
	t.Helper()
	var parts []string
	for _, i := range order {
		parts = append(parts, raffinePrintings[i])
	}
	path := filepath.Join(t.TempDir(), "default-cards.json")
	if err := os.WriteFile(path, []byte("["+strings.Join(parts, ",")+"]"), 0o644); err != nil {
		t.Fatal(err)
	}
	idx := NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	return idx
}

func TestDefaultPrintingAvoidsGimmickReprints(t *testing.T) {
	orders := [][]int{
		{0, 1, 2, 3, 4, 5, 6},
		{6, 5, 4, 3, 2, 1, 0}, // the Secret Lair first, the plain printing last
		{4, 0, 6, 2, 5, 1, 3},
	}
	want := uuid.MustParse(pid(1))
	for _, order := range orders {
		idx := loadRaffine(t, order)
		byName, ok := idx.FindByName("Raffine, Scheming Seer")
		if !ok || byName.ID != want {
			t.Errorf("order %v: FindByName = %v %s, want the snc #213 printing", order, ok, byName.SetCode+" "+byName.CollectorNo)
		}
		byOracle, ok := idx.FindByOracleID(uuid.MustParse(raffineOracle))
		if !ok || byOracle.ID != want {
			t.Errorf("order %v: FindByOracleID = %v %s, want the snc #213 printing", order, ok, byOracle.SetCode+" "+byOracle.CollectorNo)
		}
	}
}

// Newest wins inside the best tier, so a real reprint in a later
// ordinary set still beats an older one.
func TestDefaultPrintingTakesTheNewestOfTheBestTier(t *testing.T) {
	idx := NewIndex()
	for _, c := range []Card{
		{ID: uuid.MustParse(pid(11)), OracleID: uuid.MustParse(raffineOracle), Name: "Sol Ring", SetType: "commander", ReleasedAt: "2021-04-23"},
		{ID: uuid.MustParse(pid(12)), OracleID: uuid.MustParse(raffineOracle), Name: "Sol Ring", SetType: "commander", ReleasedAt: "2023-11-17"},
		{ID: uuid.MustParse(pid(13)), OracleID: uuid.MustParse(raffineOracle), Name: "Sol Ring", SetType: "masters", ReleasedAt: "2024-08-01"},
		{ID: uuid.MustParse(pid(14)), OracleID: uuid.MustParse(raffineOracle), Name: "Sol Ring", SetType: "commander", ReleasedAt: "2022-01-01"},
	} {
		idx.Put(c)
	}
	got, _ := idx.FindByName("Sol Ring")
	if got.ID != uuid.MustParse(pid(12)) {
		t.Errorf("default = %s, want the 2023 commander printing", got.ID)
	}
}

// Printings the ranking cannot tell apart keep the old rule: the later
// record wins. Fixtures and cards with no release date depend on it.
func TestRankTieKeepsTheLaterRecord(t *testing.T) {
	idx := NewIndex()
	idx.Put(Card{ID: uuid.MustParse(pid(21)), Name: "Plain"})
	idx.Put(Card{ID: uuid.MustParse(pid(22)), Name: "Plain"})
	got, _ := idx.FindByName("Plain")
	if got.ID != uuid.MustParse(pid(22)) {
		t.Errorf("tie: got %s, want the later record", got.ID)
	}
}

// A card only ever printed in a gimmick set still resolves to it.
func TestAGimmickOnlyCardStillResolves(t *testing.T) {
	idx := NewIndex()
	idx.Put(Card{ID: uuid.MustParse(pid(31)), Name: "Lair Only", SetType: "box", Promo: true, ReleasedAt: "2026-01-01"})
	got, ok := idx.FindByName("Lair Only")
	if !ok || got.ID != uuid.MustParse(pid(31)) {
		t.Errorf("got %v %v", ok, got.ID)
	}
}

func TestFindPrintingOf(t *testing.T) {
	idx := loadRaffine(t, []int{0, 1, 2, 3, 4, 5, 6})
	def, _ := idx.FindByName("Raffine, Scheming Seer")
	other := Card{ID: uuid.MustParse(pid(90)), OracleID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Name: "Someone Else", SetCode: "snc", CollectorNo: "7"}
	idx.Put(other)
	sld := uuid.MustParse(pid(5))

	t.Run("by printing id", func(t *testing.T) {
		got, ok := idx.FindPrintingOf(def, sld, "", "")
		if !ok || got.ID != sld {
			t.Fatalf("got %v %s", ok, got.ID)
		}
	})
	t.Run("by set and collector number, case folded", func(t *testing.T) {
		got, ok := idx.FindPrintingOf(def, uuid.Nil, "SLZ", "93")
		if !ok || got.ID != sld {
			t.Fatalf("got %v %s", ok, got.ID)
		}
	})
	t.Run("an id the index lacks falls back to set and number", func(t *testing.T) {
		got, ok := idx.FindPrintingOf(def, uuid.New(), "slz", "93")
		if !ok || got.ID != sld {
			t.Fatalf("got %v %s", ok, got.ID)
		}
	})
	t.Run("nothing named", func(t *testing.T) {
		if _, ok := idx.FindPrintingOf(def, uuid.Nil, "", ""); ok {
			t.Error("a deck that names no printing must keep the default")
		}
		if _, ok := idx.FindPrintingOf(def, uuid.Nil, "snc", ""); ok {
			t.Error("a set with no number names no printing")
		}
	})
	t.Run("unknown printing", func(t *testing.T) {
		if _, ok := idx.FindPrintingOf(def, uuid.New(), "zzz", "1"); ok {
			t.Error("unknown printing resolved")
		}
	})
	t.Run("a printing of another card is refused", func(t *testing.T) {
		if _, ok := idx.FindPrintingOf(def, other.ID, "", ""); ok {
			t.Error("by id: a different card's printing was honoured")
		}
		if _, ok := idx.FindPrintingOf(def, uuid.Nil, "snc", "7"); ok {
			t.Error("by set and number: a different card's printing was honoured")
		}
	})
	t.Run("a known but unusable printing does not fall through to its set number", func(t *testing.T) {
		tok := Card{ID: uuid.MustParse(pid(91)), OracleID: def.OracleID, Name: def.Name, Layout: "token", SetCode: "tsnc", CollectorNo: "1"}
		idx.Put(tok)
		if _, ok := idx.FindPrintingOf(def, tok.ID, "slz", "93"); ok {
			t.Error("a token record was honoured")
		}
	})
}
