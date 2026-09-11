package deck

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// faces_test.go — the import seam (ADR 0034). toGameCard is the only
// import-path function the face model changes, and these fixtures are
// transcribed from the real Scryfall records, NULLS INCLUDED. The
// nulls are the whole bug: a transform or modal_dfc printing ships
// mana_cost, colors and image_uris as null at the top level and puts
// the real values on card_faces.

// seaGateRestorationPrint is the ZNR modal DFC, as Scryfall ships it.
// Note what is NOT set: ManaCost and Colors at the top level, exactly
// as the dump has them.
func seaGateRestorationPrint() cards.Card {
	return cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.New(),
		Name:          "Sea Gate Restoration // Sea Gate, Reborn",
		Layout:        "modal_dfc",
		TypeLine:      "Sorcery // Land",
		ColorIdentity: []string{"U"},
		CardFaces: []cards.CardFace{
			{
				Name:       "Sea Gate Restoration",
				TypeLine:   "Sorcery",
				ManaCost:   "{4}{U}{U}{U}",
				Colors:     []string{"U"},
				OracleText: "Draw cards equal to the number of cards in your hand...",
			},
			{
				Name:       "Sea Gate, Reborn",
				TypeLine:   "Land",
				ManaCost:   "",
				OracleText: "As Sea Gate, Reborn enters, you may pay 3 life. If you don't, it enters tapped.\n{T}: Add {U}.",
			},
		},
	}
}

// jacePrint is the transform shape: the BACK face has no mana cost
// at all and expresses its colour as CR 105.2c's colour indicator.
func jacePrint() cards.Card {
	return cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.New(),
		Name:          "Jace, Vryn's Prodigy // Jace, Telepath Unbound",
		Layout:        "transform",
		TypeLine:      "Legendary Creature — Human Wizard // Legendary Planeswalker — Jace",
		ColorIdentity: []string{"U"},
		CardFaces: []cards.CardFace{
			{
				Name:      "Jace, Vryn's Prodigy",
				TypeLine:  "Legendary Creature — Human Wizard",
				ManaCost:  "{1}{U}",
				Colors:    []string{"U"},
				Power:     "0",
				Toughness: "2",
			},
			{
				Name:           "Jace, Telepath Unbound",
				TypeLine:       "Legendary Planeswalker — Jace",
				ManaCost:       "",
				ColorIndicator: []string{"U"},
				Loyalty:        "5",
			},
		},
	}
}

// fireIcePrint is the split shape, where the top level carries a
// JOINED cost that ParseCost rightly rejects and the faces carry real
// ones. Split is out of scope for the picker, but the spine still has
// to make the card cost SOMETHING rather than nothing.
func fireIcePrint() cards.Card {
	return cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.New(),
		Name:          "Fire // Ice",
		Layout:        "split",
		TypeLine:      "Instant // Instant",
		ManaCost:      "{1}{R} // {1}{U}",
		ColorIdentity: []string{"R", "U"},
		CardFaces: []cards.CardFace{
			{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}"},
			{Name: "Ice", TypeLine: "Instant", ManaCost: "{1}{U}"},
		},
	}
}

func importOne(c cards.Card) game.Card {
	return (&List{Mainboard: []cards.Card{c}}).ToGameCards()[0]
}

// TestModalDFCImportsAsItsFrontFace is the headline. Everything the
// engine reads off a freshly imported Sea Gate Restoration must be
// the SORCERY's, not the concatenation's.
func TestModalDFCImportsAsItsFrontFace(t *testing.T) {
	got := importOne(seaGateRestorationPrint())

	if got.Layout != "modal_dfc" {
		t.Errorf("Layout = %q, want modal_dfc", got.Layout)
	}
	if len(got.Faces) != 2 {
		t.Fatalf("Faces = %d, want 2", len(got.Faces))
	}
	if got.ActiveFace != 0 {
		t.Errorf("ActiveFace = %d, want 0", got.ActiveFace)
	}
	if got.Name != "Sea Gate Restoration" {
		t.Errorf("Name = %q, want the front face's", got.Name)
	}
	if got.TypeLine != "Sorcery" {
		t.Errorf("TypeLine = %q, want %q — the composite is what made "+
			"IsLand() fire on a sorcery", got.TypeLine, "Sorcery")
	}
	if got.ManaCost != "{4}{U}{U}{U}" {
		t.Errorf("ManaCost = %q, want %q. Scryfall's top-level cost is "+
			"NULL on this layout, ParseCost(\"\") succeeds with the zero "+
			"cost, and the spell was free", got.ManaCost, "{4}{U}{U}{U}")
	}
	if got.IsLand() {
		t.Error("imported front face reports IsLand()")
	}
	if _, err := game.ParseCost(got.ManaCost); err != nil {
		t.Errorf("ParseCost(%q): %v", got.ManaCost, err)
	}

	// And the back face is carried, ready for the picker.
	back := got.Faces[1]
	if back.Name != "Sea Gate, Reborn" || back.TypeLine != "Land" {
		t.Errorf("back face = %+v, want the land half", back)
	}
}

