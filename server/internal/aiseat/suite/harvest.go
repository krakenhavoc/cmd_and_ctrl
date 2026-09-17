package suite

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// harvest.go pulls candidate windows out of decision logs into an
// inbox of UNLABELLED positions.
//
// It exists because labelling is the expensive half of a position
// suite and finding the windows worth labelling is the cheap half. A
// game writes thousands of windows, almost all of them a forced pass;
// the ones worth a human's attention are the escalated ones, the ones
// where the two layers disagreed, and the ones where something in the
// funnel failed. Those filters are this file.
//
// Nothing here writes an `expected`. A harvested position is a
// question, and Run skips it until somebody answers it.

// HarvestFilter narrows the windows worth looking at. The zero value
// takes every record that carries a full Input.
type HarvestFilter struct {
	// Escalated keeps only windows that left Layer A — the ones a
	// model tier is actually judged on.
	Escalated bool
	// Disagree keeps only windows where the model's parsed index and
	// the heuristic's differ and both are valid. These are the
	// highest-value labelling candidates: one of the two layers is
	// wrong and a human can say which.
	Disagree bool
	// Fallbacks keeps windows whose funnel fallback (Trace.Fallback)
	// or runner fallback is one of these — "malformed",
	// "out-of-range", "timeout", "no-budget", "error".
	Fallbacks []string
	// Layers keeps windows answered by one of these layers ("A",
	// "B", "C", "random").
	Layers []string
	// Seats keeps windows for these seat indices.
	Seats []int
	// Tags keeps windows whose auto-tags include one of these.
	Tags []string
	// Limit caps how many positions are written. Zero is no cap.
	Limit int
	// Seed makes the sampling deterministic when Limit cuts. Two
	// harvests of the same logs with the same seed write the same
	// files, which is what makes a harvested corpus reviewable.
	Seed int64
}

// HarvestReport is what one harvest did.
type HarvestReport struct {
	// Files is the logs that were read.
	Files []string `json:"files"`
	// Records is every line scanned; Full is how many carried an
	// Input (a compact record cannot become a position).
	Records int `json:"records"`
	Full    int `json:"full"`
	// Candidates is how many passed the filter, before Limit.
	Candidates int `json:"candidates"`
	// Written is how many position files were created, and Paths
	// names them.
	Written int      `json:"written"`
	Paths   []string `json:"paths,omitempty"`
}

// Harvest reads the named decision logs and writes one unlabelled
// position per window that passes the filter.
func Harvest(paths []string, filter HarvestFilter, outDir string) (HarvestReport, error) {
	rep := HarvestReport{Files: paths}
	var candidates []Position

	for _, path := range paths {
		err := decisionlog.Scan(path, func(rec decisionlog.Record) error {
			rep.Records++
			if rec.Input == nil || len(rec.Input.Moves) == 0 {
				// A compact record (escalated mode's forced passes)
				// has the move list but no view; a position without a
				// view cannot be rendered, replayed or labelled.
				return nil
			}
			rep.Full++
			if !keep(rec, filter) {
				return nil
			}
			candidates = append(candidates, positionFrom(rec, path))
			return nil
		})
		if err != nil {
			return rep, fmt.Errorf("suite: harvest %s: %w", path, err)
		}
	}
	rep.Candidates = len(candidates)
	candidates = sample(candidates, filter.Limit, filter.Seed)

	if err := os.MkdirAll(outDir, 0o700); err != nil {
		return rep, fmt.Errorf("suite: create %s: %w", outDir, err)
	}
	used := map[string]bool{}
	for _, p := range candidates {
		p.ID = uniqueID(p.ID, used, outDir)
		used[p.ID] = true
		written, err := p.Write(outDir)
		if err != nil {
			return rep, err
		}
		rep.Written++
		rep.Paths = append(rep.Paths, written)
	}
	return rep, nil
}

