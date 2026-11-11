// Package draftratings is the outside card ratings the draft shows
// beside a card and its bots draft by: 17Lands' games-in-hand win rate
// (GIH WR), one small file per set (ADR 0222 §7, amendment 2026-11-11,
// #5703).
//
// # Where the numbers come from
//
// 17Lands is a community project that publishes aggregate card
// statistics from Arena drafts. The owner confirmed on 2026-11-11 that
// the data may be used, with credit on the page: every screen that shows
// a rating says "Card ratings from 17Lands" and links to 17lands.com
// (client/src/lib/labels.ts). Nothing here is gated or sold, and no
// person's data is in a file: a row is a card name and four aggregates.
//
// scripts/draft-ratings-refresh.sh writes data/<set>.json. It keeps a
// card only when 17Lands has a win rate for it (500 or more games in
// hand), and a set only when that leaves 40 cards, so a set 17Lands
// does not cover, or covers thinly, has no file and the draft shows the
// in-house rating (draft.Rater) for it. The files are embedded in the
// binary; a refresh is a commit.
//
// # What a rating is
//
// A Rating is a card's numbers in one set's file, with a letter Tier
// read off that set's own spread: a card's rank among the set's rated
// cards, so "B+" means the same share of the set in a set of weak cards
// as in a strong one. Z is the card's distance from the set's mean win
// rate in standard deviations; the draft bots turn it into their own
// scale (draft.RatingFromZ).
//
// Cards are found by oracle ID where a card index was bound (Bind: the
// index resolves 17Lands' spelling, a double-faced card's front face or
// full name, and the flavour names Arena uses), and by name otherwise.
package draftratings

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

//go:embed data/*.json
var embedded embed.FS

// Credit is the line every screen that shows these numbers carries, and
// URL where it links. The client's label (L.draftRatingCredit) says the
// same; this is the server's record of the owner's condition.
const (
	Credit = "Card ratings from 17Lands"
	URL    = "https://www.17lands.com/"
)

// Entry is one card's row of a set's file, as the script writes it.
type Entry struct {
	// Name is the card's name as 17Lands spells it.
	Name string `json:"n"`
	// WinRate is GIH WR as a fraction: the win rate of games with the
	// card in hand, opening hand or drawn.
	WinRate float64 `json:"w"`
	// Games is the number of games behind WinRate ("# GIH").
	Games int `json:"g"`
	// AvgSeen is the average pick number the card was last seen at, and
	// AvgPick the pick number it was taken at on average.
	AvgSeen float64 `json:"s"`
	AvgPick float64 `json:"p"`
}

// file is a set's file.
type file struct {
	Set      string  `json:"set"`
	Source   string  `json:"source"`
	Format   string  `json:"format"`
	Start    string  `json:"start"`
	Fetched  string  `json:"fetched"`
	MinGames int     `json:"min_games"`
	Cards    []Entry `json:"cards"`
}

// Rating is a card's 17Lands numbers in one set.
type Rating struct {
	// Set is the set whose file the numbers are from, lowercase.
	Set string
	// Format is the 17Lands format they were measured in
	// ("PremierDraft"), and Fetched the date they were fetched.
	Format  string
	Fetched string
	Entry
	// Tier is the letter grade from the set's own spread (Tiers).
	Tier string
	// Z is the win rate's distance from the set's mean, in standard
	// deviations of the set's rated cards.
	Z float64
}

// Percent is the win rate in percent, one decimal, as 17Lands shows it.
func (r Rating) Percent() float64 { return math.Round(r.WinRate*1000) / 10 }

// Tiers are the letter grades, best first, and the share of a set's
// rated cards at or above which a card earns each: the best 3% are A+,
// the next 7% A, and so on down to F for the weakest 3%. A card's share
// is the fraction of the set's rated cards with a lower win rate, plus
// half of those tied with it.
var Tiers = []struct {
	Grade string
	// From is the lowest share, in [0,1], that earns the grade.
	From float64
}{
	{"A+", 0.97}, {"A", 0.90}, {"A-", 0.80}, {"B+", 0.70}, {"B", 0.55}, {"B-", 0.40},
	{"C+", 0.25}, {"C", 0.15}, {"C-", 0.08}, {"D", 0.03}, {"F", 0},
}

// set is a loaded file with what is computed from it.
type set struct {
	file
	mean, sd float64
	ratings  []Rating
	byName   map[string]int
	byOracle map[uuid.UUID]int
}

// Book is every set's ratings.
type Book struct {
	sets map[string]*set
	// order is the set codes, most recent first by the date 17Lands
	// opened the set.
	order []string
}

// Load is the Book of the files embedded in the binary.
func Load() (*Book, error) { return LoadFS(embedded, "data") }

// MustLoad is Load for the server's start: the files are embedded and
// checked by this package's tests, so an error is a build mistake.
func MustLoad() *Book {
	b, err := Load()
	if err != nil {
		panic(err)
	}
	return b
}

