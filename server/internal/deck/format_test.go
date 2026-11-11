package deck

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// format_test.go: the deck check per format (ADR 0222 §2). The Standard
// rules are CR 100.2a ("each deck has a minimum deck size of 60 cards
// ... no more than four of any card with a particular English name
// other than basic land cards") and CR 100.4a ("a sideboard may contain
// no more than fifteen cards. The four-card limit ... applies to the
// combined deck and sideboard"), and the Standard card pool.

// standardCard is a card legal in Standard and in Commander.
func standardCard(name, typeLine string, colors ...string) cards.Card {
	return cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.New(),
		Name:          name,
		TypeLine:      typeLine,
		ColorIdentity: colors,
		Legalities:    map[string]string{"commander": "legal", "standard": "legal"},
	}
}

func copies(c cards.Card, n int) []cards.Card {
	out := make([]cards.Card, n)
	for i := range out {
		out[i] = c
	}
	return out
}

// standardDeck is a legal sixty: nine cards four times and 24 Mountains.
func standardDeck() *List {
	list := &List{}
	for i := 0; i < 9; i++ {
		list.Mainboard = append(list.Mainboard, copies(standardCard(fmt.Sprintf("Spell %d", i), "Instant", "R"), 4)...)
	}
	list.Mainboard = append(list.Mainboard, copies(standardCard("Mountain", "Basic Land — Mountain", "R"), 24)...)
	return list
}

func TestStandardHappyPath(t *testing.T) {
	list := standardDeck()
	if len(list.Mainboard) != 60 {
		t.Fatalf("fixture has %d cards", len(list.Mainboard))
	}
	if err := Validate(list, game.FormatStandard); err != nil {
		t.Fatalf("a legal Standard deck: %v", err)
	}
	// More than sixty is legal: the size is a minimum (CR 100.2a).
	list.Mainboard = append(list.Mainboard, copies(standardCard("Mountain", "Basic Land — Mountain", "R"), 15)...)
	if err := Validate(list, game.FormatStandard); err != nil {
		t.Fatalf("a 75-card Standard deck: %v", err)
	}
}

func TestStandardNeedsSixtyCards(t *testing.T) {
	list := standardDeck()
	list.Mainboard = list.Mainboard[:59]
	err := Validate(list, game.FormatStandard)
	assertHasViolation(t, err, CodeWrongCardCount)
	if msg := err.Error(); !strings.Contains(msg, "59") || !strings.Contains(msg, "at least 60") {
		t.Errorf("message %q should say 59 and at least 60", msg)
	}
}

func TestStandardFourCopies(t *testing.T) {
	// Five of one card in the deck.
	list := standardDeck()
	list.Mainboard[4] = list.Mainboard[0]
	assertHasViolation(t, Validate(list, game.FormatStandard), CodeTooManyCopies)

	// Four in the deck and one in the sideboard: the limit is across
	// both (CR 100.4a).
	list = standardDeck()
	list.Sideboard = []cards.Card{list.Mainboard[0]}
	err := Validate(list, game.FormatStandard)
	assertHasViolation(t, err, CodeTooManyCopies)
	if !strings.Contains(err.Error(), "across the deck and sideboard") {
		t.Errorf("message %q should say the limit is across deck and sideboard", err)
	}

	// Three and one is four: legal.
	list = standardDeck()
	list.Sideboard = []cards.Card{list.Mainboard[0]}
	list.Mainboard[0] = standardCard("Mountain", "Basic Land — Mountain", "R")
	if err := Validate(list, game.FormatStandard); err != nil {
		t.Errorf("three in the deck and one in the sideboard: %v", err)
	}

	// Any number of basic lands (CR 100.2a), and a card whose own text
	// lifts the limit.
	list = standardDeck()
	rats := standardCard("Relentless Rats", "Creature — Rat", "B")
	rats.OracleText = "A deck can have any number of cards named Relentless Rats."
	list.Mainboard = append(list.Mainboard, copies(rats, 12)...)
	if err := Validate(list, game.FormatStandard); err != nil {
		t.Errorf("24 basics and twelve Relentless Rats: %v", err)
	}
	// "Up to seven" is seven, not four and not unlimited.
	dwarves := standardCard("Seven Dwarves", "Creature — Dwarf", "R")
	dwarves.OracleText = "A deck can have up to seven cards named Seven Dwarves."
	list = standardDeck()
	list.Mainboard = append(list.Mainboard, copies(dwarves, 7)...)
	if err := Validate(list, game.FormatStandard); err != nil {
		t.Errorf("seven Seven Dwarves: %v", err)
	}
	list.Mainboard = append(list.Mainboard, dwarves)
	assertHasViolation(t, Validate(list, game.FormatStandard), CodeTooManyCopies)
}

