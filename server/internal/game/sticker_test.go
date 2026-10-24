package game

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// sticker_test.go — ADR 0208: sticker sheets as data, stickers on
// objects, their layers, their zone moves and their snapshot (CR 123).

// testSheet is a sheet with three name stickers, one ability sticker
// (ticket cost 2), two power and toughness stickers (2 and 5) and an art
// sticker.
func testSheet(name string) StickerSheet {
	oracle := uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)).String()
	return StickerSheet{
		OracleID: oracle,
		Name:     name,
		Stickers: []StickerPrint{
			{Kind: StickerName, Text: "Night", Ordinal: 0},
			{Kind: StickerName, Text: "Brushwagg", Ordinal: 1},
			{Kind: StickerName, Text: "Ringmaster", Ordinal: 2},
			{Kind: StickerAbility, Text: "Menace", TicketCost: 2, Ordinal: 0},
			{Kind: StickerPT, Text: "2/3", TicketCost: 2, Power: 2, Toughness: 3, Ordinal: 0},
			{Kind: StickerPT, Text: "10/10", TicketCost: 5, Power: 10, Toughness: 10, Ordinal: 1},
			{Kind: StickerArt, Ordinal: 0},
		},
	}
}

// withStickerAbilities makes every ability sticker give menace, for the
// span of one test (the catalog's hook, which this package's tests do
// not load).
func withStickerAbilities(t *testing.T) {
	t.Helper()
	prev := CatalogStickerAbility
	CatalogStickerAbility = func(string, int) (StickerAbilityDef, bool) {
		return StickerAbilityDef{Keywords: []string{"menace"}}, true
	}
	t.Cleanup(func() { CatalogStickerAbility = prev })
}

// stickerGame is a started two-seat game whose seat 0 has access to one
// test sheet and `tickets` ticket counters.
func stickerGame(t *testing.T, tickets int) (*Game, *Player, StickerSheet) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0]
	sheet := testSheet("Night Brushwagg Ringmaster")
	g.WithWriteLock(func() {
		me.StickerSheets = []StickerSheet{sheet}
		if tickets > 0 {
			setPlayerCounterLocked(me, CounterTicket, tickets)
		}
	})
	return g, me, sheet
}

func stickerEffective(t *testing.T, g *Game, id uuid.UUID) Characteristic {
	t.Helper()
	return scopedEffectChar(t, g, id)
}

func placeTestSticker(t *testing.T, g *Game, me *Player, on uuid.UUID, sheet StickerSheet, index, after int) {
	t.Helper()
	if !g.PlaceStickerForTest(me.ID, on, StickerID(sheet.OracleID, index), after) {
		t.Fatalf("sticker %d did not go on", index)
	}
}

// CR 103.2d, 123.2a: a deck's sheets reach the game as data, the library
// is untouched, and three of them are chosen at random, the same three
// for the same seed.
func TestStickerSheetsSeatAsDataAndThreeAreChosen(t *testing.T) {
	chosen := func(seed uint64) ([]string, int, int) {
		g := NewGame()
		for i := range 2 {
			deck := buildTestDeck(fmt.Sprintf("Commander %d", i+1))
			if i == 0 {
				for j := range 10 {
					sheet := testSheet(fmt.Sprintf("Sheet %d", j))
					deck = append(deck, Card{InstanceID: uuid.New(), Name: sheet.Name, TypeLine: "Stickers", StickerSheet: &sheet})
				}
			}
			if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
				t.Fatalf("AddPlayer: %v", err)
			}
		}
		if err := g.Start(rand.New(rand.NewPCG(seed, 2))); err != nil {
			t.Fatalf("Start: %v", err)
		}
		me := g.Seats[0]
		var names []string
		for _, s := range me.StickerSheets {
			names = append(names, s.Name)
		}
		return names, len(me.StickerSheetPool), me.Library.Size() + me.Hand.Size()
	}
	a, pool, cards := chosen(7)
	if pool != 10 || len(a) != StickerSheetsChosen {
		t.Fatalf("pool %d, chosen %v: want 10 and 3", pool, a)
	}
	if want := len(buildTestDeck("x")) - 1; cards != want {
		t.Errorf("library+hand = %d, want %d: a sheet went into a zone", cards, want)
	}
	if unique := slices.Compact(slices.Sorted(slices.Values(a))); len(unique) != 3 {
		t.Errorf("chosen %v: a sheet was chosen twice", a)
	}
	b, _, _ := chosen(7)
	if !slices.Equal(a, b) {
		t.Errorf("the same seed chose %v and %v", a, b)
	}
}

