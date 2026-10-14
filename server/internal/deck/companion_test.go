package deck

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// companion_test.go — ADR 0172: the companion slot in every import
// format, and the deck check (companion.go).

const jeganthaOracle = "f0587c55-06c1-4930-a62e-668f37464ce9"

// jegantha is Jegantha, the Wellspring as the Scryfall record has it:
// the fields the parsers and the deck check read.
func jegantha() cards.Card {
	c := basicLegal("Jegantha, the Wellspring", "Legendary Creature — Elemental Elk", "W", "U", "B", "R", "G")
	c.OracleID = uuid.MustParse(jeganthaOracle)
	c.ManaCost = "{4}{R/G}"
	c.Keywords = []string{"Companion"}
	c.OracleText = "Companion — No card in your starting deck has more than one of the same mana symbol in its mana cost. (If this card is your chosen companion, you may put it into your hand from outside the game for {3} as a sorcery.)\n{T}: Add {W}{U}{B}{R}{G}. This mana can't be spent to pay generic mana costs."
	return c
}

// costed is a legal card with a mana cost and a fresh oracle ID.
func costed(name, typeLine, cost string, colors ...string) cards.Card {
	c := basicLegal(name, typeLine, colors...)
	c.OracleID = uuid.New()
	c.ManaCost = cost
	return c
}

// companionDeck is a legal 100-card five-colour deck that fulfils
// Jegantha's condition: a commander and 99 cards whose costs never
// repeat a symbol, with Jegantha declared as the companion. extra
// replaces that many of the 99, so a test can put a breaking card in
// and keep the count at 100.
func companionDeck(extra ...cards.Card) *List {
	cmd := costed("Five Colour Commander", "Legendary Creature — Dragon", "{2}{W}{U}{B}{R}{G}", "W", "U", "B", "R", "G")
	list := &List{Commanders: []cards.Card{cmd}}
	for i := range 99 - len(extra) {
		list.Mainboard = append(list.Mainboard, costed(fmt.Sprintf("Spell %d", i), "Instant", "{1}{R}", "R"))
	}
	list.Mainboard = append(list.Mainboard, extra...)
	j := jegantha()
	list.Companion = &j
	return list
}

func violationCodes(err error) []string {
	ve, ok := err.(*ValidationError)
	if !ok {
		return nil
	}
	var out []string
	for _, v := range ve.Violations {
		out = append(out, v.Code)
	}
	return out
}