// keep applies the filter to one record.
func keep(rec decisionlog.Record, f HarvestFilter) bool {
	if f.Escalated && !rec.Escalated() {
		return false
	}
	if f.Disagree && !disagreed(rec) {
		return false
	}
	if len(f.Fallbacks) > 0 && !matchesFallback(rec, f.Fallbacks) {
		return false
	}
	if len(f.Layers) > 0 && !matchesAny(f.Layers, rec.Trace.Layer) {
		return false
	}
	if len(f.Seats) > 0 {
		found := false
		for _, s := range f.Seats {
			if s == rec.SeatIndex {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(f.Tags) > 0 {
		tags := autoTags(rec)
		found := false
		for _, want := range f.Tags {
			for _, got := range tags {
				if strings.EqualFold(strings.TrimSpace(want), got) {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func matchesFallback(rec decisionlog.Record, want []string) bool {
	if matchesAny(want, rec.Trace.Fallback, rec.Final.RunnerFallback) {
		return true
	}
	return rec.Trace.TimedOut && matchesAny(want, aiseat.FallbackTimeout)
}

// disagreed reports whether the model and the heuristic wanted
// different moves, both of them valid.
//
// It reads ParsedIndex rather than the final index on purpose: a
// model answer that was parsed and then discarded (out of range,
// or overruled) is still what the model wanted, and a window where
// the two layers pulled apart is worth labelling whichever one won.
func disagreed(rec decisionlog.Record) bool {
	if rec.Trace.ParsedIndex == nil || rec.Input == nil {
		return false
	}
	m, h, n := *rec.Trace.ParsedIndex, rec.Trace.HeuristicIndex, len(rec.Input.Moves)
	if m < 0 || m >= n || h < 0 || h >= n {
		return false
	}
	return m != h
}

func matchesAny(want []string, got ...string) bool {
	for _, w := range want {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		for _, g := range got {
			if g != "" && strings.EqualFold(w, g) {
				return true
			}
		}
	}
	return false
}

// autoTags guesses a position's tags from the shape of the window.
// They are a starting point for the labeller, not a claim: a harvest
// knows what kinds of move were on offer and nothing about why the
// window is interesting.
func autoTags(rec decisionlog.Record) []string {
	if rec.Input == nil {
		return nil
	}
	seen := map[string]bool{}
	for _, mv := range rec.Input.Moves {
		switch mv.Kind {
		case legal.KindBlock:
			seen["block"] = true
		case legal.KindAttack:
			seen["attack"] = true
		case legal.KindChoice:
			seen["choice"] = true
		case legal.KindLand:
			seen["land"] = true
		case legal.KindMulligan:
			seen["mulligan"] = true
		case legal.KindCast:
			seen["cast"] = true
		case legal.KindPass, legal.KindActivate, legal.KindMana:
			// Present in almost every window; tagging on them would
			// make the tag mean nothing.
		}
	}
	if len(rec.Input.View.StackItems) > 0 {
		seen["stack-response"] = true
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// positionFrom projects a record into an unlabelled position.
func positionFrom(rec decisionlog.Record, logPath string) Position {
	in := *rec.Input
	p := Position{
		ID:   harvestID(rec),
		V:    PositionVersion,
		Tags: autoTags(rec),
		Source: Source{
			GameID: rec.Game,
			Seat:   rec.SeatIndex,
			Seq:    rec.Seq,
			Turn:   rec.Turn,
			Step:   rec.Step,
			Log:    logPath,
		},
		Input:     in,
		Expected:  Expected{Accept: []Matcher{}},
		AtCapture: captureOf(rec),
	}
	return p
}

// captureOf records what the funnel picked when this window was
// played, so a labeller can see whether they are agreeing with the
// machine or overruling it.
func captureOf(rec decisionlog.Record) AtCapture {
	var ac AtCapture
	n := len(rec.Input.Moves)
	if h := rec.Trace.HeuristicIndex; h >= 0 && h < n {
		ac.Heuristic = &Pick{Index: h, Label: rec.Input.Moves[h].Label}
	}
	if rec.Trace.Layer == model.LayerC && rec.Trace.ParsedIndex != nil {
		m := *rec.Trace.ParsedIndex
		pick := &ModelPick{Index: m, ID: rec.Trace.Model, Layer: rec.Trace.Layer}
		if m >= 0 && m < n {
			pick.Label = rec.Input.Moves[m].Label
		}
		ac.Model = pick
	}
	return ac
}

// harvestID is <game[:8]>-s<seat>-seq<seq>: short enough to type,
// long enough to find the window again in the log it came from.
func harvestID(rec decisionlog.Record) string {
	game := rec.Game
	if len(game) > 8 {
		game = game[:8]
	}
	if game == "" {
		game = "unknown"
	}
	return fmt.Sprintf("%s-s%d-seq%d", game, rec.SeatIndex, rec.Seq)
}

// uniqueID disambiguates the handful of collisions the id scheme
// admits: a window that committed nothing has seq 0, and a seat can
// have several of those in a row.
func uniqueID(id string, used map[string]bool, outDir string) string {
	exists := func(cand string) bool {
		if used[cand] {
			return true
		}
		_, err := os.Stat(filepath.Join(outDir, cand+".json"))
		return err == nil
	}
	if !exists(id) {
		return id
	}
	for n := 2; ; n++ {
		cand := fmt.Sprintf("%s-%d", id, n)
		if !exists(cand) {
			return cand
		}
	}
}

// sample cuts the candidate list to limit, deterministically.
//
// Deterministic because a corpus that changes shape every time it is
// harvested cannot be reviewed: the same logs and the same seed have
// to produce the same inbox, or the reviewer is chasing a moving
// target. The order of what survives is the log's own, so the inbox
// still reads chronologically.
func sample(in []Position, limit int, seed int64) []Position {
	if limit <= 0 || len(in) <= limit {
		return in
	}
	idx := make([]int, len(in))
	for i := range idx {
		idx[i] = i
	}
	rng := rand.New(rand.NewPCG(uint64(seed), 0x5EED))
	rng.Shuffle(len(idx), func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	idx = idx[:limit]
	sort.Ints(idx)
	out := make([]Position, 0, limit)
	for _, i := range idx {
		out = append(out, in[i])
	}
	return out
}
