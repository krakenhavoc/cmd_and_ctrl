package coverage

import (
	"regexp"
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// devoidLine is a printed devoid keyword line: "Devoid (This card has
// no color.)". Devoid is never printed beside another keyword, so the
// line starts with it.
var devoidLine = regexp.MustCompile(`(?im)^devoid\b`)

// TestEveryCatalogDevoidCardIsColourless is #2152's probe over the
// whole catalog: every registered face whose printed text has devoid
// declares it, and is colourless through the engine even when it is
// built as a fixture — no stamped colours, no imported keywords, a
// cost full of coloured pips. That is the road a card that never went
// through the deck importer takes (a test fixture, a restore point
// written before devoid was a canonical token), and it is the one the
// importer's keyword cannot cover. The reverse holds too: a Spec that
// declares devoid prints it, so no card is colourless by accident.
//
// The printed text is the checked-in oracle fixture, so this runs in
// every PR build. Every devoid card in the real dump is swept on the
// import road by deck's TestRealDumpDevoidCardsImportColourless.
func TestEveryCatalogDevoidCardIsColourless(t *testing.T) {
	fx := loadOracleFixture(t)
	checked := 0
	for _, s := range effects.All() {
		base, face := BaseOracleID(s.OracleID)
		prints := false
		if oc, ok := fx[base]; ok {
			faces := oc.AllFaces()
			prints = face < len(faces) && devoidLine.MatchString(faces[face].Text)
		}
		declares := slices.Contains(s.PrintedKeywords, game.KeywordDevoid)
		switch {
		case prints && !declares:
			t.Errorf("%s prints devoid and does not declare it: add game.KeywordDevoid to its PrintedKeywords "+
				"(CR 702.114a), or a fixture of it is the colour of its cost", s.Name)
		case declares && !prints:
			t.Errorf("%s declares devoid in PrintedKeywords and its printed text has no devoid line", s.Name)
		}
		if !prints {
			continue
		}
		checked++
		if face != 0 {
			// A back face's key is reached through SetFace on a card
			// with faces; the declaration above is what it reads.
			continue
		}
		fixture := game.Card{Name: s.Name, OracleID: base, TypeLine: "Instant", ManaCost: "{W}{U}{B}{R}{G}"}
		if !fixture.IsColorless() || len(fixture.Effective().Colors) != 0 {
			t.Errorf("%s (devoid) as a fixture with coloured pips: colours %v, want none",
				s.Name, fixture.EffectiveColors())
		}
	}
	if checked == 0 {
		t.Fatal("no catalogued card prints devoid: the probe proved nothing (did the fixture move?)")
	}
	t.Logf("%d catalogued devoid faces checked", checked)
}
