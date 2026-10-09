package deck

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// partner_test.go — "Partner with [name]" as a deck-construction
// permission (#2142, CR 702.124j): "You may designate two legendary
// cards as your commander rather than one if each has a 'partner with
// [name]' ability with the other's name."

// frodoCard and samCard carry their Scryfall oracle text: Frodo's
// printing has no reminder text, Sam's has it.
func frodoCard() cards.Card {
	c := basicLegal("Frodo, Adventurous Hobbit", "Legendary Creature — Halfling Scout", "W", "B")
	c.OracleText = "Partner with Sam, Loyal Attendant\nVigilance\nWhenever Frodo attacks, if you gained 3 or more life this turn, the Ring tempts you. Then if Frodo is your Ring-bearer and the Ring has tempted you two or more times this game, draw a card."
	return c
}

func samCard() cards.Card {
	c := basicLegal("Sam, Loyal Attendant", "Legendary Creature — Halfling Peasant", "G", "W")
	c.OracleText = "Partner with Frodo, Adventurous Hobbit (When this creature enters, target player may put Frodo into their hand from their library, then shuffle.)\nAt the beginning of combat on your turn, create a Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")\nActivated abilities of Foods you control cost {1} less to activate."
	return c
}

func toothyCard() cards.Card {
	c := basicLegal("Toothy, Imaginary Friend", "Legendary Creature — Illusion", "U")
	c.OracleText = "Partner with Pir, Imaginative Rascal (When this creature enters, target player may put Pir into their hand from their library, then shuffle.)\nWhenever you draw a card, put a +1/+1 counter on Toothy."
	return c
}

// pairDeck is a 100-card deck on two commanders whose 98 others are
// in the pair's combined identity.
func pairDeck(a, b cards.Card) *List {
	list := &List{Commanders: []cards.Card{a, b}}
	for i := 0; i < 38; i++ {
		list.Mainboard = append(list.Mainboard, basicLegal(fmt.Sprintf("Spell %d", i), "Instant", "W"))
	}
	plains := basicLegal("Plains", "Basic Land — Plains", "W")
	for i := 0; i < 60; i++ {
		list.Mainboard = append(list.Mainboard, plains)
	}
	return list
}

