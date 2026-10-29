package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// search_prompt_keyed.go — #4253 slice 5: the library-search prompt
// (PendingChoiceSearchLibrary) written as data, so a table waiting on a
// search is a restore point.
//
// The prompt resumed through searchResumeFrame, which held the whole
// SearchLibrarySpec. Three of its fields were closures: Pred, Validate and
// Then. What answering the prompt reads of them is narrower than it looks:
//
//   - Pred is read once, when the candidates are listed
//     (SearchLibraryThenForEffect). The prompt carries the list
//     (SearchCards, SearchGraveyardCards, SearchHandCards), and the answer
//     is checked against that list and the live zones, never against the
//     predicate again. All the answer needs from it is whether the search
//     names a quality (CR 701.23b, the floor), which is a bool. So the
//     prompt does not hold the predicate at all: queueSearchChoiceLocked
//     records the bool (namesQuality) and drops Pred, on the live table
//     as on a restored one, and a predicate that read X, a colour, the
//     library or a text-changed word (ThroughTextOf) has already done so.
//   - Validate judges the picked set, so the answer does read it. Its
//     data form is a registered set rule (RegisterSearchRule, ledgered
//     "searchrule <key>") and the number and captured text lens it reads
//     (SearchRule): Myriad Landscape's "share a land type", "different
//     names", "total mana value N or less".
//   - Then is the card's continuation. Its data form, ThenTail, already
//     exists (slices 3 and 4).
//
// A search prompt is written (pendingChoices[].searchPrompt: the spec's
// data half, the quality bool, the rule, the undecided flag and a TakeAll
// search's graveyard finds), and PickThen carries SearchPromptKey, so a
// binary from before this slice, which reads a PickThen it does not know
// on this kind as a revealed-hand pick's, refuses the file.
//
// #4361 S4 finished it: Validate and Then had no production setter left,
// so they are gone from SearchLibrarySpec, and the frame holds exactly
// what a restore point writes (searchResumeFrame), never the spec with
// its predicate. The live table and a restored one answer through the
// same data, and no search frame is counted.

// SearchPromptKey is what a written search prompt's PickThen says. An
// on-disk identity: never renamed, never reused.
const SearchPromptKey = "search-library/prompt"

// SearchPromptWordsKey protects captured rule words from older readers,
// which must refuse this prompt rather than silently validate printed words.
const SearchPromptWordsKey = "search-library/prompt/source-words"

// SearchPromptZonesKey protects an and/or search's zone step (#4519) from
// older readers: one that does not know it would restore the prompt as
// an ordinary one, showing the library's candidates before the searcher
// said they searched the library, and shuffling by the take alone. It
// must refuse the file instead. An on-disk identity.
const SearchPromptZonesKey = "search-library/prompt/zones"

func searchPromptKey(s *searchPromptSnapshot) string {
	if s != nil && s.Zones != "" {
		return SearchPromptZonesKey
	}
	if s != nil && len(s.RuleWords) != 0 {
		return SearchPromptWordsKey
	}
	return SearchPromptKey
}

// SearchRuleFunc is a registered set rule over a search's picks: is
// `picked` (never empty) a legal answer, given the rule's number? It runs
// under the game lock, inside the enumerator's read lock too, so it is a
// function of the cards and the number alone and must capture nothing.
// HasSubtypeWord and HasColorWord on the cards read the searching ability
// through the text lens captured by SearchRule.ThroughTextOf.
type SearchRuleFunc func(picked []Card, n int) bool

// SearchRule names a registered set rule, its number and captured text
// lens. The key is unexported, so the only way to hold one is
// RegisterSearchRule.
type SearchRule struct {
	key string
	// N is the number the rule reads: "total mana value N or less".
	N int
	// words is the captured source text substitution, never a live source.
	words []TextWordChange
}

// Key is the rule's on-disk key.
func (r SearchRule) Key() string { return r.key }

// IsZero reports whether no rule is named.
func (r SearchRule) IsZero() bool { return r.key == "" }

// WithN is the rule reading n.
func (r SearchRule) WithN(n int) SearchRule { r.N = n; return r }

// ThroughTextOf captures the words of the resolving spell or ability for
// this rule. The source may already have left; pass TextReaderForEffect
// so its last-known words are used. The captured data survives the search
// prompt and never rereads a source that could have become a new object.
func (r SearchRule) ThroughTextOf(reader *Card) SearchRule {
	r.words = nil
	if reader != nil && !r.IsZero() {
		r.words = reader.TextWords.entries()
	}
	return r
}

// searchRules is guarded by effectRegistryMu.
var searchRules = map[string]SearchRuleFunc{}

// RegisterSearchRule registers a search's set rule under key. Call it
// once, from a package-level var. Panics on a bad or duplicate key. The
// key is an on-disk identity: never renamed, never reused, ledgered as
// "searchrule <key>".
func RegisterSearchRule(key string, fn SearchRuleFunc) SearchRule {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("search rule", key)
	if fn == nil {
		panic(fmt.Sprintf("game: search rule %q has no function", key))
	}
	if _, dup := searchRules[key]; dup {
		panic(fmt.Sprintf("game: search rule %q is already registered", key))
	}
	searchRules[key] = fn
	return SearchRule{key: key}
}

func lookupSearchRule(key string) (SearchRuleFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := searchRules[key]
	return fn, ok
}

// KnownSearchRule reports whether this binary has a set rule registered
// under key.
func KnownSearchRule(key string) bool { _, ok := lookupSearchRule(key); return ok }

// registeredSearchRuleKeysLocked lists the ledger lines, "searchrule
// <key>". Caller holds effectRegistryMu.
func registeredSearchRuleKeysLocked() []string {
	out := make([]string, 0, len(searchRules))
	for k := range searchRules {
		out = append(out, "searchrule "+k)
	}
	sort.Strings(out)
	return out
}

