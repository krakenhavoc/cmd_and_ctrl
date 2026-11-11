package cubes

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
)

var testNS = uuid.MustParse("00000000-0000-0000-0000-000000005441")

func testID(parts ...any) uuid.UUID { return uuid.NewSHA1(testNS, []byte(fmt.Sprint(parts...))) }

// testIndex holds n made-up cards "Card 000"..., one of them with a
// second printing, and the five basic lands.
func testIndex(n int) *cards.Index {
	idx := cards.NewIndex()
	for i := range n {
		name := fmt.Sprintf("Card %03d", i)
		idx.Put(cards.Card{
			ID: testID("print", name), OracleID: testID("oracle", name), Name: name,
			SetCode: "cub", SetName: "Cube test", SetType: "expansion", CollectorNo: fmt.Sprint(i + 1),
			Lang: "en", Layout: "normal", TypeLine: "Creature — Test", BorderColor: "black",
			ReleasedAt: "2026-03-01", Rarity: "common", ManaCost: "{1}{G}", Colors: []string{"G"},
		})
	}
	idx.Put(cards.Card{
		ID: testID("print", "Card 000", "alt"), OracleID: testID("oracle", "Card 000"), Name: "Card 000",
		SetCode: "alt", SetName: "Alternate", SetType: "expansion", CollectorNo: "77",
		Lang: "en", Layout: "normal", TypeLine: "Creature — Test", BorderColor: "black",
		ReleasedAt: "2020-03-01", Rarity: "common", ManaCost: "{1}{G}", Colors: []string{"G"},
	})
	for i, b := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		idx.Put(cards.Card{
			ID: testID("print", b), OracleID: testID("oracle", b), Name: b,
			SetCode: "cub", SetName: "Cube test", SetType: "expansion", CollectorNo: fmt.Sprint(900 + i),
			Lang: "en", Layout: "normal", TypeLine: "Basic Land — " + b, BorderColor: "black",
			ReleasedAt: "2026-03-01", Rarity: "common",
		})
	}
	return idx
}

// fullBut is a catalog verdict that calls every card Full but the named
// ones.
func fullBut(names ...string) func(uuid.UUID) bool {
	not := map[uuid.UUID]bool{}
	for _, n := range names {
		not[testID("oracle", n)] = true
	}
	return func(id uuid.UUID) bool { return !not[id] }
}

// list is "1 Card 000" ... for the first n cards.
func list(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "1 Card %03d\n", i)
	}
	return b.String()
}

func check(t *testing.T, idx *cards.Index, isFull func(uuid.UUID) bool, text string, singleton bool) Report {
	t.Helper()
	entries, err := ParseList(text)
	if err != nil {
		t.Fatalf("ParseList: %v", err)
	}
	return Check(idx, isFull, entries, singleton)
}

// A card that is not Full is named and blocks the cube, and so is a
// name the card index does not know. Dropping them leaves a cube that
// is ready.
func TestACardThatIsNotFullIsNamedAndBlocks(t *testing.T) {
	idx := testIndex(200)
	isFull := fullBut("Card 007", "Card 150")
	rep := check(t, idx, isFull, list(180)+"1 Not A Real Card\n", true)
	if rep.Ready {
		t.Fatal("a cube with cards that are not Full is ready")
	}
	if want := []string{"Card 007", "Card 150"}; !reflect.DeepEqual(rep.NotFull, want) {
		t.Fatalf("not_full = %v, want %v", rep.NotFull, want)
	}
	if want := []string{"Not A Real Card"}; !reflect.DeepEqual(rep.Unknown, want) {
		t.Fatalf("unknown = %v, want %v", rep.Unknown, want)
	}
	if rep.Cards != 181 {
		t.Fatalf("cards = %d, want 181", rep.Cards)
	}

	// The owner drops the three rows the check named.
	var kept []string
	for _, line := range strings.Split(strings.TrimSpace(rep.Text()), "\n") {
		name := strings.SplitN(line, " ", 2)[1]
		if name != "Card 007" && name != "Card 150" && name != "Not A Real Card" {
			kept = append(kept, line)
		}
	}
	rep = check(t, idx, isFull, strings.Join(kept, "\n"), true)
	if !rep.Ready || rep.Cards != 178 || len(rep.NotFull)+len(rep.Unknown) != 0 {
		t.Fatalf("after dropping the blockers: ready %v, cards %d, not_full %v, unknown %v", rep.Ready, rep.Cards, rep.NotFull, rep.Unknown)
	}
	// A nil verdict (no catalog) calls nothing Full.
	if rep := Check(idx, nil, mustParse(t, list(3)), true); rep.Ready || len(rep.NotFull) != 3 {
		t.Fatalf("with no catalog: ready %v, not_full %v", rep.Ready, rep.NotFull)
	}
}

