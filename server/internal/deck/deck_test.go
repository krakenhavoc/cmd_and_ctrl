package deck

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// indexWith is a test helper that builds a cards.Index pre-populated
// with the given cards. Reaches into the package to set the
// unexported byName map so tests don't need to round-trip through
// Load + a fixture file.
func indexWith(cs ...cards.Card) *cards.Index {
	idx := cards.NewIndex()
	for _, c := range cs {
		idx.Put(c)
	}
	return idx
}

// basicLegal returns a Card with commander-legal status and a valid
// type line — the default for tests that want "just a legal card".
func basicLegal(name string, typeLine string, colors ...string) cards.Card {
	id := uuid.New()
	return cards.Card{
		ID:            id,
		Name:          name,
		TypeLine:      typeLine,
		ColorIdentity: colors,
		Legalities:    map[string]string{"commander": "legal"},
	}
}

// TestParseTextBasic covers the common plain-text shapes.
func TestParseTextBasic(t *testing.T) {
	src := `
# My deck
Commander:
1 Atraxa, Praetors' Voice

Mainboard:
1 Sol Ring
4 Forest
1x Lightning Bolt
1 Mana Crypt (LTC) 367

Sideboard:
1 Relic of Progenitus
`
	entries, err := ParseText(src)
	if err != nil {
		t.Fatalf("ParseText: %v", err)
	}
	want := []Entry{
		{Name: "Atraxa, Praetors' Voice", Count: 1, IsCommander: true},
		{Name: "Sol Ring", Count: 1},
		{Name: "Forest", Count: 4},
		{Name: "Lightning Bolt", Count: 1},
		{Name: "Mana Crypt", Count: 1},
		{Name: "Relic of Progenitus", Count: 1, IsSideboard: true},
	}
	assertEntries(t, want, entries)
}

func TestParseTextCommanderMarker(t *testing.T) {
	src := `
1 Atraxa, Praetors' Voice *CMDR*
1 Sol Ring
`
	entries, err := ParseText(src)
	if err != nil {
		t.Fatalf("ParseText: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries: got %d, want 2", len(entries))
	}
	if !entries[0].IsCommander {
		t.Errorf("*CMDR* marker did not promote card to commander: %+v", entries[0])
	}
	if entries[0].Name != "Atraxa, Praetors' Voice" {
		t.Errorf("commander name: got %q", entries[0].Name)
	}
}

func TestParseTextSBPrefix(t *testing.T) {
	src := `
1 Sol Ring
SB: 1 Relic of Progenitus
SB 1 Tormod's Crypt
`
	entries, _ := ParseText(src)
	// mainboard + 2 sideboard
	if len(entries) != 3 {
		t.Fatalf("entries: got %d, want 3", len(entries))
	}
	if entries[1].Name != "Relic of Progenitus" || !entries[1].IsSideboard {
		t.Errorf("SB:: got %+v", entries[1])
	}
	if entries[2].Name != "Tormod's Crypt" || !entries[2].IsSideboard {
		t.Errorf("SB : got %+v", entries[2])
	}
}

func TestParseTextIgnoresNoise(t *testing.T) {
	src := `
// Moxfield section comment
--- divider ---
not a card
0 Zero is skipped
`
	entries, err := ParseText(src)
	if err != nil {
		t.Fatalf("ParseText: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries from noise, got %d: %+v", len(entries), entries)
	}
}

