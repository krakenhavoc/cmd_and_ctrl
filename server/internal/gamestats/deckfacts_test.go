package gamestats

import (
	"math"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

func deckCard(name, typeLine, cost string, identity ...string) cards.Card {
	return cards.Card{ID: uuid.New(), OracleID: uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)), Name: name,
		TypeLine: typeLine, ManaCost: cost, ColorIdentity: identity}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestFactsOfCurveTypesAndAverage(t *testing.T) {
	list := []DeckCard{
		{Card: deckCard("Atraxa", "Legendary Creature — Phyrexian Angel Horror", "{G}{W}{U}{B}", "W", "U", "B", "G"), Count: 1, Commander: true},
		{Card: deckCard("Forest", "Basic Land — Forest", ""), Count: 10},
		{Card: deckCard("Command Tower", "Land", ""), Count: 1},
		{Card: deckCard("Sol Ring", "Artifact", "{1}"), Count: 1},
		{Card: deckCard("Fireball", "Sorcery", "{X}{R}"), Count: 1},
		{Card: deckCard("Steel Hellkite", "Artifact Creature — Dragon", "{6}"), Count: 1},
		{Card: deckCard("Emrakul", "Legendary Creature — Eldrazi", "{15}"), Count: 1},
		{Card: deckCard("Bala Ged Recovery", "Sorcery // Land", "", "G"), Count: 1},
		{Card: deckCard("Zero", "Artifact", ""), Count: 2},
	}
	// The modal double-faced card's cost is on its faces.
	list[7].Card.CardFaces = []cards.CardFace{{TypeLine: "Sorcery", ManaCost: "{2}{G}"}, {TypeLine: "Land"}}
	f := FactsOf(list, 3)

	if f.Cards != 19 || f.Unresolved != 3 {
		t.Fatalf("cards %d unresolved %d, want 19 and 3", f.Cards, f.Unresolved)
	}
	wantCurve := []int{2, 2, 0, 1, 1, 0, 1, 1} // Zero ×2; Sol Ring and Fireball (X is 0); MDFC 3; Atraxa 4; Hellkite 6; Emrakul 7+
	for i, n := range wantCurve {
		if f.ManaCurve[i] != n {
			t.Errorf("curve[%d] = %d, want %d (%v)", i, f.ManaCurve[i], n, f.ManaCurve)
		}
	}
	if f.NonLand != 8 {
		t.Errorf("nonland = %d, want 8", f.NonLand)
	}
	// (4 + 1 + 1 + 6 + 15 + 3 + 0 + 0) / 8
	if f.AvgManaValue == nil || !near(*f.AvgManaValue, 30.0/8) {
		t.Errorf("avg mv = %v, want %v", f.AvgManaValue, 30.0/8)
	}
	types := map[string]int{}
	for _, tc := range f.Types {
		types[tc.Type] = tc.Cards
	}
	want := map[string]int{"land": 11, "creature": 3, "artifact": 4, "sorcery": 2, "enchantment": 0}
	for k, n := range want {
		if types[k] != n {
			t.Errorf("type %s = %d, want %d", k, types[k], n)
		}
	}
	if len(f.Types) != len(CardTypes) || f.Types[0].Type != "land" {
		t.Errorf("types not in CardTypes order: %v", f.Types)
	}
	if got := f.ColourIdentity; len(got) != 4 || got[0] != "W" || got[3] != "G" {
		t.Errorf("identity = %v, want W U B G", got)
	}
	if f.RolesAvailable {
		t.Error("roles available before slice 4")
	}
	if f.Hash == "" || len(f.Hash) != 16 {
		t.Errorf("hash = %q", f.Hash)
	}
}

