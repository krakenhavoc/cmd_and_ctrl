// Package moderation is the name and chat filter of a public table
// (#2984, ADR 0075 section 2.6): a small hand-kept denylist and the
// normalisation that makes it hard to dodge by typing it differently.
// No external service, no network, no state.
//
// What it does, and the three entry points:
//
//   - Name(s) refuses a display name or a table name that contains a
//     listed word. The caller shows the error to the person who typed it.
//   - Chat(s) masks a listed word in a chat line with asterisks and sends
//     the rest. Masking, not rejecting, so a false positive costs one word
//     rather than a message, and the sender is not told which word to
//     rephrase around.
//   - Key(s) folds a name to the form a block on a guest is keyed on, so
//     "Bob", "b0b" and "B O B" are one guest to a table that blocked Bob.
//
// What it folds before it looks: case; diacritics and a few look-alike
// letters from other alphabets; the usual leetspeak digits (0 1 3 4 5 7)
// and the symbols @ $ ! | + when they sit inside a word; zero-width and
// combining characters; and a stretched letter (niiigger) counts as the
// plain letter. Letters spaced out one or two at a time (n i g g e r,
// n.i.g.g.e.r) are read as the word.
//
// What it does not do, deliberately: it has no notion of meaning, so a
// slur spelled to dodge all of the above, in another language or in
// another script, gets through. The list is in denylist.txt, which owns
// the rules for adding to it, and a test holds every Magic card name
// clear of it (see filter_cards_test.go).
package moderation

import (
	_ "embed"
	"errors"
	"strings"
	"unicode"
)

//go:embed denylist.txt
var denylistSource string

// ErrNameNotAllowed is what Name returns for a refused name. Its text is
// shown to the person, so it says what to do and does not say which word.
var ErrNameNotAllowed = errors.New("that name isn't allowed here; pick another")

// letter is one folded letter of the input, with where it came from.
type letter struct {
	c          byte // 'a'..'z'
	start, end int  // byte range of the source rune in the input
	tok        int  // index of the token it belongs to
}

// token is a run of letters with no separator between them.
type token struct{ first, last int } // indexes into the letters

// run is one letter repeated n times in a row.
type run struct {
	c byte
	n int
}

// diacritics folds an accented Latin letter to its base letter, and the
// Cyrillic and Greek letters that look like a Latin one. Anything not
// listed and not ASCII is treated as a separator.
var diacritics = func() map[rune]byte {
	m := map[rune]byte{}
	add := func(base byte, chars string) {
		for _, r := range chars {
			m[r] = base
		}
	}
	add('a', "àáâãäåāăąǎ"+"аα")
	add('c', "çćĉċč"+"с")
	add('d', "ďđ")
	add('e', "èéêëēĕėęě"+"еε")
	add('g', "ĝğġģ")
	add('h', "ĥħ")
	add('i', "ìíîïĩīĭįıǐ"+"іι")
	add('j', "ĵ")
	add('k', "ķ"+"кκ")
	add('l', "ĺļľŀł")
	add('n', "ñńņňŉ")
	add('o', "òóôõöøōŏőǒ"+"оο")
	add('p', "р"+"ρ")
	add('r', "ŕŗř")
	add('s', "śŝşšß")
	add('t', "ţťŧ"+"тτ")
	add('u', "ùúûüũūŭůűųǔ")
	add('v', "ν")
	add('w', "ŵ")
	add('x', "х")
	add('y', "ýÿŷ"+"у")
	add('z', "źżž")
	return m
}()

var leetDigits = map[rune]byte{'0': 'o', '1': 'i', '3': 'e', '4': 'a', '5': 's', '7': 't'}

// leetSymbols maps a symbol to the letter it stands for. needBoth is true
// for the ones that are ordinary punctuation at the end of a sentence
// ("wow!", "C++"): those map only between two word characters.
var leetSymbols = map[rune]struct {
	c        byte
	needBoth bool
}{
	'@': {'a', false},
	'$': {'s', false},
	'!': {'i', true},
	'|': {'i', true},
	'+': {'t', true},
}

