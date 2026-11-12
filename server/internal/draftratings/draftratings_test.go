package draftratings

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// setFile is a file of n cards "<prefix> 0".."<prefix> n-1" whose win
// rate climbs 0.40, 0.41, ... so a card's rank is its number.
func setFile(code, start string, prefix string, n int) string {
	var rows []string
	for i := 0; i < n; i++ {
		rows = append(rows, fmt.Sprintf(`{"n":"%s %d","w":%.4f,"g":600,"s":5,"p":6}`, prefix, i, 0.40+float64(i)*0.002))
	}
	return fmt.Sprintf(`{"set":%q,"source":"17lands","format":"PremierDraft","start":%q,"fetched":"2026-11-11","min_games":500,"cards":[%s]}`,
		code, start, strings.Join(rows, ","))
}

func TestTheEmbeddedFilesLoadAndAreWellFormed(t *testing.T) {
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Sets()) == 0 {
		t.Fatal("no ratings file is embedded")
	}
	for _, code := range b.Sets() {
		s := b.sets[code]
		if s.Source != "17lands" || s.Format == "" || s.Fetched == "" || s.Start == "" {
			t.Errorf("%s: incomplete header %+v", code, s.file)
		}
		if len(s.Cards) < 40 {
			t.Errorf("%s: %d cards, the refresh script writes no file under 40", code, len(s.Cards))
		}
		seen := map[string]bool{}
		for _, c := range s.Cards {
			if c.Games < s.MinGames {
				t.Errorf("%s: %s has %d games, under the file's own bar of %d", code, c.Name, c.Games, s.MinGames)
			}
			if c.WinRate <= 0.2 || c.WinRate >= 0.8 {
				t.Errorf("%s: %s win rate %.4f is not a fraction of games", code, c.Name, c.WinRate)
			}
			if k := nameKey(c.Name); seen[k] {
				t.Errorf("%s: %s is listed twice", code, c.Name)
			} else {
				seen[k] = true
			}
		}
	}
}

func TestLoadRefusesAFileThatIsNotItsSet(t *testing.T) {
	for name, fsys := range map[string]fstest.MapFS{
		"wrong set": {"data/aaa.json": {Data: []byte(setFile("bbb", "2026-01-01", "X", 3))}},
		"no cards":  {"data/aaa.json": {Data: []byte(`{"set":"aaa","cards":[]}`)}},
		"not json":  {"data/aaa.json": {Data: []byte(`nope`)}},
	} {
		if _, err := LoadFS(fsys, "data"); err == nil {
			t.Errorf("%s: loaded", name)
		}
	}
}

func TestTiersAreReadOffTheSetsOwnSpread(t *testing.T) {
	b, err := LoadFS(fstest.MapFS{"data/aaa.json": {Data: []byte(setFile("aaa", "2026-01-01", "Card", 100))}}, "data")
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	prevRank := -1
	order := map[string]int{}
	for i, tier := range Tiers {
		order[tier.Grade] = len(Tiers) - i // F lowest
	}
	for i := 0; i < 100; i++ {
		r, ok := b.In("aaa", uuid.Nil, fmt.Sprintf("Card %d", i))
		if !ok {
			t.Fatalf("Card %d not found", i)
		}
		counts[r.Tier]++
		if order[r.Tier] < prevRank {
			t.Errorf("Card %d is %s, below the card under it", i, r.Tier)
		}
		prevRank = order[r.Tier]
	}
	for grade, want := range map[string]int{"A+": 3, "A": 7, "A-": 10, "B+": 10, "B": 15, "B-": 15, "C+": 15, "C": 10, "C-": 7, "D": 5, "F": 3} {
		if counts[grade] != want {
			t.Errorf("%s: %d cards, want %d (%v)", grade, counts[grade], want, counts)
		}
	}
	top, _ := b.In("aaa", uuid.Nil, "Card 99")
	if top.Percent() != 59.8 || top.Games != 600 || top.Z < 1.5 {
		t.Errorf("top card: %+v", top)
	}
	mid, _ := b.In("aaa", uuid.Nil, "Card 50")
	if mid.Z < -0.1 || mid.Z > 0.1 {
		t.Errorf("median card z = %.2f, want about 0", mid.Z)
	}
}

