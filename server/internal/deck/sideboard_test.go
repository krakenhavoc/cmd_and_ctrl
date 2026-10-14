package deck

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sideboard_test.go — ADR 0191: the sideboard every importer already
// parsed now reaches the game, as the player's outside-the-game pool.

// legalDeck is companionDeck without the companion: a legal 100-card
// five-colour list.
func legalDeck() *List {
	l := companionDeck()
	l.Companion = nil
	return l
}

// sideCards are n distinct legal sideboard cards.
func sideCards(n int) []cards.Card {
	out := make([]cards.Card, 0, n)
	for i := range n {
		out = append(out, costed(fmt.Sprintf("Side %02d", i), "Sorcery", "{1}{R}", "R"))
	}
	return out
}

func sideNames(list *List) []string {
	var out []string
	for _, c := range list.Sideboard {
		out = append(out, c.Name)
	}
	return out
}

// seatAndCount seats the list's cards for a player and returns the game
// seat.
func seatAndCount(t *testing.T, list *List) *game.Player {
	t.Helper()
	g := game.NewGame()
	p, err := g.AddPlayer("P1", list.ToGameCards())
	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	return p
}

func TestToGameCardsStampsTheSideboardAndSeatsItOutsideTheGame(t *testing.T) {
	list := legalDeck()
	list.Sideboard = sideCards(3)
	cardsOut := list.ToGameCards()
	if len(cardsOut) != 103 {
		t.Fatalf("%d game cards, want 100 + 3", len(cardsOut))
	}
	n := 0
	for _, c := range cardsOut {
		if c.IsSideboard {
			n++
			if c.IsCommander || c.IsCompanion {
				t.Errorf("%s is a sideboard card and something else", c.Name)
			}
		}
	}
	if n != 3 {
		t.Errorf("%d sideboard stamps, want 3", n)
	}
	p := seatAndCount(t, list)
	if got := len(p.SideboardCards()); got != 3 {
		t.Errorf("seated sideboard = %d, want 3", got)
	}
	if got := p.Library.Size() + p.Command.Size(); got != 100 {
		t.Errorf("library + command = %d, want the deck's 100", got)
	}
	if got := p.SideboardCards()[0].Name; got != "Side 00" {
		t.Errorf("first sideboard card %q, want list order", got)
	}
}

// A list with a companion AND a sideboard seats both, each in its own
// container.
func TestCompanionAndSideboardSeatTogether(t *testing.T) {
	list := companionDeck()
	list.Sideboard = sideCards(2)
	p := seatAndCount(t, list)
	if _, ok := p.CompanionCard(); !ok {
		t.Error("the companion was not seated")
	}
	if got := len(p.SideboardCards()); got != 2 {
		t.Errorf("sideboard = %d, want 2", got)
	}
}