// CR 123.6b/c and 123.6a: a word goes after N words, at the end when the
// name has fewer, and a blank is not a word.
func TestStickerInsertNameWord(t *testing.T) {
	for _, tc := range []struct {
		name, word string
		after      int
		want       string
	}{
		{"Grizzly Bears", "Night", 0, "Night Grizzly Bears"},
		{"Grizzly Bears", "Night", 1, "Grizzly Night Bears"},
		{"Grizzly Bears", "Night", 2, "Grizzly Bears Night"},
		{"Grizzly Bears", "Night", 9, "Grizzly Bears Night"},
		{"", "Night", 0, "Night"},
		{"_____ Goblin", "Hot Dog", 1, "_____ Goblin Hot Dog"},
		{"_____ Goblin", "Hot Dog", 0, "Hot Dog _____ Goblin"},
	} {
		if got := InsertNameWord(tc.name, tc.word, tc.after); got != tc.want {
			t.Errorf("InsertNameWord(%q, %q, %d) = %q, want %q", tc.name, tc.word, tc.after, got, tc.want)
		}
	}
	if n := NameWordCount("_____ _____ Rocketship"); n != 1 {
		t.Errorf("NameWordCount = %d, want 1", n)
	}
}

// A name sticker is layer 3, an ability sticker layer 6 and a power and
// toughness sticker layer 7b; counters still count on top (CR 123.6-123.8).
func TestStickersApplyInTheirLayers(t *testing.T) {
	withStickerAbilities(t)
	g, me, sheet := stickerGame(t, 4)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)
	placeTestSticker(t, g, me, bear, sheet, 0, 1) // "Night" after one word
	placeTestSticker(t, g, me, bear, sheet, 3, 0) // menace
	placeTestSticker(t, g, me, bear, sheet, 4, 0) // 2/3
	c := stickerEffective(t, g, bear)
	if c.Name != "Grizzly Night Bears" {
		t.Errorf("name = %q", c.Name)
	}
	if !hasAbility(c.Abilities, "menace") {
		t.Errorf("abilities = %v, want menace", c.Abilities)
	}
	if c.Power != 2 || c.Toughness != 3 {
		t.Errorf("P/T = %d/%d, want 2/3", c.Power, c.Toughness)
	}
	if got := PlayerTickets(me); got != 0 {
		t.Errorf("tickets = %d, want 0 (2 + 2 paid)", got)
	}
}

// CR 613.7: of two power and toughness stickers, the later one wins.
func TestLaterPowerToughnessStickerWins(t *testing.T) {
	g, me, sheet := stickerGame(t, 7)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)
	placeTestSticker(t, g, me, bear, sheet, 5, 0) // 10/10
	placeTestSticker(t, g, me, bear, sheet, 4, 0) // 2/3, later
	if c := stickerEffective(t, g, bear); c.Power != 2 || c.Toughness != 3 {
		t.Errorf("P/T = %d/%d, want the later 2/3", c.Power, c.Toughness)
	}
}

