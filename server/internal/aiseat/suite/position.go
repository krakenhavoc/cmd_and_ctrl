// Package suite is the labelled position suite: a corpus of frozen
// decision windows with a HUMAN's answer attached, and the machinery
// to ask a policy the same questions and count how often it agrees.
//
// # Why a suite and not a win rate
//
// An arena says which bot wins more games. It cannot say WHY, it
// cannot be run in CI (minutes of wall clock, and a model endpoint),
// and it moves under you: change the heuristic and every position the
// arena ever visited changes too. A suite is the complement — the
// same fixed questions, asked of every policy, every commit. It is
// cheap enough to be a regression gate on the heuristic and precise
// enough to say "the model got worse at blocks" rather than "the
// model lost".
//
// # Labels, never indices (ADR 0052 §3)
//
// A position stores the answer as MATCHERS over the frozen move list
// — a label, a regexp, a legal.Kind, or an action type plus a subset
// of its params — and the index is derived at load time by matching.
// Storing the index would be correct until the day legal's
// enumeration order changes, and then every position in the suite
// would silently start grading a different move while still reporting
// a number. A matcher either still matches or it does not, and one
// that matches nothing fails Load LOUDLY, naming the position.
//
// # The import rule
//
// This package sits under aiseat/ and is bound by ADR 0033 §3 like
// every other subpackage: no internal/game, ever, test files
// included. It reads frozen aiseat.Input JSON and calls policies.
// The binary that needs a game to harvest from is cmd/boteval.
package suite

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// PositionVersion is the schema version every position file carries.
// A file with a version this build does not know fails Load rather
// than being guessed at: a suite that silently mis-grades is worse
// than one that will not start.
const PositionVersion = 1

// Source is where a position came from. It is provenance, not input:
// nothing in Run reads it, and a reviewer chasing "is this label
// still right?" needs every field of it.
type Source struct {
	GameID     string `json:"game_id,omitempty"`
	Seat       int    `json:"seat"`
	Seq        uint64 `json:"seq,omitempty"`
	Turn       int    `json:"turn,omitempty"`
	Step       string `json:"step,omitempty"`
	Log        string `json:"log,omitempty"`
	Reviewer   string `json:"reviewer,omitempty"`
	ReviewedAt string `json:"reviewed_at,omitempty"`
}

// Pick is one policy's answer at capture time.
type Pick struct {
	Index int    `json:"index"`
	Label string `json:"label,omitempty"`
}

// ModelPick is Pick plus which model said it and which layer answered.
type ModelPick struct {
	Index int    `json:"index"`
	Label string `json:"label,omitempty"`
	ID    string `json:"id,omitempty"`
	Layer string `json:"layer,omitempty"`
}

// AtCapture records what the funnel picked when the window was
// harvested. It is context for the labeller and a way to find the
// windows where the two layers disagreed; Run never grades against it.
type AtCapture struct {
	Heuristic *Pick      `json:"heuristic,omitempty"`
	Model     *ModelPick `json:"model,omitempty"`
}

// Matcher names a move without naming its index.
//
// Exactly one of Label, LabelRe, Kind or Type is set. ParamsSubset
// may accompany Type, and narrows it to moves whose Params contain
// every key/value the subset names — which is how "sacrifice THAT
// creature" is expressed when two moves share a label.
type Matcher struct {
	// Label matches legal.Move.Label exactly.
	Label string `json:"label,omitempty"`
	// LabelRe matches legal.Move.Label as an RE2 regexp.
	LabelRe string `json:"label_re,omitempty"`
	// Kind matches legal.Move.Kind — the coarse matcher, and the
	// right one for "make the land drop" or "pass".
	Kind legal.Kind `json:"kind,omitempty"`
	// Type matches legal.Move.Type, the action type on the wire.
	Type string `json:"type,omitempty"`
	// ParamsSubset narrows Type: every key in this object must be
	// present in the move's Params with an equal value. Nested
	// objects are compared the same way, recursively.
	ParamsSubset json.RawMessage `json:"params_subset,omitempty"`

	re *regexp.Regexp
}

