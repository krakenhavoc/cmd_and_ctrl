package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// podSize is the number of drafters at a 17Lands table (Arena's Premier,
// Traditional and Quick Draft all seat eight), so the pack a drafter
// holds at pick k comes back to them at pick k+8.
const podSize = 8

// card is what the two datasets say about one card name.
type card struct {
	// inHand and inHandWins count the games in which the card was in
	// the opening hand or drawn, and how many of them were won.
	inHand, inHandWins int
	// picks is how often a drafter took the card, and pickSum the sum
	// of the pick numbers (1-based) it was taken at.
	picks, pickSum int
	// seen is how many packs a drafter saw it in, and seenSum the sum of
	// the pick numbers (1-based) it was last seen at in each.
	seen, seenSum int
}

// Stats is a set's numbers, built from its game and draft datasets.
type Stats struct {
	cards map[string]*card
	// first is the day of the earliest draft in either dataset,
	// "YYYY-MM-DD", or "" before any row is read.
	first string
	// games and picks count the rows read.
	games, picks int
}

// NewStats is an empty Stats.
func NewStats() *Stats { return &Stats{cards: map[string]*card{}} }

func (s *Stats) card(name string) *card {
	c, ok := s.cards[name]
	if !ok {
		c = &card{}
		s.cards[name] = c
	}
	return c
}

func (s *Stats) noteTime(t string) {
	if len(t) < 10 {
		return
	}
	if day := t[:10]; s.first == "" || day < s.first {
		s.first = day
	}
}

// First is the day of the earliest draft read, "YYYY-MM-DD".
func (s *Stats) First() string { return s.first }

// Open opens a dataset file, gzip-compressed (as 17Lands publishes it)
// or plain CSV, told apart by its first two bytes.
func Open(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReaderSize(f, 1<<20)
	magic, err := br.Peek(2)
	if err != nil && !errors.Is(err, io.EOF) {
		_ = f.Close()
		return nil, err
	}
	if len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		zr, err := gzip.NewReader(br)
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		return readCloser{zr, func() error { _ = zr.Close(); return f.Close() }}, nil
	}
	return readCloser{br, f.Close}, nil
}

type readCloser struct {
	io.Reader
	close func() error
}

func (r readCloser) Close() error { return r.close() }

// header indexes a dataset's header row by column name.
type header map[string]int

func readHeader(r *csv.Reader) (header, error) {
	row, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("header: %w", err)
	}
	h := header{}
	for i, name := range row {
		h[name] = i
	}
	return h, nil
}

func (h header) need(names ...string) error {
	for _, n := range names {
		if _, ok := h[n]; !ok {
			return fmt.Errorf("no %q column", n)
		}
	}
	return nil
}

