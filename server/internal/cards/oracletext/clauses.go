package oracletext

import (
	"regexp"
	"slices"
	"strings"
)

// ClauseKind is what a printed paragraph is, as far as the stack cares.
type ClauseKind int

const (
	// ClauseOther is anything that is not a triggered or activated
	// ability line: a keyword line, a static ability, a spell's effect.
	ClauseOther ClauseKind = iota
	// ClauseTriggered starts with "When", "Whenever" or "At" (after an
	// ability word, if any), carries a reflexive "When you do", or is a
	// Saga chapter (CR 714.2b).
	ClauseTriggered
	// ClauseActivated is "<cost>: <effect>", a loyalty ability included.
	ClauseActivated
)

// Clause is one printed ability of a face: a paragraph of its Oracle
// text, with the "•" mode lines printed under it.
type Clause struct {
	// Text is the paragraph exactly as printed, reminder text and
	// ability word included, its mode lines joined with "\n".
	Text string
	Kind ClauseKind

	norm        string          // normalised paragraph, mode lines excluded
	words       map[string]bool // stemmed content words of the whole clause
	costKeys    []string        // activated: normalised cost keys
	effectWords map[string]bool // activated: stemmed content words after the colon
	reminder    bool            // the paragraph prints reminder text
}

var (
	triggerStartRe = regexp.MustCompile(`(?i)^(when|whenever|at)\b`)
	reflexiveRe    = regexp.MustCompile(`(?i)\bwhen you do\b`)
	chapterRe      = regexp.MustCompile(`^[IVX]+(, [IVX]+)* — `)
)

// isTriggerLine reports whether a line (reminder text removed) is a
// triggered ability.
func isTriggerLine(line string) bool {
	return triggerStartRe.MatchString(line) ||
		triggerStartRe.MatchString(AbilityWordRe.ReplaceAllString(line, "")) ||
		chapterRe.MatchString(line) ||
		reflexiveRe.MatchString(line)
}

// Clauses splits a face's Oracle text into its printed abilities.
func Clauses(text string, names []string) []Clause {
	var out []Clause
	parent := -1 // the paragraph the next "•" line belongs to
	for _, raw := range strings.Split(text, "\n") {
		raw = strings.TrimSpace(raw)
		line := strings.TrimSpace(StripReminder(raw))
		if line == "" {
			continue
		}
		if bullet, ok := strings.CutPrefix(line, "•"); ok && parent >= 0 {
			bullet = strings.TrimSpace(bullet)
			// A Siege's "• Khans — At the beginning of combat …" under
			// "As this enters, choose Khans or Dragons." is a triggered
			// ability of its own, not a mode of an ability.
			if out[parent].Kind == ClauseOther && isTriggerLine(bullet) {
				out = append(out, newClause(strings.TrimSpace(strings.TrimPrefix(raw, "•")), bullet, names))
				continue
			}
			prev := &out[parent]
			prev.Text += "\n" + raw
			for w := range stems(ContentWords(Normalize(bullet, names))) {
				prev.words[w] = true
				if prev.effectWords != nil {
					prev.effectWords[w] = true
				}
			}
			continue
		}
		out = append(out, newClause(raw, line, names))
		parent = len(out) - 1
	}
	return out
}

func newClause(raw, line string, names []string) Clause {
	c := Clause{
		Text:     raw,
		norm:     Normalize(line, names),
		words:    stems(ContentWords(Normalize(line, names))),
		reminder: raw != line,
	}
	// Activated first: "{5}, {T}: Mill two cards. … When you do, …" is
	// an activated ability with a reflexive trigger inside it.
	if cost, effect, ok := SplitCost(line); ok && IsCostPrefix(cost) {
		c.Kind = ClauseActivated
		c.costKeys = CostKeys(cost, names)
		c.effectWords = stems(ContentWords(Normalize(effect, names)))
	} else if isTriggerLine(line) {
		c.Kind = ClauseTriggered
	}
	return c
}

// stems folds the common inflections away, so a label's "attacked"
// and "cycled" find the card's "attacks" and "cycle". Crude, and the
// same on both sides of every comparison, which is all it needs.
func stems(words map[string]bool) map[string]bool {
	out := make(map[string]bool, len(words))
	for w := range words {
		out[stem(w)] = true
	}
	return out
}

func stem(w string) string {
	for _, suf := range []string{"ing", "ed", "es", "s", "e"} {
		if t, ok := strings.CutSuffix(w, suf); ok && len(t) >= 3 {
			return t
		}
	}
	// "died" → "die", the stem "dies" reaches.
	if t, ok := strings.CutSuffix(w, "d"); ok && strings.HasSuffix(t, "e") {
		return stem(t)
	}
	return w
}

// labelWords is the stemmed content words of a label's text. A
// label's parentheses are not reminder text: "gain 3 life (a creature
// died)" says which of two "gain 3 life" triggers it is.
func labelWords(s string, names []string) map[string]bool {
	w := stems(ContentWords(Normalize(strings.NewReplacer("(", " ", ")", " ").Replace(s), names)))
	delete(w, "~")
	return w
}

// matchFloor is the share of a label's words a clause must contain to
// be that label's ability. The coverage drift guard uses the same
// floor for "this label is the ability printed at its cost".
const matchFloor = 0.5