// A sticker's ability is a layer-6 effect at its own timestamp (CR
// 613.7k): a "loses all abilities" from before it leaves it, one from
// after it takes it away.
func TestAbilityStickerAgainstAbilityRemoval(t *testing.T) {
	withStickerAbilities(t)
	g, me, sheet := stickerGame(t, 2)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)
	registerScopedEffectForTest(t, g, bear, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	placeTestSticker(t, g, me, bear, sheet, 3, 0)
	if c := stickerEffective(t, g, bear); !hasAbility(c.Abilities, "menace") {
		t.Fatalf("a removal from before the sticker took its menace: %v", c.Abilities)
	}
	registerScopedEffectForTest(t, g, bear, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	if c := stickerEffective(t, g, bear); hasAbility(c.Abilities, "menace") {
		t.Errorf("a removal from after the sticker left its menace: %v", c.Abilities)
	}
}

// CR 123.3b, 123.3c: no sticker on an object the player doesn't own, none
// they can't pay for, and none already on something they own.
func TestPlaceStickerRefusals(t *testing.T) {
	g, me, sheet := stickerGame(t, 1)
	theirs := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	mine := pushScopedTestCreature(g, me.ID, 2, 2)
	if g.PlaceStickerForTest(me.ID, theirs, StickerID(sheet.OracleID, 0), 0) {
		t.Error("a sticker went on an object its player does not own")
	}
	if g.PlaceStickerForTest(me.ID, mine, StickerID(sheet.OracleID, 4), 0) {
		t.Error("a sticker went on with one ticket for a cost of two")
	}
	if got := PlayerTickets(me); got != 1 {
		t.Errorf("a refused sticker charged tickets: %d left", got)
	}
	placeTestSticker(t, g, me, mine, sheet, 0, 0)
	other := pushScopedTestCreature(g, me.ID, 2, 2)
	if g.PlaceStickerForTest(me.ID, other, StickerID(sheet.OracleID, 0), 0) {
		t.Error("a sticker already on an object went on a second one")
	}
}

// CR 123.5: kept on a move to a public zone, gone on a move to a hidden
// one, and then free to put on something else.
func TestStickersFollowPublicMovesOnly(t *testing.T) {
	g, me, sheet := stickerGame(t, 0)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)
	placeTestSticker(t, g, me, bear, sheet, 1, 0)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	var inYard Card
	g.ReadSnapshot(func() { inYard, _ = g.LookupCardForEffect(bear) })
	if len(inYard.Stickers) != 1 || inYard.Effective().Name != "Brushwagg Grizzly Bears" {
		t.Fatalf("in the graveyard: stickers %v, name %q", inYard.Stickers, inYard.Effective().Name)
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(me.Graveyard, me.Hand, bear); err != nil {
			t.Fatalf("move: %v", err)
		}
	})
	var inHand Card
	g.ReadSnapshot(func() { inHand, _ = g.LookupCardForEffect(bear) })
	if len(inHand.Stickers) != 0 {
		t.Errorf("a card in a hand kept %v", inHand.Stickers)
	}
	var free bool
	g.ReadSnapshot(func() {
		for _, s := range g.AvailableStickersLocked(me.ID, nil) {
			free = free || s.Index == 1
		}
	})
	if !free {
		t.Error("the sticker that came off is not available again")
	}
}

// CR 123.1, 707.2: a token made from a stickered template has none.
func TestTokenNeverInheritsStickers(t *testing.T) {
	g, me, sheet := stickerGame(t, 0)
	tmpl := Card{Name: "Copy", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Stickers: []PlacedSticker{{Sheet: sheet.OracleID, Index: 0, Print: sheet.Stickers[0]}}}
	before := g.Battlefield.Size()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, tmpl, 1); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	if g.Battlefield.Size() != before+1 {
		t.Fatal("no token")
	}
	tok := g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
	if len(tok.Stickers) != 0 {
		t.Errorf("the token has stickers %v", tok.Stickers)
	}
}

// The stickers on an object and the sheets survive a snapshot round trip.
func TestStickersSurviveASnapshot(t *testing.T) {
	g, me, sheet := stickerGame(t, 2)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)
	placeTestSticker(t, g, me, bear, sheet, 4, 0)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if c := stickerEffective(t, restored, bear); c.Power != 2 || c.Toughness != 3 {
		t.Errorf("restored P/T = %d/%d, want 2/3", c.Power, c.Toughness)
	}
	if got := len(restored.Seats[0].StickerSheets); got != 1 {
		t.Errorf("restored sheets = %d, want 1", got)
	}
}
