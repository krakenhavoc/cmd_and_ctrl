package deck

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
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
//	Companion:                 // or "// Companion" (ADR 0172 §1)
//	Mainboard:                 // or "Deck:"
//	Sideboard:
//
// Anything that doesn't parse as a quantity-prefixed row inside a
// section is ignored — the parser is deliberately forgiving so
// pasted lists with stray blank lines or separators still go
// through.
//
// Moxfield's text export (and MTGO's Commander export) names no
// commander at all: it ends with the commander rows after a blank
// line, with no header, below the mainboard and any `SIDEBOARD:`
// block. So when no row anywhere is marked commander, a final block of
// one or two rows that has no header of its own is read as the
// commanders (trailingCommanderBlock). A list like that used to parse
// with no commander and fail `missing_commander`, so the rule changes
// nothing that was accepted before. ADR 0172 records the export shape.
func ParseText(src string) ([]Entry, error) {
	var entries []Entry
	// block numbers each blank-line-separated run of lines; headed
	// says whether that run opened with a section header.
	var blocks []int
	headed := map[int]bool{}
	block, blankSeen := 0, false
	section := sectionMain
	sc := bufio.NewScanner(strings.NewReader(src))
	// Bump the buffer: 99-card decklists fit comfortably, but pasting
	// a 1500-card cube with flavor text shouldn't crash the parser.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			blankSeen = true
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if blankSeen {
			block++
			blankSeen = false
		}
		if s, ok := parseSectionHeader(line); ok {
			section = s
			headed[block] = true
			continue
		}
		if entry, ok, err := parseLine(line, section); err != nil {
			return nil, fmt.Errorf("deck: line %d: %w", lineNo, err)
		} else if ok {
			entries = append(entries, entry)
			blocks = append(blocks, block)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("deck: scan: %w", err)
	}
	trailingCommanderBlock(entries, blocks, headed)
	return entries, nil
}

// trailingCommanderBlock marks the last block's rows as commanders when
// the list names no commander, the list has more than one block, and
// the last block is one or two rows with no section header of its own.
// See ParseText for the export it reads.
func trailingCommanderBlock(entries []Entry, blocks []int, headed map[int]bool) {
	if len(entries) == 0 {
		return
	}
	for _, e := range entries {
		if e.IsCommander {
			return
		}
	}
	last := blocks[len(blocks)-1]
	if last == blocks[0] || headed[last] {
		return
	}
	first := len(entries)
	for first > 0 && blocks[first-1] == last {
		first--
	}
	if n := len(entries) - first; n < 1 || n > 2 {
		return
	}
	for i := first; i < len(entries); i++ {
		entries[i].IsCommander = true
		entries[i].IsSideboard = false
		entries[i].IsCompanion = false
	}
}

type section int

const (
	sectionMain section = iota
	sectionCommander
	sectionSideboard
	sectionCompanion
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
	case "companion", "companions":
		return sectionCompanion, true
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
// A trailing "(SET) number" names the printing (ADR 0192 §2): Resolve
// shows that printing's art when the index has it and it is the same
// card. The name alone still identifies the card, and a suffix that
// names nothing the index holds changes nothing.
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

	// Split off the set / collector suffix: "Sol Ring (C21) 243" →
	// "Sol Ring" plus set "C21" and number "243".
	var setCode, collectorNo string
	if idx := strings.Index(name, " ("); idx >= 0 {
		setCode, collectorNo = parsePrintingSuffix(name[idx+1:])
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
		IsCompanion: sec == sectionCompanion && !isCommander && !isSideboard,
		SetCode:     setCode,
		CollectorNo: collectorNo,
	}, true, nil
}

// parsePrintingSuffix reads "(SNC) 213" or "(SNC) 213 *F*" off the end
// of a row, the way Moxfield, Archidekt and MTGO write a printing. It
// returns empty strings for anything else, and an empty number when
// the row names a set only (or carries a foil marker where the number
// would be).
func parsePrintingSuffix(s string) (set, number string) {
	rest, ok := strings.CutPrefix(strings.TrimSpace(s), "(")
	if !ok {
		return "", ""
	}
	set, rest, ok = strings.Cut(rest, ")")
	set = strings.TrimSpace(set)
	if !ok || set == "" || strings.ContainsAny(set, " \t") {
		return "", ""
	}
	fields := strings.Fields(rest)
	if len(fields) > 0 && !strings.HasPrefix(fields[0], "*") && !strings.HasPrefix(fields[0], "[") {
		number = fields[0]
	}
	return set, number
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

// Text renders a resolved list as the plain-text dialect ParseText
// reads, so a deck fetched from a link can be stored as a list (ADR
// 0110 owner decision 7) and re-seated without the network. Copies are
// grouped by card name in first-seen order.
func (l *List) Text() string {
	var b strings.Builder
	section := func(header string, cs []cards.Card) {
		if len(cs) == 0 {
			return
		}
		// Copies group by printing, not by name, so a list that holds
		// two printings of one card (basic lands, mostly) keeps both.
		// ADR 0192 §2: a printing with a set and number is written
		// "(SET) number", which ParseText reads back.
		type row struct {
			c cards.Card
			n int
		}
		var order []string
		rows := map[string]*row{}
		for _, c := range cs {
			k := c.Name + "\x00" + c.SetCode + "\x00" + c.CollectorNo
			if rows[k] == nil {
				rows[k] = &row{c: c}
				order = append(order, k)
			}
			rows[k].n++
		}
		b.WriteString(header + "\n")
		for _, k := range order {
			r := rows[k]
			if r.c.SetCode != "" && r.c.CollectorNo != "" {
				fmt.Fprintf(&b, "%d %s (%s) %s\n", r.n, r.c.Name, strings.ToUpper(r.c.SetCode), r.c.CollectorNo)
				continue
			}
			fmt.Fprintf(&b, "%d %s\n", r.n, r.c.Name)
		}
	}
	section("Commander:", l.Commanders)
	// ADR 0172 §1: the companion is saved in its own section, so a
	// deck fetched from a link reseats with it. A companion read off a
	// lone sideboard card is written here too, which makes the saved
	// text say what the table will see.
	if l.Companion != nil {
		section("Companion:", []cards.Card{*l.Companion})
	}
	section("Mainboard:", l.Mainboard)
	section("Sideboard:", l.Sideboard)
	return b.String()
}
