package deck

// repeated_cumulative_keywords_test.go — #1510, CR 702.108b. A card
// that PRINTS a cumulative keyword twice ("Prowess, prowess" — Thor
// Odinson, Ruric Thar, Biomagus, Cursed Firebreathing Yogurt) used to
// import with one instance: Scryfall's `keywords` array is a SET, and
// the oracle-line scan collapsed the repeat into one map key. Both
// halves are fixed here: the importer now counts a cumulative
// keyword's own repeats off the oracle line (keywordLineCounts,
// printedKeywords), and a non-cumulative repeat still collapses to
// one — the pre-#1510 behaviour, which is still correct for every
// combat keyword and the like.
//
// See internal/game/prowess_test.go for the same fix's engine-side
// tests (the printedCharacteristic merge, on the battlefield).

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ruricTharBiomagus is a Scryfall-shaped record for the #1510 report:
// "Prowess, prowess" and nothing else. No catalog entry — a creature
// whose whole text is a doubled keyword line needs none, the same
// Monastery Swiftspear case #706 already proves for a single prowess.
func ruricTharBiomagus() cards.Card {
	return cards.Card{
		ID:         uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		OracleID:   uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		Name:       "Ruric Thar, Biomagus",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Ogre Shaman",
		ManaCost:   "{3}{R}{G}",
		Power:      "5",
		Toughness:  "5",
		Keywords:   []string{"Prowess"},
		OracleText: "Prowess, prowess",
	}
}

// thorOdinson is the same report's other card: a doubled prowess
// alongside two ordinary, non-cumulative keywords, to prove the
// repeat count is read per keyword rather than applied to the whole
// line.
func thorOdinson() cards.Card {
	return cards.Card{
		ID:         uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		OracleID:   uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		Name:       "Thor Odinson",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — God",
		ManaCost:   "{3}{R}{R}",
		Power:      "6",
		Toughness:  "6",
		Keywords:   []string{"Flying", "Vigilance", "Prowess"},
		OracleText: "Flying, vigilance, prowess, prowess",
	}
}

// doubledFlier is a synthetic record — no printed card reads this way
// — built only to prove a NON-cumulative keyword printed twice on one
// line still collapses to a single instance (CR 702.9 has no
// CR 702.108b-style "each instance" clause, and nothing about a
// second "flying" does anything).
func doubledFlier() cards.Card {
	return cards.Card{
		ID:         uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		OracleID:   uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		Name:       "Doubled Flier",
		Layout:     "normal",
		TypeLine:   "Creature — Bird",
		ManaCost:   "{3}{U}",
		Power:      "2",
		Toughness:  "2",
		Keywords:   []string{"Flying"},
		OracleText: "Flying, flying",
	}
}

func TestRepeatedPrintedCumulativeKeywordImportsAsTwoInstances(t *testing.T) {
	list := &List{Mainboard: []cards.Card{ruricTharBiomagus()}}
	got := list.ToGameCards()
	if len(got) != 1 {
		t.Fatalf("imported %d cards, want 1", len(got))
	}
	c := got[0]
	if !sameSet(c.Keywords, []string{game.KeywordProwess, game.KeywordProwess}) {
		t.Fatalf("keywords = %v, want [prowess prowess]", c.Keywords)
	}
	if n := game.ProwessCount(&c); n != 2 {
		t.Errorf("ProwessCount(imported Ruric Thar, Biomagus) = %d, want 2 (CR 702.108b)", n)
	}
}

func TestRepeatedPrintedCumulativeKeywordAlongsideOrdinaryOnes(t *testing.T) {
	list := &List{Mainboard: []cards.Card{thorOdinson()}}
	got := list.ToGameCards()
	if len(got) != 1 {
		t.Fatalf("imported %d cards, want 1", len(got))
	}
	c := got[0]
	want := []string{"flying", "vigilance", game.KeywordProwess, game.KeywordProwess}
	if !sameSet(c.Keywords, want) {
		t.Fatalf("keywords = %v, want %v (flying and vigilance once each, prowess twice)", c.Keywords, want)
	}
	if n := game.ProwessCount(&c); n != 2 {
		t.Errorf("ProwessCount(imported Thor Odinson) = %d, want 2", n)
	}
	if !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "vigilance") {
		t.Errorf("Thor Odinson lost a non-cumulative printed keyword: %v", c.Keywords)
	}
}

// TestRepeatedNonCumulativeKeywordStillDedupes is the control: a
// doubled keyword line does not, by itself, mean two instances — only
// a CUMULATIVE one (game.KeywordIsCumulative) does. A second "flying"
// is still redundant (CR 702.9 has no per-instance clause).
func TestRepeatedNonCumulativeKeywordStillDedupes(t *testing.T) {
	list := &List{Mainboard: []cards.Card{doubledFlier()}}
	got := list.ToGameCards()
	if len(got) != 1 {
		t.Fatalf("imported %d cards, want 1", len(got))
	}
	c := got[0]
	if !sameSet(c.Keywords, []string{"flying"}) {
		t.Errorf("keywords = %v, want [flying] (a second \"flying\" is not cumulative)", c.Keywords)
	}
}