func TestCompanionDeckThatFulfilsJeganthaValidates(t *testing.T) {
	if err := Validate(companionDeck()); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

// The companion is not one of the 100 (CR 702.139, ruling: "Your
// companion is not one of your one hundred cards"): a 100-card list
// plus a companion is the right size, and the companion is not judged
// by singleton or by the 100's colour identity.
func TestCompanionIsOutsideTheHundred(t *testing.T) {
	list := companionDeck()
	if n := len(list.Commanders) + len(list.Mainboard); n != 100 {
		t.Fatalf("fixture has %d cards", n)
	}
	// A second Jegantha in the 99 is a different card from the
	// companion; singleton covers only the 100.
	list.Mainboard[0] = jegantha()
	if err := Validate(list); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

func TestJeganthaConditionEdgeCases(t *testing.T) {
	split := costed("Fire // Ice", "Instant // Instant", "{1}{R} // {1}{U}", "U", "R")
	split.Layout = "split"
	split.CardFaces = []cards.CardFace{{Name: "Fire", ManaCost: "{1}{R}"}, {Name: "Ice", ManaCost: "{1}{U}"}}

	adventure := costed("Bonecrusher Giant // Stomp", "Creature — Giant // Instant — Adventure", "{2}{R} // {1}{R}", "R")
	adventure.Layout = "adventure"
	adventure.CardFaces = []cards.CardFace{{Name: "Bonecrusher Giant", ManaCost: "{2}{R}"}, {Name: "Stomp", ManaCost: "{1}{R}"}}

	dfc := costed("Valki, God of Lies // Tibalt, Cosmic Impostor", "Legendary Creature — God // Legendary Planeswalker — Tibalt", "", "B", "R")
	dfc.Layout = "modal_dfc"
	dfc.CardFaces = []cards.CardFace{{Name: "Valki, God of Lies", ManaCost: "{1}{B}"}, {Name: "Tibalt, Cosmic Impostor", ManaCost: "{5}{B}{R}"}}

	dfcTwice := costed("Twice Front // Back", "Creature // Creature", "", "G")
	dfcTwice.Layout = "transform"
	dfcTwice.CardFaces = []cards.CardFace{{Name: "Twice Front", ManaCost: "{G}{G}"}, {Name: "Back", ManaCost: ""}}

	cases := []struct {
		name  string
		card  cards.Card
		fails bool
	}{
		{"hybrid twice", costed("Kitchen Finks", "Creature", "{1}{G/W}{G/W}", "G", "W"), true},
		{"X twice", costed("Crackle with Power", "Sorcery", "{X}{X}{X}{R}", "R"), true},
		{"coloured twice", costed("Counterspell", "Instant", "{U}{U}", "U"), true},
		{"different symbols", costed("Hybrid Neighbour", "Creature", "{R}{R/G}", "R", "G"), false},
		{"split card sees each half's symbols", split, true},
		{"adventurer is its main face", adventure, false},
		{"double-faced card is its front face", dfc, false},
		{"double-faced front face repeats", dfcTwice, true},
		{"land has no cost", basicLegal("Command Tower", "Land"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(companionDeck(tc.card))
			failed := slices.Contains(violationCodes(err), CodeCompanionConditionNotMet)
			if failed != tc.fails {
				t.Errorf("condition failed = %v, want %v (%v)", failed, tc.fails, err)
			}
		})
	}
}

// CR 702.139b: in Commander the starting deck is checked before the
// commander is set aside, so a commander that repeats a symbol breaks
// the condition.
func TestJeganthaConditionCountsTheCommander(t *testing.T) {
	list := companionDeck()
	list.Commanders[0].ManaCost = "{W}{W}{U}{B}{R}{G}"
	err := Validate(list)
	ve, _ := err.(*ValidationError)
	if ve == nil {
		t.Fatal("Validate passed a commander with {W}{W}")
	}
	for _, v := range ve.Violations {
		if v.Code == CodeCompanionConditionNotMet && v.Card == "Five Colour Commander" {
			return
		}
	}
	t.Errorf("no condition violation naming the commander: %+v", ve.Violations)
}

// The violation list names each offender, one row each, up to ten.
func TestJeganthaConditionListsAtMostTenOffenders(t *testing.T) {
	var bad []cards.Card
	for i := range 15 {
		bad = append(bad, costed(fmt.Sprintf("Double %d", i), "Instant", "{R}{R}", "R"))
	}
	n := 0
	for _, code := range violationCodes(Validate(companionDeck(bad...))) {
		if code == CodeCompanionConditionNotMet {
			n++
		}
	}
	if n != maxCompanionOffenders {
		t.Errorf("%d offender rows, want %d", n, maxCompanionOffenders)
	}
}

func TestCompanionMustFitTheCommandersColourIdentity(t *testing.T) {
	list := companionDeck()
	list.Commanders[0].ColorIdentity = []string{"R"}
	list.Commanders[0].ManaCost = "{2}{R}"
	codes := violationCodes(Validate(list))
	if !slices.Contains(codes, CodeCompanionColorIdentity) {
		t.Errorf("codes %v, want %s", codes, CodeCompanionColorIdentity)
	}
}

// Owner decision 5: the commander's own card can't also be the
// companion, not even as a second copy.
func TestCompanionSameCardAsCommanderIsRefused(t *testing.T) {
	list := companionDeck()
	list.Commanders = []cards.Card{jegantha()}
	codes := violationCodes(Validate(list))
	if !slices.Contains(codes, CodeCompanionIsCommander) {
		t.Errorf("codes %v, want %s", codes, CodeCompanionIsCommander)
	}
}

// The owner lifted Resolve's refusal: Jegantha (or any companion) is an
// ordinary commander. Its companion ability does nothing, so the deck
// check has no companion to judge.
func TestCompanionCardAsCommanderResolvesAndValidates(t *testing.T) {
	j := jegantha()
	lurrus := basicLegal("Lurrus of the Dream-Den", "Legendary Creature — Cat Nightmare", "W", "B")
	lurrus.Keywords = []string{"Companion", "Lifelink"}
	lurrus.OracleText = "Companion — Each permanent card in your starting deck has mana value 2 or less.\nLifelink"
	idx := indexWith(j, lurrus, basicLegal("Sol Ring", "Artifact"))
	for _, name := range []string{j.Name, lurrus.Name} {
		list, err := Resolve(idx, "t", []Entry{{Name: name, Count: 1, IsCommander: true}, {Name: "Sol Ring", Count: 1}})
		if err != nil {
			t.Fatalf("%s as commander: Resolve: %v", name, err)
		}
		if list.Companion != nil || len(list.Commanders) != 1 {
			t.Fatalf("%s as commander: companion %v, commanders %d", name, list.Companion, len(list.Commanders))
		}
		for _, code := range violationCodes(Validate(list)) {
			if strings.HasPrefix(code, "companion") || code == CodeUnsupportedMechanic {
				t.Errorf("%s as commander: unexpected %s", name, code)
			}
		}
	}
}

func TestCompanionViolations(t *testing.T) {
	notOne := companionDeck()
	sol := basicLegal("Sol Ring", "Artifact")
	notOne.Companion = &sol

	unsupported := companionDeck()
	lurrus := costed("Lurrus of the Dream-Den", "Legendary Creature — Cat Nightmare", "{1}{W/B}{W/B}", "W", "B")
	lurrus.Keywords = []string{"Companion"}
	unsupported.Companion = &lurrus

	banned := companionDeck()
	banned.Companion.Legalities = map[string]string{"commander": "banned"}

	two := companionDeck()
	two.AddCompanion(jegantha())

	for _, tc := range []struct {
		name string
		list *List
		want string
	}{
		{"no companion ability", notOne, CodeNotACompanion},
		{"condition not registered", unsupported, CodeUnsupportedMechanic},
		{"not legal", banned, CodeCompanionNotLegal},
		{"a second companion", two, CodeTooManyCompanions},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if codes := violationCodes(Validate(tc.list)); !slices.Contains(codes, tc.want) {
				t.Errorf("codes %v, want %s", codes, tc.want)
			}
		})
	}
}

