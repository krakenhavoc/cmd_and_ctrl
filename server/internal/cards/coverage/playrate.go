package coverage

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// playrate.go — catalog coverage by play rate (#3352).
//
// The census counts what the catalog holds; it cannot say whether those
// are the cards strangers put in their decks. Scryfall's `edhrec_rank`
// is the play-rate signal the dump already carries: rank 1 is the most
// played Commander card. This file measures how much of each top-N the
// catalog covers and lists the gaps, so the next sweep batch is picked
// from data rather than memory.
//
// The input is the Scryfall default-cards dump, which CI does not have,
// so the report is not part of the census block (that block is checked
// in CI and a dump-gated check would skip there). `cmd/playrate` prints
// it for a human or an agent at a prompt:
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go run ./cmd/playrate [-top 500,1000,2000] [-list 1000]
//
// A card is covered when its bare oracle ID is a registry key. A back
// face registered alone ("<id>#1", the MDFC land backs) is half a card
// and does not count.

// DefaultPlayRateTops are the cut-offs reported when none are asked for.
var DefaultPlayRateTops = []int{250, 500, 1000, 2000, 5000}

// RankedCard is one oracle card with its EDHREC rank.
type RankedCard struct {
	OracleID string
	Name     string
	Rank     int
}

// ReadRankedCards streams a Scryfall default-cards dump and returns one
// entry per oracle ID that has an edhrec_rank, ordered by rank then name.
// A card with several printings keeps its lowest rank.
func ReadRankedCards(r io.Reader) ([]RankedCard, error) {
	dec := json.NewDecoder(r)
	if tok, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("read dump: %w", err)
	} else if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, fmt.Errorf("read dump: want a JSON array, got %v", tok)
	}
	best := map[string]RankedCard{}
	for dec.More() {
		var c struct {
			OracleID  string `json:"oracle_id"`
			Name      string `json:"name"`
			Rank      int    `json:"edhrec_rank"`
			CardFaces []struct {
				OracleID string `json:"oracle_id"`
			} `json:"card_faces"`
		}
		if err := dec.Decode(&c); err != nil {
			return nil, fmt.Errorf("read dump: %w", err)
		}
		id := c.OracleID
		if id == "" && len(c.CardFaces) > 0 {
			id = c.CardFaces[0].OracleID
		}
		if id == "" || c.Rank <= 0 {
			continue
		}
		if old, ok := best[id]; !ok || c.Rank < old.Rank {
			best[id] = RankedCard{OracleID: id, Name: c.Name, Rank: c.Rank}
		}
	}
	out := make([]RankedCard, 0, len(best))
	for _, c := range best {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Rank != out[j].Rank {
			return out[i].Rank < out[j].Rank
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// CoveredOracleIDs returns the set of bare oracle IDs the registry holds.
func CoveredOracleIDs() map[string]bool {
	out := map[string]bool{}
	for _, s := range effects.All() {
		if !faceSuffix.MatchString(s.OracleID) {
			out[s.OracleID] = true
		}
	}
	return out
}

// PlayRateRow is the coverage of one top-N.
type PlayRateRow struct {
	Top     int // rank cut-off
	Ranked  int // cards with an edhrec_rank at or under Top
	Covered int // of those, how many the catalog holds
}

// Percent is Covered as a percentage of Ranked, 0 when nothing is ranked.
func (r PlayRateRow) Percent() float64 {
	if r.Ranked == 0 {
		return 0
	}
	return 100 * float64(r.Covered) / float64(r.Ranked)
}

// PlayRateReport is coverage by play rate over a set of cut-offs.
type PlayRateReport struct {
	Rows []PlayRateRow
	// Missing are the uncovered cards at or under the largest list cut-off,
	// in rank order.
	Missing []RankedCard
}

// BuildPlayRateReport measures `ranked` (sorted by rank) against `covered`.
// tops are the cut-offs to count; missingTop is the cut-off for the
// Missing list (0 for none).
func BuildPlayRateReport(ranked []RankedCard, covered map[string]bool, tops []int, missingTop int) PlayRateReport {
	var rep PlayRateReport
	tops = append([]int(nil), tops...)
	sort.Ints(tops)
	for _, top := range tops {
		row := PlayRateRow{Top: top}
		for _, c := range ranked {
			if c.Rank > top {
				break
			}
			row.Ranked++
			if covered[c.OracleID] {
				row.Covered++
			}
		}
		rep.Rows = append(rep.Rows, row)
	}
	for _, c := range ranked {
		if c.Rank > missingTop {
			break
		}
		if !covered[c.OracleID] {
			rep.Missing = append(rep.Missing, c)
		}
	}
	return rep
}

// String renders the report as plain text.
func (r PlayRateReport) String() string {
	var b strings.Builder
	b.WriteString("Catalog coverage by EDHREC rank\n\n")
	b.WriteString("  top    ranked  covered  missing  coverage\n")
	for _, row := range r.Rows {
		fmt.Fprintf(&b, "  %-6d %6d  %7d  %7d  %7.1f%%\n",
			row.Top, row.Ranked, row.Covered, row.Ranked-row.Covered, row.Percent())
	}
	if len(r.Missing) > 0 {
		fmt.Fprintf(&b, "\nMissing from the catalog (%d), by rank:\n", len(r.Missing))
		for _, c := range r.Missing {
			fmt.Fprintf(&b, "  %5d  %s  %s\n", c.Rank, c.Name, c.OracleID)
		}
	}
	return b.String()
}
