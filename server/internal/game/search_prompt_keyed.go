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
// The prompt resumed through searchResumeFrame, which holds the whole
// SearchLibrarySpec. Three of its fields are closures: Pred, Validate and
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
//     "searchrule <key>") and the one number it reads (SearchRule.N): Myriad
//     Landscape's "share a land type", "different names", "total mana
//     value N or less".
//   - Then is the card's continuation. Its data form, ThenTail, already
//     exists (slices 3 and 4).
//
// A prompt whose spec has no Validate and no Then is written
// (pendingChoices[].searchPrompt: the spec's data half, the quality bool,
// the rule, the undecided flag and a TakeAll search's graveyard finds), and
// PickThen carries SearchPromptKey, so a binary from before this slice,
// which reads a PickThen it does not know on this kind as a revealed-hand
// pick's, refuses the file. A spec with a closure Validate or Then keeps
// the closure form, and its frame is still counted.

// SearchPromptKey is what a written search prompt's PickThen says. An
// on-disk identity: never renamed, never reused.
const SearchPromptKey = "search-library/prompt"

// SearchRuleFunc is a registered set rule over a search's picks: is
// `picked` (never empty) a legal answer, given the rule's number? It runs
// under the game lock, inside the enumerator's read lock too, so it is a
// function of the cards and the number alone and must capture nothing.
type SearchRuleFunc func(picked []Card, n int) bool

// SearchRule names a registered set rule and the number it reads. The key
// is unexported, so the only way to hold one is RegisterSearchRule.
type SearchRule struct {
	key string
	// N is the number the rule reads: "total mana value N or less".
	N int
}

// Key is the rule's on-disk key.
func (r SearchRule) Key() string { return r.key }

// IsZero reports whether no rule is named.
func (r SearchRule) IsZero() bool { return r.key == "" }

// WithN is the rule reading n.
func (r SearchRule) WithN(n int) SearchRule { r.N = n; return r }

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

// setAllows reports whether the spec's set rule, closure or registered,
// accepts `picked`. An unknown rule key accepts nothing: a binary that
// cannot judge the answer must not take one (restore refuses such a file
// before it gets here).
func (spec SearchLibrarySpec) setAllows(picked []Card) bool {
	if spec.Validate != nil && !spec.Validate(picked) {
		return false
	}
	if spec.Rule.IsZero() {
		return true
	}
	fn, ok := lookupSearchRule(spec.Rule.key)
	return ok && fn(picked, spec.Rule.N)
}

// forPrompt is the spec a search prompt holds: the predicate's one
// remaining fact recorded (namesQuality) and the predicate dropped, since
// nothing after the candidates are listed reads it. The live table and a
// restored one then answer through the same spec.
func (spec SearchLibrarySpec) forPrompt() SearchLibrarySpec {
	spec.namesQuality = spec.statedQuality()
	spec.Pred = nil
	return spec
}

// searchPromptIsData reports whether c is a search prompt whose frame can
// be written: no closure set rule and no closure continuation.
func searchPromptIsData(c *PendingChoice) bool {
	if c == nil || c.Kind != PendingChoiceSearchLibrary || c.searchResume == nil {
		return false
	}
	spec := c.searchResume.spec
	return spec.Pred == nil && spec.Validate == nil && spec.Then == nil
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
	// Undecided is searchResumeFrame.undecided: a "you may search" not
	// yet begun (#4134).
	Undecided bool `json:"undecided,omitempty"`
	// GraveyardFinds is a TakeAll search's graveyard matches (#4033),
	// found with the searcher's picks.
	GraveyardFinds []uuid.UUID `json:"graveyardFinds,omitempty"`
}

func snapshotSearchPrompt(c *PendingChoice) *searchPromptSnapshot {
	if !searchPromptIsData(c) {
		return nil
	}
	spec := c.searchResume.spec
	return &searchPromptSnapshot{
		Spec:           searchTailSpecOf(spec),
		Quality:        spec.statedQuality(),
		Rule:           spec.Rule.key,
		RuleN:          spec.Rule.N,
		Undecided:      c.searchResume.undecided,
		GraveyardFinds: copyUUIDs(spec.graveyardFinds),
	}
}

func restoreSearchPrompt(s *searchPromptSnapshot) *searchResumeFrame {
	spec := s.Spec.searchSpec()
	spec.namesQuality = s.Quality
	spec.Rule = SearchRule{key: s.Rule, N: s.RuleN}
	spec.graveyardFinds = copyUUIDs(s.GraveyardFinds)
	return &searchResumeFrame{spec: spec, undecided: s.Undecided}
}

// unknownSearchPromptKeys is checkEffectKeys' half for a search prompt:
// the marker, the rule and the card's tail.
func unknownSearchPromptKeys(c *pendingChoiceSnapshot) []string {
	if c.SearchPrompt == nil || c.PickThen != SearchPromptKey {
		return []string{fmt.Sprintf("search prompt %q", c.PickThen)}
	}
	var unknown []string
	if r := c.SearchPrompt.Rule; r != "" && !KnownSearchRule(r) {
		unknown = append(unknown, "search rule "+r)
	}
	return append(unknown, unknownKeyedTail(c.SearchPrompt.Spec.ThenTail)...)
}