// Owner decision 3: a lone companion card in the sideboard is read as
// the companion, with a warning; anything else in the sideboard keeps
// the sideboard warning, and an explicit companion always wins.
func TestCompanionFromALoneSideboardCard(t *testing.T) {
	j := jegantha()
	cmd := costed("Five Colour Commander", "Legendary Creature — Dragon", "{2}{W}{U}{B}{R}{G}", "W", "U", "B", "R", "G")
	other := basicLegal("Relic of Progenitus", "Artifact")
	idx := indexWith(j, cmd, other)

	lone, err := Resolve(idx, "t", []Entry{{Name: cmd.Name, Count: 1, IsCommander: true}, {Name: j.Name, Count: 1, IsSideboard: true}})
	if err != nil {
		t.Fatal(err)
	}
	if lone.Companion == nil || lone.Companion.Name != j.Name || !lone.CompanionFromSideboard || len(lone.Sideboard) != 0 {
		t.Fatalf("lone sideboard: companion %v, from sideboard %v, sideboard %d", lone.Companion, lone.CompanionFromSideboard, len(lone.Sideboard))
	}
	if codes := violationCodes(Validate(lone)); !slices.Contains(codes, CodeCompanionFromSideboard) || slices.Contains(codes, CodeSideboardUnsupported) {
		t.Errorf("lone sideboard codes %v, want %s and no sideboard warning", codes, CodeCompanionFromSideboard)
	}

	shared, err := Resolve(idx, "t", []Entry{{Name: cmd.Name, Count: 1, IsCommander: true}, {Name: j.Name, Count: 1, IsSideboard: true}, {Name: other.Name, Count: 1, IsSideboard: true}})
	if err != nil {
		t.Fatal(err)
	}
	if shared.Companion != nil || len(shared.Sideboard) != 2 {
		t.Errorf("shared sideboard: companion %v, sideboard %d; want none and 2", shared.Companion, len(shared.Sideboard))
	}

	notCompanion, err := Resolve(idx, "t", []Entry{{Name: cmd.Name, Count: 1, IsCommander: true}, {Name: other.Name, Count: 1, IsSideboard: true}})
	if err != nil {
		t.Fatal(err)
	}
	if notCompanion.Companion != nil {
		t.Errorf("a lone non-companion sideboard card became the companion")
	}

	declared, err := Resolve(idx, "t", []Entry{{Name: cmd.Name, Count: 1, IsCommander: true}, {Name: j.Name, Count: 1, IsCompanion: true}, {Name: j.Name, Count: 1, IsSideboard: true}})
	if err != nil {
		t.Fatal(err)
	}
	if declared.Companion == nil || declared.CompanionFromSideboard || len(declared.Sideboard) != 0 {
		t.Errorf("declared + sideboard copy: companion %v, from sideboard %v, sideboard %d; want declared, the copy dropped", declared.Companion, declared.CompanionFromSideboard, len(declared.Sideboard))
	}
}

