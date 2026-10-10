package deck

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// partner_test.go — the partner abilities as deck-construction
// permissions (CR 702.124): "Partner with [name]" (#2142, CR
// 702.124j), and partner, partner—[text], choose a Background and
// Doctor's companion (#2874, CR 702.124h, i, k, m).

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

// karlachCard, agentCard and the rest carry their Scryfall oracle
// text and type lines (#2874, CR 702.124k): Karlach chooses a
// Background, Agent of the Iron Throne is one.
func karlachCard() cards.Card {
	c := basicLegal("Karlach, Fury of Avernus", "Legendary Creature — Tiefling Barbarian", "R")
	c.OracleText = "Whenever you attack, if it's the first combat phase of the turn, untap all attacking creatures. They gain first strike until end of turn. After this phase, there is an additional combat phase.\nChoose a Background (You can have a Background as a second commander.)"
	return c
}

func agentCard() cards.Card {
	c := basicLegal("Agent of the Iron Throne", "Legendary Enchantment — Background", "B")
	c.OracleText = "Commander creatures you own have \"Whenever an artifact or creature you control is put into a graveyard from the battlefield, each opponent loses 1 life.\""
	return c
}

func facelessOneCard() cards.Card {
	c := basicLegal("Faceless One", "Legendary Enchantment Creature — Background", "W", "U", "B", "R", "G")
	c.OracleText = "If Faceless One is your commander, choose a color before the game begins. Faceless One is the chosen color.\nChoose a Background (You can have a Background as a second commander.)"
	return c
}

func plainPartnerCard(name, color string) cards.Card {
	c := basicLegal(name, "Legendary Creature — Human", color)
	c.OracleText = "Flying\nPartner (You can have two commanders if both have partner.)"
	return c
}

func friendsForeverCard(name, color string) cards.Card {
	c := basicLegal(name, "Legendary Creature — Human", color)
	c.OracleText = "Vigilance\nPartner—Friends forever (You can have two commanders if both have this ability.)"
	return c
}

func companionCard() cards.Card {
	c := basicLegal("Nardole, Resourceful Cyborg", "Legendary Artifact Creature — Scientist", "U")
	c.OracleText = "Undying\n{T}: Add {U} for each counter on Nardole. Spend this mana only to cast noncreature spells.\nDoctor's companion (You can have two commanders if the other is the Doctor.)"
	return c
}

func doctorCard() cards.Card {
	return basicLegal("The Second Doctor", "Legendary Creature — Time Lord Doctor", "W")
}

func TestPartnerAbilitiesReadEveryKeywordLine(t *testing.T) {
	cases := map[string][]partnerAbility{
		karlachCard().OracleText:                {{kind: chooseABackground}},
		companionCard().OracleText:              {{kind: doctorsCompanion}},
		"Doctor’s companion":                    {{kind: doctorsCompanion}},
		plainPartnerCard("x", "W").OracleText:   {{kind: partnerPlain}},
		friendsForeverCard("x", "W").OracleText: {{kind: partnerText, arg: "Friends forever"}},
		"Partner—Survivors":                     {{kind: partnerText, arg: "Survivors"}},
		samCard().OracleText:                    {{kind: partnerWith, arg: "Frodo, Adventurous Hobbit"}},
		"Choose target creature with partner.\nThe partner keyword is cool.": nil,
		agentCard().OracleText: nil,
	}
	for text, want := range cases {
		got := partnerAbilitiesIn(text)
		if !slices.Equal(got, want) {
			t.Errorf("%q: got %+v, want %+v", text, got, want)
		}
	}
}

