package deck

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// meld_import_test.go — ADR 0145, #2699: a meld card imports with its
// pair's combined back face, a deck of meld cards validates without a
// layout banner, and the combined back face itself is refused.

var (
	urzaLPID      = uuid.MustParse("600259ab-2bb7-4e09-b957-b07dbd37a521")
	mightstoneID  = uuid.MustParse("02aea379-b444-46a3-82f4-3038f698d4f4")
	urzaPWPrintID = uuid.MustParse("40a01679-1a23-4b8f-9a4b-0e5d5e0f5a6c")
)

func urzaMeldParts() []cards.RelatedCard {
	return []cards.RelatedCard{
		{ID: urzaPWPrintID, Component: "meld_result", Name: "Urza, Planeswalker"},
		{ID: mightstoneID, Component: "meld_part", Name: "The Mightstone and Weakstone"},
		{ID: urzaLPID, Component: "meld_part", Name: "Urza, Lord Protector"},
		{ID: uuid.New(), Component: "token", Name: "Soldier"},
	}
}

func urzaLordProtectorPrint() cards.Card {
	return cards.Card{
		ID: urzaLPID, OracleID: uuid.MustParse("df2af646-3e5b-43a3-8f3e-50565889f456"),
		Name: "Urza, Lord Protector", Layout: "meld",
		TypeLine: "Legendary Creature — Human Artificer", ManaCost: "{1}{W}{U}",
		Colors: []string{"W", "U"}, ColorIdentity: []string{"W", "U"},
		Power: "2", Toughness: "4", Keywords: []string{"Meld"},
		OracleText: "Artifact, instant, and sorcery spells you cast cost {1} less to cast.\n{7}: If you both own and control Urza, Lord Protector and an artifact named The Mightstone and Weakstone, exile them, then meld them into Urza, Planeswalker. Activate only as a sorcery.",
		Legalities: map[string]string{"commander": "legal"},
		AllParts:   urzaMeldParts(),
	}
}

func mightstonePrint() cards.Card {
	return cards.Card{
		ID: mightstoneID, OracleID: uuid.MustParse("c396db03-bf11-4e20-b630-4f9aa8fd78da"),
		Name: "The Mightstone and Weakstone", Layout: "meld",
		TypeLine: "Legendary Artifact — Powerstone", ManaCost: "{5}",
		OracleText: "When The Mightstone and Weakstone enters, choose one —\n• Draw two cards.\n• Target creature gets -5/-5 until end of turn.\n{T}: Add {C}{C}. This mana can't be spent to cast nonartifact spells.\n(Melds with Urza, Lord Protector.)",
		Legalities: map[string]string{"commander": "legal"},
		AllParts:   urzaMeldParts(),
	}
}

func urzaPlaneswalkerPrint() cards.Card {
	return cards.Card{
		ID: urzaPWPrintID, OracleID: uuid.MustParse("759406d7-44ae-4260-9ef5-3bb2c92f751a"),
		Name: "Urza, Planeswalker", Layout: "meld",
		TypeLine: "Legendary Planeswalker — Urza", Loyalty: "7",
		Colors: []string{"U", "W"}, ColorIdentity: []string{"U", "W"},
		OracleText: "You may activate the loyalty abilities of Urza twice each turn rather than only once.\n+2: Artifact, instant, and sorcery spells you cast this turn cost {2} less to cast. You gain 2 life.",
		Legalities: map[string]string{"commander": "legal"},
		AllParts:   urzaMeldParts(),
	}
}

// TestIndexKeepsOnlyMeldPartsAndAttachesTheBackFace: Put (like Load)
// keeps the meld entries of all_parts, drops the rest, and points each
// meld front at the back face's record whichever arrives first.
func TestIndexKeepsOnlyMeldPartsAndAttachesTheBackFace(t *testing.T) {
	idx := indexWith(urzaLordProtectorPrint(), urzaPlaneswalkerPrint(), mightstonePrint())
	for _, name := range []string{"Urza, Lord Protector", "The Mightstone and Weakstone"} {
		c, ok := idx.FindByName(name)
		if !ok {
			t.Fatalf("%s not indexed", name)
		}
		if len(c.AllParts) != 3 {
			t.Errorf("%s keeps %d parts, want the 3 meld entries", name, len(c.AllParts))
		}
		if c.MeldResult == nil || c.MeldResult.Name != "Urza, Planeswalker" {
			t.Errorf("%s has no back face attached: %+v", name, c.MeldResult)
		}
		if c.IsMeldBackFace() {
			t.Errorf("%s reads as a back face", name)
		}
	}
	pw, _ := idx.FindByName("Urza, Planeswalker")
	if !pw.IsMeldBackFace() || pw.MeldResult != nil {
		t.Errorf("the back face's own record: back=%v result=%+v", pw.IsMeldBackFace(), pw.MeldResult)
	}
}

