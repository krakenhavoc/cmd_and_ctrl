package effects

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_realdump_manual_test.go — #1127, and the guard ADR 0078
// decision 8 asks for. Opt-in against the real Scryfall bulk dump,
// gated on CMDCTRL_SCRYFALL_DUMP so ordinary CI never touches it (the
// file is ~630MB and is not in the repo):
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/cards/effects/ -run RealDumpEveryTokenTemplate -v
//
// e2e-nightly.yml has the dump and runs it there.
//
// WHY A TOKEN TEMPLATE IS CHECKED AGAINST A PRINTED TOKEN CARD. A
// template's Name, type line, P/T and Colors are the token's
// CHARACTERISTICS, and the printed token card is the published record
// of what they are. #1127 is what happens without this test: 21 rows
// declared no colour where every printing of that token is coloured,
// which is a rules bug — a colourless Soldier dodges protection from
// white, is missed by a white lord, survives "destroy all nonwhite
// creatures" and counts wrongly anywhere `Colors` is read — and it sat
// in the table until somebody matched the whole table against the dump
// by hand.
//
// It is also the precondition for ADR 0078's art: a template that
// matches no printing can be given no picture, so this list and that
// one are the same list.
//
// The match is ADR 0078 decision 3's identity rule, minus the art
// half: pool filters (3a) then name / type line / P/T / colours (3b).
// Keywords are decision 3c's soft preference and are deliberately NOT
// compared — the template's Keywords are what the ENGINE grants and
// Scryfall's are what that piece of card prints, and the two disagree
// harmlessly (a Wurmcoil half carries deathtouch; the shared base row
// carries nothing). When #1115 builds the resolver, this test should
// call it instead of re-implementing the rule here.

// knownUnmatchedTokens are the templates that deliberately match no
// printed token card, each with the reason. The test fails on a set
// difference in EITHER direction: a template that stops matching is a
// regression, and a template on this list that starts matching is a
// stale exemption to delete.
var knownUnmatchedTokens = map[string]string{
	// Hallowed Haunting's token prints as */* — "this token's power
	// and toughness are each equal to the number of Spirits you
	// control". The engine models that as a 0/0 the Haunting sizes
	// with a layer 7a set (b29SpiritClericSizing), and the 0/0 is
	// deliberately NOT PrintedPTKnown, which is what keeps a shrunken
	// Cleric out of the CR 704.5f state-based action. Decision
	// (#1127): keep the 0/0. ADR 0078's rule refuses to equate `*`
	// with a number on purpose — a characteristic-defining token is
	// not the same object as a printed 0/0 — and bending it here
	// would either claim a printed 0/0 or need a `*` the template
	// table has no slot for.
	"0/0 white Spirit Cleric": "printed as */*; the engine's 0/0 is the stand-in the Haunting sizes",

	// Spawning Pit is a 2004 artifact and its token — "create a 2/2
	// colorless Spawn artifact creature token" — was never printed as
	// a token card; token cards start in 2007. The template is right
	// and the dump simply has no such record. (The dump's "Spawn" is
	// a 3/3 red one from a different card, and "Eldrazi Spawn" is its
	// own token this table already carries.)
	"2/2 colorless Spawn artifact": "no token card was ever printed for Spawning Pit's Spawn",

	// Wurmcoil Engine's current oracle says "3/3 colorless Phyrexian
	// Wurm artifact creature token"; the printed token card, from
	// 2010, is named "Wurm" — it predates the 2021 Phyrexian subtype
	// pass. Decision (#1127): keep "Phyrexian Wurm". The engine
	// follows the current oracle because that is what the RULES read
	// — "Phyrexian" and "Wurm" are both subtypes a lord or a removal
	// spell can name — and a token card is a prop, not a rules text.
	// The cost is that this row keeps the text fallback until someone
	// pins an id through ADR 0078 decision 6's override seam.
	"3/3 colorless Phyrexian Wurm artifact": "printed token card says \"Wurm\"; the engine follows the current oracle",
}

// printedToken is the slice of a Scryfall record this match needs.
type printedToken struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Lang        string            `json:"lang"`
	Layout      string            `json:"layout"`
	TypeLine    string            `json:"type_line"`
	Power       string            `json:"power"`
	Toughness   string            `json:"toughness"`
	Colors      []string          `json:"colors"`
	BorderColor string            `json:"border_color"`
	SetType     string            `json:"set_type"`
	ImageURIs   map[string]string `json:"image_uris"`
}

