package coverage

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// oracle.go — the third drift guard (#1276): a registered ability
// that disagrees with the printed card.
//
// # The failure being hunted
//
// The Wandering Emperor sat in the catalog from S27 to #1208 with
// three loyalty abilities at the wrong costs, one of them invented,
// and four tests asserting the wrong card. caveats.go could not see
// it — it reads the catalog, not the card — and a reviewer has only
// the Label to match against the oracle text, which is exactly the
// string that was wrong.
//
// # What it checks
//
// Three questions, each a string comparison against the card's
// printed oracle text:
//
//  1. COST. Every ActivatedAbility whose Label is "<cost>: <effect>"
//     names a cost some line of the card prints. A keyword label with
//     no colon ("Equip {2}", "Crew 3", "Ninjutsu {U}{B}") must begin
//     some line or keyword of the card.
//  2. EFFECT. Where the label has an effect, most of its words appear
//     in the effect of an oracle line with THAT cost. This is the half
//     that catches the Emperor: her +1, −1 and −2 all exist on the
//     card, so a cost-only check passes a permuted set. The measure is
//     word containment (label words found in the oracle line / label
//     words), because labels abbreviate; a label that shares a cost
//     with a line and less than half its words is a different ability.
//  3. MISSING. Every loyalty line and every non-mana "<cost>: <effect>"
//     line on the face a Spec registers has an ability registered for
//     it — or a caveat that names it, or a pinned row in
//     knownOracleMismatches (oracle_test.go).
//
// It is deliberately NOT "the label equals the oracle line": labels
// abbreviate, spell the card's name where the oracle says "this
// creature", and split a modal line into one ability per mode.
//
// # Scope, stated honestly
//
//   - Mana abilities ("{T}: Add …") are out of scope in both
//     directions. Their labels carry no cost ("Add {C}{C}") and basic
//     lands have no Spec at all, so there is nothing to compare.
//   - The MISSING check reads the face a Spec is registered under
//     (OracleID for face 0, "<id>#1" for face 1). A second face with
//     no Spec of its own is not checked for missing abilities; the
//     COST and EFFECT checks search every face.
//   - Abilities granted in quotes ('Equipped creature has "{T}: …"')
//     are not lines of their own and are not checked.
//
// # Where the text comes from
//
// testdata/oracle_text.json: every catalogued oracle ID's printed
// text, generated from the Scryfall dump. CI has no dump, so the check
// reads the fixture, and TestOracleFixtureIsCurrent (dump-gated, run
// nightly) fails when the fixture and the dump disagree. Regenerate:
//
//	CMDCTRL_SCRYFALL_DUMP=../data/scryfall/default-cards.json \
//	  go test ./internal/cards/coverage/ -run TestOracleFixtureIsCurrent -update-oracle

// OracleFixturePath is the checked-in oracle text, relative to this
// package.
const OracleFixturePath = "testdata/oracle_text.json"

// OracleCard is one card's printed text as the fixture records it: a
// single-faced card's text in Text, a multi-faced card's per face in
// Faces (and Text empty).
type OracleCard struct {
	Name  string       `json:"name"`
	Text  string       `json:"text,omitempty"`
	Faces []OracleFace `json:"faces,omitempty"`
}

// AllFaces returns the card's faces, a single-faced card as one.
func (c OracleCard) AllFaces() []OracleFace {
	if len(c.Faces) > 0 {
		return c.Faces
	}
	return []OracleFace{{Name: c.Name, Text: c.Text}}
}

// OracleFace is one printed face. A single-faced card has one.
type OracleFace struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// LoadOracleFixture reads the fixture, keyed by base oracle ID.
func LoadOracleFixture(path string) (map[string]OracleCard, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]OracleCard{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return out, nil
}

