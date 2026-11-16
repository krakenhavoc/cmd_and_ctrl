package draft

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// A drafted creature carries its rules text, power, toughness and
// keywords; a non-creature carries text but no power or toughness, and a
// card with no keywords has none on the wire (#5848).
func TestACardCarriesRulesTextPowerToughnessAndKeywords(t *testing.T) {
	fx := loadBotFixture(t)

	angel := fx.card(t, "Serra Angel", "x")
	if angel.OracleText != "Flying\nVigilance (Attacking doesn't cause this creature to tap.)" {
		t.Errorf("oracle text %q", angel.OracleText)
	}
	if angel.Power != "4" || angel.Toughness != "4" {
		t.Errorf("power/toughness %q/%q", angel.Power, angel.Toughness)
	}
	if !reflect.DeepEqual(angel.Keywords, []string{"Flying", "Vigilance"}) {
		t.Errorf("keywords %v", angel.Keywords)
	}

	swords := fx.card(t, "Swords to Plowshares", "y")
	if swords.OracleText == "" {
		t.Error("a non-creature has no rules text")
	}
	if swords.Power != "" || swords.Toughness != "" {
		t.Errorf("a non-creature has power/toughness %q/%q", swords.Power, swords.Toughness)
	}
	raw, err := json.Marshal(swords)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"power"`, `"toughness"`, `"keywords"`} {
		if strings.Contains(string(raw), key) {
			t.Errorf("a non-creature without keywords sends %s: %s", key, raw)
		}
	}
	raw, _ = json.Marshal(angel)
	for _, key := range []string{`"oracle_text":`, `"power":"4"`, `"toughness":"4"`, `"keywords":["Flying","Vigilance"]`} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("the wire of a creature lacks %s: %s", key, raw)
		}
	}
}

// The fields reach a seat's view of its pack and picks.
func TestAViewsCardsCarryTheNewFields(t *testing.T) {
	fx := loadBotFixture(t)
	t.Cleanup(func() { UseCards(nil); UseRatings(nil) })
	UseCards(fx.idx)
	UseRatings(nil)

	d, err := New(Config{Name: "d", Set: botFixtureSet, Mode: PickOne, Seats: 2, Seed: 5}, t0)
	if err != nil {
		t.Fatal(err)
	}
	host, _ := d.AddSeat("host", uuid.Nil)
	if err := d.Start(host.ID, fx.src, t0); err != nil {
		t.Fatal(err)
	}
	v, ok := d.ViewFor(host.ID, t0, nil)
	if !ok || len(v.Pack) == 0 {
		t.Fatal("no pack in the view")
	}
	creatures := 0
	for _, c := range v.Pack {
		if c.Basic {
			continue
		}
		if c.OracleText == "" && !strings.Contains(c.TypeLine, "Land") && !strings.HasPrefix(c.Name, botFixtureVanilla) {
			t.Errorf("%s has no rules text in the view", c.Name)
		}
		if strings.Contains(c.TypeLine, "Creature") {
			creatures++
			if c.Power == "" || c.Toughness == "" {
				t.Errorf("creature %s has power/toughness %q/%q", c.Name, c.Power, c.Toughness)
			}
		} else if c.Power != "" || c.Toughness != "" {
			t.Errorf("%s (%s) has power/toughness", c.Name, c.TypeLine)
		}
	}
	if creatures == 0 {
		t.Error("the pack held no creature to check")
	}
}

// A two-faced card has its faces' text joined, and its power and
// toughness from the front face when the top level has none.
func TestATwoFacedCardJoinsItsFacesText(t *testing.T) {
	front := cards.CardFace{Name: "Front", OracleText: "Flying", Power: "2", Toughness: "3"}
	back := cards.CardFace{Name: "Back", OracleText: "Trample", Power: "4", Toughness: "5"}
	var c Card
	fillPrinted(&c, cards.Card{
		TypeLine: "Creature // Creature", CardFaces: []cards.CardFace{front, back},
		Keywords: []string{"Flying", "Trample"},
	})
	if c.OracleText != "Flying\n//\nTrample" {
		t.Errorf("oracle text %q", c.OracleText)
	}
	if c.Power != "2" || c.Toughness != "3" {
		t.Errorf("power/toughness %q/%q", c.Power, c.Toughness)
	}
	if !reflect.DeepEqual(c.Keywords, []string{"Flying", "Trample"}) {
		t.Errorf("keywords %v", c.Keywords)
	}

	// A card with its own top-level text uses it as it is.
	var d Card
	fillPrinted(&d, cards.Card{OracleText: "Draw a card.", CardFaces: []cards.CardFace{front}})
	if d.OracleText != "Draw a card." {
		t.Errorf("oracle text %q", d.OracleText)
	}
}

// The view carries the set's average win rate, from the 17Lands file of
// the draft's set, and omits it with no data and for a cube.
func TestAViewCarriesTheSetsAverageWinRate(t *testing.T) {
	fx := loadBotFixture(t)
	t.Cleanup(func() { UseCards(nil); UseRatings(nil) })
	UseCards(fx.idx)

	start := func(cfg Config) View {
		t.Helper()
		d, err := New(cfg, t0)
		if err != nil {
			t.Fatal(err)
		}
		host, _ := d.AddSeat("host", uuid.Nil)
		if err := d.Start(host.ID, fx.src, t0); err != nil {
			t.Fatal(err)
		}
		v, ok := d.ViewFor(host.ID, t0, nil)
		if !ok {
			t.Fatal("no view")
		}
		return v
	}
	cfg := Config{Name: "d", Set: botFixtureSet, Mode: PickOne, Seats: 2, Seed: 5}

	// Three cards at 45%, 48% and 51%: the mean is 48.0.
	UseRatings(fixtureBook(t, fixtureCardNames(t, fx, 3)...))
	v := start(cfg)
	if v.SetAvgWinRate != 48.0 {
		t.Errorf("set average %v, want 48.0", v.SetAvgWinRate)
	}
	raw, _ := json.Marshal(v)
	if !strings.Contains(string(raw), `"set_avg_win_rate":48`) {
		t.Errorf("the wire has no set average: %s", raw)
	}

	// No 17Lands data: omitted.
	UseRatings(nil)
	v = start(cfg)
	if v.SetAvgWinRate != 0 {
		t.Errorf("set average %v with no ratings", v.SetAvgWinRate)
	}
	raw, _ = json.Marshal(v)
	if strings.Contains(string(raw), "set_avg_win_rate") {
		t.Errorf("the wire has a set average with no ratings: %s", raw)
	}

	// A set the book has no file for: omitted.
	UseRatings(fixtureBook(t, fixtureCardNames(t, fx, 3)...))
	other := cfg
	other.Set = "zzz"
	if v := start(other); v.SetAvgWinRate != 0 {
		t.Errorf("set average %v for a set with no file", v.SetAvgWinRate)
	}
}