// Karlach and a Background are a legal pair, Resolve puts the
// Background second whichever order the list gives them, and the
// deck's colour identity is the pair's union (CR 702.124c).
func TestChooseABackgroundPairIsAllowed(t *testing.T) {
	idx := indexWith(karlachCard(), agentCard())
	list, err := Resolve(idx, "t", []Entry{
		{Name: "Agent of the Iron Throne", Count: 1, IsCommander: true},
		{Name: "Karlach, Fury of Avernus", Count: 1, IsCommander: true},
	})
	if err != nil {
		t.Fatalf("Resolve refused the pair: %v", err)
	}
	if got := commanderNames(list); !slices.Equal(got, []string{"Karlach, Fury of Avernus", "Agent of the Iron Throne"}) {
		t.Fatalf("commanders %q, want Karlach then the Background", got)
	}
	d := pairDeck(karlachCard(), agentCard())
	for i := range d.Mainboard {
		switch {
		case i == 0:
			d.Mainboard[i] = basicLegal("Black Spell", "Instant", "B")
		case i == 1:
			d.Mainboard[i] = basicLegal("Rakdos Spell", "Instant", "B", "R")
		case i < 38:
			d.Mainboard[i] = basicLegal(fmt.Sprintf("Red Spell %d", i), "Instant", "R")
		default:
			d.Mainboard[i] = basicLegal("Mountain", "Basic Land — Mountain", "R")
		}
	}
	if err := Validate(d); err != nil {
		t.Fatalf("Validate refused Karlach and a Background: %v", err)
	}
	d.Mainboard[2] = basicLegal("White Spell", "Instant", "W")
	vs := violationsOf(t, Validate(d))
	if len(vs) != 1 || vs[0].Code != CodeColorIdentity || vs[0].Card != "White Spell" {
		t.Errorf("violations %+v, want one colour identity violation for White Spell", vs)
	}
}

func commanderNames(l *List) []string {
	out := make([]string, 0, len(l.Commanders))
	for _, c := range l.Commanders {
		out = append(out, c.Name)
	}
	return out
}

// A Background is never a commander on its own (CR 702.124k), unless
// it has Choose a Background itself (Faceless One), and Faceless One
// may choose another Background.
func TestBackgroundAloneIsNotACommander(t *testing.T) {
	list := pairDeck(karlachCard(), agentCard())
	list.Commanders = []cards.Card{agentCard()}
	list.Mainboard = append(list.Mainboard, basicLegal("Plains", "Basic Land — Plains", "W"))
	vs := violationsOf(t, Validate(list))
	found := false
	for _, v := range vs {
		if v.Code == CodeNotLegalCommander && v.Card == "Agent of the Iron Throne" {
			found = true
		}
	}
	if !found {
		t.Errorf("a lone Background passed: %+v", vs)
	}

	list.Commanders = []cards.Card{facelessOneCard()}
	if err := Validate(list); err != nil {
		t.Errorf("Faceless One alone: %v", err)
	}
	if err := Validate(pairDeck(facelessOneCard(), agentCard())); err != nil {
		t.Errorf("Faceless One and a Background: %v", err)
	}
	// A Choose a Background commander alone is an ordinary commander.
	list.Commanders = []cards.Card{karlachCard()}
	list.Mainboard = pairDeck(karlachCard(), agentCard()).Mainboard
	list.Mainboard = append(list.Mainboard, basicLegal("Plains", "Basic Land — Plains", "W"))
	for i := range list.Mainboard[:38] {
		list.Mainboard[i] = basicLegal(fmt.Sprintf("Red Spell %d", i), "Instant", "R")
	}
	for i := 38; i < len(list.Mainboard); i++ {
		list.Mainboard[i] = basicLegal("Mountain", "Basic Land — Mountain", "R")
	}
	if err := Validate(list); err != nil {
		t.Errorf("Karlach alone: %v", err)
	}
}