func TestDoubleFacedNamesMatchTheirFrontFace(t *testing.T) {
	data := `{"set":"aaa","source":"17lands","format":"PremierDraft","start":"2026-01-01","fetched":"2026-11-11","min_games":500,"cards":[` +
		`{"n":"Fire // Ice","w":0.55,"g":600,"s":1,"p":1},{"n":"Delver","w":0.5,"g":600,"s":1,"p":1}]}`
	b, err := LoadFS(fstest.MapFS{"data/aaa.json": {Data: []byte(data)}}, "data")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Fire // Ice", "Fire", "fire  // ice", "FIRE"} {
		if _, ok := b.In("aaa", uuid.Nil, name); !ok {
			t.Errorf("%q not found", name)
		}
	}
	if _, ok := b.In("aaa", uuid.Nil, "Ice"); ok {
		t.Error("the back face found the card")
	}
}

func TestBindFindsACardByIdentityWhenTheNameDiffers(t *testing.T) {
	oracle := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	idx := cards.NewIndex()
	idx.Put(cards.Card{ID: uuid.New(), OracleID: oracle, Name: "Real Name", FlavorName: "Arena Title", SetCode: "aaa", Lang: "en"})
	data := `{"set":"aaa","source":"17lands","format":"PremierDraft","start":"2026-01-01","fetched":"2026-11-11","min_games":500,"cards":[` +
		`{"n":"Arena Title","w":0.55,"g":600,"s":1,"p":1}]}`
	b, err := LoadFS(fstest.MapFS{"data/aaa.json": {Data: []byte(data)}}, "data")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := b.In("aaa", oracle, "Real Name"); ok {
		t.Fatal("found by a name nobody wrote, before Bind")
	}
	b.Bind(idx)
	if _, ok := b.In("aaa", oracle, "Real Name"); !ok {
		t.Error("Bind did not map the card's title to its oracle ID")
	}
}

func TestACubeCardIsRatedByTheMostRecentSetThatHasIt(t *testing.T) {
	old := strings.ReplaceAll(setFile("old", "2024-01-01", "Shared", 50), `"w":0.4000`, `"w":0.4100`)
	fsys := fstest.MapFS{
		"data/old.json": {Data: []byte(old)},
		"data/new.json": {Data: []byte(setFile("new", "2026-01-01", "Shared", 45))},
		"data/mid.json": {Data: []byte(setFile("mid", "2025-01-01", "Only Mid", 45))},
	}
	b, err := LoadFS(fsys, "data")
	if err != nil {
		t.Fatal(err)
	}
	if got := b.Sets(); strings.Join(got, ",") != "new,mid,old" {
		t.Fatalf("Sets = %v, want most recent first", got)
	}
	// "Shared 47" is in old (50 cards) and in new only up to 44.
	if r, ok := b.Latest(uuid.Nil, "Shared 47"); !ok || r.Set != "old" {
		t.Errorf("Shared 47 is from %q, want old", r.Set)
	}
	if r, ok := b.Latest(uuid.Nil, "Shared 3"); !ok || r.Set != "new" {
		t.Errorf("Shared 3 is from %q, want new, the most recent", r.Set)
	}
	// For prefers the card's own set, then falls back to the most recent.
	if r, ok := b.For("old", uuid.Nil, "Shared 3"); !ok || r.Set != "old" {
		t.Errorf("For(old) = %q, want the card's own set", r.Set)
	}
	if r, ok := b.For("mid", uuid.Nil, "Shared 3"); !ok || r.Set != "new" {
		t.Errorf("For(mid) = %q, want the latest set that has it", r.Set)
	}
	if _, ok := b.For("new", uuid.Nil, "Nothing Of This Name"); ok {
		t.Error("an unknown card has a rating")
	}
	var none *Book
	if _, ok := none.For("new", uuid.Nil, "x"); ok || none.Has("new") || none.Cards("new") != 0 || none.Sets() != nil {
		t.Error("a nil Book rated something")
	}
}