// TestTransformImportsCostAndColour is #276's other half and the
// blocker behind #343 / #325: Aang and every other transform card
// imported with no cost and no colours because Scryfall nulls both at
// the top level.
func TestTransformImportsCostAndColour(t *testing.T) {
	got := importOne(jacePrint())

	if got.ManaCost != "{1}{U}" {
		t.Errorf("ManaCost = %q, want {1}{U} — every transform card in "+
			"the game was castable for FREE without this", got.ManaCost)
	}
	if len(got.Colors) != 1 || got.Colors[0] != "U" {
		t.Errorf("Colors = %v, want [U]", got.Colors)
	}
	if got.Power != 0 || got.Toughness != 2 {
		t.Errorf("P/T = %d/%d, want 0/2", got.Power, got.Toughness)
	}
	// The front face is a creature, so it has NO starting loyalty.
	// The pre-0034 "first face that prints a number" fallback gave it
	// the back face's 5.
	if got.StartingLoyalty != 0 {
		t.Errorf("front-face StartingLoyalty = %d, want 0 — a Human "+
			"Wizard has no loyalty", got.StartingLoyalty)
	}
	if got.Faces[1].StartingLoyalty != 5 {
		t.Errorf("back-face StartingLoyalty = %d, want 5",
			got.Faces[1].StartingLoyalty)
	}
	// The back face is blue by COLOUR INDICATOR and by nothing else.
	if cols := got.Faces[1].Colors; len(cols) != 1 || cols[0] != "U" {
		t.Errorf("back-face colours = %v, want [U] from the colour "+
			"indicator", cols)
	}
	// A transform card is not castable as its back face (CR 712.4).
	if faces := got.CastableFaces(); len(faces) != 1 {
		t.Errorf("CastableFaces = %v, want just the front", faces)
	}
}

// TestJunkTypesAreGone pins the ParseTypeLine half: the concatenated
// type line tokenised into the literal types "//" and "—" on all 501
// DFC oracle IDs.
func TestJunkTypesAreGone(t *testing.T) {
	for _, print := range []cards.Card{seaGateRestorationPrint(), jacePrint(), fireIcePrint()} {
		got := importOne(print)
		eff := got.Effective()
		all := append(append(append([]string{}, eff.Types...), eff.Subtypes...), eff.Supertypes...)
		for _, ty := range all {
			switch ty {
			case "//", "—", "-", "":
				t.Errorf("%s: junk type %q in %v", print.Name, ty, all)
			}
		}
	}
}

// TestSplitImportsItsLeftHalfCost is the declared simplification,
// pinned so it cannot regress into either of its two worse
// neighbours: a joined cost (rejected by ParseCost, so #289 refuses
// the cast) or an empty one (free).
func TestSplitImportsItsLeftHalfCost(t *testing.T) {
	got := importOne(fireIcePrint())
	if got.ManaCost != "{1}{R}" {
		t.Errorf("ManaCost = %q, want the LEFT half {1}{R}", got.ManaCost)
	}
	if _, err := game.ParseCost(got.ManaCost); err != nil {
		t.Errorf("ParseCost(%q): %v — a split card must cost something", got.ManaCost, err)
	}
	if faces := got.CastableFaces(); len(faces) != 1 {
		t.Errorf("split CastableFaces = %v, want just the left half "+
			"(fusing is not designed)", faces)
	}
}