func violationsOf(t *testing.T, err error) []Violation {
	t.Helper()
	if err == nil {
		return nil
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	return ve.Violations
}

func TestPartnerWithNamesReadsTheKeywordLine(t *testing.T) {
	cases := map[string][]string{
		frodoCard().OracleText: {"Sam, Loyal Attendant"},
		samCard().OracleText:   {"Frodo, Adventurous Hobbit"},
		"Flying\nPartner with Will Kenrith\nRowan Kenrith can be your commander.": {"Will Kenrith"},
		"Partner (You can have two commanders if both have partner.)":             nil,
		"Choose target creature with partner with somebody.":                      nil,
	}
	for text, want := range cases {
		got := partnerWithNamesIn(text)
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}
}

// The pair each naming the other is a legal pair of commanders.
func TestPartnerWithPairIsAllowed(t *testing.T) {
	idx := indexWith(frodoCard(), samCard())
	list, err := Resolve(idx, "t", []Entry{
		{Name: "Frodo, Adventurous Hobbit", Count: 1, IsCommander: true},
		{Name: "Sam, Loyal Attendant", Count: 1, IsCommander: true},
	})
	if err != nil {
		t.Fatalf("Resolve refused the pair: %v", err)
	}
	if len(list.Commanders) != 2 {
		t.Fatalf("commanders %d, want 2", len(list.Commanders))
	}
	if err := Validate(pairDeck(frodoCard(), samCard())); err != nil {
		t.Errorf("Validate refused the pair: %v", err)
	}
}

// A partner-with card alone is an ordinary commander: the ability is a
// permission, never a requirement.
func TestPartnerWithCardAloneIsACommander(t *testing.T) {
	idx := indexWith(frodoCard())
	if _, err := Resolve(idx, "t", []Entry{{Name: "Frodo, Adventurous Hobbit", Count: 1, IsCommander: true}}); err != nil {
		t.Fatalf("Resolve refused a lone Frodo: %v", err)
	}
	list := pairDeck(frodoCard(), samCard())
	list.Commanders = list.Commanders[:1]
	list.Mainboard = append(list.Mainboard, basicLegal("Plains", "Basic Land — Plains", "W"))
	if err := Validate(list); err != nil {
		t.Errorf("Validate refused a lone Frodo: %v", err)
	}
}

// Two partner-with cards that name other cards are not a pair.
func TestPartnerWithRefusedWhenItNamesAnotherCard(t *testing.T) {
	vs := violationsOf(t, Validate(pairDeck(frodoCard(), toothyCard())))
	var got *Violation
	for i := range vs {
		if vs[i].Code == CodeInvalidPartnerPair {
			got = &vs[i]
		}
	}
	if got == nil {
		t.Fatalf("no %s in %+v", CodeInvalidPartnerPair, vs)
	}
	if !strings.Contains(got.Message, "Sam, Loyal Attendant") {
		t.Errorf("message %q does not say whom Frodo partners with", got.Message)
	}
}

// One card naming the other is not enough: each has to name the other.
func TestPartnerWithRefusedWhenOnlyOneNamesTheOther(t *testing.T) {
	// A card named Sam that partners with somebody else.
	samElsewhere := samCard()
	samElsewhere.OracleText = "Partner with Merry, Warden of Isengard"
	// A card named Sam with no partner ability at all.
	plainSam := samCard()
	plainSam.OracleText = "Vigilance"

	for _, sam := range []cards.Card{samElsewhere, plainSam} {
		vs := violationsOf(t, Validate(pairDeck(frodoCard(), sam)))
		found := false
		for _, v := range vs {
			if v.Code == CodeInvalidPartnerPair && v.Card == "Sam, Loyal Attendant" {
				found = true
			}
		}
		if !found {
			t.Errorf("Sam text %q: want %s naming Sam, got %+v", sam.OracleText, CodeInvalidPartnerPair, vs)
		}
	}
}

// Two commanders with no partner ability between them stay refused as
// too many.
func TestTwoCommandersWithoutPartnerWithAreTooMany(t *testing.T) {
	a := basicLegal("Commander A", "Legendary Creature — Human", "W")
	b := basicLegal("Commander B", "Legendary Creature — Human", "W")
	assertHasViolation(t, Validate(pairDeck(a, b)), CodeTooManyCommanders)
}

// No partner ability allows a third commander (CR 702.124g).
func TestThreeCommandersAreRefused(t *testing.T) {
	list := pairDeck(frodoCard(), samCard())
	list.Commanders = append(list.Commanders, toothyCard())
	list.Mainboard = list.Mainboard[1:]
	assertHasViolation(t, Validate(list), CodeTooManyCommanders)
}

// The pair's colour identities combine (CR 702.124c): Frodo is W/B and
// Sam G/W, so a green card is in, a blue one is out.
func TestPartnerPairIdentityIsTheUnion(t *testing.T) {
	list := pairDeck(frodoCard(), samCard())
	list.Mainboard[0] = basicLegal("Green Spell", "Instant", "G")
	list.Mainboard[1] = basicLegal("Black Spell", "Instant", "B")
	if err := Validate(list); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	list.Mainboard[2] = basicLegal("Blue Spell", "Instant", "U")
	vs := violationsOf(t, Validate(list))
	if len(vs) != 1 || vs[0].Code != CodeColorIdentity || vs[0].Card != "Blue Spell" {
		t.Errorf("violations %+v, want one colour identity violation for Blue Spell", vs)
	}
}

// Each commander of a pair has to be a legal commander on its own.
func TestPartnerPairChecksBothCommanders(t *testing.T) {
	sam := samCard()
	sam.TypeLine = "Creature — Halfling Peasant"
	vs := violationsOf(t, Validate(pairDeck(frodoCard(), sam)))
	found := false
	for _, v := range vs {
		if v.Code == CodeNotLegalCommander && v.Card == sam.Name {
			found = true
		}
	}
	if !found {
		t.Errorf("a nonlegendary second commander passed: %+v", vs)
	}
}

// The other partner abilities are unchanged by #2142: plain partner and
// partner—[text] are still refused at Resolve, and a Background pairing
// is two commanders that are not a partner-with pair.
func TestOtherPartnerAbilitiesAreStillRefused(t *testing.T) {
	friends := basicLegal("Friend", "Legendary Creature — Human", "W")
	friends.OracleText = "Partner—Friends forever (You can have two commanders if both have friends forever.)"
	idx := indexWith(friends)
	if _, err := Resolve(idx, "t", []Entry{{Name: "Friend", Count: 1, IsCommander: true}}); !errors.Is(err, ErrUnsupportedMechanic) {
		t.Errorf("friends forever: got %v, want ErrUnsupportedMechanic", err)
	}

	chooser := basicLegal("Chooser", "Legendary Creature — Human", "W")
	chooser.OracleText = "Choose a Background (You can have a Background as a second commander.)"
	bg := basicLegal("Some Background", "Legendary Enchantment — Background", "W")
	assertHasViolation(t, Validate(pairDeck(chooser, bg)), CodeTooManyCommanders)
}
