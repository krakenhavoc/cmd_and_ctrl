package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// format_limited_test.go: the limited format (ADR 0222 §2 and §6,
// #5440), and that adding it changed nothing about a Commander table.

// limitedDeck is 40 cards and a sideboard of side cards, the rest of a
// card pool (CR 100.4b).
func limitedDeck(side int) []Card {
	deck := make([]Card, 0, 40+side)
	for i := range 40 {
		deck = append(deck, NewCard(fmt.Sprintf("Drafted Card %d", i+1), uuid.Nil))
	}
	for i := range side {
		c := NewCard(fmt.Sprintf("Left Out %d", i+1), uuid.Nil)
		c.IsSideboard = true
		deck = append(deck, c)
	}
	return deck
}

func TestTheLimitedRow(t *testing.T) {
	l := FormatLimited.Spec()
	if l.Format != FormatLimited || l.StartingLife != 20 || l.MaxSeats != 2 || l.CommanderExpected || !l.Matches || !l.DraftOnly {
		t.Errorf("limited row = %+v", l)
	}
	if l.Deck != (DeckRules{MinCards: 40, AnySideboard: true, FromPool: true}) {
		t.Errorf("limited deck rules = %+v", l.Deck)
	}
	if _, capped := l.SideboardCap(); capped {
		t.Error("a limited sideboard has a size cap; it is the rest of the pool (CR 100.4b)")
	}
	// No other format is draft-only or has an uncapped sideboard.
	for _, f := range Formats() {
		if f.Format == FormatLimited {
			continue
		}
		if f.DraftOnly || f.Deck.AnySideboard || f.Deck.FromPool {
			t.Errorf("%s took on a limited-only rule: %+v", f.Format, f)
		}
		if n, capped := f.SideboardCap(); !capped || n != MaxSideboardSize {
			t.Errorf("%s: sideboard cap %d (capped %v), want %d", f.Format, n, capped, MaxSideboardSize)
		}
	}
}

// A limited table is opened by a draft (OpenLimited): 20 life, two
// seats, a match, and a sideboard as large as the pool left over.
func TestOpenLimited(t *testing.T) {
	g := NewGame()
	if err := g.OpenLimited(3); err != nil {
		t.Fatalf("OpenLimited: %v", err)
	}
	s := g.TableSettingsSnapshot()
	if s.EffectiveFormat() != FormatLimited || s.StartingLife != 20 || s.EffectiveBestOf() != 3 || g.SeatCap() != 2 {
		t.Fatalf("settings = %+v, seats %d", s, g.SeatCap())
	}
	// A pool of 45 leaves 22 or more out of a 23-card deck: far past the
	// fifteen a constructed sideboard holds.
	a, err := g.AddPlayer("A", limitedDeck(25))
	if err != nil {
		t.Fatalf("a limited seat with 25 sideboard cards: %v", err)
	}
	if err := g.ReplaceDeck(a.ID, limitedDeck(25)); err != nil {
		t.Fatalf("ReplaceDeck: %v", err)
	}
	if _, err := g.AddPlayer("B", limitedDeck(25)); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddPlayer("C", limitedDeck(0)); !errors.Is(err, ErrGameFull) {
		t.Fatalf("a third seat at a limited table: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	m := g.MatchSnapshot()
	if m == nil || m.BestOf != 3 || len(m.Decks) != 2 || len(m.Decks[0].Main) != 40 || len(m.Decks[0].Side) != 25 {
		t.Fatalf("match = %+v", m)
	}
	for _, p := range g.Seats {
		if p.Life != 20 || len(p.Command.Cards) != 0 {
			t.Errorf("%s: life %d, %d cards in the command zone", p.Name, p.Life, len(p.Command.Cards))
		}
	}
	if g.HasCommander() {
		t.Error("a limited table reports a commander")
	}
	// Between games the deck check is the limited row's: 40 cards, any
	// sideboard, no copy limit.
	spec := FormatLimited.Spec()
	main, side := limitedDeck(0), limitedDeck(30)[40:]
	if err := CheckMatchDeck(spec, main, side, 0, nil); err != nil {
		t.Errorf("a 40-card deck with a 30-card sideboard between games: %v", err)
	}
	if err := CheckMatchDeck(spec, main[:39], side, 0, nil); !errors.Is(err, ErrSideboardIllegal) {
		t.Errorf("a 39-card deck between games: %v", err)
	}
	five := append([]Card(nil), main...)
	for i := range 5 {
		five[i].Name = "Same Common"
	}
	if err := CheckMatchDeck(spec, five, side, 0, nil); err != nil {
		t.Errorf("five copies of a card the pool holds five of: %v (CR 100.2b has no four-card limit)", err)
	}
}

// OpenLimited is for an empty table in its lobby, and nothing else
// makes a limited table: not a settings patch, on any table.
func TestLimitedIsNotChosenByHand(t *testing.T) {
	limited, standard, commander := FormatLimited, FormatStandard, FormatCommander
	g := NewGame()
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Format: &limited}); !errors.Is(err, ErrInvalidSetting) {
		t.Fatalf("a patch to limited: %v", err)
	}
	if got := g.TableSettingsSnapshot(); got.Format != "" || got.StartingLife != StartingLife {
		t.Fatalf("a refused patch changed the table: %+v", got)
	}
	if err := g.OpenLimited(2); !errors.Is(err, ErrInvalidSetting) {
		t.Fatalf("OpenLimited(2): %v", err)
	}
	seated := NewGame()
	if _, err := seated.AddPlayer("A", buildSixtyCardDeck()); err != nil {
		t.Fatal(err)
	}
	if err := seated.OpenLimited(1); !errors.Is(err, ErrGameNotInLobby) {
		t.Fatalf("OpenLimited on a table with a seat: %v", err)
	}
	// And a limited table stays one.
	lim := NewGame()
	if err := lim.OpenLimited(1); err != nil {
		t.Fatal(err)
	}
	for _, f := range []*Format{&standard, &commander} {
		if err := lim.UpdateSettings(uuid.Nil, SettingsPatch{Format: f}); !errors.Is(err, ErrInvalidSetting) {
			t.Fatalf("a limited table patched to %s: %v", *f, err)
		}
	}
}