// EncodeOracleFixture renders the fixture one card per line, sorted
// by oracle ID, so a regeneration diff is one line per changed card
// and two branches adding different cards rarely touch the same line.
func EncodeOracleFixture(cards map[string]OracleCard) ([]byte, error) {
	ids := make([]string, 0, len(cards))
	for id := range cards {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var b strings.Builder
	b.WriteString("{\n")
	for i, id := range ids {
		var line strings.Builder
		enc := json.NewEncoder(&line)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(cards[id]); err != nil {
			return nil, err
		}
		fmt.Fprintf(&b, "%q: %s", id, strings.TrimSuffix(line.String(), "\n"))
		if i < len(ids)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("}\n")
	return []byte(b.String()), nil
}

// BaseOracleID strips a Spec key's "#<face>" suffix and returns the
// face index it names (0 when there is none).
func BaseOracleID(key string) (string, int) {
	base, suffix, ok := strings.Cut(key, "#")
	if !ok {
		return key, 0
	}
	n := 0
	if _, err := fmt.Sscanf(suffix, "%d", &n); err != nil {
		return base, 0
	}
	return base, n
}

// OracleFindingKind names which of the three questions failed.
type OracleFindingKind string

const (
	// OracleCostNotPrinted: the label's cost appears on no line.
	OracleCostNotPrinted OracleFindingKind = "cost not printed"
	// OracleEffectMismatch: the cost is printed, the effect is not the
	// one printed at that cost.
	OracleEffectMismatch OracleFindingKind = "effect does not match"
	// OracleAbilityMissing: a printed ability has nothing registered.
	OracleAbilityMissing OracleFindingKind = "printed ability not registered"
	// OracleNoText: the Spec registers abilities and the fixture has no
	// text for it.
	OracleNoText OracleFindingKind = "no oracle text in fixture"
)

// OracleFinding is one disagreement between a Spec and its card.
type OracleFinding struct {
	Kind     OracleFindingKind
	Card     string // Spec.Name
	OracleID string // Spec.OracleID, suffix included
	// Label is the registered ability's Label (cost / effect
	// findings), empty for a missing ability.
	Label string
	// Line is the oracle line the finding is about: the best match for
	// an effect mismatch, the unregistered line for a missing ability.
	Line string
	// Cost is the normalised cost the finding is keyed on.
	Cost string
	// Score is the effect match for OracleEffectMismatch.
	Score float64
}

// Key is the allow-list identity: card name, kind, and the cost (or
// label, for a keyword label) the finding is about. No line numbers
// and no scores, so an unrelated edit never moves a pin.
func (f OracleFinding) Key() string {
	subject := f.Cost
	if subject == "" {
		subject = f.Label
	}
	return f.Card + " | " + string(f.Kind) + " | " + subject
}

// Describe renders the finding for a test failure.
func (f OracleFinding) Describe() string {
	switch f.Kind {
	case OracleCostNotPrinted:
		return fmt.Sprintf("%s: ability %q has a cost the card does not print (%q)", f.Card, f.Label, f.Cost)
	case OracleEffectMismatch:
		return fmt.Sprintf("%s: ability %q is not the ability printed at %q (%.0f%% of its words; closest line %q)",
			f.Card, f.Label, f.Cost, 100*f.Score, f.Line)
	case OracleAbilityMissing:
		return fmt.Sprintf("%s: the card prints %q and nothing is registered for it", f.Card, f.Line)
	case OracleNoText:
		return fmt.Sprintf("%s (%s): registers activated abilities but %s has no text for it", f.Card, f.OracleID, OracleFixturePath)
	}
	return f.Card + ": " + string(f.Kind)
}

// effectMatchFloor is the word-containment score below which a label
// and the oracle line at its cost are different abilities. Measured on
// the catalog when the check landed: every correct label that shares
// a cost with its line scores at least 0.8, and the Emperor's three
// permuted labels score 0.14, 0.33 and 0.20.
const effectMatchFloor = 0.5

// OracleReport is the check's answer: what disagrees, and how much was
// actually compared — the counts are what TestOracleCheckIsNotVacuous
// reads, so a normaliser that silently stops matching anything fails
// loudly instead of passing on nothing.
type OracleReport struct {
	Findings []OracleFinding
	// Abilities is how many registered activated abilities were looked
	// up in the fixture.
	Abilities int
	// EffectsCompared is how many of them had an effect worth scoring
	// (a colon label whose cost is printed).
	EffectsCompared int
	// PrintedLines is how many printed cost lines the MISSING check
	// examined.
	PrintedLines int
}

// CheckAbilitiesAgainstOracle runs all three questions over specs.
// A Spec with no fixture entry is reported only when it registers an
// activated ability, the one case where the check has something to
// say and cannot say it.
func CheckAbilitiesAgainstOracle(specs []effects.Spec, fixture map[string]OracleCard) OracleReport {
	var r OracleReport
	for _, s := range specs {
		base, face := BaseOracleID(s.OracleID)
		card, ok := fixture[base]
		if !ok {
			if len(s.Activated) > 0 {
				r.Findings = append(r.Findings, OracleFinding{Kind: OracleNoText, Card: s.Name, OracleID: s.OracleID})
			}
			continue
		}
		checkSpec(&r, s, card, face)
	}
	sort.Slice(r.Findings, func(i, j int) bool { return r.Findings[i].Key() < r.Findings[j].Key() })
	return r
}

func checkSpec(r *OracleReport, s effects.Spec, card OracleCard, face int) {
	names := cardNames(s.Name, card)
	faces := card.AllFaces()
	var all []oracleLine
	for _, f := range faces {
		all = append(all, abilityLines(f.Text, names)...)
	}
	add := func(f OracleFinding) {
		f.Card, f.OracleID = s.Name, s.OracleID
		r.Findings = append(r.Findings, f)
	}

	var registered []string // normalised cost keys of every label
	for _, a := range s.Activated {
		r.Abilities++
		label := stripReminder(a.Label)
		cost, effect, hasColon := splitCost(label)
		if !hasColon {
			kw := normalize(label, names)
			registered = append(registered, kw)
			if !keywordPrinted(kw, all) {
				add(OracleFinding{Kind: OracleCostNotPrinted, Label: a.Label, Cost: kw})
			}
			continue
		}
		keys := costKeys(cost, names)
		registered = append(registered, keys...)
		var atCost []oracleLine
		for _, l := range all {
			if l.hasCost && intersects(keys, l.keys) {
				atCost = append(atCost, l)
			}
		}
		if len(atCost) == 0 {
			add(OracleFinding{Kind: OracleCostNotPrinted, Label: a.Label, Cost: keys[0]})
			continue
		}
		lw := contentWords(normalize(effect, names))
		if len(lw) == 0 {
			continue
		}
		r.EffectsCompared++
		best, bestLine := -1.0, ""
		for _, l := range atCost {
			if sc := containment(lw, l.words); sc > best {
				best, bestLine = sc, l.raw
			}
		}
		if best < effectMatchFloor {
			add(OracleFinding{Kind: OracleEffectMismatch, Label: a.Label, Line: bestLine, Cost: keys[0], Score: best})
		}
	}

	if face >= len(faces) {
		return
	}
	for _, l := range abilityLines(faces[face].Text, names) {
		if !l.hasCost || !l.checkable {
			continue
		}
		r.PrintedLines++
		if intersects(l.keys, registered) || caveatNames(s.Caveats, l, names) {
			continue
		}
		add(OracleFinding{Kind: OracleAbilityMissing, Line: l.raw, Cost: l.keys[0]})
	}
}

// oracleLine is one line of printed text, reminder text removed.
type oracleLine struct {
	raw     string   // the printed line, reminder text removed
	norm    string   // normalised whole line
	hasCost bool     // it has an unquoted "<prefix>:"
	keys    []string // normalised cost keys (see costKeys)
	words   map[string]bool
	// checkable: the prefix is a cost (loyalty or activated) and the
	// line is not a mana ability — the MISSING check's domain.
	checkable bool
}

// abilityLines splits a face's text. A modal line ("…: Choose one —")
// absorbs the "•" lines under it, so a label written per mode is
// compared against the mode's words.
func abilityLines(text string, names []string) []oracleLine {
	var out []oracleLine
	for _, rawLine := range strings.Split(text, "\n") {
		line := strings.TrimSpace(stripReminder(rawLine))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "•") && len(out) > 0 && out[len(out)-1].hasCost {
			prev := &out[len(out)-1]
			for w := range contentWords(normalize(strings.TrimPrefix(line, "•"), names)) {
				prev.words[w] = true
			}
			continue
		}
		l := oracleLine{raw: line, norm: normalize(line, names)}
		if cost, effect, ok := splitCost(line); ok {
			l.hasCost = true
			l.keys = costKeys(cost, names)
			l.words = contentWords(normalize(effect, names))
			l.checkable = isCostPrefix(cost) && !isManaEffect(effect)
		}
		out = append(out, l)
	}
	return out
}

var reminderRe = regexp.MustCompile(`\s*\([^()]*\)`)

// stripReminder removes parenthesised reminder text, innermost first.
func stripReminder(s string) string {
	for {
		t := reminderRe.ReplaceAllString(s, "")
		if t == s {
			return s
		}
		s = t
	}
}

// splitCost splits at the first colon outside quotation marks.
func splitCost(s string) (cost, effect string, ok bool) {
	quoted := false
	for i, r := range s {
		switch r {
		case '"', '“', '”':
			quoted = !quoted
		case ':':
			if !quoted {
				return s[:i], s[i+1:], true
			}
		}
	}
	return "", "", false
}

var (
	abilityWordRe = regexp.MustCompile(`^[^—:]+ — `)    // "Threshold — ", "Exhaust — "
	stationRe     = regexp.MustCompile(`^\d+\+ \| `)    // "12+ | " (station, CR 702.184)
	bracketLoyRe  = regexp.MustCompile(`^\[([^\]]*)\]`) // "[+1]" / "[0]" as some labels write it
)

// costKeys returns the normalised cost plus its forms without a
// leading ability word or station threshold, so "Exhaust — {4}" and
// "{4}" name the same cost. The full form is always keys[0].
func costKeys(cost string, names []string) []string {
	cost = strings.TrimSpace(cost)
	cost = bracketLoyRe.ReplaceAllString(cost, "$1")
	keys := []string{normalize(cost, names)}
	for _, re := range []*regexp.Regexp{abilityWordRe, stationRe} {
		if t := re.ReplaceAllString(cost, ""); t != cost {
			keys = append(keys, normalize(t, names))
		}
	}
	return keys
}

var (
	loyaltyCostRe = regexp.MustCompile(`^[+-]?(\d+|x)$`)
	costPartRe    = regexp.MustCompile(`^((\{[^}]+\})+|(sacrifice|discard|exile|pay|remove|tap|untap|return|put|reveal|mill|collect|forage|waterbend)\b.*)$`)
	manaEffectRe  = regexp.MustCompile(`(?i)\badds?\b|\bunspent mana\b`)
)

// isCostPrefix reports whether a colon prefix is an activation cost:
// a loyalty cost, or a comma-separated list every part of which is a
// cost component. Anything else with a colon ("Chapter", a Class
// level, prose) is not an ability line.
func isCostPrefix(cost string) bool {
	c := normalize(bracketLoyRe.ReplaceAllString(strings.TrimSpace(cost), "$1"), nil)
	c = stationRe.ReplaceAllString(abilityWordRe.ReplaceAllString(c, ""), "")
	if loyaltyCostRe.MatchString(c) {
		return true
	}
	for _, part := range strings.Split(c, ", ") {
		if !costPartRe.MatchString(strings.TrimSpace(part)) {
			return false
		}
	}
	return c != ""
}

func isManaEffect(effect string) bool { return manaEffectRe.MatchString(effect) }

var selfRefRe = regexp.MustCompile(`(?i)\bthis (creature|artifact|land|enchantment|permanent|vehicle|equipment|card|planeswalker|spacecraft|aura|saga|battle|spell)\b`)

// cardNames lists every way the card's text names itself: the Spec
// name, each face, and a legendary's short name ("Teferi" for
// "Teferi, Temporal Pilgrim"). Longest first, so the short name never
// eats part of the full one.
func cardNames(specName string, card OracleCard) []string {
	seen := map[string]bool{}
	var out []string
	add := func(n string) {
		for _, part := range strings.Split(n, " // ") {
			for _, cand := range []string{part, strings.SplitN(part, ",", 2)[0]} {
				cand = strings.TrimSpace(cand)
				if cand != "" && !seen[strings.ToLower(cand)] {
					seen[strings.ToLower(cand)] = true
					out = append(out, cand)
				}
			}
		}
	}
	add(specName)
	add(card.Name)
	for _, f := range card.AllFaces() {
		add(f.Name)
	}
	sort.SliceStable(out, func(i, j int) bool { return len(out[i]) > len(out[j]) })
	return out
}

// normalize is the one comparison form for costs, lines and labels:
// reminder text gone, U+2212 minus and en dash as "-", the card's own
// name and "this creature"-style self-references as "~", loyalty
// brackets dropped, whitespace collapsed, lower case, no final period.
func normalize(s string, names []string) string {
	s = stripReminder(s)
	s = strings.NewReplacer("−", "-", "–", "-", "’", "'").Replace(s)
	for _, n := range names {
		s = replaceWord(s, n, "~")
	}
	s = selfRefRe.ReplaceAllString(s, "~")
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.TrimSuffix(s, ".")
}

// replaceWord replaces name case-insensitively where it is not part of
// a longer word ("Ao" must not rewrite "Aura").
func replaceWord(s, name, with string) string {
	lower, ln := strings.ToLower(s), strings.ToLower(name)
	var b strings.Builder
	i := 0
	for {
		j := strings.Index(lower[i:], ln)
		if j < 0 {
			b.WriteString(s[i:])
			return b.String()
		}
		j += i
		end := j + len(ln)
		if isWordByteBefore(s, j) || isWordByteAt(s, end) {
			b.WriteString(s[i : j+1])
			i = j + 1
			continue
		}
		b.WriteString(s[i:j])
		b.WriteString(with)
		i = end
	}
}

func isWordByteBefore(s string, i int) bool {
	if i == 0 {
		return false
	}
	r := []rune(s[:i])
	return isWordRune(r[len(r)-1])
}

func isWordByteAt(s string, i int) bool {
	if i >= len(s) {
		return false
	}
	return isWordRune([]rune(s[i:])[0])
}

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// keywordPrinted: a colon-less label ("crew 3") begins some line or
// some comma-separated keyword on a line, ending at a word boundary
// so "crew 1" does not match "crew 10".
func keywordPrinted(kw string, lines []oracleLine) bool {
	for _, l := range lines {
		for _, seg := range append([]string{l.norm}, strings.Split(l.norm, ", ")...) {
			seg = strings.TrimSpace(seg)
			if strings.HasPrefix(seg, kw) && !isWordByteAt(seg, len(kw)) {
				return true
			}
		}
	}
	return false
}

// caveatNames: a caveat excuses a missing line when it quotes the
// line's cost, or — for a loyalty ability — says its number ("The −14
// isn't offered").
func caveatNames(caveats []string, l oracleLine, names []string) bool {
	for _, cv := range caveats {
		c := normalize(cv, names)
		for _, k := range l.keys {
			if strings.Contains(c, k) && (!loyaltyCostRe.MatchString(k) || loyaltyMentioned(c, k)) {
				return true
			}
		}
	}
	return false
}

// loyaltyMentioned finds a loyalty cost as its own token: "-7" in
// "the -7 ultimate", not in "-7/-7".
func loyaltyMentioned(c, k string) bool {
	for i := 0; ; {
		j := strings.Index(c[i:], k)
		if j < 0 {
			return false
		}
		j += i
		end := j + len(k)
		before := j == 0 || c[j-1] == ' ' || c[j-1] == '"' || c[j-1] == '('
		after := end == len(c) || strings.ContainsRune(" ,.\"')", rune(c[end]))
		if before && after {
			return true
		}
		i = j + 1
	}
}

var (
	wordRe    = regexp.MustCompile(`[\pL\pN{}/+\-~']+`)
	stopWords = map[string]bool{
		"a": true, "an": true, "the": true, "of": true, "to": true, "and": true, "or": true,
		"it": true, "its": true, "this": true, "that": true, "on": true, "in": true,
		"for": true, "with": true, "from": true, "your": true, "you": true, "target": true,
	}
)

// contentWords is the bag the EFFECT check compares: lower-case words
// minus a short stop list, with a trailing plural "s" dropped.
func contentWords(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range wordRe.FindAllString(strings.ToLower(s), -1) {
		w = strings.Trim(w, "'")
		if w == "" || stopWords[w] {
			continue
		}
		if len(w) > 3 && strings.HasSuffix(w, "s") {
			w = strings.TrimSuffix(w, "s")
		}
		out[w] = true
	}
	return out
}

// containment is |label ∩ line| / |label|.
func containment(label, line map[string]bool) float64 {
	if len(label) == 0 {
		return 1
	}
	n := 0
	for w := range label {
		if line[w] {
			n++
		}
	}
	return float64(n) / float64(len(label))
}

func intersects(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}