// Choose a Background pairs only with a Background, and a Background
// only with a Choose a Background commander. The violation names the
// card that does not meet the other's requirement.
func TestChooseABackgroundRefusesOtherPairs(t *testing.T) {
	notBackground := basicLegal("Some Legend", "Legendary Creature — Human", "R")
	nonLegendaryBG := basicLegal("Odd Background", "Enchantment — Background", "R")
	twin := agentCard()
	twin.Name = "Another Background"
	cases := []struct {
		a, b    cards.Card
		culprit string
	}{
		{karlachCard(), notBackground, "Some Legend"},
		{agentCard(), notBackground, "Some Legend"},
		{agentCard(), twin, "Another Background"},
		{karlachCard(), nonLegendaryBG, "Odd Background"},
		{karlachCard(), plainPartnerCard("Partner Guy", "R"), "Partner Guy"},
		{agentCard(), plainPartnerCard("Partner Guy", "R"), "Partner Guy"},
	}
	for _, tc := range cases {
		vs := violationsOf(t, Validate(pairDeck(tc.a, tc.b)))
		found := false
		for _, v := range vs {
			if v.Code == CodeInvalidPartnerPair && v.Card == tc.culprit {
				found = true
			}
		}
		if !found {
			t.Errorf("%s + %s: want %s naming %s, got %+v", tc.a.Name, tc.b.Name, CodeInvalidPartnerPair, tc.culprit, vs)
		}
	}
}

// Plain partner: both have it (CR 702.124h). Partner—[text]: both have
// the same one (CR 702.124i). Different partner abilities never
// combine (CR 702.124f).
func TestPartnerAndPartnerTextPairs(t *testing.T) {
	if err := Validate(pairDeck(plainPartnerCard("Thrasios", "W"), plainPartnerCard("Tymna", "W"))); err != nil {
		t.Errorf("two partners: %v", err)
	}
	if err := Validate(pairDeck(friendsForeverCard("Will", "W"), friendsForeverCard("Mike", "W"))); err != nil {
		t.Errorf("two friends forever: %v", err)
	}
	survivor := basicLegal("Abby", "Legendary Creature — Human Survivor", "W")
	survivor.OracleText = "Partner—Survivors"
	refused := [][2]cards.Card{
		{friendsForeverCard("Will", "W"), survivor},
		{plainPartnerCard("Thrasios", "W"), friendsForeverCard("Will", "W")},
		{plainPartnerCard("Thrasios", "W"), frodoCard()},
		{plainPartnerCard("Thrasios", "W"), basicLegal("No Partner", "Legendary Creature — Human", "W")},
	}
	for _, pair := range refused {
		assertHasViolation(t, Validate(pairDeck(pair[0], pair[1])), CodeInvalidPartnerPair)
	}
	// Resolve no longer refuses a partner commander.
	idx := indexWith(plainPartnerCard("Thrasios", "W"), friendsForeverCard("Will", "W"))
	if _, err := Resolve(idx, "t", []Entry{{Name: "Thrasios", Count: 1, IsCommander: true}, {Name: "Will", Count: 1, IsCommander: true}}); err != nil {
		t.Errorf("Resolve: %v", err)
	}
}

// Doctor's companion: two legendary creature cards, one with the
// ability and the other a Time Lord Doctor with no other creature types
// (CR 702.124m).
func TestDoctorsCompanionPairsWithALoneDoctor(t *testing.T) {
	if err := Validate(pairDeck(doctorCard(), companionCard())); err != nil {
		t.Errorf("the Doctor and a companion: %v", err)
	}
	notOnlyDoctor := basicLegal("Romana", "Legendary Creature — Time Lord Scientist", "W")
	twoCompanions := companionCard()
	twoCompanions.Name = "Another Companion"
	for _, other := range []cards.Card{notOnlyDoctor, twoCompanions, plainPartnerCard("Partner Guy", "W")} {
		vs := violationsOf(t, Validate(pairDeck(companionCard(), other)))
		found := false
		for _, v := range vs {
			if v.Code == CodeInvalidPartnerPair && v.Card == other.Name {
				found = true
			}
		}
		if !found {
			t.Errorf("companion + %s: want %s naming it, got %+v", other.Name, CodeInvalidPartnerPair, vs)
		}
	}
	// Two Doctors are not a pair: the ability is the companion's.
	other := doctorCard()
	other.Name = "The War Doctor"
	assertHasViolation(t, Validate(pairDeck(doctorCard(), other)), CodeTooManyCommanders)
}