// Every importer parses a sideboard, resolves it and keeps it into the
// game.
func TestEveryImporterKeepsTheSideboardIntoTheGame(t *testing.T) {
	names := []string{"Burning Wish", "Lightning Bolt", "Extrapolate the Impossible"}
	cmdName := "Test Commander"
	mk := func(entries []Entry) *List {
		t.Helper()
		var cs []cards.Card
		cs = append(cs, basicLegal(cmdName, "Legendary Creature — Test", "R"))
		for _, n := range names {
			cs = append(cs, basicLegal(n, "Sorcery", "R"))
		}
		list, err := Resolve(indexWith(cs...), "t", entries)
		if err != nil {
			t.Fatal(err)
		}
		return list
	}

	t.Run("text", func(t *testing.T) {
		for _, header := range []string{"Sideboard:", "Sideboard", "// Sideboard"} {
			src := "Commander:\n1 " + cmdName + "\n\nDeck:\n" + "\n" + header + "\n1 Burning Wish\n2 Lightning Bolt\n1 Extrapolate the Impossible\n"
			entries, err := ParseText(src)
			if err != nil {
				t.Fatal(err)
			}
			list := mk(entries)
			if got := sideNames(list); len(got) != 4 || got[0] != "Burning Wish" || got[1] != "Lightning Bolt" || got[2] != "Lightning Bolt" {
				t.Errorf("header %q: sideboard %v", header, got)
			}
			if p := seatAndCount(t, list); len(p.SideboardCards()) != 4 {
				t.Errorf("header %q: seated %d sideboard cards, want 4", header, len(p.SideboardCards()))
			}
		}
		// The SB: prefix.
		entries, err := ParseText("1 " + cmdName + " *CMDR*\nSB: 1 Burning Wish\nSB: 1 Lightning Bolt\n")
		if err != nil {
			t.Fatal(err)
		}
		if got := sideNames(mk(entries)); !slices.Equal(got, []string{"Burning Wish", "Lightning Bolt"}) {
			t.Errorf("SB: prefix: sideboard %v", got)
		}
	})

	t.Run("moxfield json", func(t *testing.T) {
		raw := `{"name":"t","boards":{
			"commanders":{"cards":{"a":{"quantity":1,"card":{"name":"` + cmdName + `"}}}},
			"mainboard":{"cards":{}},
			"sideboard":{"cards":{"b":{"quantity":1,"card":{"name":"Burning Wish"}},"c":{"quantity":2,"card":{"name":"Lightning Bolt"}}}}}}`
		_, entries, err := ParseMoxfield([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		list := mk(entries)
		if len(list.Sideboard) != 3 {
			t.Fatalf("sideboard %v, want 3 cards", sideNames(list))
		}
		if p := seatAndCount(t, list); len(p.SideboardCards()) != 3 {
			t.Errorf("seated %d sideboard cards, want 3", len(p.SideboardCards()))
		}
	})

	t.Run("moxfield text export", func(t *testing.T) {
		entries, err := ParseText(string(readFixture(t, "moxfield_3401_export_shared_sideboard.txt")))
		if err != nil {
			t.Fatal(err)
		}
		list, err := Resolve(fixtureIndex(entries), "", entries)
		if err != nil {
			t.Fatal(err)
		}
		if len(list.Sideboard) != 15 {
			t.Fatalf("sideboard %d, want the export's 15", len(list.Sideboard))
		}
		p := seatAndCount(t, list)
		if got := len(p.SideboardCards()); got != 15 {
			t.Errorf("seated %d sideboard cards, want 15", got)
		}
		if got := p.Library.Size() + p.Command.Size(); got != 100 {
			t.Errorf("deck is %d cards, want 100", got)
		}
	})

	t.Run("archidekt", func(t *testing.T) {
		raw := `{"name":"t","cards":[
			{"quantity":1,"categories":["Commander"],"card":{"oracleCard":{"name":"` + cmdName + `"}}},
			{"quantity":1,"categories":["Sideboard"],"card":{"oracleCard":{"name":"Burning Wish"}}},
			{"quantity":1,"categories":["Sideboard","Burn"],"card":{"oracleCard":{"name":"Lightning Bolt"}}},
			{"quantity":1,"categories":["Maybeboard"],"card":{"oracleCard":{"name":"Extrapolate the Impossible"}}}]}`
		_, entries, err := ParseArchidekt([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		list := mk(entries)
		if got := sideNames(list); !slices.Equal(got, []string{"Burning Wish", "Lightning Bolt"}) {
			t.Errorf("sideboard %v (the maybeboard is not one)", got)
		}
	})
}

// A saved deck keeps its source text. List.Text() is how a deck fetched
// from a link is saved, so the sideboard must survive Text -> ParseText
// -> Resolve.
func TestTheSideboardSurvivesASaveAndReload(t *testing.T) {
	list := legalDeck()
	list.Sideboard = []cards.Card{
		costed("Burning Wish", "Sorcery", "{1}{R}", "R"),
		costed("Lightning Bolt", "Instant", "{R}", "R"),
	}
	list.Sideboard = append(list.Sideboard, list.Sideboard[1])
	text := list.Text()
	if !strings.Contains(text, "Sideboard:") {
		t.Fatalf("the saved text has no sideboard section:\n%s", text)
	}
	entries, err := ParseText(text)
	if err != nil {
		t.Fatal(err)
	}
	var all []cards.Card
	all = append(all, list.Commanders...)
	all = append(all, list.Mainboard...)
	all = append(all, list.Sideboard...)
	reloaded, err := Resolve(indexWith(all...), "t", entries)
	if err != nil {
		t.Fatal(err)
	}
	if got := sideNames(reloaded); !slices.Equal(got, []string{"Burning Wish", "Lightning Bolt", "Lightning Bolt"}) {
		t.Errorf("reloaded sideboard %v", got)
	}
	if len(reloaded.Mainboard) != 99 || len(reloaded.Commanders) != 1 {
		t.Errorf("reloaded %d mainboard, %d commanders", len(reloaded.Mainboard), len(reloaded.Commanders))
	}
}

func violationFor(err error, code string) (Violation, bool) {
	ve, ok := err.(*ValidationError)
	if !ok {
		return Violation{}, false
	}
	for _, v := range ve.Violations {
		if v.Code == code {
			return v, true
		}
	}
	return Violation{}, false
}

// Owner decision 2026-10-14: Commander has no sideboard (CR 903.5e), and
// a sideboard is the outside-the-game pool anyway. The warning names
// what the section is for, and is not fatal.
func TestASideboardWarnsWithWhatItIsFor(t *testing.T) {
	list := legalDeck()
	list.Sideboard = sideCards(3)
	v, ok := violationFor(Validate(list), CodeSideboardUnsupported)
	if !ok {
		t.Fatal("no sideboard note")
	}
	if !strings.Contains(v.Message, "3 sideboard cards") || !strings.Contains(v.Message, "outside the game") ||
		strings.Contains(v.Message, "ignored") {
		t.Errorf("message %q", v.Message)
	}
	if _, fatal := violationFor(Validate(list), CodeSideboardTooLarge); fatal {
		t.Error("a small sideboard is too large")
	}
	// No sideboard, no note.
	if err := Validate(legalDeck()); err != nil {
		t.Errorf("a plain deck: %v", err)
	}
}

// An oversized sideboard is a warning and is not carried into the game:
// the deck still validates and seats, so a saved deck with a large
// sideboard or maybeboard section is never refused, and the server never
// picks which cards to keep.
func TestASideboardOverTheCapWarnsAndIsNotSeated(t *testing.T) {
	list := legalDeck()
	list.Sideboard = sideCards(MaxSideboard)
	if _, ok := violationFor(Validate(list), CodeSideboardTooLarge); ok {
		t.Errorf("a sideboard of exactly %d is too large", MaxSideboard)
	}
	if got := len(list.SeatedSideboard()); got != MaxSideboard {
		t.Errorf("a full sideboard seats %d cards, want %d", got, MaxSideboard)
	}

	list.Sideboard = sideCards(MaxSideboard + 1)
	err := Validate(list)
	v, ok := violationFor(err, CodeSideboardTooLarge)
	if !ok {
		t.Fatalf("a sideboard of %d is not flagged", MaxSideboard+1)
	}
	if !strings.Contains(v.Message, fmt.Sprint(MaxSideboard+1)) || !strings.Contains(v.Message, fmt.Sprintf("limit of %d", MaxSideboard)) ||
		!strings.Contains(v.Message, "ignored") {
		t.Errorf("message %q: want the count, the limit and that the sideboard is ignored", v.Message)
	}
	if _, also := violationFor(err, CodeSideboardUnsupported); also {
		t.Error("an oversized sideboard also got the soft note")
	}
	// Nothing about the sideboard may be fatal: the lobby treats this
	// code as a warning, and no other violation is raised.
	ve, _ := err.(*ValidationError)
	if ve == nil || len(ve.Violations) != 1 {
		t.Errorf("violations %+v, want only the warning", err)
	}

	// Nothing is carried into the game: the deck seats with an empty
	// sideboard rather than a server-chosen 15.
	if n := len(list.SeatedSideboard()); n != 0 {
		t.Errorf("an oversized sideboard seats %d cards", n)
	}
	cardsOut := list.ToGameCards()
	if len(cardsOut) != 100 {
		t.Errorf("%d game cards, want the deck's 100", len(cardsOut))
	}
	p := seatAndCount(t, list)
	if len(p.SideboardCards()) != 0 || p.Library.Size()+p.Command.Size() != 100 {
		t.Errorf("seated %d sideboard cards, library+command %d", len(p.SideboardCards()), p.Library.Size()+p.Command.Size())
	}
	// The cap is the engine's, and the engine still guards it.
	if MaxSideboard != game.MaxSideboardSize {
		t.Errorf("deck cap %d, engine cap %d", MaxSideboard, game.MaxSideboardSize)
	}
}

// ADR 0172 owner decision 3, as ADR 0191 settles the rest: a lone
// companion card in the sideboard is still the companion; any other
// sideboard stays a sideboard, companion-keyword cards included.
func TestALoneCompanionStillLeavesTheSideboardAndTheRestStays(t *testing.T) {
	j := jegantha()
	cmd := costed("Five Colour Commander", "Legendary Creature — Dragon", "{2}{W}{U}{B}{R}{G}", "W", "U", "B", "R", "G")
	other := basicLegal("Burning Wish", "Sorcery", "R")
	idx := indexWith(j, cmd, other)

	lone, err := Resolve(idx, "t", []Entry{{Name: cmd.Name, Count: 1, IsCommander: true}, {Name: j.Name, Count: 1, IsSideboard: true}})
	if err != nil {
		t.Fatal(err)
	}
	if lone.Companion == nil || len(lone.Sideboard) != 0 {
		t.Errorf("lone companion: companion %v, sideboard %d", lone.Companion, len(lone.Sideboard))
	}
	shared, err := Resolve(idx, "t", []Entry{
		{Name: cmd.Name, Count: 1, IsCommander: true},
		{Name: j.Name, Count: 1, IsSideboard: true},
		{Name: other.Name, Count: 1, IsSideboard: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if shared.Companion != nil || !slices.Equal(sideNames(shared), []string{j.Name, other.Name}) {
		t.Errorf("shared sideboard: companion %v, sideboard %v; want none and both cards", shared.Companion, sideNames(shared))
	}
	// Both cards reach the game, outside it.
	g := game.NewGame()
	p, err := g.AddPlayer("P1", shared.ToGameCards())
	if err != nil {
		t.Fatal(err)
	}
	if got := len(p.SideboardCards()); got != 2 {
		t.Errorf("seated %d sideboard cards, want 2", got)
	}
	if _, ok := p.CompanionCard(); ok {
		t.Error("a companion-keyword card in a shared sideboard became the companion")
	}
}