func TestToGameCardsSeatsTheCompanionLast(t *testing.T) {
	list := companionDeck()
	gc := list.ToGameCards()
	if len(gc) != 101 {
		t.Fatalf("%d game cards, want 101", len(gc))
	}
	last := gc[len(gc)-1]
	if !last.IsCompanion || last.IsCommander || last.Name != "Jegantha, the Wellspring" {
		t.Errorf("last card %q companion=%v commander=%v", last.Name, last.IsCompanion, last.IsCommander)
	}
	for _, c := range gc[:100] {
		if c.IsCompanion {
			t.Errorf("%q is stamped companion", c.Name)
		}
	}
}

// ---- Parsers ----

func TestParseTextCompanionSection(t *testing.T) {
	for _, header := range []string{"Companion", "Companion:", "// Companion", "COMPANION"} {
		src := "Commander\n1 Five Colour Commander\n\n" + header + "\n1 Jegantha, the Wellspring\n\nDeck\n1 Sol Ring\n"
		entries, err := ParseText(src)
		if err != nil {
			t.Fatalf("%q: %v", header, err)
		}
		want := []Entry{
			{Name: "Five Colour Commander", Count: 1, IsCommander: true},
			{Name: "Jegantha, the Wellspring", Count: 1, IsCompanion: true},
			{Name: "Sol Ring", Count: 1},
		}
		assertEntries(t, want, entries)
	}
}

func TestTextRoundTripKeepsTheCompanion(t *testing.T) {
	list := companionDeck()
	entries, err := ParseText(list.Text())
	if err != nil {
		t.Fatal(err)
	}
	var companions []string
	for _, e := range entries {
		if e.IsCompanion {
			companions = append(companions, e.Name)
		}
	}
	if !slices.Equal(companions, []string{"Jegantha, the Wellspring"}) {
		t.Errorf("companions after the round trip: %v\n%s", companions, list.Text())
	}
}