// ignorable runes do not break a word and are not a letter: combining
// marks (the decomposed form of an accent), zero-width characters and the
// soft hyphen, which are all ways to type a word the eye reads and the
// matcher does not.
func ignorable(r rune) bool {
	switch {
	case r >= 0x0300 && r <= 0x036F:
		return true
	case r >= 0x200B && r <= 0x200D, r == 0x2060, r == 0xFEFF, r == 0x00AD:
		return true
	}
	return false
}

// foldLetter returns the plain a-z letter r stands for, if it is one.
func foldLetter(r rune) (byte, bool) {
	if r >= 0xFF21 && r <= 0xFF3A { // fullwidth capitals
		r = r - 0xFF21 + 'A'
	} else if r >= 0xFF41 && r <= 0xFF5A { // fullwidth lower case
		r = r - 0xFF41 + 'a'
	}
	r = unicode.ToLower(r)
	if r >= 'a' && r <= 'z' {
		return byte(r), true
	}
	c, ok := diacritics[r]
	return c, ok
}

func wordChar(r rune) bool {
	if _, ok := foldLetter(r); ok {
		return true
	}
	_, ok := leetDigits[r]
	return ok || unicode.IsDigit(r) || unicode.IsLetter(r)
}

// normalise folds s into letters and tokens.
func normalise(s string) ([]letter, []token) {
	type rn struct {
		r          rune
		start, end int
	}
	var rs []rn
	for i, r := range s {
		rs = append(rs, rn{r: r, start: i, end: i + len(string(r))})
	}
	var letters []letter
	var tokens []token
	open := false // the last token is still being extended
	closeTok := func() { open = false }
	add := func(c byte, x rn) {
		if !open {
			tokens = append(tokens, token{first: len(letters), last: len(letters)})
			open = true
		}
		letters = append(letters, letter{c: c, start: x.start, end: x.end, tok: len(tokens) - 1})
		tokens[len(tokens)-1].last = len(letters) - 1
	}
	for i, x := range rs {
		if ignorable(x.r) {
			continue
		}
		if c, ok := foldLetter(x.r); ok {
			add(c, x)
			continue
		}
		if c, ok := leetDigits[x.r]; ok {
			add(c, x)
			continue
		}
		if sym, ok := leetSymbols[x.r]; ok {
			before := i > 0 && wordChar(rs[i-1].r)
			after := i+1 < len(rs) && wordChar(rs[i+1].r)
			if (sym.needBoth && before && after) || (!sym.needBoth && (before || after)) {
				add(sym.c, x)
				continue
			}
		}
		closeTok()
	}
	return letters, tokens
}

// runsOf run-length encodes letters[from..to].
func runsOf(letters []letter, from, to int) []run {
	var out []run
	for i := from; i <= to; i++ {
		if n := len(out); n > 0 && out[n-1].c == letters[i].c {
			out[n-1].n++
		} else {
			out = append(out, run{c: letters[i].c, n: 1})
		}
	}
	return out
}

// covers reports whether text runs (starting at ts) have the entry's
// letters in order with at least as many of each: a stretched letter is
// still the letter.
func covers(text []run, entry []run) bool {
	if len(text) < len(entry) {
		return false
	}
	for i, e := range entry {
		if text[i].c != e.c || text[i].n < e.n {
			return false
		}
	}
	return true
}

func sameWord(text, entry []run) bool {
	return len(text) == len(entry) && covers(text, entry)
}

// entry is one denylist line.
type entry struct {
	whole bool
	words [][]run // one per word; a substring entry has exactly one
	flat  []run   // all the words run together, for the spaced-out reading
}

var entries = parseDenylist(denylistSource)

func parseDenylist(src string) []entry {
	var out []entry
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		e := entry{}
		if strings.HasPrefix(line, "=") {
			e.whole = true
			line = strings.TrimSpace(line[1:])
		}
		letters, tokens := normalise(line)
		if len(tokens) == 0 || (!e.whole && len(tokens) != 1) {
			panic("moderation: bad denylist line " + line)
		}
		for _, t := range tokens {
			e.words = append(e.words, runsOf(letters, t.first, t.last))
		}
		e.flat = runsOf(letters, 0, len(letters)-1)
		out = append(out, e)
	}
	return out
}

// maxSpacedTokenLetters is the longest word that still counts as one
// piece of a spaced-out word: "n i g g e r" and "ni gg er" are read as
// the word, "bone rope" is not read as "bonerope".
const maxSpacedTokenLetters = 2