func TestRealDumpEveryTokenTemplateMatchesAPrintedToken(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to match the token table against the real dump")
	}
	pool := loadPrintedTokenPool(t, path)
	t.Logf("eligible printed token records: %d", len(pool))
	if len(pool) < 1000 {
		t.Fatalf("only %d eligible token printings — the pool filter or the dump is wrong", len(pool))
	}
	byName := make(map[string][]printedToken, len(pool))
	for _, r := range pool {
		n := normalizeTokenName(r.Name)
		byName[n] = append(byName[n], r)
	}

	templates := tokenTemplatesUnderTest()
	unmatched := map[string]string{}
	display := map[string]string{}
	for _, label := range templates {
		display[label.key] = label.display
		if matchPrintedToken(byName, label.card) == 0 {
			unmatched[label.key] = describeTemplate(label.card)
		}
	}

	for key, why := range unmatched {
		if _, ok := knownUnmatchedTokens[key]; !ok {
			t.Errorf("%s matches no printed token card (%s).\n"+
				"\tEither the template's characteristics are wrong — check the oracle text of the card "+
				"that makes it — or it is a deliberate exemption and belongs in knownUnmatchedTokens with a reason.",
				display[key], why)
		}
	}
	for key := range knownUnmatchedTokens {
		if _, ok := unmatched[key]; !ok {
			t.Errorf("%q now matches a printed token card: remove it from knownUnmatchedTokens", key)
		}
	}
	t.Logf("%d templates checked, %d unmatched (all expected)", len(templates), len(unmatched))
}

type labelledTemplate struct {
	key     string // the table key, or the constructor name
	display string // how the key reads in a failure message
	card    game.Card
}

// tokenTemplatesUnderTest is every plain row plus the behaviour
// templates that live in tokens.go — the ones with a mana or activated
// ability, which are still printed token cards and still have to be
// right.
func tokenTemplatesUnderTest() []labelledTemplate {
	out := make([]labelledTemplate, 0, len(tokenTable)+12)
	for _, k := range TokenKeys() {
		out = append(out, labelledTemplate{key: k, display: strconv.Quote(k), card: TokenCard(k)})
	}
	for _, bt := range []struct {
		name string
		card game.Card
	}{
		{"TreasureToken()", TreasureToken()},
		{"GoldToken()", GoldToken()},
		{"FoodToken()", FoodToken()},
		{"ClueToken()", ClueToken()},
		{"BloodToken()", BloodToken()},
		{"PowerstoneToken()", PowerstoneToken()},
		{"EldraziSpawnToken()", EldraziSpawnToken()},
		{"BlueShapeshifterToken()", BlueShapeshifterToken()},
		{"ColorlessShapeshifterToken()", ColorlessShapeshifterToken()},
	} {
		out = append(out, labelledTemplate{key: bt.name, display: bt.name, card: bt.card})
	}
	return out
}

// matchPrintedToken counts the pool records whose identity equals the
// template's: name, (card types, subtypes), printed P/T, colours.
func matchPrintedToken(byName map[string][]printedToken, tmpl game.Card) int {
	_, wantTypes, wantSubs := game.ParseTypeLine(tmpl.TypeLine)
	wantP, wantT := "", ""
	if containsString(wantTypes, "Creature") {
		wantP, wantT = strconv.Itoa(tmpl.Power), strconv.Itoa(tmpl.Toughness)
	}
	n := 0
	for _, r := range byName[normalizeTokenName(tmpl.Name)] {
		_, gotTypes, gotSubs := game.ParseTypeLine(r.TypeLine)
		if !sameStringSet(gotTypes, wantTypes) || !sameStringSet(gotSubs, wantSubs) {
			continue
		}
		if r.Power != wantP || r.Toughness != wantT {
			continue
		}
		if !sameStringSet(r.Colors, tmpl.Colors) {
			continue
		}
		n++
	}
	return n
}

// loadPrintedTokenPool streams the dump and keeps the records ADR
// 0078 decision 3a admits: an English, black-bordered `layout: token`
// printing with art, from a set whose type is not memorabilia,
// minigame or funny.
func loadPrintedTokenPool(t *testing.T, path string) []printedToken {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	dec := json.NewDecoder(f)
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("read opening token: %v", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		t.Fatalf("expected a top-level JSON array, got %v", tok)
	}
	var pool []printedToken
	for dec.More() {
		var r printedToken
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("decode record: %v", err)
		}
		switch {
		case r.Layout != "token",
			r.Lang != "en",
			len(r.ImageURIs) == 0,
			r.BorderColor != "black",
			r.SetType == "memorabilia", r.SetType == "minigame", r.SetType == "funny":
			continue
		}
		pool = append(pool, r)
	}
	return pool
}

// normalizeTokenName is index.go's normalizeName: trim, collapse
// internal whitespace, lowercase.
func normalizeTokenName(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as, bs := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(as)
	sort.Strings(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

func describeTemplate(c game.Card) string {
	colors := "colourless"
	if len(c.Colors) > 0 {
		colors = strings.Join(c.Colors, "")
	}
	return fmt.Sprintf("name %q, type line %q, %d/%d, colours %s",
		c.Name, c.TypeLine, c.Power, c.Toughness, colors)
}
