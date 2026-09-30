// Package tokenart resolves a Scryfall TOKEN printing for a token
// about to be created, so the client has something to draw besides
// its name (ADR 0078, issue #1115).
//
// It is the seam ADR 0078 decision 6 describes: `internal/game`
// cannot import `internal/cards` (the catalog already imports game,
// so that is the direction that has to compile), and neither can
// `internal/cards` import `internal/game` without creating a second
// cycle through the effects catalog. This package sits one level
// above both — it imports `cards` for the Index and `game` for
// Card / ParseTypeLine / TokenArtRequest — and hands `main.go` a
// plain function to assign to `game.TokenArtResolver`.
package tokenart

import (
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// candidate is one pool printing inside an identity bucket: just
// enough to run decision 3c's keyword tie-break and decision 4's
// lowest-UUID pick without holding onto the whole cards.Card.
type candidate struct {
	id       uuid.UUID
	keywords map[string]bool
}

// Resolver answers ADR 0078's matching rule against one loaded
// cards.Index. Safe for concurrent use.
//
// buckets is built once, at New, and never mutated afterward — reads
// need no lock. Only the memo (decision 5) and the warned-once set
// (decision 8) change after construction, so they get their own
// mutex.
type Resolver struct {
	buckets map[string][]candidate
	log     *slog.Logger

	mu     sync.RWMutex
	memo   map[string]string // "" is a legitimate memoised MISS
	warned map[string]bool

	misses atomic.Int64
}

// New builds a Resolver over idx's current token-art pool
// (Index.TokenPrintingPool). idx may be nil — the pool is then empty
// and every Resolve call misses, which is the graceful-degradation
// path CI (no Scryfall dump) and the game package's own tests rely
// on. log may be nil; it defaults to slog.Default().
func New(idx *cards.Index, log *slog.Logger) *Resolver {
	if log == nil {
		log = slog.Default()
	}
	var pool []cards.Card
	if idx != nil {
		pool = idx.TokenPrintingPool()
	}
	return &Resolver{
		buckets: buildBuckets(pool),
		log:     log,
		memo:    make(map[string]string),
		warned:  make(map[string]bool),
	}
}

// HookFunc is the main.go one-liner: nil idx (no dump directory
// configured at all — CMDCTRL_DATA_DIR="") yields a nil func, which
// is exactly what leaves game.TokenArtResolver nil and every token
// on the today's text-fallback path, per ADR 0078 decision 6's "nil
// means no resolver is wired". An idx with an empty pool (dump file
// missing or not yet refreshed) still gets a real Resolver — one
// that always misses, which is the same visible behaviour by a
// different road.
func HookFunc(idx *cards.Index, log *slog.Logger) func(game.TokenArtRequest) string {
	if idx == nil {
		return nil
	}
	return New(idx, log).Resolve
}

// MissCount is decision 8's plain counter: how many Resolve calls,
// across the process's lifetime, found no candidate.
func (r *Resolver) MissCount() int64 { return r.misses.Load() }

// Resolve implements the game.TokenArtResolver hook shape: a
// Scryfall printing id for req's template, or "" for "no art —
// render the name" (decision 8).
func (r *Resolver) Resolve(req game.TokenArtRequest) string {
	key := templateIdentityKey(req.Template)
	memoKey := key + "\x00" + keywordJoin(req.Template.Keywords)

	r.mu.RLock()
	result, cached := r.memo[memoKey]
	r.mu.RUnlock()

	if !cached {
		result = r.pick(key, req.Template.Keywords)
		r.mu.Lock()
		r.memo[memoKey] = result
		r.mu.Unlock()
	}

	// Counted on every creation that renders as text (decision 8), not
	// once per distinct template — a Treasure-heavy game's "how many
	// tokens rendered as text today" answer has to track the table,
	// not the catalog. The memo above only skips the BUCKET SCAN; the
	// WARN below is still once per template (warnOnce's own guard).
	if result == "" {
		r.misses.Add(1)
		r.warnOnce(key, req.Template)
	}
	return result
}

// pick runs decision 3c (keyword preference) and decision 4 (lowest
// UUID) over the identity bucket named by key. cands within a bucket
// are already sorted ascending by UUID string (buildBuckets), so
// "first survivor" is always the right pick.
func (r *Resolver) pick(key string, templateKeywords []string) string {
	cands := r.buckets[key]
	if len(cands) == 0 {
		return ""
	}
	wanted := keywordSet(templateKeywords)
	var narrowed []candidate
	for _, c := range cands {
		if keywordSetsEqual(c.keywords, wanted) {
			narrowed = append(narrowed, c)
		}
	}
	pool := cands
	if len(narrowed) > 0 {
		pool = narrowed
	}
	return pool[0].id.String()
}

// warnOnce logs a WARN the first time a given identity key misses in
// this process (decision 8) — once per distinct template, not once
// per token creation.
func (r *Resolver) warnOnce(key string, t game.Card) {
	r.mu.Lock()
	if r.warned[key] {
		r.mu.Unlock()
		return
	}
	r.warned[key] = true
	r.mu.Unlock()
	r.log.Warn("token art: no matching Scryfall printing",
		"name", t.Name,
		"type_line", t.TypeLine,
		"power", t.Power,
		"toughness", t.Toughness,
		"colors", t.Colors,
	)
}

// buildBuckets is decision 5's one-time bucketing pass: every pool
// printing, grouped by its identity key, sorted ascending by UUID
// string so decision 4's "lowest UUID" is always bucket[0] once
// decision 3c's keyword filter has run.
func buildBuckets(pool []cards.Card) map[string][]candidate {
	buckets := make(map[string][]candidate, len(pool)/4+1)
	for _, c := range pool {
		key := poolIdentityKey(c)
		buckets[key] = append(buckets[key], candidate{
			id:       c.ID,
			keywords: keywordSet(c.Keywords),
		})
	}
	for _, bucket := range buckets {
		sort.Slice(bucket, func(i, j int) bool {
			return bucket[i].id.String() < bucket[j].id.String()
		})
	}
	return buckets
}

// templateIdentityKey is decision 3b run against a token template —
// a game.Card whose Power/Toughness are ints.
func templateIdentityKey(t game.Card) string {
	pt := ""
	if isCreatureTypeLine(t.TypeLine) {
		pt = strconv.Itoa(t.Power) + "/" + strconv.Itoa(t.Toughness)
	}
	return identityKey(t.Name, t.TypeLine, pt, t.Colors)
}

// poolIdentityKey is decision 3b run against a Scryfall pool record
// — power/toughness are Scryfall's own strings ("*" for a CDA, ""
// for a non-creature), never converted to int, so a variable-P/T
// printing can never accidentally match a numeric template (decision
// 3d's Spirit Cleric case).
func poolIdentityKey(c cards.Card) string {
	pt := ""
	if isCreatureTypeLine(c.TypeLine) {
		pt = c.Power + "/" + c.Toughness
	}
	return identityKey(c.Name, c.TypeLine, pt, c.Colors)
}

// identityKey is decision 3b's four-field comparison collapsed into
// one string: normalised name, the (types ∪ subtypes) set with
// supertypes discarded (game.ParseTypeLine — this is what makes the
// template's "Token " prefix and a printing's "Legendary" alike cost
// nothing), the power/toughness pair, and the colour set.
func identityKey(name, typeLine, pt string, colors []string) string {
	_, types, subtypes := game.ParseTypeLine(typeLine)
	typeSet := make([]string, 0, len(types)+len(subtypes))
	typeSet = append(typeSet, types...)
	typeSet = append(typeSet, subtypes...)
	sort.Strings(typeSet)

	colorSet := append([]string(nil), colors...)
	sort.Strings(colorSet)

	return normalizeName(name) + "\x00" + strings.Join(typeSet, ",") + "\x00" + pt + "\x00" + strings.Join(colorSet, ",")
}

// isCreatureTypeLine reports whether the parsed type line carries the
// card type "Creature" — the signal that separates "no printed P/T"
// (an artifact token) from "0/0, printed" (a Construct token), since
// both leave a game.Card's Power/Toughness ints at zero.
func isCreatureTypeLine(typeLine string) bool {
	_, types, _ := game.ParseTypeLine(typeLine)
	for _, t := range types {
		if t == "Creature" {
			return true
		}
	}
	return false
}

// normalizeName lower-cases and collapses whitespace. Deliberately a
// local copy rather than a shared helper: cards.normalizeName is
// unexported, and matching its exact behaviour matters less than
// applying the SAME function to both the pool side and the query
// side, which this does.
func normalizeName(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), " ")
}

// keywordSet lower-cases a keyword list into a set for order- and
// case-insensitive comparison (decision 3c).
func keywordSet(keywords []string) map[string]bool {
	if len(keywords) == 0 {
		return nil
	}
	set := make(map[string]bool, len(keywords))
	for _, k := range keywords {
		set[strings.ToLower(k)] = true
	}
	return set
}

func keywordSetsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// keywordJoin is keywordSet flattened to a stable string for the
// memo key.
func keywordJoin(keywords []string) string {
	set := keywordSet(keywords)
	if len(set) == 0 {
		return ""
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}
