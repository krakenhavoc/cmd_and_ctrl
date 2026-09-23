// Package coverage measures the live card-effect catalog so that the
// documentation about it cannot quietly stop being true.
//
// # Why this package exists
//
// docs/decklists/card-coverage-roadmap.md is the map of the next 4000
// Commander cards, and every number in it used to be typed by hand by
// whoever last touched the file. Nothing checked them. They drifted
// twice in the first month:
//
//   - A batch author DERIVED a count by adding up what they believed
//     had landed ("319") when the registry actually held 320. That is
//     why the doc now carries the line "Always MEASURE this line,
//     never derive it" — it is a scar, not a style note.
//
//   - The measured number was then reported as a card count when it is
//     a KEY count. Some of the registry's keys are BACK FACES,
//     registered as "<oracle_id>#1" (game.CatalogKey for face 1) —
//     see TestBackFaceSpecsAreKeyedByFace. Conflating the two
//     overstated "how many cards do we have" by roughly 13%.
//
//     Since S32 the back-face keyspace has two tenants and they mean
//     different things for coverage. The sixty MDFC LAND backs have
//     deliberately unregistered fronts, so each is half a card and
//     each is still a gap card in an unstarted batch. The Siege back
//     faces are the other half of a front that IS registered, so
//     those cards are whole. The census cannot tell them apart from
//     the key alone and does not try: it reports the key count and
//     says what the split is in prose.
//
// So the census is computed here, from effects.All(), and a test
// compares it against a delimited generated block in the roadmap. The
// numbers stop being a claim and start being a measurement.
//
// # What lives here and what does not
//
// This package holds the measuring, not the asserting: Census and
// CensusBlock are ordinary functions with no testing dependency, and
// the drift guards that fail the build are in census_test.go and
// caveats_test.go. That split is deliberate — the measurement is
// useful to a human at a prompt ("what does the catalog actually
// hold right now?") and should not require reading a test to find.
//
// Nothing here touches the 629 MB Scryfall bulk dump. The dump is not
// in CI and a check gated on it would skip there, which is the same
// as not existing. Everything below is derived from the registry,
// which is compiled in.
package coverage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// faceSuffix matches the composite registry key a non-front face
// registers under. game.CatalogKeyForFace builds it as
// `oracleID + "#" + strconv.Itoa(face)` for any face other than 0, so
// the pattern is a "#" and one or more digits, anchored at the end.
// Face 0 keeps the bare oracle ID, which is what makes the whole
// scheme a no-op for every single-faced card.
//
// This is the one place the convention is re-stated outside the game
// package. If CatalogKeyForFace ever changes shape,
// TestFaceSuffixMatchesCatalogKey fails here rather than the census
// silently reporting sixty extra "cards".
var faceSuffix = regexp.MustCompile(`#[0-9]+$`)

// Census is the measured state of the catalog at this commit.
//
// Keys is the raw len(effects.All()). It is NOT a card count and is
// never presented as one: WholeCards and BackFaces are the split that
// answers a human question, and they are reported together precisely
// because reporting either alone is how the doc went wrong.
type Census struct {
	// Keys is the number of registered specs — one per registry key.
	Keys int

	// WholeCards is the number of keys that are a bare oracle ID:
	// one key, one whole card.
	WholeCards int

	// BackFaces is the number of keys carrying a "#N" face suffix.
	// Usually that is ONE FACE of a card whose other face is not
	// registered — half a card's worth of coverage, and the card is
	// still a gap card. Since S32 the Sieges are the exception: their
	// fronts are registered too, so those cards are whole and are
	// counted once in WholeCards and once here.
	BackFaces int

	// Full, Caveats and Unreviewed count the Completeness declaration
	// across every key (ADR 0042). They sum to Keys. Unreviewed is
	// reported as its own number and never folded into Full — see
	// effects/completeness.go on why that distinction is the whole
	// point of the field.
	Full       int
	Caveats    int
	Unreviewed int
}

// TakeCensus measures the live registry.
//
// It reads effects.All() — the shipped registry — and nothing else.
// Note that this package is NOT package effects: effects' own test
// binary registers a non-card probe spec ("test-flicker-etb-probe" in
// flicker_test.go), which is why the roadmap's hand-written history
// keeps having to say "the test binary reports one more". Measuring
// from outside that package removes the correction entirely instead
// of documenting it.
func TakeCensus() Census {
	var c Census
	for _, s := range effects.All() {
		c.Keys++
		if faceSuffix.MatchString(s.OracleID) {
			c.BackFaces++
		} else {
			c.WholeCards++
		}
		switch s.Completeness {
		case effects.CompletenessFull:
			c.Full++
		case effects.CompletenessCaveats:
			c.Caveats++
		default:
			c.Unreviewed++
		}
	}
	return c
}

