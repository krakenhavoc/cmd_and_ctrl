package deck

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// #3605 / ADR 0192: a deck's source can name the printing of a card,
// and the card the table shows is that printing.

var (
	raffineOracle = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sncID         = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	slzID         = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	solOracle     = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	sol1ID        = uuid.MustParse("00000000-0000-0000-0000-000000000011")
	sol2ID        = uuid.MustParse("00000000-0000-0000-0000-000000000012")
)

// printingIndex holds Raffine in two printings (the Secret Lair is the
// one a name alone would otherwise find only if it ranked first, so it
// is put LAST to be the "later record") and Sol Ring in two sets.
func printingIndex() *cards.Index {
	idx := cards.NewIndex()
	for _, c := range []cards.Card{
		{ID: sncID, OracleID: raffineOracle, Name: "Raffine, Scheming Seer", SetCode: "snc", CollectorNo: "213", SetType: "expansion", ReleasedAt: "2022-04-29"},
		{ID: slzID, OracleID: raffineOracle, Name: "Raffine, Scheming Seer", SetCode: "slz", CollectorNo: "93", SetType: "box", ReleasedAt: "2026-09-02"},
		{ID: sol1ID, OracleID: solOracle, Name: "Sol Ring", SetCode: "c21", CollectorNo: "263", SetType: "commander", ReleasedAt: "2021-04-23"},
		{ID: sol2ID, OracleID: solOracle, Name: "Sol Ring", SetCode: "ltc", CollectorNo: "1", SetType: "commander", ReleasedAt: "2023-06-23"},
	} {
		c.Legalities = map[string]string{"commander": "legal"}
		c.TypeLine = "Artifact"
		idx.Put(c)
	}
	return idx
}

func resolved(t *testing.T, entries []Entry, err error) *List {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	list, rerr := Resolve(printingIndex(), "t", entries)
	if rerr != nil {
		t.Fatal(rerr)
	}
	return list
}

func TestResolveWithNoPrintingTakesTheDefault(t *testing.T) {
	list, err := Resolve(printingIndex(), "t", []Entry{{Name: "Raffine, Scheming Seer", Count: 1, IsCommander: true}})
	if err != nil {
		t.Fatal(err)
	}
	if got := list.Commanders[0].ID; got != sncID {
		t.Errorf("default printing %s, want snc", got)
	}
}