func TestParseMoxfieldBasic(t *testing.T) {
	raw := []byte(`{
		"name": "Atraxa Superfriends",
		"commanders": { "Atraxa, Praetors' Voice": {"quantity": 1} },
		"mainboard": { "Sol Ring": {"quantity": 1}, "Forest": {"quantity": 4} },
		"sideboard": { "Relic of Progenitus": {"quantity": 1} }
	}`)
	name, entries, err := ParseMoxfield(raw)
	if err != nil {
		t.Fatalf("ParseMoxfield: %v", err)
	}
	if name != "Atraxa Superfriends" {
		t.Errorf("name: got %q", name)
	}

	// Map iteration is nondeterministic, so compare via a set.
	got := make(map[string]Entry, len(entries))
	for _, e := range entries {
		got[e.Name] = e
	}
	if !got["Atraxa, Praetors' Voice"].IsCommander {
		t.Error("commander entry not flagged as IsCommander")
	}
	if got["Forest"].Count != 4 {
		t.Errorf("Forest count: got %d, want 4", got["Forest"].Count)
	}
	if !got["Relic of Progenitus"].IsSideboard {
		t.Error("sideboard entry not flagged as IsSideboard")
	}
}

func TestParseMoxfieldRejectsCompanion(t *testing.T) {
	raw := []byte(`{
		"commanders": { "Atraxa": {"quantity": 1} },
		"mainboard": { "Sol Ring": {"quantity": 1} },
		"companions": { "Lurrus of the Dream-Den": {"quantity": 1} }
	}`)
	_, _, err := ParseMoxfield(raw)
	if !errors.Is(err, ErrUnsupportedMechanic) {
		t.Errorf("companion: got %v, want ErrUnsupportedMechanic", err)
	}
}

func TestParseMoxfieldBadJSON(t *testing.T) {
	_, _, err := ParseMoxfield([]byte("not json"))
	if err == nil {
		t.Error("bad JSON: expected error, got nil")
	}
}

func TestParseMoxfieldEmptyShape(t *testing.T) {
	_, _, err := ParseMoxfield([]byte(`{"name": "empty"}`))
	if err == nil || !strings.Contains(err.Error(), "no commanders or mainboard") {
		t.Errorf("empty shape: got %v, want shape error", err)
	}
}

// TestResolveUnknownNamesBatched covers the "report all unknowns at
// once" contract.
func TestResolveUnknownNamesBatched(t *testing.T) {
	idx := indexWith(basicLegal("Sol Ring", "Artifact"))
	entries := []Entry{
		{Name: "Sol Ring", Count: 1},
		{Name: "Not A Real Card", Count: 1},
		{Name: "Also Not A Real Card", Count: 1},
	}
	_, err := Resolve(idx, "test", entries)
	var uce *UnknownCardError
	if !errors.As(err, &uce) {
		t.Fatalf("Resolve: got %v, want UnknownCardError", err)
	}
	if len(uce.Names) != 2 {
		t.Errorf("unknowns: got %d, want 2", len(uce.Names))
	}
}

// TestResolvePartnerSniff covers the oracle-text heuristic for
// catching cards with unsupported mechanics. Uses a synthetic card
// because the partner keyword depends on real oracle text.
func TestResolvePartnerSniff(t *testing.T) {
	partner := basicLegal("Thrasios, Triton Hero", "Legendary Creature — Merfolk Wizard", "U", "G")
	partner.OracleText = "Whenever Thrasios deals combat damage. Partner (You can have two commanders if both have partner.)"
	idx := indexWith(partner)
	_, err := Resolve(idx, "t", []Entry{{Name: "Thrasios, Triton Hero", Count: 1, IsCommander: true}})
	if !errors.Is(err, ErrUnsupportedMechanic) {
		t.Errorf("partner: got %v, want ErrUnsupportedMechanic", err)
	}
}

// TestResolvePartnerSniffCommanderOnly guards against the false-
// positive where a non-commander mainboard card whose oracle text
// references the partner keyword (e.g. a tutor that targets partner
// creatures) would have rejected the whole deck.
func TestResolvePartnerSniffCommanderOnly(t *testing.T) {
	ref := basicLegal("Partner Reference", "Instant", "U")
	ref.OracleText = "Choose target creature with Partner (...)."
	cmd := basicLegal("Test Commander", "Legendary Creature — Human", "U")
	idx := indexWith(ref, cmd)
	_, err := Resolve(idx, "t", []Entry{
		{Name: "Test Commander", Count: 1, IsCommander: true},
		{Name: "Partner Reference", Count: 1},
	})
	if err != nil {
		t.Errorf("mainboard partner reference: got %v, want nil", err)
	}
}