// setAllows reports whether the spec's registered set rule accepts
// `picked`. An unknown rule key accepts nothing: a binary that cannot
// judge the answer must not take one (restore refuses such a file before
// it gets here).
func (spec SearchLibrarySpec) setAllows(picked []Card) bool {
	if spec.Rule.IsZero() {
		return true
	}
	fn, ok := lookupSearchRule(spec.Rule.key)
	if !ok {
		return false
	}
	if len(spec.Rule.words) != 0 {
		reader := textReaderOf(textWordsFromEntries(spec.Rule.words))
		lensed := make([]Card, len(picked))
		for i, c := range picked {
			lensed[i] = ThroughTextOf(reader, c)
		}
		picked = lensed
	}
	return fn(picked, spec.Rule.N)
}

// newSearchResumeFrame is the frame a search prompt holds for spec: the
// predicate's one remaining fact recorded (quality) and the predicate
// dropped, since nothing after the candidates are listed reads it.
func newSearchResumeFrame(spec SearchLibrarySpec, undecided bool) *searchResumeFrame {
	return &searchResumeFrame{
		data:           searchTailSpecOf(spec),
		quality:        spec.statedQuality(),
		rule:           spec.Rule,
		graveyardFinds: copyUUIDs(spec.graveyardFinds),
		undecided:      undecided,
		zones:          spec.initialZones(),
	}
}

// spec is the SearchLibrarySpec the frame stands for, with no predicate:
// a fresh copy each call, so a caller may change it.
func (f *searchResumeFrame) spec() SearchLibrarySpec {
	spec := f.data.searchSpec()
	spec.namesQuality = f.quality
	spec.Rule = f.rule
	spec.graveyardFinds = copyUUIDs(f.graveyardFinds)
	return spec
}

// searchPromptSnapshot is a search prompt's frame on disk.
type searchPromptSnapshot struct {
	Spec searchTailSpec `json:"spec"`
	// Quality is SearchLibrarySpec.statedQuality: the search names a
	// quality, so it may find nothing (CR 701.23b).
	Quality bool `json:"quality,omitempty"`
	// Rule and RuleN are SearchLibrarySpec.Rule.
	Rule  string `json:"rule,omitempty"`
	RuleN int    `json:"ruleN,omitempty"`
	// RuleWords is the source text lens captured when the search began.
	RuleWords []TextWordChange `json:"ruleWords,omitempty"`
	// Undecided is searchResumeFrame.undecided: a "you may search" not
	// yet begun (#4134).
	Undecided bool `json:"undecided,omitempty"`
	// Zones is searchResumeFrame.zones, an and/or search's zone step
	// (#4519): "asking" while the searcher has not said which zones they
	// search, then "library" (the library was chosen, so it is shuffled)
	// or "other" (the graveyard or hand alone). Absent for a search that
	// does not ask, and for a prompt written before #4519, which restores
	// as one. Its key is SearchPromptZonesKey.
	Zones string `json:"zones,omitempty"`
	// GraveyardFinds is a TakeAll search's graveyard matches (#4033),
	// found with the searcher's picks.
	GraveyardFinds []uuid.UUID `json:"graveyardFinds,omitempty"`
}

// wire is the zone step's on-disk name; "" for a search that does not ask.
func (z searchZones) wire() string {
	switch z {
	case searchZonesAsking:
		return "asking"
	case searchZonesLibrary:
		return "library"
	case searchZonesOther:
		return "other"
	}
	return ""
}

// searchZonesOfWire is wire's inverse. A name this binary does not know
// cannot reach it: the prompt's key (SearchPromptZonesKey) is checked first.
func searchZonesOfWire(s string) searchZones {
	switch s {
	case "asking":
		return searchZonesAsking
	case "library":
		return searchZonesLibrary
	case "other":
		return searchZonesOther
	}
	return searchZonesUnasked
}

func snapshotSearchPrompt(c *PendingChoice) *searchPromptSnapshot {
	if c == nil || c.Kind != PendingChoiceSearchLibrary || c.searchResume == nil {
		return nil
	}
	f := c.searchResume
	return &searchPromptSnapshot{
		Spec:           searchTailSpecOf(f.data.searchSpec()),
		Quality:        f.quality,
		Rule:           f.rule.key,
		RuleN:          f.rule.N,
		RuleWords:      cloneTextWordChanges(f.rule.words),
		Undecided:      f.undecided,
		Zones:          f.zones.wire(),
		GraveyardFinds: copyUUIDs(f.graveyardFinds),
	}
}

func restoreSearchPrompt(s *searchPromptSnapshot) *searchResumeFrame {
	return &searchResumeFrame{
		data:           searchTailSpecOf(s.Spec.searchSpec()),
		quality:        s.Quality,
		rule:           SearchRule{key: s.Rule, N: s.RuleN, words: cloneTextWordChanges(s.RuleWords)},
		graveyardFinds: copyUUIDs(s.GraveyardFinds),
		undecided:      s.Undecided,
		zones:          searchZonesOfWire(s.Zones),
	}
}

// unknownSearchPromptKeys is checkEffectKeys' half for a search prompt:
// the marker, the rule and the card's tail.
func unknownSearchPromptKeys(c *pendingChoiceSnapshot) []string {
	if c.SearchPrompt == nil || c.PickThen != searchPromptKey(c.SearchPrompt) {
		return []string{fmt.Sprintf("search prompt %q", c.PickThen)}
	}
	var unknown []string
	if r := c.SearchPrompt.Rule; r != "" && !KnownSearchRule(r) {
		unknown = append(unknown, "search rule "+r)
	}
	return append(unknown, unknownKeyedTail(c.SearchPrompt.Spec.ThenTail)...)
}