// Commander is untouched: the sideboard cap, the seat cap, the life,
// and a snapshot with no format field in it.
func TestLimitedChangesNothingAboutCommanderOrStandard(t *testing.T) {
	for _, f := range []Format{FormatCommander, FormatStandard} {
		g := NewGame()
		if f != FormatCommander {
			setFormat(t, g, f)
		}
		over := limitedDeck(MaxSideboardSize + 1)
		if _, err := g.AddPlayer("A", over); !errors.Is(err, ErrSideboardTooLarge) {
			t.Errorf("%s: a seat with %d sideboard cards: %v, want ErrSideboardTooLarge", f, MaxSideboardSize+1, err)
		}
		p, err := g.AddPlayer("A", limitedDeck(MaxSideboardSize))
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if err := g.ReplaceDeck(p.ID, over); !errors.Is(err, ErrSideboardTooLarge) {
			t.Errorf("%s: ReplaceDeck with %d sideboard cards: %v", f, MaxSideboardSize+1, err)
		}
	}

	g := NewGame()
	for i := range 4 {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildSixtyCardDeck()); err != nil {
			t.Fatalf("seat %d at a Commander table: %v", i+1, err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatal(err)
	}
	if g.Match != nil {
		t.Error("a Commander table has a match")
	}
	for _, p := range g.Seats {
		if p.Life != StartingLife {
			t.Errorf("%s starts at %d life, want %d", p.Name, p.Life, StartingLife)
		}
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"format"`, `"bestOf"`, `"best_of"`, "limited"} {
		if strings.Contains(string(raw), field) {
			t.Errorf("a Commander table's snapshot contains %s", field)
		}
	}
	// A snapshot with no format field reads as Commander.
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if got := restored.TableSettingsSnapshot(); got.EffectiveFormat() != FormatCommander || got.FormatSpec().MaxSeats != MaxPlayers {
		t.Errorf("a restored Commander table reads as %s", got.EffectiveFormat())
	}
}
