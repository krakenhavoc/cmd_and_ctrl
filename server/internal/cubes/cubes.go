// Package cubes is a signed-in person's saved cubes (ADR 0222 §8,
// #5441): named lists of cards a draft can deal its packs from in
// place of a set's. It is migration 0039.
//
// A cube is a list. It has no commander and no sideboard: whatever
// section of a pasted list or a deck site's deck a row sat in, it is a
// card of the cube. It may hold one copy of a card (basic lands aside)
// when its owner leaves Singleton on, the default, and any number
// otherwise.
//
// # The check
//
// A cube is checked the way a deck is (Check): every card must be in
// the catalog as Full, and the check names the ones that are not, so
// the owner can drop them. A cube with any card that blocks cannot be
// drafted. The list is kept as text and checked against the catalog
// each time it is read, never stored as a verdict, so a card that is
// built later stops blocking the cube with no edit.
//
// # Size
//
// A draft needs seats x 3 packs x pack size cards (Need). MaxCards is
// the most a cube may hold.
//
// This is personal data (ADR 0149). Account deletion calls
// ForgetAccount inside its own transaction.
package cubes

import (
	"bufio"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
)

// The bounds of a cube.
const (
	// MaxCards is the most cards a cube may hold, copies counted: three
	// times the 360 an eight-seat draft of 15-card packs deals. A list
	// over it is refused when it is saved.
	MaxCards = 1080
	// MaxCubes is the most cubes one person may keep.
	MaxCubes = 20
	// MaxNameLen is the longest name a cube may have, in characters.
	MaxNameLen = 60
)

// Errors a save refuses with.
var (
	ErrNotFound  = errors.New("cubes: not found")
	ErrNameTaken = errors.New("cubes: you already have a cube with that name")
	ErrTooMany   = errors.New("cubes: you have as many cubes as you can keep")
	ErrTooLarge  = fmt.Errorf("cubes: a cube holds at most %d cards", MaxCards)
	ErrEmpty     = errors.New("cubes: that list names no cards")
)

// Cube is one cubes row.
type Cube struct {
	ID      uuid.UUID
	OwnerID uuid.UUID
	Name    string
	// Singleton says the list may hold one copy of a card, basic lands
	// aside.
	Singleton bool
	// List is the card list as text, one "N Name" row a card (Text).
	List      string
	CardCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Need is how many cards a draft of this many seats and this pack size
// deals: three packs a seat (CR 905.1a, draft.Rounds).
func Need(seats, packSize int) int { return seats * draft.Rounds * packSize }

// SeatsFor is the largest draft a cube of n cards can deal at a pack
// size, capped at draft.MaxSeats, and 0 when it cannot deal the
// smallest.
func SeatsFor(n, packSize int) int {
	if packSize <= 0 {
		return 0
	}
	seats := min(n/(draft.Rounds*packSize), draft.MaxSeats)
	if seats < draft.MinSeats {
		return 0
	}
	return seats
}

// ParseList reads a cube list: a pasted deck list in the dialect
// deck.ParseText reads, or the plain list of card names, one a line,
// that cube sites export. A row with no count is one copy.
func ParseList(text string) ([]deck.Entry, error) {
	var b strings.Builder
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			// Blank lines mean nothing in a cube; dropping them keeps
			// the deck reader from taking a short last block for a
			// commander.
			continue
		}
		if !counted(line) && !heading(line) {
			line = "1 " + line
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("cubes: reading the list: %w", err)
	}
	return deck.ParseText(b.String())
}

// counted reports whether a row opens with a count ("2 ", "2x ") or the
// sideboard prefix a deck list uses.
func counted(line string) bool {
	if strings.HasPrefix(line, "SB:") || strings.HasPrefix(line, "SB ") {
		return true
	}
	head, _, ok := strings.Cut(strings.ReplaceAll(line, "\t", " "), " ")
	if !ok {
		return false
	}
	head = strings.TrimSuffix(strings.TrimSuffix(head, "x"), "X")
	_, err := strconv.Atoi(head)
	return err == nil
}

// heading reports whether a row is a comment or a section heading of a
// deck list ("# ...", "// Sideboard", "Mainboard:"), which the deck
// reader handles. No card's name ends in a colon.
func heading(line string) bool {
	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") || strings.HasSuffix(line, ":")
}

// Row is one card of a checked cube.
type Row struct {
	// Name is the card's name as the index spells it, or as the list
	// did for a name the index does not know.
	Name  string
	Count int
	// Card is the printing the cube deals, the zero Card for an unknown
	// name.
	Card  cards.Card
	Known bool
	// Printing is true when the list named this printing; Text then
	// writes it back.
	Printing bool
}