// span is a half-open-free inclusive range of letters.
type span struct{ from, to int }

// find returns every listed word the input contains.
func find(s string) []span {
	letters, tokens := normalise(s)
	if len(letters) == 0 {
		return nil
	}
	var hits []span

	// Pass 1: inside a token, and a whole-word entry against whole tokens.
	tokRuns := make([][]run, len(tokens))
	for i, t := range tokens {
		tokRuns[i] = runsOf(letters, t.first, t.last)
	}
	for _, e := range entries {
		if e.whole {
			for t := 0; t+len(e.words) <= len(tokens); t++ {
				ok := true
				for k, w := range e.words {
					if !sameWord(tokRuns[t+k], w) {
						ok = false
						break
					}
				}
				if ok {
					hits = append(hits, span{tokens[t].first, tokens[t+len(e.words)-1].last})
				}
			}
			continue
		}
		for t := range tokens {
			rs := tokRuns[t]
			for i := 0; i+len(e.flat) <= len(rs); i++ {
				if covers(rs[i:], e.flat) {
					hits = append(hits, runSpan(tokens[t].first, rs, i, len(e.flat)))
				}
			}
		}
	}

	// Pass 2: the word spelled out across tokens, over the whole input
	// run together. Only tokens of at most two letters take part.
	all := runsOf(letters, 0, len(letters)-1)
	for _, e := range entries {
		for i := 0; i+len(e.flat) <= len(all); i++ {
			if !covers(all[i:], e.flat) {
				continue
			}
			sp := runSpan(0, all, i, len(e.flat))
			if letters[sp.from].tok == letters[sp.to].tok {
				continue // pass 1 owns a match inside one token
			}
			if spacedOK(letters, tokens, sp, e.whole) {
				hits = append(hits, sp)
			}
		}
	}
	return hits
}

// runSpan is the letter range of runs[i..i+n) when runs were built from
// letters starting at base.
func runSpan(base int, runs []run, i, n int) span {
	from := base
	for k := 0; k < i; k++ {
		from += runs[k].n
	}
	to := from
	for k := i; k < i+n; k++ {
		to += runs[k].n
	}
	return span{from, to - 1}
}

// spacedOK decides whether a match across tokens is a word spelled out.
func spacedOK(letters []letter, tokens []token, sp span, whole bool) bool {
	a, b := letters[sp.from].tok, letters[sp.to].tok
	for t := a; t <= b; t++ {
		if tokens[t].last-tokens[t].first+1 > maxSpacedTokenLetters {
			return false
		}
	}
	if whole {
		return sp.from == tokens[a].first && sp.to == tokens[b].last
	}
	return true
}

// Flagged reports whether s contains a listed word.
func Flagged(s string) bool { return len(find(s)) > 0 }

// Name returns ErrNameNotAllowed when s contains a listed word.
func Name(s string) error {
	if Flagged(s) {
		return ErrNameNotAllowed
	}
	return nil
}

// Chat returns s with every listed word replaced by asterisks, and
// whether it changed anything. Spacing outside the masked words is kept.
func Chat(s string) (string, bool) {
	letters, _ := normalise(s)
	hits := find(s)
	if len(hits) == 0 {
		return s, false
	}
	masked := make([]bool, len(s))
	for _, h := range hits {
		for i := letters[h.from].start; i < letters[h.to].end; i++ {
			masked[i] = true
		}
	}
	var b strings.Builder
	for i, r := range s {
		switch {
		case !masked[i]:
			b.WriteRune(r)
		case ignorable(r):
			// dropped: it carried no visible character
		case unicode.IsSpace(r):
			b.WriteRune(r)
		default:
			b.WriteByte('*')
		}
	}
	return b.String(), true
}

// Key folds a name to the form a table's block on a guest is keyed on:
// the folded letters, run together, with a stretched letter counted once.
// Two spellings a person would call the same name give the same key. A
// name with no letters in it keys on its trimmed, lower-cased self.
func Key(s string) string {
	letters, _ := normalise(s)
	if len(letters) == 0 {
		return strings.ToLower(strings.TrimSpace(s))
	}
	var b strings.Builder
	for i, l := range letters {
		if i > 0 && letters[i-1].c == l.c {
			continue
		}
		b.WriteByte(l.c)
	}
	return b.String()
}