// MatchTriggered answers, for each triggered row label of one
// definition, the printed triggered ability it implements, or "" when
// no single clause is certain. The labels are the rows' Keys, written
// "<card> — <what happens>" (docs/adding-cards.md).
//
// A label that names a keyword the card prints with its reminder text
// ("echo {3}{R}", "cascade", "Suspend — …") is that keyword's line.
// Otherwise its clause is the triggered one containing at least half
// the words of what happens and more of them than any other, or the
// one of such equals the other rows have not claimed (settleTies).
func MatchTriggered(labels []string, clauses []Clause, names []string) []string {
	var cands []Clause
	for _, c := range clauses {
		if c.Kind == ClauseTriggered {
			cands = append(cands, c)
		}
	}
	out, ties := make([]string, len(labels)), make([][]string, len(labels))
	for i, label := range labels {
		head, effect, dashed := strings.Cut(label, " — ")
		if !dashed {
			head, effect = "", label
		}
		kws := []string{Normalize(effect, names), Normalize(head, names)}
		if before, _, ok := strings.Cut(kws[0], ":"); ok {
			kws = append(kws, before)
		}
		if text := keywordClause(kws, clauses, true); text != "" {
			out[i] = text
			continue
		}
		w := labelWords(effect, names)
		if len(w) == 0 {
			// A bare name says nothing about which clause; only a card
			// with one trigger and one row can answer it.
			if len(cands) == 1 && len(labels) == 1 {
				out[i] = cands[0].Text
			}
			continue
		}
		out[i], ties[i] = bestClause(w, cands, func(c Clause) map[string]bool { return c.words })
	}
	settleTies(labels, out, ties)
	return out
}

// MatchActivated is MatchTriggered for activated row labels, which
// are written as the printed line: "<cost>: <effect>", or a keyword
// with its cost ("Equip {2}", "Crew 3"). A label's clause prints the
// same cost and, when the label says an effect, at least half its
// words.
func MatchActivated(labels []string, clauses []Clause, names []string) []string {
	out, ties := make([]string, len(labels)), make([][]string, len(labels))
	for i, label := range labels {
		l := strings.TrimSpace(StripReminder(label))
		cost, effect, ok := SplitCost(l)
		if !ok {
			out[i] = keywordClause([]string{Normalize(l, names)}, clauses, false)
			continue
		}
		keys := CostKeys(cost, names)
		var atCost []Clause
		for _, c := range clauses {
			if c.Kind == ClauseActivated && Intersects(keys, c.costKeys) {
				atCost = append(atCost, c)
			}
		}
		w := labelWords(effect, names)
		if len(w) == 0 {
			if len(atCost) == 1 {
				out[i] = atCost[0].Text
			}
			continue
		}
		out[i], ties[i] = bestClause(w, atCost, func(c Clause) map[string]bool { return c.effectWords })
	}
	settleTies(labels, out, ties)
	return out
}

// settleTies answers a row left between tied clauses when the other
// rows have claimed all of them but one. Archangel of Wrath's "kicked,
// 2 damage to any target" ties its two enters triggers, and the
// "kicked twice" row has the second, so the first row has the first.
// Rows sharing a label never settle each other.
func settleTies(labels, out []string, ties [][]string) {
	seen := map[string]int{}
	for _, l := range labels {
		seen[l]++
	}
	first := slices.Clone(out) // a settled row claims nothing for another
	for i := range labels {
		if out[i] != "" || len(ties[i]) < 2 || seen[labels[i]] != 1 {
			continue
		}
		var left []string
		for _, t := range ties[i] {
			claimed := false
			for j := range labels {
				if j != i && first[j] == t && labels[j] != labels[i] {
					claimed = true
					break
				}
			}
			if !claimed {
				left = append(left, t)
			}
		}
		if len(left) == 1 {
			out[i] = left[0]
		}
	}
}

// bestClause is the clause whose words contain the most of w, at least
// matchFloor of them. When clauses with different text contain as many
// there is no answer, and tied is those clauses, for settleTies.
func bestClause(w map[string]bool, cands []Clause, words func(Clause) map[string]bool) (text string, tied []string) {
	top := -1.0
	for _, c := range cands {
		top = max(top, Containment(w, words(c)))
	}
	if top < matchFloor {
		return "", nil
	}
	for _, c := range cands {
		if Containment(w, words(c)) == top && !slices.Contains(tied, c.Text) {
			tied = append(tied, c.Text)
		}
	}
	if len(tied) == 1 {
		return tied[0], nil
	}
	return "", tied
}

// keywordClause is the one paragraph that begins with one of kws, a
// colon-less keyword label ("equip {2}", "cycling {2}", "echo {3}{r}"),
// after an ability word if it has one; "" when none or several do.
//
// Only a line that is not itself a triggered or activated ability
// answers. For a triggered row the line must also print its reminder
// text, as a keyword whose ability triggers does ("Echo {3}{R} (At the
// beginning of …)"), so a label "draw a card" is never taken for a
// spell face's "Draw a card."
func keywordClause(kws []string, clauses []Clause, triggered bool) string {
	found := ""
	for _, kw := range kws {
		if kw == "" || strings.Contains(kw, "~") {
			continue
		}
		for _, c := range clauses {
			if c.Kind != ClauseOther || (triggered && !c.reminder) || !keywordLineStarts(c.norm, kw) {
				continue
			}
			if found != "" && found != c.Text {
				return ""
			}
			found = c.Text
		}
	}
	return found
}

// keywordLineStarts: the normalised line, or the line after its
// ability word, begins with kw and the keyword ends there.
func keywordLineStarts(norm, kw string) bool {
	for _, s := range []string{norm, AbilityWordRe.ReplaceAllString(norm, "")} {
		if !strings.HasPrefix(s, kw) || IsWordByteAt(s, len(kw)) {
			continue
		}
		if rest := s[len(kw):]; !strings.HasPrefix(rest, ",") {
			return true
		}
	}
	return false
}