// Report is what a check of a cube found. JSON field names are the
// client's contract; client/src/lib/cubes.ts mirrors this struct.
type Report struct {
	// Cards is how many cards the list holds, copies counted.
	Cards int `json:"cards"`
	// Unknown are the names the card index does not know, NotFull the
	// cards the catalog does not hold as Full, and Repeated the cards a
	// singleton cube holds more than once. All sorted by name. Any of
	// them blocks the cube.
	Unknown  []string `json:"unknown"`
	NotFull  []string `json:"not_full"`
	Repeated []string `json:"repeated"`
	// TooLarge is true for a list over MaxCards.
	TooLarge bool `json:"too_large"`
	// Ready is true when no card blocks the cube and it is within
	// MaxCards. Whether it is large enough is a matter of the draft
	// (Need); MaxSeats says how large a draft of 15-card packs it can
	// deal, 0 for none.
	Ready    bool `json:"ready"`
	MaxSeats int  `json:"max_seats"`

	rows []Row
}

// Rows is the checked list, one row a card and printing, in the order
// the list named them.
func (r Report) Rows() []Row { return r.rows }

// Deal is the cards a draft of this cube hands out: every copy of
// every row. It is only meaningful for a Ready report.
func (r Report) Deal() []cards.Card {
	out := make([]cards.Card, 0, r.Cards)
	for _, row := range r.rows {
		for range row.Count {
			out = append(out, row.Card)
		}
	}
	return out
}

// Text renders the checked list as the text a cube is stored as: one
// "N Name" row a card, with the printing where the list named one. It
// reads back through ParseList to the same rows.
func (r Report) Text() string {
	var b strings.Builder
	for _, row := range r.rows {
		fmt.Fprintf(&b, "%d %s", row.Count, row.Name)
		if row.Known && row.Printing && row.Card.SetCode != "" {
			fmt.Fprintf(&b, " (%s)", strings.ToUpper(row.Card.SetCode))
			if row.Card.CollectorNo != "" {
				b.WriteString(" " + row.Card.CollectorNo)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// IsBasicLand reports whether a card is a basic land ("Basic Land —
// Forest", "Basic Snow Land — Snow-Covered Forest"). A basic land is
// exempt from both checks: the engine plays one with no catalog entry,
// so the catalog registers none (as draftpacks.Pool.Blockers has it),
// and singleton has never meant one Forest.
func IsBasicLand(c cards.Card) bool {
	return strings.HasPrefix(c.TypeLine, "Basic ") && strings.Contains(c.TypeLine, "Land")
}

// Check checks a cube list against the card index and the catalog.
// Every row counts, whichever board of a deck it was listed on. isFull
// is the catalog's verdict by oracle ID (draftpacks.CatalogIsFull).
func Check(idx *cards.Index, isFull func(uuid.UUID) bool, entries []deck.Entry, singleton bool) Report {
	rep := Report{Unknown: []string{}, NotFull: []string{}, Repeated: []string{}}
	type key struct {
		name     string
		printing uuid.UUID
	}
	at := map[key]int{}
	copies := map[string]int{}   // by card name, for singleton
	basic := map[string]bool{}   // names that are basic lands
	blocked := map[string]bool{} // names already reported as not Full
	for _, e := range entries {
		if e.Count <= 0 {
			continue
		}
		rep.Cards += e.Count
		row := Row{Name: strings.TrimSpace(e.Name), Count: e.Count}
		if idx != nil {
			if c, ok := idx.FindByName(e.Name); ok {
				row.Known, row.Name = true, c.Name
				named := e.PrintingID != uuid.Nil || e.SetCode != ""
				if p, found := idx.FindPrintingOf(c, e.PrintingID, e.SetCode, e.CollectorNo); found {
					row.Printing = named
					c = p
				}
				row.Card = c
			}
		}
		k := key{name: strings.ToLower(row.Name)}
		if row.Printing {
			k.printing = row.Card.ID
		}
		if i, ok := at[k]; ok {
			rep.rows[i].Count += row.Count
		} else {
			at[k] = len(rep.rows)
			rep.rows = append(rep.rows, row)
		}
		switch {
		case !row.Known:
			if copies[k.name] == 0 {
				rep.Unknown = append(rep.Unknown, row.Name)
			}
		case IsBasicLand(row.Card):
			basic[k.name] = true
		case isFull == nil || !isFull(row.Card.OracleID):
			if !blocked[k.name] {
				blocked[k.name] = true
				rep.NotFull = append(rep.NotFull, row.Name)
			}
		}
		copies[k.name] += row.Count
	}
	if singleton {
		seen := map[string]bool{}
		for _, row := range rep.rows {
			name := strings.ToLower(row.Name)
			if copies[name] > 1 && !basic[name] && !seen[name] {
				seen[name] = true
				rep.Repeated = append(rep.Repeated, row.Name)
			}
		}
	}
	sort.Strings(rep.Unknown)
	sort.Strings(rep.NotFull)
	sort.Strings(rep.Repeated)
	rep.TooLarge = rep.Cards > MaxCards
	rep.Ready = rep.Cards > 0 && !rep.TooLarge &&
		len(rep.Unknown) == 0 && len(rep.NotFull) == 0 && len(rep.Repeated) == 0
	rep.MaxSeats = SeatsFor(rep.Cards, draft.DefaultCubePackSize)
	return rep
}