func TestStandardCardPool(t *testing.T) {
	for _, status := range []string{"not_legal", "banned", "restricted", ""} {
		list := standardDeck()
		out := standardCard("Rotated Out", "Instant", "R")
		if status == "" {
			delete(out.Legalities, "standard")
		} else {
			out.Legalities["standard"] = status
		}
		list.Mainboard[0] = out
		list.Mainboard[1] = out // twice in the deck, one violation
		vs := violationsOf(t, Validate(list, game.FormatStandard))
		n := 0
		for _, v := range vs {
			if v.Code == CodeNotLegalInFormat && v.Card == "Rotated Out" {
				n++
				if !strings.Contains(v.Message, "standard") {
					t.Errorf("message %q should name the format", v.Message)
				}
			}
		}
		if n != 1 {
			t.Errorf("status %q: %d not-legal violations for the card, want 1 (%+v)", status, n, vs)
		}
	}

	// A sideboard card is held to the pool too.
	list := standardDeck()
	out := standardCard("Rotated Out", "Instant", "R")
	out.Legalities["standard"] = "not_legal"
	list.Sideboard = []cards.Card{out}
	assertHasViolation(t, Validate(list, game.FormatStandard), CodeNotLegalInFormat)

	// A card legal in Standard and not in Commander is fine here.
	list = standardDeck()
	list.Mainboard[0].Legalities = map[string]string{"standard": "legal", "commander": "banned"}
	if err := Validate(list, game.FormatStandard); err != nil {
		t.Errorf("a card banned only in Commander: %v", err)
	}
}

func TestStandardHasNoCommander(t *testing.T) {
	list := standardDeck()
	cmd := standardCard("Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", "R")
	list.Commanders = []cards.Card{cmd}
	err := Validate(list, game.FormatStandard)
	assertHasViolation(t, err, CodeCommanderNotAllowed)
	codes := violationCodes(err)
	for _, commanderOnly := range []string{CodeMissingCommander, CodeColorIdentity, CodeSingleton, CodeNotLegalCommander, CodeSideboardUnsupported} {
		if slices.Contains(codes, commanderOnly) {
			t.Errorf("a Standard check raised the Commander violation %q", commanderOnly)
		}
	}
}

func TestStandardSideboard(t *testing.T) {
	side := func(n int) []cards.Card {
		out := make([]cards.Card, 0, n)
		for i := range n {
			out = append(out, standardCard(fmt.Sprintf("Side %02d", i), "Sorcery", "R"))
		}
		return out
	}
	list := standardDeck()
	list.Sideboard = side(15)
	if err := Validate(list, game.FormatStandard); err != nil {
		t.Errorf("a fifteen-card sideboard: %v", err)
	}
	// It seats outside the game, as ADR 0191 has it.
	if got := len(list.SeatedSideboard()); got != 15 {
		t.Errorf("%d sideboard cards seated, want 15", got)
	}
	list.Sideboard = side(16)
	err := Validate(list, game.FormatStandard)
	assertHasViolation(t, err, CodeSideboardOverLimit)
	if slices.Contains(violationCodes(err), CodeSideboardTooLarge) {
		t.Error("a Standard sideboard over the limit is fatal, not Commander's warning")
	}
}