// Generated-block delimiters. The roadmap is a hand-authored
// document with one machine-owned paragraph in it, and these two
// lines are the fence between them. Everything outside is prose a
// person wrote and the tooling must never touch; everything inside is
// rewritten wholesale by `-update`.
//
// The separation is the point. internal/legal's TestTimingAgreement
// learned the lesson the hard way: a regeneration that is allowed to
// rewrite hand-authored claims turns the test into a tautology — it
// stops asserting that the document is true and starts asserting that
// the document is whatever the code says today, which no longer
// catches anything.
//
// BeginMarker's own text is part of what it says, not just where it
// sits: since Discussion #1231 the block is refreshed by CI on every
// push to develop and main, and a PR is not expected — and must not
// try — to regenerate it itself (see census_test.go's
// staleCensusActionFor). `-update` stays the correct manual tool, so
// the marker keeps naming it, but the marker no longer reads as an
// instruction aimed at whoever is looking at a stale block in a PR.
const (
	BeginMarker = "<!-- BEGIN GENERATED CATALOG CENSUS — refreshed by CI on every push to develop and main; do not regenerate in a PR (manual refresh: go test ./internal/cards/coverage/ -update) -->"
	EndMarker   = "<!-- END GENERATED CATALOG CENSUS -->"
)

// RoadmapPath is the document the census is published in, relative to
// this package's directory (which is `go test`'s working directory).
const RoadmapPath = "../../../../docs/decklists/card-coverage-roadmap.md"

// Block renders the census as the markdown that belongs between the
// two markers, delimiters included. Deterministic: same registry,
// same bytes.
func (c Census) Block() string {
	var b strings.Builder
	b.WriteString(BeginMarker + "\n")
	b.WriteString("\n")
	b.WriteString("**The catalog, as measured on this commit.** Not typed by hand and not\n")
	b.WriteString("derived from what a batch believes it landed — `TakeCensus` in\n")
	b.WriteString("`server/internal/cards/coverage` reads `effects.All()`, and\n")
	b.WriteString("`TestRoadmapCensusIsCurrent` fails the build when this block and the\n")
	b.WriteString("registry disagree.\n")
	b.WriteString("\n")
	b.WriteString("| Measured | Count |\n")
	b.WriteString("|---|---:|\n")
	fmt.Fprintf(&b, "| Registry keys (`len(effects.All())`) | **%d** |\n", c.Keys)
	fmt.Fprintf(&b, "| — whole cards (bare `oracle_id`) | **%d** |\n", c.WholeCards)
	fmt.Fprintf(&b, "| — back faces (`<oracle_id>#1`) | %d |\n", c.BackFaces)
	fmt.Fprintf(&b, "| Declared `full` | %d |\n", c.Full)
	fmt.Fprintf(&b, "| Declared `caveats` | %d |\n", c.Caveats)
	fmt.Fprintf(&b, "| Declared `unreviewed` | %d |\n", c.Unreviewed)
	b.WriteString("\n")
	b.WriteString("A back face is usually half a card: the modal-DFC land cycle registers\n")
	b.WriteString("only its sixty land backs, and those cards are still gap cards on their\n")
	b.WriteString("batch issues. The exception is the Sieges, whose fronts are registered\n")
	b.WriteString("too — a Siege is one whole card spread over two keys. Whole cards is\n")
	b.WriteString("still the number to quote when someone asks how many cards the engine\n")
	b.WriteString("automates; it undercounts by the number of Sieges.\n")
	b.WriteString("\n")
	b.WriteString(EndMarker)
	return b.String()
}

// Splice replaces the generated block inside `doc` with `block`,
// returning the new document. ok is false when the markers are
// missing or out of order — the caller must fail loudly rather than
// appending, because a roadmap that lost its markers is a roadmap
// nothing is checking.
func Splice(doc, block string) (string, bool) {
	start := strings.Index(doc, BeginMarker)
	if start < 0 {
		return doc, false
	}
	end := strings.Index(doc, EndMarker)
	if end < start {
		return doc, false
	}
	return doc[:start] + block + doc[end+len(EndMarker):], true
}

// Extract returns the generated block currently in `doc`, delimiters
// included. ok is false under exactly the conditions Splice rejects.
func Extract(doc string) (string, bool) {
	start := strings.Index(doc, BeginMarker)
	if start < 0 {
		return "", false
	}
	end := strings.Index(doc, EndMarker)
	if end < start {
		return "", false
	}
	return doc[start : end+len(EndMarker)], true
}