// TestResolveUnsupportedMechanicErrorShape covers the new structured
// error that Resolve returns for partner/companion so the HTTP layer
// can emit `{"violations": [...]}` without pattern-matching messages.
func TestResolveUnsupportedMechanicErrorShape(t *testing.T) {
	partner := basicLegal("Thrasios, Triton Hero", "Legendary Creature", "U", "G")
	partner.OracleText = "Partner (You can have two commanders if both have partner.)"
	idx := indexWith(partner)
	_, err := Resolve(idx, "t", []Entry{{Name: "Thrasios, Triton Hero", Count: 1, IsCommander: true}})
	var ume *UnsupportedMechanicError
	if !errors.As(err, &ume) {
		t.Fatalf("Resolve: got %v, want *UnsupportedMechanicError", err)
	}
	if ume.Card != "Thrasios, Triton Hero" {
		t.Errorf("card: got %q, want Thrasios", ume.Card)
	}
	vs := ume.Violations()
	if len(vs) != 1 || vs[0].Code != CodeUnsupportedMechanic {
		t.Errorf("Violations: got %+v, want one %s", vs, CodeUnsupportedMechanic)
	}
	// errors.Is must still unwrap to the sentinel so existing tests
	// and switch statements keep working.
	if !errors.Is(err, ErrUnsupportedMechanic) {
		t.Errorf("errors.Is: unwrap to ErrUnsupportedMechanic failed")
	}
}

// TestValidateDFCLegendaryCommander covers the case where a
// legendary commander's top-level TypeLine is empty (Scryfall's
// modal-double-faced schema stamps per-face lines only). The
// per-face predicate should still accept it.
func TestValidateDFCLegendaryCommander(t *testing.T) {
	cmd := basicLegal("Esika, God of the Tree // The Prismatic Bridge", "", "W", "U", "B", "R", "G")
	cmd.CardFaces = []cards.CardFace{
		{Name: "Esika, God of the Tree", TypeLine: "Legendary Creature — God"},
		{Name: "The Prismatic Bridge", TypeLine: "Legendary Enchantment"},
	}
	list := &List{Commanders: []cards.Card{cmd}}
	basic := basicLegal("Plains", "Basic Land — Plains", "W")
	for i := 0; i < 99; i++ {
		list.Mainboard = append(list.Mainboard, basic)
	}
	err := Validate(list)
	// We expect no commander-legality violation. Other violations
	// (e.g. singleton) are fine — we only care that the DFC legendary
	// isn't rejected as "not a commander".
	if err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			for _, v := range ve.Violations {
				if v.Code == CodeNotLegalCommander {
					t.Errorf("DFC legendary rejected: %+v", v)
				}
			}
		}
	}
}

// --- validation ---

// buildValidDeck returns a List that passes every validation rule.
// Used as the starting point for per-violation tests: mutate one
// field, expect one violation.
func buildValidDeck() *List {
	commander := basicLegal("Atraxa, Praetors' Voice", "Legendary Creature — Phyrexian Angel", "W", "U", "B", "G")
	// 99 cards in mainboard: one of each unique + 60 basics (which
	// are exempt from the singleton rule).
	list := &List{
		Commanders: []cards.Card{commander},
	}
	for i := 0; i < 39; i++ {
		name := fmt.Sprintf("Spell %d", i)
		list.Mainboard = append(list.Mainboard, basicLegal(name, "Instant", "W", "U", "B", "G"))
	}
	basic := basicLegal("Plains", "Basic Land — Plains", "W")
	for i := 0; i < 60; i++ {
		list.Mainboard = append(list.Mainboard, basic)
	}
	return list
}

func TestValidateHappyPath(t *testing.T) {
	list := buildValidDeck()
	if err := Validate(list); err != nil {
		t.Errorf("Validate: got %v, want nil", err)
	}
}

