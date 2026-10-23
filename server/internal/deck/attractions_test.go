package deck

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// attractions_test.go — ADR 0207 §2: a deck declares its Attraction
// deck, which seats face down in the command zone and is not one of the
// 100.

// attractionCards are n distinct legal Attraction cards, each lighting
// one number.
func attractionCards(n int) []cards.Card {
	out := make([]cards.Card, 0, n)
	for i := range n {
		c := basicLegal(fmt.Sprintf("Attraction %02d", i), "Artifact — Attraction")
		c.AttractionLights = []int{i%6 + 1}
		out = append(out, c)
	}
	return out
}

func TestPrintsAttraction(t *testing.T) {
	for typeLine, want := range map[string]bool{
		"Artifact — Attraction":       true,
		"Artifact — Attraction Equip": true,
		"Artifact — Contraption":      false,
		"Artifact Creature — Robot":   false,
		"Sorcery":                     false,
	} {
		if got := PrintsAttraction(basicLegal("X", typeLine)); got != want {
			t.Errorf("%q: %v, want %v", typeLine, got, want)
		}
	}
}

// Every importer's Attraction cards reach the Attraction deck, wherever
// the source listed them: under an "Attractions" heading, Moxfield's
// attractions board, or no heading at all (CR 717.2).
func TestEveryImporterPutsAttractionsInTheAttractionDeck(t *testing.T) {
	cmd := basicLegal("Test Commander", "Legendary Creature — Test", "R")
	attractions := attractionCards(10)
	idx := indexWith(append([]cards.Card{cmd, basicLegal("Lightning Bolt", "Instant", "R")}, attractions...)...)
	var rows strings.Builder
	for _, c := range attractions {
		fmt.Fprintf(&rows, "1 %s\n", c.Name)
	}
	for _, src := range []string{
		"Commander:\n1 Test Commander\n\nDeck:\n1 Lightning Bolt\n\nAttractions:\n" + rows.String(),
		"Commander:\n1 Test Commander\n\nDeck:\n1 Lightning Bolt\n\nAttraction Deck\n" + rows.String(),
		"Commander:\n1 Test Commander\n\nDeck:\n1 Lightning Bolt\n" + rows.String(),
	} {
		entries, err := ParseText(src)
		if err != nil {
			t.Fatal(err)
		}
		list, err := Resolve(idx, "t", entries)
		if err != nil {
			t.Fatal(err)
		}
		if len(list.Attractions) != 10 || len(list.Mainboard) != 1 {
			t.Errorf("%q: %d Attractions and %d main deck cards, want 10 and 1", src[:40], len(list.Attractions), len(list.Mainboard))
		}
	}
	t.Run("moxfield", func(t *testing.T) {
		var b strings.Builder
		for i, c := range attractions {
			if i > 0 {
				b.WriteString(",")
			}
			fmt.Fprintf(&b, `"a%d":{"quantity":1,"card":{"name":%q}}`, i, c.Name)
		}
		raw := `{"name":"t","boards":{"commanders":{"cards":{"c":{"quantity":1,"card":{"name":"Test Commander"}}}},` +
			`"mainboard":{"cards":{"m":{"quantity":1,"card":{"name":"Lightning Bolt"}}}},` +
			`"attractions":{"cards":{` + b.String() + `}}}}`
		_, entries, err := ParseMoxfield([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		list, err := Resolve(idx, "t", entries)
		if err != nil {
			t.Fatal(err)
		}
		if len(list.Attractions) != 10 {
			t.Errorf("moxfield attractions board: %d Attractions, want 10", len(list.Attractions))
		}
	})
}

// The Attraction deck rides List.Text, so a saved deck reseats it.
func TestAttractionsRoundTripThroughText(t *testing.T) {
	list := legalDeck()
	list.Attractions = attractionCards(10)
	text := list.Text()
	if !strings.Contains(text, "Attractions:\n") {
		t.Fatalf("no Attractions section in:\n%s", text)
	}
	entries, err := ParseText(text)
	if err != nil {
		t.Fatal(err)
	}
	all := append(append(append([]cards.Card{}, list.Commanders...), list.Mainboard...), list.Attractions...)
	back, err := Resolve(indexWith(all...), "t", entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Attractions) != 10 || len(back.Mainboard) != 99 {
		t.Errorf("round trip: %d Attractions, %d main deck cards", len(back.Attractions), len(back.Mainboard))
	}
}

// A legal Attraction deck is not one of the 100, and seats with its
// lights and the Astrotorium back.
func TestALegalAttractionDeckSeatsInTheCommandZone(t *testing.T) {
	list := legalDeck()
	list.Attractions = attractionCards(10)
	if err := Validate(list); err != nil {
		t.Fatalf("a legal deck with ten Attractions: %v", err)
	}
	out := list.ToGameCards()
	if len(out) != 110 {
		t.Fatalf("%d game cards, want 100 + 10", len(out))
	}
	for _, c := range out[100:] {
		if !c.AstrotoriumBack || len(c.AttractionLights) != 1 {
			t.Errorf("%s: back %v lights %v", c.Name, c.AstrotoriumBack, c.AttractionLights)
		}
	}
	for _, c := range out[:100] {
		if c.AstrotoriumBack {
			t.Errorf("%s has the Astrotorium back", c.Name)
		}
	}
	p := seatAndCount(t, list)
	if len(p.AttractionDeckCards()) != 10 || p.Library.Size()+p.Command.Size() != 100 {
		t.Errorf("seated %d Attractions, library+command %d", len(p.AttractionDeckCards()), p.Library.Size()+p.Command.Size())
	}
}

// CR 717.2a: fewer than ten, or two with one name, is said (non-fatal)
// and the Attraction deck is left out of the game; the deck still seats.
func TestAnIllegalAttractionDeckIsSaidAndNotSeated(t *testing.T) {
	dup := attractionCards(10)
	dup[9] = dup[0]
	for name, deck := range map[string][]cards.Card{"nine": attractionCards(9), "a duplicate name": dup} {
		t.Run(name, func(t *testing.T) {
			list := legalDeck()
			list.Attractions = deck
			codes := violationCodes(Validate(list))
			if !slices.Equal(codes, []string{CodeAttractionDeckIllegal}) {
				t.Fatalf("codes %v, want only %s", codes, CodeAttractionDeckIllegal)
			}
			if n := len(list.SeatedAttractions()); n != 0 {
				t.Errorf("an illegal Attraction deck seats %d cards", n)
			}
			if p := seatAndCount(t, list); p.AttractionDeckCards() != nil {
				t.Error("an illegal Attraction deck reached the game")
			}
		})
	}
}

// An Attraction not legal in Commander (the acorn ones) is refused like
// any deck card.
func TestANotLegalAttractionIsRefused(t *testing.T) {
	list := legalDeck()
	list.Attractions = attractionCards(10)
	list.Attractions[3].Legalities = map[string]string{"commander": "not_legal"}
	if codes := violationCodes(Validate(list)); !slices.Contains(codes, CodeNotLegalInFormat) {
		t.Errorf("codes %v, want %s", codes, CodeNotLegalInFormat)
	}
}
