// Package oracletext reads a card's printed Oracle text: the one
// comparison form for costs, lines and labels (moved here from
// cards/coverage, whose drift guards still use it), and the map from a
// catalog row to the printed ability it implements, which the stack
// shows instead of the row's engine label (#3830).
//
// It imports the game package only, never cards/effects, so the
// protocol view can read it without depending on the catalog.
package oracletext

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var reminderRe = regexp.MustCompile(`\s*\([^()]*\)`)

// StripReminder removes parenthesised reminder text, innermost first.
func StripReminder(s string) string {
	for {
		t := reminderRe.ReplaceAllString(s, "")
		if t == s {
			return s
		}
		s = t
	}
}

// SplitCost splits at the first colon outside quotation marks.
func SplitCost(s string) (cost, effect string, ok bool) {
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
	// AbilityWordRe matches a leading ability word: "Threshold — ",
	// "Exhaust — ".
	AbilityWordRe = regexp.MustCompile(`^[^—:]+ — `)
	// StationRe matches a station threshold: "12+ | " (CR 702.184).
	StationRe = regexp.MustCompile(`^\d+\+ \| `)
	// BracketLoyRe matches "[+1]" / "[0]", as some labels write a
	// loyalty cost.
	BracketLoyRe = regexp.MustCompile(`^\[([^\]]*)\]`)
	// LoyaltyCostRe matches a normalised loyalty cost.
	LoyaltyCostRe = regexp.MustCompile(`^[+-]?(\d+|x)$`)

	costPartRe   = regexp.MustCompile(`^((\{[^}]+\})+|(sacrifice|discard|exile|pay|remove|tap|untap|return|put|reveal|mill|collect|forage|waterbend)\b.*)$`)
	manaEffectRe = regexp.MustCompile(`(?i)\badds?\b|\bunspent mana\b`)
)

// CostKeys returns the normalised cost plus its forms without a
// leading ability word or station threshold, so "Exhaust — {4}" and
// "{4}" name the same cost. The full form is always keys[0].
func CostKeys(cost string, names []string) []string {
	cost = strings.TrimSpace(cost)
	cost = BracketLoyRe.ReplaceAllString(cost, "$1")
	keys := []string{Normalize(cost, names)}
	for _, re := range []*regexp.Regexp{AbilityWordRe, StationRe} {
		if t := re.ReplaceAllString(cost, ""); t != cost {
			keys = append(keys, Normalize(t, names))
		}
	}
	return keys
}

// IsCostPrefix reports whether a colon prefix is an activation cost:
// a loyalty cost, or a comma-separated list every part of which is a
// cost component. Anything else with a colon ("Chapter", a Class
// level, prose) is not an ability line.
func IsCostPrefix(cost string) bool {
	c := Normalize(BracketLoyRe.ReplaceAllString(strings.TrimSpace(cost), "$1"), nil)
	c = StationRe.ReplaceAllString(AbilityWordRe.ReplaceAllString(c, ""), "")
	if LoyaltyCostRe.MatchString(c) {
		return true
	}
	for _, part := range strings.Split(c, ", ") {
		if !costPartRe.MatchString(strings.TrimSpace(part)) {
			return false
		}
	}
	return c != ""
}

// IsManaEffect reports whether an ability's effect adds mana.
func IsManaEffect(effect string) bool { return manaEffectRe.MatchString(effect) }

var selfRefRe = regexp.MustCompile(`(?i)\bthis (creature|artifact|land|enchantment|permanent|vehicle|equipment|card|planeswalker|spacecraft|aura|saga|battle|spell)\b`)

// CardNames lists every way a card's text names itself, from the names
// given (the Spec name, the card name, each face's): each " // " half,
// and a legendary's short name ("Teferi" for "Teferi, Temporal
// Pilgrim"). Longest first, so the short name never eats part of the
// full one.
func CardNames(names ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range names {
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
	sort.SliceStable(out, func(i, j int) bool { return len(out[i]) > len(out[j]) })
	return out
}

// Normalize is the one comparison form for costs, lines and labels:
// reminder text gone, U+2212 minus and en dash as "-", the card's own
// name and "this creature"-style self-references as "~", loyalty
// brackets dropped, whitespace collapsed, lower case, no final period.
func Normalize(s string, names []string) string {
	s = StripReminder(s)
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
		if isWordByteBefore(s, j) || IsWordByteAt(s, end) {
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

// IsWordByteAt reports whether a letter or digit starts at byte i of s.
func IsWordByteAt(s string, i int) bool {
	if i >= len(s) {
		return false
	}
	return isWordRune([]rune(s[i:])[0])
}

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

var (
	wordRe    = regexp.MustCompile(`[\pL\pN{}/+\-~']+`)
	stopWords = map[string]bool{
		"a": true, "an": true, "the": true, "of": true, "to": true, "and": true, "or": true,
		"it": true, "its": true, "this": true, "that": true, "on": true, "in": true,
		"for": true, "with": true, "from": true, "your": true, "you": true, "target": true,
	}
)

// ContentWords is the bag an effect comparison reads: lower-case words
// minus a short stop list, with a trailing plural "s" dropped.
func ContentWords(s string) map[string]bool {
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

// Containment is |label ∩ line| / |label|.
func Containment(label, line map[string]bool) float64 {
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

// Intersects reports whether a and b share an element.
func Intersects(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}