// withPrefix is every card name with a column "<prefix><name>", and the
// column's index, sorted by name.
func (h header) withPrefix(prefix string) ([]string, map[string]int) {
	idx := map[string]int{}
	var names []string
	for col, i := range h {
		if name, ok := strings.CutPrefix(col, prefix); ok && name != "" {
			idx[name] = i
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, idx
}

func newReader(r io.Reader) *csv.Reader {
	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	return cr
}

// count reads a card-count cell: "0", "1", "2", ... (a float such as
// "1.0" is read too). An empty cell is 0.
func count(v string) int {
	switch v {
	case "", "0":
		return 0
	case "1":
		return 1
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return int(f)
	}
	return 0
}

// truth reads the dataset's "won" cell.
func truth(v string) (bool, error) {
	switch v {
	case "True", "true", "1", "1.0":
		return true, nil
	case "False", "false", "0", "0.0":
		return false, nil
	}
	return false, fmt.Errorf("won is %q", v)
}

// ReadGames reads a game dataset (game_data_public.<SET>.<FORMAT>.csv):
// one row per game, with "won" and, per card, the copies in the opening
// hand ("opening_hand_<name>") and drawn ("drawn_<name>"). A card is in
// hand in a game when either is above 0, which is 17Lands' "games in
// hand" (GIH); tutored copies are not counted, as on 17Lands.
func (s *Stats) ReadGames(r io.Reader) error {
	cr := newReader(r)
	h, err := readHeader(cr)
	if err != nil {
		return err
	}
	if err := h.need("won", "draft_time"); err != nil {
		return err
	}
	names, opening := h.withPrefix("opening_hand_")
	_, drawn := h.withPrefix("drawn_")
	type cols struct {
		c             *card
		opening, draw int
	}
	var cs []cols
	for _, n := range names {
		d, ok := drawn[n]
		if !ok {
			return fmt.Errorf("opening_hand_%s has no drawn_ column", n)
		}
		cs = append(cs, cols{s.card(n), opening[n], d})
	}
	won, at := h["won"], h["draft_time"]
	for line := 2; ; line++ {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		w, err := truth(row[won])
		if err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
		s.games++
		s.noteTime(row[at])
		for _, c := range cs {
			if count(row[c.opening]) > 0 || count(row[c.draw]) > 0 {
				c.c.inHand++
				if w {
					c.c.inHandWins++
				}
			}
		}
	}
}

// pickRow is one pick of a draft: the pack and pick number (0-based, as
// in the dataset), the card taken, and the cards in the pack.
type pickRow struct {
	pack, pick int
	taken      string
	inPack     []int // indices into the draft reader's names
}

// ReadDrafts reads a draft dataset (draft_data_public.<SET>.<FORMAT>.csv):
// one row per pick, with the draft's id, the pack and pick number
// (0-based), the card taken ("pick") and the copies of each card in the
// pack ("pack_card_<name>"). The rows of one draft are consecutive.
//
// Average pick (17Lands' ATA) is the mean pick number, 1-based, a card
// was taken at. Average seen (ALSA) is the mean pick number, 1-based, at
// which a pack the drafter held last showed the card: the pack held at
// pick k comes back at pick k+8, so a card still in it then was last
// seen at k+8, and otherwise at k.
func (s *Stats) ReadDrafts(r io.Reader) error {
	cr := newReader(r)
	h, err := readHeader(cr)
	if err != nil {
		return err
	}
	if err := h.need("draft_id", "draft_time", "pack_number", "pick_number", "pick"); err != nil {
		return err
	}
	names, inPack := h.withPrefix("pack_card_")
	cs := make([]*card, len(names))
	colOf := make([]int, len(names))
	for i, n := range names {
		cs[i] = s.card(n)
		colOf[i] = inPack[n]
	}
	id, at, packCol, pickCol, takenCol := h["draft_id"], h["draft_time"], h["pack_number"], h["pick_number"], h["pick"]
	var draft string
	var rows []pickRow
	for line := 2; ; line++ {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		pack, err1 := strconv.Atoi(row[packCol])
		pick, err2 := strconv.Atoi(row[pickCol])
		if err1 != nil || err2 != nil || pack < 0 || pick < 0 {
			return fmt.Errorf("line %d: pack %q pick %q", line, row[packCol], row[pickCol])
		}
		if row[id] != draft {
			s.flushDraft(rows, cs)
			draft, rows = row[id], rows[:0]
		}
		s.picks++
		s.noteTime(row[at])
		pr := pickRow{pack: pack, pick: pick, taken: row[takenCol]}
		for i, col := range colOf {
			if count(row[col]) > 0 {
				pr.inPack = append(pr.inPack, i)
			}
		}
		rows = append(rows, pr)
	}
	s.flushDraft(rows, cs)
	return nil
}

// flushDraft adds one draft's picks to the stats.
func (s *Stats) flushDraft(rows []pickRow, cs []*card) {
	type key struct{ pack, pick int }
	byPick := make(map[key]*pickRow, len(rows))
	for i := range rows {
		r := &rows[i]
		byPick[key{r.pack, r.pick}] = r
		if r.taken != "" {
			c := s.card(r.taken)
			c.picks++
			c.pickSum += r.pick + 1
		}
	}
	for i := range rows {
		r := &rows[i]
		if r.pick >= podSize {
			continue // a pack already seen at pick-8
		}
		var later map[int]bool
		if back, ok := byPick[key{r.pack, r.pick + podSize}]; ok {
			later = make(map[int]bool, len(back.inPack))
			for _, j := range back.inPack {
				later[j] = true
			}
		}
		for _, j := range r.inPack {
			last := r.pick + 1
			if later[j] {
				last += podSize
			}
			cs[j].seen++
			cs[j].seenSum += last
		}
	}
}

// Row is one card's numbers in the shape of 17Lands' card ratings
// endpoint (GET /card_ratings/data), so scripts/draft-ratings-refresh.sh
// builds a ratings file from it exactly as from a live answer. A rate
// with nothing behind it is null.
type Row struct {
	Name               string   `json:"name"`
	EverDrawnGameCount int      `json:"ever_drawn_game_count"`
	EverDrawnWinRate   *float64 `json:"ever_drawn_win_rate"`
	AvgSeen            *float64 `json:"avg_seen"`
	SeenCount          int      `json:"seen_count"`
	AvgPick            *float64 `json:"avg_pick"`
	PickCount          int      `json:"pick_count"`
}

func ratio(n, d int, places int) *float64 {
	if d == 0 {
		return nil
	}
	p := math.Pow(10, float64(places))
	v := math.Round(float64(n)/float64(d)*p) / p
	return &v
}

// basicLands are left out of the rows, as the live endpoint leaves them
// out: the datasets count them like any card, but a basic is never a
// pick worth rating and the draft shows no rating for one.
var basicLands = map[string]bool{
	"Plains": true, "Island": true, "Swamp": true, "Mountain": true, "Forest": true, "Wastes": true,
	"Snow-Covered Plains": true, "Snow-Covered Island": true, "Snow-Covered Swamp": true,
	"Snow-Covered Mountain": true, "Snow-Covered Forest": true, "Snow-Covered Wastes": true,
}

// Rows are every card's numbers but the basic lands', sorted by name. A
// card name the draft dataset has as a pick but no column for (it
// should not happen) is listed too.
func (s *Stats) Rows() []Row {
	names := make([]string, 0, len(s.cards))
	for n := range s.cards {
		if !basicLands[n] {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	out := make([]Row, 0, len(names))
	for _, n := range names {
		c := s.cards[n]
		out = append(out, Row{
			Name:               n,
			EverDrawnGameCount: c.inHand,
			EverDrawnWinRate:   ratio(c.inHandWins, c.inHand, 6),
			AvgSeen:            ratio(c.seenSum, c.seen, 4),
			SeenCount:          c.seen,
			AvgPick:            ratio(c.pickSum, c.picks, 4),
			PickCount:          c.picks,
		})
	}
	return out
}

// WriteRows writes the rows as one JSON array.
func WriteRows(w io.Writer, rows []Row) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	return enc.Encode(rows)
}