// String renders a matcher for an error message or a report.
func (m Matcher) String() string {
	switch {
	case m.Label != "":
		return fmt.Sprintf("label %q", m.Label)
	case m.LabelRe != "":
		return fmt.Sprintf("label_re %q", m.LabelRe)
	case m.Kind != "":
		return fmt.Sprintf("kind %q", m.Kind)
	case m.Type != "" && len(m.ParamsSubset) > 0:
		return fmt.Sprintf("type %q with params %s", m.Type, string(m.ParamsSubset))
	case m.Type != "":
		return fmt.Sprintf("type %q", m.Type)
	default:
		return "empty matcher"
	}
}

// compile validates the matcher's shape and prepares its regexp.
func (m *Matcher) compile() error {
	forms := 0
	for _, set := range []bool{m.Label != "", m.LabelRe != "", m.Kind != "", m.Type != ""} {
		if set {
			forms++
		}
	}
	switch {
	case forms == 0:
		return fmt.Errorf("a matcher must set one of label, label_re, kind or type")
	case forms > 1:
		return fmt.Errorf("a matcher sets %d of label/label_re/kind/type; it may set exactly one", forms)
	case len(m.ParamsSubset) > 0 && m.Type == "":
		return fmt.Errorf("params_subset is only meaningful beside type")
	}
	if len(m.ParamsSubset) > 0 {
		var probe map[string]any
		if err := json.Unmarshal(m.ParamsSubset, &probe); err != nil {
			return fmt.Errorf("params_subset is not a JSON object: %w", err)
		}
	}
	if m.LabelRe != "" {
		re, err := regexp.Compile(m.LabelRe)
		if err != nil {
			return fmt.Errorf("label_re %q: %w", m.LabelRe, err)
		}
		m.re = re
	}
	return nil
}

// Matches reports whether mv is one of the moves this matcher names.
func (m Matcher) Matches(mv legal.Move) bool {
	switch {
	case m.Label != "":
		return mv.Label == m.Label
	case m.re != nil:
		return m.re.MatchString(mv.Label)
	case m.LabelRe != "":
		// compile() was never run (a hand-built matcher). Compile now
		// rather than silently matching nothing.
		re, err := regexp.Compile(m.LabelRe)
		if err != nil {
			return false
		}
		return re.MatchString(mv.Label)
	case m.Kind != "":
		return mv.Kind == m.Kind
	case m.Type != "":
		if mv.Type != m.Type {
			return false
		}
		return paramsContain(mv.Params, m.ParamsSubset)
	default:
		return false
	}
}

// paramsContain reports whether every key/value in subset appears in
// params with an equal value. An empty subset matches anything.
func paramsContain(params, subset json.RawMessage) bool {
	if len(subset) == 0 {
		return true
	}
	var want, got map[string]any
	if err := json.Unmarshal(subset, &want); err != nil {
		return false
	}
	if len(params) == 0 {
		return len(want) == 0
	}
	if err := json.Unmarshal(params, &got); err != nil {
		return false
	}
	return subsetOf(want, got)
}

func subsetOf(want, got map[string]any) bool {
	for k, wv := range want {
		gv, ok := got[k]
		if !ok {
			return false
		}
		wm, wIsMap := wv.(map[string]any)
		gm, gIsMap := gv.(map[string]any)
		if wIsMap && gIsMap {
			if !subsetOf(wm, gm) {
				return false
			}
			continue
		}
		// Everything else compares by its JSON rendering, which makes
		// numbers, strings, bools, nulls and arrays all work without a
		// type switch per shape.
		wb, werr := json.Marshal(wv)
		gb, gerr := json.Marshal(gv)
		if werr != nil || gerr != nil || string(wb) != string(gb) {
			return false
		}
	}
	return true
}

// Expected is the label: which moves are right, which are
// specifically wrong, and whether declining is defensible.
//
// An EMPTY Accept means the position has not been labelled yet — a
// harvested position straight out of the inbox. Run reports those as
// OutcomeSkipped and never counts them as agreement, because a suite
// that scores its own unlabelled positions is measuring nothing.
type Expected struct {
	Accept    []Matcher `json:"accept"`
	Reject    []Matcher `json:"reject,omitempty"`
	DeclineOK bool      `json:"decline_ok,omitempty"`
}