func mustParse(t *testing.T, text string) []deck.Entry {
	t.Helper()
	entries, err := ParseList(text)
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

// Singleton is the default and is enforced: a second copy is named and
// blocks, whether it is a count of two, two rows or two printings.
// Basic lands are exempt. With singleton off the same list is ready.
func TestSingletonIsEnforcedOrWaived(t *testing.T) {
	idx := testIndex(200)
	text := list(120) +
		"2 Card 150\n" + // a count of two
		"1 Card 003\n" + // a second row of a card already listed
		"1 Card 000 (ALT) 77\n" + // a second printing of one
		"6 Forest\n1 Forest\n"
	rep := check(t, idx, fullBut(), text, true)
	if rep.Ready {
		t.Fatal("a singleton cube with repeated cards is ready")
	}
	if want := []string{"Card 000", "Card 003", "Card 150"}; !reflect.DeepEqual(rep.Repeated, want) {
		t.Fatalf("repeated = %v, want %v (basic lands are exempt)", rep.Repeated, want)
	}
	rep = check(t, idx, fullBut(), text, false)
	if !rep.Ready || len(rep.Repeated) != 0 {
		t.Fatalf("with singleton off: ready %v, repeated %v", rep.Ready, rep.Repeated)
	}
	if rep.Cards != 120+2+1+1+7 {
		t.Fatalf("cards = %d, want %d", rep.Cards, 120+2+1+1+7)
	}
	if got := len(rep.Deal()); got != rep.Cards {
		t.Fatalf("Deal hands out %d cards, want every copy (%d)", got, rep.Cards)
	}
	// The two printings of Card 000 are two rows and deal two printings.
	printings := map[uuid.UUID]int{}
	for _, c := range rep.Deal() {
		if c.Name == "Card 000" {
			printings[c.ID]++
		}
	}
	if len(printings) != 2 {
		t.Fatalf("Card 000 deals %d printings, want the two the list named", len(printings))
	}
}

// Size: a draft needs seats x 3 packs x pack size, and the check says
// how large a draft of 15-card packs the cube can deal.
func TestACubeTooSmallForTheSeats(t *testing.T) {
	if got := Need(8, 15); got != 360 {
		t.Fatalf("Need(8, 15) = %d, want 360", got)
	}
	if got := Need(4, draft.DefaultCubePackSize); got != 180 {
		t.Fatalf("Need(4, default) = %d, want 180", got)
	}
	idx := testIndex(400)
	for _, tc := range []struct{ cards, maxSeats int }{
		{89, 0},  // not even two seats
		{90, 2},  // exactly two
		{179, 3}, // one short of four
		{180, 4},
		{359, 7},
		{360, 8},
		{400, 8}, // never more than a draft seats
	} {
		rep := check(t, idx, fullBut(), list(tc.cards), true)
		if rep.MaxSeats != tc.maxSeats {
			t.Errorf("%d cards: max_seats = %d, want %d", tc.cards, rep.MaxSeats, tc.maxSeats)
		}
		// Too small for a draft is not a fault of the list.
		if !rep.Ready {
			t.Errorf("%d cards: the list is sound but not ready", tc.cards)
		}
		for seats := draft.MinSeats; seats <= draft.MaxSeats; seats++ {
			if fits := rep.Cards >= Need(seats, 15); fits != (seats <= tc.maxSeats) {
				t.Errorf("%d cards, %d seats: fits = %v, max_seats = %d", tc.cards, seats, fits, tc.maxSeats)
			}
		}
	}
	// The cap: a list over MaxCards is not ready.
	entries := mustParse(t, fmt.Sprintf("%d Forest\n", MaxCards+1))
	if rep := Check(idx, fullBut(), entries, false); rep.Ready || !rep.TooLarge {
		t.Fatalf("a list of %d cards: ready %v, too_large %v", MaxCards+1, rep.Ready, rep.TooLarge)
	}
	if rep := Check(idx, fullBut(), nil, true); rep.Ready {
		t.Fatal("an empty list is ready")
	}
}

// A cube is a list: a plain list of names reads as one copy each, and
// the commander, sideboard and companion rows of a deck list are cards
// of the cube like any other.
func TestParseListReadsPlainNamesAndFlattensADeck(t *testing.T) {
	idx := testIndex(20)
	rep := check(t, idx, fullBut(), "Card 001\nCard 002\n\n# a comment\n3x Card 003\nCard 004 (CUB) 5\n", false)
	if rep.Cards != 6 || len(rep.Unknown) != 0 {
		t.Fatalf("plain names: cards %d, unknown %v", rep.Cards, rep.Unknown)
	}
	rep = check(t, idx, fullBut(), "Commander:\n1 Card 001\n\nMainboard:\n1 Card 002\n1 Card 003\n\nSideboard:\n1 Card 004\nSB: 1 Card 005\n\n1 Card 006\n", true)
	if rep.Cards != 6 || !rep.Ready {
		t.Fatalf("a deck list: cards %d, ready %v, unknown %v", rep.Cards, rep.Ready, rep.Unknown)
	}
	// Text reads back to the same rows.
	again := check(t, idx, fullBut(), rep.Text(), true)
	if !reflect.DeepEqual(again.Rows(), rep.Rows()) {
		t.Fatalf("Text does not read back:\n%v\n%v", again.Rows(), rep.Rows())
	}
}

func openStore(t *testing.T) (*SQLStore, *db.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLStore(d), d
}

func addUser(t *testing.T, d *db.DB, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, id.String(), name); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestStoreKeepsACubeForItsOwnerAlone(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	ann, bob := addUser(t, d, "Ann"), addUser(t, d, "Bob")

	c, err := s.Create(ctx, Cube{OwnerID: ann, Name: "Vintage", Singleton: true, List: "1 Card 001\n", CardCount: 1})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := s.Create(ctx, Cube{OwnerID: ann, Name: "Vintage", List: "1 Card 002\n", CardCount: 1}); !errors.Is(err, ErrNameTaken) {
		t.Fatalf("a second cube of one name: err = %v, want ErrNameTaken", err)
	}
	// Another person may use the name, and sees only their own.
	if _, err := s.Create(ctx, Cube{OwnerID: bob, Name: "Vintage", List: "1 Card 003\n", CardCount: 1}); err != nil {
		t.Fatalf("Bob's cube of the same name: %v", err)
	}
	if _, err := s.Get(ctx, bob, c.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Bob reading Ann's cube: err = %v, want ErrNotFound", err)
	}
	if err := s.Delete(ctx, bob, c.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Bob deleting Ann's cube: err = %v, want ErrNotFound", err)
	}
	if _, err := s.Update(ctx, Cube{ID: c.ID, OwnerID: bob, Name: "Mine now", List: "x"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Bob changing Ann's cube: err = %v, want ErrNotFound", err)
	}

	got, err := s.Get(ctx, ann, c.ID)
	if err != nil || got.Name != "Vintage" || !got.Singleton || got.List != "1 Card 001\n" || got.CardCount != 1 {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	got.Name, got.Singleton, got.List, got.CardCount = "Legacy", false, "2 Card 001\n", 2
	upd, err := s.Update(ctx, got)
	if err != nil || upd.Name != "Legacy" || upd.Singleton || upd.CardCount != 2 {
		t.Fatalf("Update = %+v, %v", upd, err)
	}
	all, err := s.List(ctx, ann)
	if err != nil || len(all) != 1 || all[0].Name != "Legacy" {
		t.Fatalf("List = %+v, %v", all, err)
	}
	if err := s.Delete(ctx, ann, c.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if all, _ := s.List(ctx, ann); len(all) != 0 {
		t.Fatalf("after Delete Ann has %d cubes", len(all))
	}
}

func TestStoreCapsAPersonsCubesAndForgetsAnAccount(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	ann, bob := addUser(t, d, "Ann"), addUser(t, d, "Bob")
	for i := range MaxCubes {
		if _, err := s.Create(ctx, Cube{OwnerID: ann, Name: fmt.Sprintf("Cube %d", i), List: "1 Forest\n", CardCount: 1}); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}
	if _, err := s.Create(ctx, Cube{OwnerID: ann, Name: "One too many", List: "1 Forest\n", CardCount: 1}); !errors.Is(err, ErrTooMany) {
		t.Fatalf("cube %d: err = %v, want ErrTooMany", MaxCubes+1, err)
	}
	if _, err := s.Create(ctx, Cube{OwnerID: bob, Name: "Bob's", List: "1 Forest\n", CardCount: 1}); err != nil {
		t.Fatal(err)
	}

	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	n, err := ForgetAccount(ctx, tx, ann.String())
	if err != nil || n != MaxCubes {
		t.Fatalf("ForgetAccount = %d, %v; want %d", n, err, MaxCubes)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if all, _ := s.List(ctx, ann); len(all) != 0 {
		t.Fatalf("Ann still has %d cubes", len(all))
	}
	if all, _ := s.List(ctx, bob); len(all) != 1 {
		t.Fatalf("Bob has %d cubes, want his one", len(all))
	}
}