// TestSingleFacedImportIsUntouched is the other half of the promise:
// nothing about an ordinary card changes.
func TestSingleFacedImportIsUntouched(t *testing.T) {
	got := importOne(cards.Card{
		ID: uuid.New(), OracleID: uuid.New(),
		Name: "Grizzly Bears", Layout: "normal",
		TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Power: "2", Toughness: "2",
	})
	if got.Faces != nil {
		t.Errorf("single-faced card got %d faces, want nil", len(got.Faces))
	}
	if got.ActiveFace != 0 {
		t.Errorf("ActiveFace = %d, want 0", got.ActiveFace)
	}
	if got.Name != "Grizzly Bears" || got.TypeLine != "Creature — Bear" ||
		got.ManaCost != "{1}{G}" || got.Power != 2 || got.Toughness != 2 {
		t.Errorf("single-faced import changed: %+v", got)
	}
}

// TestUnsupportedLayoutIsWarnedNotRefused is the honest-fallback
// half of ADR 0034: a card whose layout the engine only half-plays
// imports, works as its front face, and SAYS SO. Refusing would
// reject a whole deck over a cosmetic simplification; saying nothing
// is what produced #265.
func TestUnsupportedLayoutIsWarnedNotRefused(t *testing.T) {
	list := &List{
		Mainboard: []cards.Card{
			jacePrint(),               // transform
			fireIcePrint(),            // split
			seaGateRestorationPrint(), // modal_dfc — NOT a warning
			{Name: "Island", Layout: "normal", TypeLine: "Basic Land — Island"},
		},
	}
	vs := unsupportedLayoutViolations(list)
	byCode := map[string]int{}
	joined := ""
	for _, v := range vs {
		byCode[v.Code]++
		joined += v.Message + "\n"
	}
	if byCode[CodeUnsupportedLayout] != 2 {
		t.Fatalf("got %d layout warnings, want 2 (transform + split): %v",
			byCode[CodeUnsupportedLayout], vs)
	}
	if !strings.Contains(joined, "Jace, Vryn's Prodigy") {
		t.Errorf("warning does not name the FRONT face: %q", joined)
	}
	if strings.Contains(joined, "Sea Gate") {
		t.Error("a modal DFC was warned about; both its faces are " +
			"playable and the picker chooses between them")
	}
	if strings.Contains(joined, "Island") {
		t.Error("a single-faced card was warned about")
	}
}

// TestLayoutWarningGroupsRatherThanSpams: a deck running a dozen
// transform cards gets one banner naming them, not a dozen banners.
func TestLayoutWarningGroupsRatherThanSpams(t *testing.T) {
	a, b := jacePrint(), jacePrint()
	b.Name = "Delver of Secrets // Insectile Aberration"
	b.CardFaces[0].Name = "Delver of Secrets"
	list := &List{Mainboard: []cards.Card{a, a, a, b}}
	vs := unsupportedLayoutViolations(list)
	if len(vs) != 1 {
		t.Fatalf("got %d warnings, want 1 grouped by layout: %v", len(vs), vs)
	}
	msg := vs[0].Message
	if !strings.Contains(msg, "Delver of Secrets") || !strings.Contains(msg, "Jace, Vryn's Prodigy") {
		t.Errorf("grouped warning does not name both cards: %q", msg)
	}
	if strings.Count(msg, "Jace, Vryn's Prodigy") != 1 {
		t.Errorf("a card running three copies is named %d times, want once: %q",
			strings.Count(msg, "Jace, Vryn's Prodigy"), msg)
	}
}

// TestReversibleCardIsNotPlayable pins the index filter. All 81
// reversible_card printings carry mana_cost: null and type_line:
// null, and 71 byName keys resolved to one before this.
func TestReversibleCardIsNotPlayable(t *testing.T) {
	idx := cards.NewIndex()
	real := cards.Card{
		ID: uuid.New(), OracleID: uuid.New(),
		Name: "Wrath of God", Layout: "normal", SetType: "expansion",
		TypeLine: "Sorcery", ManaCost: "{2}{W}{W}",
	}
	novelty := cards.Card{
		ID: uuid.New(), OracleID: uuid.New(),
		Name: "Wrath of God", Layout: "reversible_card", SetType: "expansion",
	}
	// Insert the placeholder LAST — "most recent printing wins" is
	// the tiebreak among same-class records, so an unfiltered
	// reversible_card would take the key here.
	idx.Put(real)
	idx.Put(novelty)

	got, ok := idx.FindByName("Wrath of God")
	if !ok {
		t.Fatal("Wrath of God is not in the index")
	}
	if got.Layout != "normal" {
		t.Errorf("byName resolved to the %q printing, want the real one — "+
			"a reversible_card record has no cost and no type line, so "+
			"the card would import as a free typeless blank", got.Layout)
	}
}
