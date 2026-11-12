package deck

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prototype_import_test.go — CR 718 (#5656). Scryfall's `prototype`
// layout is one face whose inset is only in the oracle text; the
// importer gives the card the inset as a second face.

func blitzAutomatonPrint() cards.Card {
	return cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.MustParse("19f344ee-031d-4346-9553-c5d32b02df43"),
		Name:          "Blitz Automaton",
		Layout:        "prototype",
		TypeLine:      "Artifact Creature — Construct",
		ManaCost:      "{7}",
		Power:         "6",
		Toughness:     "4",
		Colors:        []string{},
		ColorIdentity: []string{"R"},
		Keywords:      []string{"Haste", "Prototype"},
		OracleText:    "Prototype {2}{R} — 3/2 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nHaste",
	}
}

func TestPrototypeCardImportsWithItsInsetAsASecondFace(t *testing.T) {
	got := importOne(blitzAutomatonPrint())

	if !game.IsPrototypeCard(got) {
		t.Fatalf("layout %q with %d faces is not a prototype card", got.Layout, len(got.Faces))
	}
	// CR 718.4: in a library or a hand the card has only its normal
	// characteristics.
	if got.ActiveFace != 0 || got.ManaCost != "{7}" || got.Power != 6 || got.Toughness != 4 || len(got.Effective().Colors) != 0 {
		t.Errorf("imported card = face %d %s %d/%d %v, want the colourless {7} 6/4",
			got.ActiveFace, got.ManaCost, got.Power, got.Toughness, got.Effective().Colors)
	}
	if !game.HasKeyword(&got, "haste") {
		t.Error("the imported card has no haste")
	}
	inset := got.Faces[game.PrototypeFace]
	if inset.ManaCost != "{2}{R}" || inset.Power != 3 || inset.Toughness != 2 || len(inset.Colors) != 1 || inset.Colors[0] != "R" {
		t.Errorf("inset face = %s %d/%d %v, want a red {2}{R} 3/2", inset.ManaCost, inset.Power, inset.Toughness, inset.Colors)
	}
	// CR 718.5: everything else is the same on both.
	if inset.Name != got.Faces[0].Name || inset.TypeLine != got.Faces[0].TypeLine || inset.OracleText != got.Faces[0].OracleText {
		t.Errorf("inset face differs in name, type line or text: %+v", inset)
	}

	got.SetFace(game.PrototypeFace)
	if !got.Prototyped() || got.ManaValue() != 3 || !game.HasKeyword(&got, "haste") {
		t.Errorf("prototyped: %v mv %d haste %v, want a mana value 3 creature with haste",
			got.Prototyped(), got.ManaValue(), game.HasKeyword(&got, "haste"))
	}
	// One catalog entry for both (CR 718.5).
	if key := game.CatalogKey(got); key != "19f344ee-031d-4346-9553-c5d32b02df43" {
		t.Errorf("prototyped catalog key = %q, want the bare oracle ID", key)
	}
}

// A prototype-layout record whose text has no inset stays single-faced.
func TestPrototypeLayoutWithoutAnInsetIsSingleFaced(t *testing.T) {
	p := blitzAutomatonPrint()
	p.OracleText = "Haste"
	if got := importOne(p); len(got.Faces) != 0 {
		t.Errorf("Faces = %d, want none", len(got.Faces))
	}
}