func TestFactsOfColourShare(t *testing.T) {
	list := []DeckCard{
		{Card: deckCard("A", "Instant", "{W}{W}"), Count: 1},
		{Card: deckCard("B", "Instant", "{W/U}"), Count: 2},      // half each, twice
		{Card: deckCard("C", "Instant", "{B/P}{2/R}"), Count: 1}, // Phyrexian and two-brid count as their colour
		{Card: deckCard("D", "Instant", "{C}{S}{X}{3}"), Count: 1},
		{Card: deckCard("Plains", "Basic Land — Plains", ""), Count: 5},
	}
	f := FactsOf(list, 0)
	want := map[string]float64{"W": 3, "U": 1, "B": 1, "R": 1, "G": 0}
	for _, cs := range f.ColourShare {
		if !near(cs.Symbols, want[cs.Colour]) {
			t.Errorf("%s symbols = %v, want %v", cs.Colour, cs.Symbols, want[cs.Colour])
		}
		if cs.Share == nil || !near(*cs.Share, want[cs.Colour]/6) {
			t.Errorf("%s share = %v, want %v", cs.Colour, cs.Share, want[cs.Colour]/6)
		}
	}
	if order := f.ColourShare; order[0].Colour != "W" || order[4].Colour != "G" {
		t.Errorf("colour order %v", order)
	}
	if none := FactsOf([]DeckCard{{Card: deckCard("Sol Ring", "Artifact", "{1}"), Count: 1}}, 0); none.ColourShare[0].Share != nil {
		t.Error("a colourless deck has a share")
	}
	if empty := FactsOf(nil, 0); empty.AvgManaValue != nil || empty.Hash != "" || len(empty.ManaCurve) != CurveTop+1 {
		t.Errorf("empty deck: %+v", empty)
	}
}

// The deck page's "since last edit" matches seats by this hash, so it
// must be seatDeckHash's: oracle ids with their counts, commanders in.
func TestFactsOfHashIsTheSeatHash(t *testing.T) {
	cmdr := deckCard("Atraxa", "Legendary Creature", "{G}{W}{U}{B}")
	bolt := deckCard("Bolt", "Instant", "{R}")
	f := FactsOf([]DeckCard{{Card: cmdr, Count: 1, Commander: true}, {Card: bolt, Count: 2}}, 0)
	want := DeckHash(map[string]int{cmdr.OracleID.String(): 1, bolt.OracleID.String(): 2})
	if f.Hash != want {
		t.Errorf("hash %s, want %s", f.Hash, want)
	}
	// A new printing of the same card is the same version.
	reprint := bolt
	reprint.ID = uuid.New()
	g := FactsOf([]DeckCard{{Card: cmdr, Count: 1, Commander: true}, {Card: reprint, Count: 2}}, 0)
	if g.Hash != f.Hash {
		t.Error("a new printing changed the version")
	}
}

func TestManaValueOfSplitAndFaces(t *testing.T) {
	split := deckCard("Fire // Ice", "Instant // Instant", "{1}{R} // {1}{U}")
	split.Layout = "split"
	if mv := ManaValue(split); mv != 4 {
		t.Errorf("split mv = %d, want 4", mv)
	}
	adventure := deckCard("Bonecrusher", "Creature // Instant", "{2}{R} // {R}")
	adventure.Layout = "adventure"
	if mv := ManaValue(adventure); mv != 3 {
		t.Errorf("adventure mv = %d, want its front's 3", mv)
	}
	if !IsBasicLand(deckCard("Snow-Covered Forest", "Basic Snow Land — Forest", "")) || IsBasicLand(deckCard("Command Tower", "Land", "")) {
		t.Error("IsBasicLand")
	}
}

func TestFactsOfRolesCountCopiesUnderEachRole(t *testing.T) {
	ring := deckCard("Sol Ring", "Artifact", "{1}")
	counter := deckCard("Counterspell", "Instant", "{U}{U}")
	list := []DeckCard{{Card: ring, Count: 2}, {Card: counter, Count: 1}, {Card: deckCard("Forest", "Basic Land — Forest", ""), Count: 5}}
	roles := map[uuid.UUID][]string{ring.OracleID: {cards.RoleRamp, cards.RoleManaRock}, counter.OracleID: {cards.RoleCounterspell}}
	f := FactsOfRoles(list, 0, func(o uuid.UUID) []string { return roles[o] })
	if !f.RolesAvailable || len(f.Roles) != len(cards.AllRoles) {
		t.Fatalf("roles %+v", f)
	}
	for i, r := range f.Roles {
		if r.Role != cards.AllRoles[i] {
			t.Errorf("role %d = %s, want AllRoles order", i, r.Role)
		}
	}
	got := map[string]int{}
	for _, r := range f.Roles {
		got[r.Role] = r.Cards
	}
	if got[cards.RoleRamp] != 2 || got[cards.RoleManaRock] != 2 || got[cards.RoleCounterspell] != 1 || got[cards.RoleRemoval] != 0 {
		t.Errorf("counts %v", got)
	}
	// Roles never move the version hash.
	if plain := FactsOf(list, 0); plain.Hash != f.Hash || plain.RolesAvailable || plain.Roles != nil {
		t.Errorf("without roles: %+v", plain)
	}
}