// TestMeldCardImportsWithItsCombinedBackFace is the import acceptance:
// the front plays as itself and carries the back face's printed
// characteristics for the meld to use.
func TestMeldCardImportsWithItsCombinedBackFace(t *testing.T) {
	idx := indexWith(urzaLordProtectorPrint(), urzaPlaneswalkerPrint(), mightstonePrint())
	list, err := Resolve(idx, "Urza", []Entry{
		{Name: "Urza, Lord Protector", Count: 1, IsCommander: true},
		{Name: "The Mightstone and Weakstone", Count: 1},
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	gs := list.ToGameCards()
	for _, gc := range gs {
		if gc.Layout != game.LayoutMeld || gc.Faces != nil {
			t.Errorf("%s: layout %q faces %d", gc.Name, gc.Layout, len(gc.Faces))
		}
		if !gc.IsMeldCard() {
			t.Fatalf("%s imported without its meld print", gc.Name)
		}
		m := gc.Meld
		if m.ResultOracleID != "759406d7-44ae-4260-9ef5-3bb2c92f751a" || m.ResultScryfallID != urzaPWPrintID.String() {
			t.Errorf("%s: result ids %q %q", gc.Name, m.ResultOracleID, m.ResultScryfallID)
		}
		if m.Result.Name != "Urza, Planeswalker" || m.Result.StartingLoyalty != 7 || m.Result.TypeLine != "Legendary Planeswalker — Urza" {
			t.Errorf("%s: result face %+v", gc.Name, m.Result)
		}
	}
	if gs[0].Name != "Urza, Lord Protector" || gs[0].ManaCost != "{1}{W}{U}" || gs[0].Power != 2 {
		t.Errorf("the front does not play as itself: %+v", gs[0])
	}
	if !game.CanMeld(gs[0], gs[1]) {
		t.Error("the imported pair should meld")
	}
}

// TestMeldCardIsNotWarned: meld left the layout banners with ADR 0145.
func TestMeldCardIsNotWarned(t *testing.T) {
	list := &List{Mainboard: []cards.Card{urzaLordProtectorPrint(), mightstonePrint()}}
	if vs := unsupportedLayoutViolations(list); len(vs) != 0 {
		t.Errorf("meld cards drew a layout warning: %v", vs)
	}
	if vs := meldBackFaceViolations(list); len(vs) != 0 {
		t.Errorf("meld fronts were refused: %v", vs)
	}
}

// TestADeckOfMeldCardsValidates: a legal 100-card deck with Urza, Lord
// Protector in command and the Mightstone in the 99 passes Validate.
func TestADeckOfMeldCardsValidates(t *testing.T) {
	list := &List{Commanders: []cards.Card{urzaLordProtectorPrint()}}
	list.Mainboard = append(list.Mainboard, mightstonePrint())
	for i := 0; i < 38; i++ {
		list.Mainboard = append(list.Mainboard, basicLegal("Spell "+string(rune('A'+i%26))+string(rune('a'+i/26)), "Instant", "W", "U"))
	}
	basic := basicLegal("Island", "Basic Land — Island", "U")
	for i := 0; i < 60; i++ {
		list.Mainboard = append(list.Mainboard, basic)
	}
	if err := Validate(list); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

// TestTheCombinedBackFaceIsNotADeckCard: Urza, Planeswalker is half of
// two oversized card backs (CR 712.4b), so naming it in a deck is a
// fatal violation that says which cards to add instead — in the 99 and
// in the command zone alike.
func TestTheCombinedBackFaceIsNotADeckCard(t *testing.T) {
	list := buildValidDeck()
	list.Mainboard[0] = urzaPlaneswalkerPrint()
	err := Validate(list)
	assertHasViolation(t, err, CodeMeldBackFace)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatal("not a ValidationError")
	}
	for _, v := range ve.Violations {
		if v.Code == CodeMeldBackFace && (!strings.Contains(v.Message, "The Mightstone and Weakstone") || !strings.Contains(v.Message, "Urza, Lord Protector")) {
			t.Errorf("message does not name the two halves: %q", v.Message)
		}
	}

	list = buildValidDeck()
	list.Commanders = []cards.Card{urzaPlaneswalkerPrint()}
	assertHasViolation(t, Validate(list), CodeMeldBackFace)
}