// The trailing commander block (Moxfield's and MTGO's export): no row
// marked commander, a final headerless block of one or two rows.
func TestParseTextTrailingCommanderBlock(t *testing.T) {
	entries, err := ParseText("1 Sol Ring\n1 Arcane Signet\n\n1 Tymna the Weaver\n1 Thrasios, Triton Hero\n")
	if err != nil {
		t.Fatal(err)
	}
	assertEntries(t, []Entry{
		{Name: "Sol Ring", Count: 1},
		{Name: "Arcane Signet", Count: 1},
		{Name: "Tymna the Weaver", Count: 1, IsCommander: true},
		{Name: "Thrasios, Triton Hero", Count: 1, IsCommander: true},
	}, entries)

	for name, src := range map[string]string{
		"a commander is already marked": "Commander\n1 Atraxa\nDeck\n1 Sol Ring\n\n1 Arcane Signet\n",
		"one block only":                "1 Sol Ring\n1 Arcane Signet\n",
		"the last block has a header":   "1 Sol Ring\n\nSideboard\n1 Relic of Progenitus\n",
		"three rows in the last block":  "1 Sol Ring\n\n1 A\n1 B\n1 C\n",
	} {
		entries, err := ParseText(src)
		if err != nil {
			t.Fatal(err)
		}
		marked := 0
		for _, e := range entries {
			if e.IsCommander {
				marked++
			}
		}
		want := 0
		if name == "a commander is already marked" {
			want = 1
		}
		if marked != want {
			t.Errorf("%s: %d commanders, want %d", name, marked, want)
		}
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// fixtureIndex is a card index holding every name in entries as a legal
// card, Jegantha as itself, so a real export resolves without the dump.
func fixtureIndex(entries []Entry) *cards.Index {
	j := jegantha()
	cs := []cards.Card{j}
	for _, e := range entries {
		if strings.EqualFold(e.Name, j.Name) {
			continue
		}
		cs = append(cs, basicLegal(e.Name, "Legendary Creature — Test", "W", "U", "B", "R", "G"))
	}
	return indexWith(cs...)
}

func countEntries(entries []Entry, pred func(Entry) bool) (n int, names []string) {
	for _, e := range entries {
		if pred(e) {
			n += e.Count
			names = append(names, e.Name)
		}
	}
	return n, names
}

// Moxfield's text export (ADR 0172, "Import formats"): the companion
// arrives under SIDEBOARD: when Moxfield lists it on both boards, and
// the commander comes last after a blank line.
func TestMoxfieldExportLoneSideboardCompanion(t *testing.T) {
	entries, err := ParseText(string(readFixture(t, "moxfield_3401_export_lone_sideboard.txt")))
	if err != nil {
		t.Fatal(err)
	}
	if _, cmds := countEntries(entries, func(e Entry) bool { return e.IsCommander }); !slices.Equal(cmds, []string{"Sisay, Weatherlight Captain"}) {
		t.Fatalf("commanders %v", cmds)
	}
	list, err := Resolve(fixtureIndex(entries), "", entries)
	if err != nil {
		t.Fatal(err)
	}
	if list.Companion == nil || list.Companion.Name != "Jegantha, the Wellspring" || !list.CompanionFromSideboard {
		t.Fatalf("companion %v, from sideboard %v", list.Companion, list.CompanionFromSideboard)
	}
	if n := len(list.Commanders) + len(list.Mainboard); n != 100 {
		t.Errorf("%d cards in the deck, want 100", n)
	}
}

// Jegantha among other sideboard cards is not read as the companion.
func TestMoxfieldExportSharedSideboard(t *testing.T) {
	entries, err := ParseText(string(readFixture(t, "moxfield_3401_export_shared_sideboard.txt")))
	if err != nil {
		t.Fatal(err)
	}
	list, err := Resolve(fixtureIndex(entries), "", entries)
	if err != nil {
		t.Fatal(err)
	}
	if list.Companion != nil || len(list.Sideboard) != 15 || len(list.Commanders) != 1 {
		t.Errorf("companion %v, sideboard %d, commanders %d; want none, 15, 1", list.Companion, len(list.Sideboard), len(list.Commanders))
	}
}

// A companion on Moxfield's companions board alone is left out of the
// text export, so the list has no companion at all.
func TestMoxfieldExportDropsACompanionOnlyCompanion(t *testing.T) {
	entries, err := ParseText(string(readFixture(t, "moxfield_3401_export_companion_dropped.txt")))
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := countEntries(entries, func(e Entry) bool { return e.IsCompanion || strings.HasPrefix(e.Name, "Jegantha") }); n != 0 {
		t.Errorf("found a companion row in an export that has none")
	}
	if n, _ := countEntries(entries, func(e Entry) bool { return e.IsCommander }); n != 1 {
		t.Errorf("%d commanders, want the trailing one", n)
	}
}

func TestMoxfieldJSONCompanionsBoard(t *testing.T) {
	for _, tc := range []struct {
		file      string
		sideboard int // sideboard rows of the companion before Resolve
	}{
		{"moxfield_3401_sisay_companion_and_sideboard.json", 1},
		{"moxfield_3401_sisay_companion_only.json", 0},
	} {
		t.Run(tc.file, func(t *testing.T) {
			_, entries, err := ParseMoxfield(readFixture(t, tc.file))
			if err != nil {
				t.Fatal(err)
			}
			if n, _ := countEntries(entries, func(e Entry) bool { return e.IsCompanion }); n != 1 {
				t.Fatalf("%d companion rows, want 1", n)
			}
			if n, _ := countEntries(entries, func(e Entry) bool { return e.IsSideboard && strings.HasPrefix(e.Name, "Jegantha") }); n != tc.sideboard {
				t.Fatalf("%d sideboard Jegantha rows, want %d", n, tc.sideboard)
			}
			list, err := Resolve(fixtureIndex(entries), "", entries)
			if err != nil {
				t.Fatal(err)
			}
			if list.Companion == nil || list.CompanionFromSideboard || len(list.Sideboard) != 0 {
				t.Errorf("companion %v, from sideboard %v, sideboard %d", list.Companion, list.CompanionFromSideboard, len(list.Sideboard))
			}
			if n := len(list.Commanders) + len(list.Mainboard); n != 100 {
				t.Errorf("%d cards, want 100", n)
			}
		})
	}
}

func TestArchidektCompanionCategory(t *testing.T) {
	_, entries, err := ParseArchidekt(readFixture(t, "archidekt_3401_companion_category.json"))
	if err != nil {
		t.Fatal(err)
	}
	// Jegantha's categories are ["Companion", "Commander"]: the first
	// recognised one wins, so it is the companion, not a commander.
	if _, names := countEntries(entries, func(e Entry) bool { return e.IsCompanion }); !slices.Equal(names, []string{"Jegantha, the Wellspring"}) {
		t.Errorf("companions %v", names)
	}
	if _, names := countEntries(entries, func(e Entry) bool { return e.IsCommander }); !slices.Equal(names, []string{"Infinite Guideline Station"}) {
		t.Errorf("commanders %v", names)
	}
	list, err := Resolve(fixtureIndex(entries), "", entries)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(list.Commanders) + len(list.Mainboard); n != 100 || list.Companion == nil {
		t.Errorf("%d cards, companion %v; want 100 and Jegantha", n, list.Companion)
	}
}

func TestArchidektLoneSideboardCompanion(t *testing.T) {
	_, entries, err := ParseArchidekt(readFixture(t, "archidekt_3401_companion_in_sideboard.json"))
	if err != nil {
		t.Fatal(err)
	}
	list, err := Resolve(fixtureIndex(entries), "", entries)
	if err != nil {
		t.Fatal(err)
	}
	if list.Companion == nil || !list.CompanionFromSideboard {
		t.Fatalf("companion %v, from sideboard %v", list.Companion, list.CompanionFromSideboard)
	}
	if n := len(list.Commanders) + len(list.Mainboard); n != 100 {
		t.Errorf("%d cards, want 100", n)
	}
}
