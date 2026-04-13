package deck

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// ParseText parses a plain-text decklist. The dialect is the union
// of what Moxfield, Archidekt, TappedOut, and Scryfall all accept:
//
//	# any comment line
//	1 Sol Ring                 // quantity + name
//	1x Sol Ring                // quantity with x suffix
//	1 Sol Ring (C21) 243       // quantity + name + set + collector no
//	SB: 1 Sol Ring             // sideboard prefix
//	1 Atraxa, Praetors' Voice *CMDR*  // commander tag suffix
//
// Section headers are also recognised:
//
//	Commander:                 // or "Commanders:" or "// Commander"
//	Mainboard:                 // or "Deck:"
//	Sideboard:
//
// Anything that doesn't parse as a quantity-prefixed row inside a
// section is ignored — the parser is deliberately forgiving so
// pasted lists with stray blank lines or separators still go
// through.
func ParseText(src string) ([]Entry, error) {
	var entries []Entry
	section := sectionMain
	sc := bufio.NewScanner(strings.NewReader(src))
	// Bump the buffer: 99-card decklists fit comfortably, but pasting
	// a 1500-card cube with flavor text shouldn't crash the parser.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if s, ok := parseSectionHeader(line); ok {
			section = s
			continue
		}
		if entry, ok, err := parseLine(line, section); err != nil {
			return nil, fmt.Errorf("deck: line %d: %w", lineNo, err)
		} else if ok {
			entries = append(entries, entry)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("deck: scan: %w", err)
	}
	return entries, nil
}

type section int

const (
	sectionMain section = iota
	sectionCommander
	sectionSideboard
)

// parseSectionHeader recognises the various ways Moxfield / Archidekt
// / TappedOut label sections. Returns (section, true) on match; the
// caller then flips to that section for subsequent lines.
func parseSectionHeader(line string) (section, bool) {
	// Strip leading `// ` (some exports use it as a section-comment
	// convention).
	line = strings.TrimPrefix(line, "// ")
	line = strings.TrimSpace(line)
	// Headers end in ':' — strip for comparison.
	lower := strings.ToLower(strings.TrimSuffix(line, ":"))
	switch lower {
	case "commander", "commanders":
		return sectionCommander, true
	case "deck", "mainboard", "main", "main deck", "maindeck":
		return sectionMain, true
	case "sideboard", "side":
		return sectionSideboard, true
	}
	return 0, false
}

// parseLine parses a single row into an Entry. Rows that don't start
// with an integer quantity are skipped (ok=false, err=nil) — those
// are usually blank lines, dividers, or accidentally-copied flavour
// text. The only hard parse error is a malformed quantity (e.g.
// "1 1/2 Sol Ring").
//
// Supported shapes:
//
//	1 Sol Ring
//	1x Sol Ring
//	4 Forest
//	1 Atraxa, Praetors' Voice *CMDR*
//	SB: 1 Sol Ring
//	1 Sol Ring (C21) 243
//
// Set and collector-number suffixes are discarded — resolution by
// name is enough at S05, and set-specific variants don't matter for
// a sandbox.
func parseLine(line string, sec section) (Entry, bool, error) {
	// Sideboard prefix overrides section for this one row.
	isSideboard := sec == sectionSideboard
	if rest, ok := strings.CutPrefix(line, "SB:"); ok {
		isSideboard = true
		line = strings.TrimSpace(rest)
	} else if rest, ok := strings.CutPrefix(line, "SB "); ok {
		isSideboard = true
		line = strings.TrimSpace(rest)
	}

	// Split the quantity off the front. Accept "1", "1x", "1X".
	head, rest, ok := cutFirstSpace(line)
	if !ok {
		return Entry{}, false, nil
	}
	head = strings.TrimSuffix(head, "x")
	head = strings.TrimSuffix(head, "X")
	qty, err := strconv.Atoi(head)
	if err != nil {
		// Not a quantity-prefixed line — skip silently.
		return Entry{}, false, nil
	}
	if qty <= 0 {
		return Entry{}, false, nil
	}

	// Strip "*CMDR*" / "*COMMANDER*" suffix — TappedOut / Archidekt
	// style commander marker.
	name := strings.TrimSpace(rest)
	isCommander := sec == sectionCommander
	if strippedName, hit := stripCommanderMarker(name); hit {
		name = strippedName
		isCommander = true
	}

	// Strip set / collector suffix: "Sol Ring (C21) 243" → "Sol Ring".
	if idx := strings.Index(name, " ("); idx >= 0 {
		name = strings.TrimSpace(name[:idx])
	}

	if name == "" {
		return Entry{}, false, fmt.Errorf("empty card name")
	}
	return Entry{
		Name:        name,
		Count:       qty,
		IsCommander: isCommander,
		IsSideboard: isSideboard && !isCommander,
	}, true, nil
}

// cutFirstSpace splits at the first ASCII space or tab run and
// returns the head, tail, and whether a split happened. Used instead
// of strings.Fields so we preserve the tail exactly — commas,
// apostrophes, etc. — without allocating a slice.
func cutFirstSpace(s string) (head, rest string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			return s[:i], strings.TrimLeft(s[i:], " \t"), true
		}
	}
	return s, "", false
}

// stripCommanderMarker removes the "*CMDR*" / "*COMMANDER*" suffix
// some exports stamp on the commander row. Returns the cleaned name
// and whether the marker was present.
func stripCommanderMarker(name string) (string, bool) {
	trimmed := strings.TrimSpace(name)
	for _, marker := range []string{"*CMDR*", "*Commander*", "*COMMANDER*"} {
		if s, ok := strings.CutSuffix(trimmed, marker); ok {
			return strings.TrimSpace(s), true
		}
	}
	return name, false
}