// Position is one labelled decision window.
type Position struct {
	ID   string   `json:"id"`
	V    int      `json:"v"`
	Tags []string `json:"tags,omitempty"`
	Note string   `json:"note,omitempty"`
	// Gate names the policies for which a miss on this position
	// FAILS `go test` rather than merely being counted. A gated
	// position is a pinned bug or a pinned invariant; gate a policy
	// only once it demonstrably passes.
	Gate      []string     `json:"gate,omitempty"`
	Source    Source       `json:"source"`
	Input     aiseat.Input `json:"input"`
	Expected  Expected     `json:"expected"`
	AtCapture AtCapture    `json:"at_capture"`

	// accept and reject are the derived indices — computed at Load
	// by matching, never stored. Path is where the file came from.
	accept []int
	reject []int
	path   string
}

// Path is the file this position was loaded from, empty for one built
// in memory.
func (p Position) Path() string { return p.path }

// Accept is the indices of the moves the label calls right.
func (p Position) Accept() []int { return append([]int(nil), p.accept...) }

// Reject is the indices of the moves the label calls specifically wrong.
func (p Position) Reject() []int { return append([]int(nil), p.reject...) }

// Labelled reports whether anybody has answered this position yet.
func (p Position) Labelled() bool { return len(p.Expected.Accept) > 0 }

// Gates reports whether a miss by the named policy fails the build.
func (p Position) Gates(policy string) bool {
	for _, g := range p.Gate {
		if strings.EqualFold(strings.TrimSpace(g), policy) {
			return true
		}
	}
	return false
}

// AcceptLabels is the accepted moves' labels, for a report.
func (p Position) AcceptLabels() []string {
	out := make([]string, 0, len(p.accept))
	for _, i := range p.accept {
		out = append(out, p.Input.Moves[i].Label)
	}
	return out
}

// Load reads every *.json file in dir as a position, derives the
// accept and reject indices, and returns them sorted by id.
//
// It is strict on purpose. A position that will not load is a
// position that would otherwise grade silently wrong: an unknown
// schema version, a matcher that names no move in the frozen list
// ("stale" — the enumerator moved under the label), a duplicate id.
// All of them are errors, and every error names the position.
func Load(dir string) ([]Position, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("suite: read %s: %w", dir, err)
	}
	var (
		out  []Position
		seen = map[string]string{}
	)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		p, err := LoadFile(path)
		if err != nil {
			return nil, err
		}
		if prev, dup := seen[p.ID]; dup {
			return nil, fmt.Errorf("suite: position id %q appears in both %s and %s; ids are the suite's primary key", p.ID, prev, path)
		}
		seen[p.ID] = path
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// LoadFile reads and validates one position file.
func LoadFile(path string) (Position, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // the suite reads the directory it was pointed at
	if err != nil {
		return Position{}, fmt.Errorf("suite: read %s: %w", path, err)
	}
	var p Position
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return Position{}, fmt.Errorf("suite: %s: %w", path, err)
	}
	p.path = path
	if err := p.prepare(); err != nil {
		return Position{}, err
	}
	return p, nil
}

// prepare validates the position and derives the matcher indices.
func (p *Position) prepare() error {
	where := p.ID
	if where == "" {
		where = p.path
	}
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("suite: %s: a position needs an id", p.path)
	}
	if p.V != PositionVersion {
		return fmt.Errorf("suite: position %q (%s): schema version %d, this build reads %d", where, p.path, p.V, PositionVersion)
	}
	if len(p.Input.Moves) == 0 {
		return fmt.Errorf("suite: position %q (%s): the frozen input has no legal moves", where, p.path)
	}
	accept, err := p.resolve(p.Expected.Accept, "accept")
	if err != nil {
		return err
	}
	reject, err := p.resolve(p.Expected.Reject, "reject")
	if err != nil {
		return err
	}
	for _, a := range accept {
		for _, r := range reject {
			if a == r {
				return fmt.Errorf("suite: position %q (%s): move %d %q is both accepted and rejected",
					where, p.path, a, p.Input.Moves[a].Label)
			}
		}
	}
	p.accept, p.reject = accept, reject
	return nil
}