func TestValidateWrongCardCount(t *testing.T) {
	list := buildValidDeck()
	list.Mainboard = list.Mainboard[:50]
	err := Validate(list)
	assertHasViolation(t, err, CodeWrongCardCount)
}

func TestValidateMissingCommander(t *testing.T) {
	list := buildValidDeck()
	list.Commanders = nil
	err := Validate(list)
	assertHasViolation(t, err, CodeMissingCommander)
}

func TestValidateTooManyCommanders(t *testing.T) {
	list := buildValidDeck()
	list.Commanders = append(list.Commanders, list.Commanders[0])
	err := Validate(list)
	assertHasViolation(t, err, CodeTooManyCommanders)
}

func TestValidateNotLegalCommander(t *testing.T) {
	list := buildValidDeck()
	// Replace the commander with a non-legendary creature.
	list.Commanders = []cards.Card{basicLegal("Llanowar Elves", "Creature — Elf Druid", "G")}
	err := Validate(list)
	assertHasViolation(t, err, CodeNotLegalCommander)
}

func TestValidateColorIdentity(t *testing.T) {
	list := &List{
		Commanders: []cards.Card{basicLegal("Mono-White Commander", "Legendary Creature — Human", "W")},
	}
	// 99 singletons, all in-identity Plains, then one offender.
	basic := basicLegal("Plains", "Basic Land — Plains", "W")
	for i := 0; i < 99; i++ {
		list.Mainboard = append(list.Mainboard, basic)
	}
	// Replace the last card with a red card.
	list.Mainboard[0] = basicLegal("Lightning Bolt", "Instant", "R")
	err := Validate(list)
	assertHasViolation(t, err, CodeColorIdentity)
}

func TestValidateSingletonViolation(t *testing.T) {
	list := buildValidDeck()
	// Duplicate a non-basic card.
	list.Mainboard[0] = basicLegal("Sol Ring", "Artifact")
	list.Mainboard[1] = basicLegal("Sol Ring", "Artifact")
	err := Validate(list)
	assertHasViolation(t, err, CodeSingleton)
}

func TestValidateBasicLandsNotSingleton(t *testing.T) {
	list := buildValidDeck()
	// Already has 60 Plains — should pass.
	if err := Validate(list); err != nil {
		t.Errorf("basics: got %v, want nil", err)
	}
}

func TestValidateBannedCard(t *testing.T) {
	list := buildValidDeck()
	list.Mainboard[0] = cards.Card{
		Name:          "Iona, Shield of Emeria",
		TypeLine:      "Legendary Creature",
		ColorIdentity: []string{"W", "U", "B", "G"},
		Legalities:    map[string]string{"commander": "banned"},
	}
	err := Validate(list)
	assertHasViolation(t, err, CodeNotLegalInFormat)
}

func TestValidateSideboardWarning(t *testing.T) {
	list := buildValidDeck()
	list.Sideboard = []cards.Card{basicLegal("Relic of Progenitus", "Artifact")}
	err := Validate(list)
	// Sideboard is a warning only — still surfaces as a violation so
	// the client can show a banner, but nothing else should fail.
	assertHasViolation(t, err, CodeSideboardUnsupported)
}

// --- JSON wire shape ---

func TestValidationErrorMarshals(t *testing.T) {
	ve := &ValidationError{Violations: []Violation{
		{Code: CodeColorIdentity, Card: "Lightning Bolt", Message: "off-color"},
	}}
	raw, err := json.Marshal(ve)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"code":"color_identity_violation"`) {
		t.Errorf("marshal lost code: %s", raw)
	}
}

// --- helpers ---

func assertEntries(t *testing.T, want, got []Entry) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("len: got %d, want %d; got=%+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func assertHasViolation(t *testing.T, err error, code string) {
	t.Helper()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	for _, v := range ve.Violations {
		if v.Code == code {
			return
		}
	}
	t.Errorf("missing violation %q in %+v", code, ve.Violations)
}