// LoadFS reads every *.json file of dir in fsys. A file that does not
// parse, has no cards, or names a different set than its filename is an
// error: a bad refresh must not be served.
func LoadFS(fsys fs.FS, dir string) (*Book, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	b := &Book{sets: map[string]*set{}}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := fs.ReadFile(fsys, dir+"/"+e.Name())
		if err != nil {
			return nil, err
		}
		var f file
		if err := json.Unmarshal(raw, &f); err != nil {
			return nil, fmt.Errorf("draftratings: %s: %w", e.Name(), err)
		}
		code := strings.TrimSuffix(e.Name(), ".json")
		if f.Set != code {
			return nil, fmt.Errorf("draftratings: %s names set %q", e.Name(), f.Set)
		}
		if len(f.Cards) == 0 {
			return nil, fmt.Errorf("draftratings: %s has no cards", e.Name())
		}
		s := &set{file: f, byName: map[string]int{}, byOracle: map[uuid.UUID]int{}}
		s.compute()
		b.sets[code] = s
		b.order = append(b.order, code)
	}
	sort.Slice(b.order, func(i, j int) bool {
		a, c := b.sets[b.order[i]], b.sets[b.order[j]]
		if a.Start != c.Start {
			return a.Start > c.Start
		}
		return a.Set < c.Set
	})
	return b, nil
}

// compute fills in the set's mean, spread, tiers and name index.
func (s *set) compute() {
	n := float64(len(s.Cards))
	for _, c := range s.Cards {
		s.mean += c.WinRate / n
	}
	for _, c := range s.Cards {
		s.sd += (c.WinRate - s.mean) * (c.WinRate - s.mean) / n
	}
	s.sd = math.Sqrt(s.sd)
	s.ratings = make([]Rating, len(s.Cards))
	for i, c := range s.Cards {
		below, tied := 0.0, 0.0
		for _, o := range s.Cards {
			switch {
			case o.WinRate < c.WinRate:
				below++
			case o.WinRate == c.WinRate:
				tied++
			}
		}
		share := (below + (tied-1)/2 + 0.5) / n
		z := 0.0
		if s.sd > 0 {
			z = (c.WinRate - s.mean) / s.sd
		}
		s.ratings[i] = Rating{
			Set: s.Set, Format: s.Format, Fetched: s.Fetched, Entry: c,
			Tier: tierOf(share), Z: z,
		}
		s.byName[nameKey(c.Name)] = i
	}
}

func tierOf(share float64) string {
	for _, t := range Tiers {
		if share >= t.From {
			return t.Grade
		}
	}
	return Tiers[len(Tiers)-1].Grade
}

// nameKey is a card name as it is compared: the front face of a
// double-faced or split card, lowercase, single-spaced.
func nameKey(name string) string {
	front, _, _ := strings.Cut(name, "//")
	return strings.ToLower(strings.Join(strings.Fields(front), " "))
}

// Bind resolves every card name of every file to an oracle ID with the
// card index, so a card is found by identity: 17Lands' spelling of a
// double-faced card, or Arena's title for a reskinned one, still lands
// on the card the draft holds. A name the index does not know stays
// findable by name. Bind is not safe to call while the Book is read; the
// server calls it once at start.
func (b *Book) Bind(idx *cards.Index) {
	if b == nil || idx == nil {
		return
	}
	for _, s := range b.sets {
		clear(s.byOracle)
		for i, c := range s.Cards {
			if card, ok := idx.FindByName(c.Name); ok && card.OracleID != uuid.Nil {
				s.byOracle[card.OracleID] = i
			}
		}
	}
}

// Sets are the codes of the sets that have a file, most recent first.
func (b *Book) Sets() []string {
	if b == nil {
		return nil
	}
	return append([]string(nil), b.order...)
}

// Has reports whether a set has a file.
func (b *Book) Has(set string) bool {
	if b == nil {
		return false
	}
	_, ok := b.sets[strings.ToLower(set)]
	return ok
}

// Cards is how many cards a set's file rates, 0 for a set without one.
func (b *Book) Cards(set string) int {
	if b == nil {
		return 0
	}
	if s, ok := b.sets[strings.ToLower(set)]; ok {
		return len(s.Cards)
	}
	return 0
}

// find is the card in one set's file, by oracle ID and then by name.
func (s *set) find(oracle uuid.UUID, name string) (Rating, bool) {
	if oracle != uuid.Nil {
		if i, ok := s.byOracle[oracle]; ok {
			return s.ratings[i], true
		}
	}
	if i, ok := s.byName[nameKey(name)]; ok {
		return s.ratings[i], true
	}
	return Rating{}, false
}

// In is the card's rating in the file of the named set.
func (b *Book) In(set string, oracle uuid.UUID, name string) (Rating, bool) {
	if b == nil {
		return Rating{}, false
	}
	s, ok := b.sets[strings.ToLower(set)]
	if !ok {
		return Rating{}, false
	}
	return s.find(oracle, name)
}

// Latest is the card's rating from the most recent set whose file has
// it: what a cube card, or a reprint the card's own set has no rate for,
// is shown with.
func (b *Book) Latest(oracle uuid.UUID, name string) (Rating, bool) {
	if b == nil {
		return Rating{}, false
	}
	for _, code := range b.order {
		if r, ok := b.sets[code].find(oracle, name); ok {
			return r, true
		}
	}
	return Rating{}, false
}

// For is the rating a draft shows for a card of the named set: the
// card's own set first, then the most recent set that has it.
func (b *Book) For(set string, oracle uuid.UUID, name string) (Rating, bool) {
	if r, ok := b.In(set, oracle, name); ok {
		return r, true
	}
	return b.Latest(oracle, name)
}