// resolve turns matchers into indices, and fails on any that matches
// nothing. That failure is the whole point of storing labels: a stale
// matcher means the enumerator changed under the label, and a suite
// that skipped it would keep reporting an agreement number computed
// over a question nobody is asking any more.
func (p *Position) resolve(ms []Matcher, field string) ([]int, error) {
	var out []int
	seen := map[int]bool{}
	for i := range ms {
		m := &ms[i]
		if err := m.compile(); err != nil {
			return nil, fmt.Errorf("suite: position %q (%s): expected.%s[%d]: %w", p.ID, p.path, field, i, err)
		}
		hits := 0
		for j, mv := range p.Input.Moves {
			if !m.Matches(mv) {
				continue
			}
			hits++
			if !seen[j] {
				seen[j] = true
				out = append(out, j)
			}
		}
		if hits == 0 {
			return nil, fmt.Errorf("suite: position %q (%s): expected.%s[%d] (%s) is STALE — it matches none of the %d frozen moves. "+
				"Either the label was wrong or the enumerator's labels moved; re-render the position and re-label it",
				p.ID, p.path, field, i, m.String(), len(p.Input.Moves))
		}
	}
	sort.Ints(out)
	return out, nil
}

// Accepts reports whether index i is one of the labelled answers.
func (p Position) Accepts(i int) bool { return containsInt(p.accept, i) }

// Rejects reports whether index i is one the label names as wrong.
func (p Position) Rejects(i int) bool { return containsInt(p.reject, i) }

func containsInt(xs []int, i int) bool {
	for _, x := range xs {
		if x == i {
			return true
		}
	}
	return false
}

// envelope is the on-disk shape, and it exists for one reason: the
// frozen Input is 25 KiB of board view that no human will ever read,
// and indenting it triples the file for nothing. Marshal indents
// everything a labeller edits — id, tags, note, gate, source,
// expected — and writes `input` on a single compact line.
//
// It mirrors Position's own tags. TestPositionFilesRoundTrip pins the
// two together, because a field added to one and not the other would
// silently stop being written.
type envelope struct {
	ID        string          `json:"id"`
	V         int             `json:"v"`
	Tags      []string        `json:"tags,omitempty"`
	Note      string          `json:"note,omitempty"`
	Gate      []string        `json:"gate,omitempty"`
	Source    Source          `json:"source"`
	Input     json.RawMessage `json:"input"`
	Expected  Expected        `json:"expected"`
	AtCapture AtCapture       `json:"at_capture"`
}

// inputSentinel stands in for the compact input while the envelope is
// being indented, and is spliced out afterwards. It is not a string
// any move label or card name can be.
const inputSentinel = `"@@suite-input@@"`

// Marshal renders the position as its on-disk bytes.
func (p Position) Marshal() ([]byte, error) {
	in, err := json.Marshal(p.Input)
	if err != nil {
		return nil, fmt.Errorf("suite: marshal input of %s: %w", p.ID, err)
	}
	blob, err := json.MarshalIndent(envelope{
		ID:        p.ID,
		V:         p.V,
		Tags:      p.Tags,
		Note:      p.Note,
		Gate:      p.Gate,
		Source:    p.Source,
		Input:     json.RawMessage(inputSentinel),
		Expected:  p.Expected,
		AtCapture: p.AtCapture,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("suite: marshal %s: %w", p.ID, err)
	}
	if n := strings.Count(string(blob), inputSentinel); n != 1 {
		return nil, fmt.Errorf("suite: marshal %s: the input placeholder appeared %d times, want 1", p.ID, n)
	}
	return append([]byte(strings.Replace(string(blob), inputSentinel, string(in), 1)), '\n'), nil
}

// Write marshals a position to <dir>/<id>.json, 0600. Harvest uses
// it; so does anyone building a position by hand.
func (p Position) Write(dir string) (string, error) {
	if strings.TrimSpace(p.ID) == "" {
		return "", fmt.Errorf("suite: a position needs an id before it can be written")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("suite: create %s: %w", dir, err)
	}
	blob, err := p.Marshal()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, p.ID+".json")
	if err := os.WriteFile(path, blob, 0o600); err != nil {
		return "", fmt.Errorf("suite: write %s: %w", path, err)
	}
	return path, nil
}