// A colour identity means nothing in Standard: any colours together.
func TestStandardIgnoresColourIdentity(t *testing.T) {
	list := standardDeck()
	list.Mainboard[0] = standardCard("Off Colour", "Instant", "U", "B")
	if err := Validate(list, game.FormatStandard); err != nil {
		t.Errorf("a two-colour splash: %v", err)
	}
}

// Commander's check is what it was: the empty format and "commander"
// are the same check, a Standard-shaped deck fails it the way it always
// did, and a legal Commander deck fails Standard's.
func TestCommanderCheckIsUnchanged(t *testing.T) {
	if err := Validate(buildValidDeck(), game.FormatCommander); err != nil {
		t.Fatalf("a legal Commander deck: %v", err)
	}
	if err := Validate(buildValidDeck(), ""); err != nil {
		t.Fatalf("the empty format is Commander: %v", err)
	}
	a := violationCodes(Validate(standardDeck(), game.FormatCommander))
	b := violationCodes(Validate(standardDeck(), ""))
	if !slices.Equal(a, b) {
		t.Errorf("empty format gave %v, commander gave %v", b, a)
	}
	for _, want := range []string{CodeMissingCommander, CodeWrongCardCount, CodeSingleton} {
		if !slices.Contains(a, want) {
			t.Errorf("a sixty-card four-of deck at a Commander table: missing %q in %v", want, a)
		}
	}
	for _, standardOnly := range []string{CodeTooManyCopies, CodeCommanderNotAllowed, CodeSideboardOverLimit} {
		if slices.Contains(a, standardOnly) {
			t.Errorf("the Commander check raised the constructed violation %q", standardOnly)
		}
	}
	// A legal Commander deck is not a Standard deck.
	std := violationCodes(Validate(buildValidDeck(), game.FormatStandard))
	if !slices.Contains(std, CodeCommanderNotAllowed) || !slices.Contains(std, CodeNotLegalInFormat) {
		t.Errorf("a Commander deck at a Standard table: %v", std)
	}
	if err := Validate(buildValidDeck(), "vintage"); err == nil {
		t.Error("an unknown format validated")
	}
	if err := Validate(nil, game.FormatStandard); err == nil {
		t.Error("a nil list validated")
	}
}

// A constructed paste often ends in an unheaded block: the sideboard.
// ParseText guesses a short one is a Commander export's commander;
// EntriesFor reads the guess as the sideboard for a format with none,
// and leaves a row the list itself marked alone.
func TestEntriesForReadsAGuessedCommanderAsTheSideboard(t *testing.T) {
	entries, err := ParseText("4 Shock\n20 Mountain\n\n2 Negate\n")
	if err != nil {
		t.Fatal(err)
	}
	last := entries[len(entries)-1]
	if !last.IsCommander || !last.CommanderInferred {
		t.Fatalf("ParseText should guess the last block is the commander: %+v", last)
	}
	// Commander: unchanged, the same slice.
	same := EntriesFor(game.FormatCommander, entries)
	if &same[0] != &entries[0] || !same[len(same)-1].IsCommander {
		t.Error("EntriesFor changed a Commander list")
	}
	std := EntriesFor(game.FormatStandard, entries)
	got := std[len(std)-1]
	if got.IsCommander || !got.IsSideboard || got.Name != "Negate" || got.Count != 2 {
		t.Errorf("Standard: last row = %+v, want a sideboard row", got)
	}
	if !entries[len(entries)-1].IsCommander {
		t.Error("EntriesFor changed its argument")
	}

	// A marked commander stays one, for the deck check to refuse.
	marked, err := ParseText("Commander:\n1 Krenko, Mob Boss\n\nDeck:\n4 Shock\n")
	if err != nil {
		t.Fatal(err)
	}
	std = EntriesFor(game.FormatStandard, marked)
	if !std[0].IsCommander {
		t.Errorf("a row under a Commander heading was moved: %+v", std[0])
	}
}
