package game

import (
	"strings"

	"github.com/google/uuid"
)

// block_requirement_filtered.go is #6087's block requirement: "must be
// blocked by a Dalek if able" (Ace's Baseball Bat). It is "must be blocked
// if able" (BlockRequirementMustBeBlocked) narrowed by a blocker FILTER, the
// narrowing Marble Priest's Lure already carries (BlockRequirement.Filter):
// one requirement on the attacker, obeyed when at least one blocker the
// filter matches blocks it. A blocker the filter does not match may block
// the attacker freely and obeys nothing.
//
// CR 509.1c: "The defending player checks each creature they control to see
// whether it's affected by any requirements (effects that say a creature
// must block, or that it must block if some condition is met). If the number
// of requirements that are being obeyed is fewer than the maximum possible
// number of requirements that could be obeyed without disobeying any
// restrictions, the declaration of blockers is illegal."
//
// So the search (bestBlockRequirementAdditionLocked) has to see it: a filtered
// requirement is worth one, once, to the first MATCHING blocker an attacker
// takes. The flow gives such an attacker a second node: matching blockers
// reach the attacker through it, and its first unit carries the bonus, the
// way an unfiltered "must be blocked" carries it on the attacker's first
// unit. One filter per attacker is modelled in the flow (the first one it
// carries); a second, different filter on the same attacker is still counted
// exactly in the witness, so the search can only under-find it, the weaker
// direction. No printed card puts two on one creature.
//
// # Subtype filters
//
// A filter key "subtype:<Type>" matches a blocker with that subtype, read
// off its effective characteristics at the declaration (changeling
// included). It is self-describing data, so a card needs no registration,
// and the type can be written through the source's text (CR 612.2).

// blockerFilterSubtypePrefix starts a subtype filter key.
const blockerFilterSubtypePrefix = "subtype:"

// BlockerFilterSubtype is the filter key for blockers with subtype `sub`:
// "must be blocked by a Dalek if able" carries BlockerFilterSubtype("Dalek").
func BlockerFilterSubtype(sub string) string { return blockerFilterSubtypePrefix + sub }

// blockerFilterSubtype is the subtype a subtype filter key names, or "".
func blockerFilterSubtype(key string) string {
	if sub, ok := strings.CutPrefix(key, blockerFilterSubtypePrefix); ok {
		return sub
	}
	return ""
}

// BlockerFilterNoun is the filter as a sentence names one blocker it matches
// ("a Dalek", "a Wall", "a creature with flying"), or "" for no filter.
func BlockerFilterNoun(key string) string {
	switch key {
	case "":
		return ""
	case BlockerFilterWall:
		return "a Wall"
	case BlockerFilterFlying:
		return "a creature with flying"
	}
	if sub := blockerFilterSubtype(key); sub != "" {
		return articleForWord(sub) + " " + sub
	}
	return "a certain creature"
}

// articleForWord is "a" or "an" for a word.
func articleForWord(w string) string {
	if w != "" && strings.ContainsRune("AEIOUaeiou", rune(w[0])) {
		return "an"
	}
	return "a"
}

// blockerFilterMatch is the predicate behind a filter key: a subtype key's
// own, else the registered one. Nil for an unknown key.
func blockerFilterMatch(key string) func(blocker *Card) bool {
	if sub := blockerFilterSubtype(key); sub != "" {
		return func(b *Card) bool { return b.HasSubtype(sub) }
	}
	blockerFilters.RLock()
	defer blockerFilters.RUnlock()
	return blockerFilters.byKey[key]
}

// filteredMustBeBlocked reports whether r is a "must be blocked by <filter>
// if able" requirement.
func (r BlockRequirement) filteredMustBeBlocked() bool {
	return r.Kind == BlockRequirementMustBeBlocked && r.Filter != ""
}

// mustBeBlockedObeyedBy reports whether "must be blocked [by <filter>] if
// able" on `atk` is obeyed by `assign`, and the first blocker that obeys it.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) mustBeBlockedObeyedBy(r BlockRequirement, atk *Card, assign blockAssignment) (bool, uuid.UUID) {
	for i := range g.Battlefield.Cards {
		b := &g.Battlefield.Cards[i]
		if !assign.has(b.InstanceID, atk.InstanceID) {
			continue
		}
		if r.Filter == "" || r.bindsBlocker(b) {
			return true, b.InstanceID
		}
	}
	return false, uuid.Nil
}

// filteredMustBeBlockedOf is the first filter key among `reqs`' filtered
// "must be blocked" requirements, how many of them carry that key, and
// whether `assign` already has a matching blocker on `atk` — the search's
// view of one attacker.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) filteredMustBeBlockedOf(atk *Card, reqs []BlockRequirement, assign blockAssignment) (key string, n int, obeyed bool) {
	for _, r := range reqs {
		if !r.filteredMustBeBlocked() {
			continue
		}
		if key == "" {
			key = r.Filter
			obeyed, _ = g.mustBeBlockedObeyedBy(r, atk, assign)
		}
		if r.Filter == key {
			n++
		}
	}
	return key, n, obeyed
}

// unfilteredMustBeBlockedCount is how many plain "must be blocked if able"
// requirements `reqs` carry: the ones any blocker obeys.
func unfilteredMustBeBlockedCount(reqs []BlockRequirement) int {
	n := 0
	for _, r := range reqs {
		if r.Kind == BlockRequirementMustBeBlocked && r.Filter == "" {
			n++
		}
	}
	return n
}