func TestArchidektNamesThePrinting(t *testing.T) {
	raw := `{"name":"d","cards":[
	 {"quantity":1,"categories":["Commander"],"card":{"uid":"` + slzID.String() + `","collectorNumber":"93","edition":{"editioncode":"slz"},"oracleCard":{"name":"Raffine, Scheming Seer"}}},
	 {"quantity":1,"categories":[],"card":{"uid":"not-a-uuid","collectorNumber":"1","edition":{"editioncode":"LTC"},"oracleCard":{"name":"Sol Ring"}}}
	]}`
	_, entries, err := ParseArchidekt([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].PrintingID != slzID || entries[0].SetCode != "slz" || entries[0].CollectorNo != "93" {
		t.Errorf("commander entry %+v", entries[0])
	}
	if entries[1].PrintingID != uuid.Nil {
		t.Errorf("a malformed uid must read as no id, got %s", entries[1].PrintingID)
	}
	list := resolved(t, entries, nil)
	if got := list.Commanders[0].ID; got != slzID {
		t.Errorf("commander printing %s, want the one the deck names (slz)", got)
	}
	if got := list.Mainboard[0].ID; got != sol2ID {
		t.Errorf("Sol Ring %s, want LTC #1 by set and number", got)
	}
}

func TestMoxfieldNamesThePrinting(t *testing.T) {
	raw := `{"name":"d","boards":{
	 "commanders":{"cards":{"a":{"quantity":1,"card":{"name":"Raffine, Scheming Seer","scryfall_id":"` + slzID.String() + `","set":"slz","cn":"93"}}}},
	 "mainboard":{"cards":{"b":{"quantity":1,"card":{"name":"Sol Ring","scryfall_id":"` + sol2ID.String() + `","set":"ltc","cn":"1"}},
	                       "c":{"quantity":1,"card":{"name":"Raffine, Scheming Seer"}}}}}}`
	_, entries, err := ParseMoxfield([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	list := resolved(t, entries, nil)
	if got := list.Commanders[0].ID; got != slzID {
		t.Errorf("commander printing %s, want slz", got)
	}
	for _, c := range list.Mainboard {
		switch c.Name {
		case "Sol Ring":
			if c.ID != sol2ID {
				t.Errorf("Sol Ring %s, want the named printing", c.ID)
			}
		default:
			if c.ID != sncID {
				t.Errorf("%s %s: a row with no printing keeps the default", c.Name, c.ID)
			}
		}
	}
}

func TestTextListNamesThePrinting(t *testing.T) {
	src := "Commander:\n1 Raffine, Scheming Seer (SLZ) 93\n\nMainboard:\n" +
		"1 Sol Ring (LTC) 1 *F*\n" +
		"1 Sol Ring (C21)\n" + // a set with no number names no printing
		"1 Sol Ring (ZZZ) 9\n" + // a printing the index lacks
		"1 Sol Ring\n"
	entries, err := ParseText(src)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].SetCode != "SLZ" || entries[0].CollectorNo != "93" {
		t.Errorf("commander entry %+v", entries[0])
	}
	if entries[1].SetCode != "LTC" || entries[1].CollectorNo != "1" {
		t.Errorf("foil marker row %+v", entries[1])
	}
	if entries[2].SetCode != "C21" || entries[2].CollectorNo != "" {
		t.Errorf("set-only row %+v", entries[2])
	}
	list := resolved(t, entries, nil)
	if got := list.Commanders[0].ID; got != slzID {
		t.Errorf("commander printing %s, want slz", got)
	}
	// LTC #1 by number; the rest fall to the default (the newest commander printing: LTC).
	for i, c := range list.Mainboard {
		if c.ID != sol2ID {
			t.Errorf("Sol Ring row %d resolved to %s", i, c.ID)
		}
	}
}

// A different card's printing is ignored: a deck cannot swap the card
// it holds by naming someone else's set and number.
func TestAPrintingOfAnotherCardIsIgnored(t *testing.T) {
	entries := []Entry{{Name: "Sol Ring", Count: 1, PrintingID: slzID}, {Name: "Sol Ring", Count: 1, SetCode: "slz", CollectorNo: "93"}}
	list, err := Resolve(printingIndex(), "t", entries)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range list.Mainboard {
		if c.Name != "Sol Ring" || c.OracleID != solOracle {
			t.Errorf("got %s %s", c.Name, c.OracleID)
		}
	}
}

// ADR 0192 §2: a deck fetched from a link is saved as List.Text(), so
// the text must carry the printing, and an old saved list with none
// must still parse to the default.
func TestTextCarriesThePrintingThroughASave(t *testing.T) {
	idx := printingIndex()
	list, err := Resolve(idx, "t", []Entry{
		{Name: "Raffine, Scheming Seer", Count: 1, IsCommander: true, PrintingID: slzID},
		{Name: "Sol Ring", Count: 1, PrintingID: sol1ID},
		{Name: "Sol Ring", Count: 2, PrintingID: sol2ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := list.Text()
	for _, want := range []string{"1 Raffine, Scheming Seer (SLZ) 93", "1 Sol Ring (C21) 263", "2 Sol Ring (LTC) 1"} {
		if !strings.Contains(text, want) {
			t.Errorf("saved text lacks %q:\n%s", want, text)
		}
	}
	entries, err := ParseText(text)
	if err != nil {
		t.Fatal(err)
	}
	back, err := Resolve(idx, "t", entries)
	if err != nil {
		t.Fatal(err)
	}
	if back.Commanders[0].ID != slzID {
		t.Errorf("commander came back as %s", back.Commanders[0].ID)
	}
	var ids []uuid.UUID
	for _, c := range back.Mainboard {
		ids = append(ids, c.ID)
	}
	if len(ids) != 3 || ids[0] != sol1ID || ids[1] != sol2ID || ids[2] != sol2ID {
		t.Errorf("mainboard printings %v", ids)
	}

	// A list saved before printings were recorded has no suffix.
	old, err := ParseText("Commander:\n1 Raffine, Scheming Seer\n\nMainboard:\n1 Sol Ring\n")
	if err != nil {
		t.Fatal(err)
	}
	oldList, err := Resolve(idx, "t", old)
	if err != nil {
		t.Fatal(err)
	}
	if oldList.Commanders[0].ID != sncID || oldList.Mainboard[0].ID != sol2ID {
		t.Errorf("an old list resolved to %s / %s, want the defaults", oldList.Commanders[0].ID, oldList.Mainboard[0].ID)
	}
}
