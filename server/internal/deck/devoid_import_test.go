package deck

// devoid_import_test.go — the INGESTION half of devoid (#2152, CR
// 702.114a). Scryfall says a devoid card's colours are [] and lists
// "Devoid" in `keywords`; the importer stamps the keyword, and the
// engine reads it so the empty list means colourless rather than
// "not stamped". Hand-built records, keeping only the fields the
// importer reads; TestRealDumpDevoidCardsImportColourless runs the
// same check against every devoid card in the real dump.

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDevoidImportsColourless(t *testing.T) {
	binding := cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.MustParse("ed622e71-f348-4e46-8fb9-05aae430983a"),
		Name:          "Ugin's Binding",
		Layout:        "normal",
		TypeLine:      "Instant",
		ManaCost:      "{2}{U}",
		Colors:        []string{},
		ColorIdentity: []string{"U"},
		Keywords:      []string{"Devoid"},
		OracleText: "Devoid (This card has no color.)\n" +
			"Return target nonland permanent you don't control to its owner's hand.",
	}
	got := toGameCard(binding, false)
	if !slices.Contains(got.Keywords, game.KeywordDevoid) {
		t.Fatalf("keywords = %v, want devoid stamped", got.Keywords)
	}
	if !got.IsColorless() || got.HasColor("U") {
		t.Errorf("imported Ugin's Binding colours = %v, want none", got.EffectiveColors())
	}
	if c := got.Effective().Colors; len(c) != 0 {
		t.Errorf("imported Ugin's Binding Effective().Colors = %v, want none", c)
	}
	// CR 903.4: identity is the mana symbols, which devoid leaves alone.
	if !reflect.DeepEqual(got.ColorIdentity, []string{"U"}) {
		t.Errorf("colour identity = %v, want [U]", got.ColorIdentity)
	}
}

// TestDevoidModalDFCFaceImportsColourless: Drowner of Truth // Drowned
// Jungle, the one devoid double-faced card. Scryfall gives both faces
// `colors: []`, and the face colour fallback used to derive green-blue
// from the front face's hybrid pips — a STAMPED colour list, which the
// engine trusts over the keyword. The face that prints devoid is now
// left unstamped, so the keyword decides.
func TestDevoidModalDFCFaceImportsColourless(t *testing.T) {
	drowner := cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.MustParse("db19a27a-ee22-4931-ae3c-0ce21f456ea6"),
		Name:          "Drowner of Truth // Drowned Jungle",
		Layout:        game.LayoutModalDFC,
		TypeLine:      "Creature — Eldrazi Drone // Land",
		ColorIdentity: []string{"G", "U"},
		Keywords:      []string{"Devoid"},
		CardFaces: []cards.CardFace{
			{
				Name: "Drowner of Truth", TypeLine: "Creature — Eldrazi Drone", ManaCost: "{5}{G/U}{G/U}",
				Colors: []string{}, Power: "7", Toughness: "6",
				OracleText: "Devoid (This card has no color.)\n" +
					"When you cast this spell, if {C} was spent to cast it, create two 0/1 colorless Eldrazi Spawn creature tokens with \"Sacrifice this token: Add {C}.\"",
			},
			{
				Name: "Drowned Jungle", TypeLine: "Land", Colors: []string{},
				OracleText: "This land enters tapped.\n{T}: Add {G} or {U}.",
			},
		},
	}
	got := toGameCard(drowner, false)
	if len(got.Faces) != 2 {
		t.Fatalf("faces = %d, want 2", len(got.Faces))
	}
	if len(got.Faces[0].Colors) != 0 {
		t.Errorf("devoid front face stamped colours %v, want none", got.Faces[0].Colors)
	}
	if !slices.Contains(got.Keywords, game.KeywordDevoid) {
		t.Fatalf("front face keywords = %v, want devoid", got.Keywords)
	}
	if !got.IsColorless() {
		t.Errorf("Drowner of Truth colours = %v, want none", got.EffectiveColors())
	}
	land := got
	land.SetFace(1)
	if slices.Contains(land.Keywords, game.KeywordDevoid) {
		t.Errorf("the land face prints no devoid, got keywords %v", land.Keywords)
	}
	if !land.IsColorless() {
		t.Errorf("Drowned Jungle colours = %v, want none", land.EffectiveColors())
	}
	if !reflect.DeepEqual(got.ColorIdentity, []string{"G", "U"}) {
		t.Errorf("colour identity = %v, want [G U]", got.ColorIdentity)
	}

	// A coloured face with no colours stamped and no devoid still
	// derives its colour from its cost: the fallback adventure and
	// split faces depend on.
	if c := faceColors(cards.CardFace{ManaCost: "{1}{R}", OracleText: "Deal 2 damage."}); !reflect.DeepEqual(c, []string{"R"}) {
		t.Errorf("faceColors(no colours, no devoid) = %v, want [R] from the cost", c)
	}
}

// TestRealDumpDevoidCardsImportColourless is the real-dump half: every
// Commander-legal printing whose Scryfall `keywords` list devoid
// imports colourless, with its colour identity untouched.
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/deck/ -run RealDumpDevoid -v
func TestRealDumpDevoidCardsImportColourless(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	seen := map[uuid.UUID]bool{}
	var coloured []string
	streamDump(t, path, func(c cards.Card) {
		if !slices.ContainsFunc(c.Keywords, func(k string) bool { return k == "Devoid" }) || seen[c.OracleID] {
			return
		}
		seen[c.OracleID] = true
		got := toGameCard(c, false)
		if !got.IsColorless() || len(got.Effective().Colors) != 0 {
			coloured = append(coloured, got.Name)
		}
		if !reflect.DeepEqual(got.ColorIdentity, c.ColorIdentity) {
			t.Errorf("%s: identity %v, want Scryfall's %v", got.Name, got.ColorIdentity, c.ColorIdentity)
		}
	})
	t.Logf("%d devoid oracle IDs", len(seen))
	if len(seen) == 0 {
		t.Fatal("no devoid card in the dump: the sweep proved nothing")
	}
	if len(coloured) > 0 {
		t.Errorf("%d devoid cards import coloured: %v", len(coloured), coloured)
	}
}
