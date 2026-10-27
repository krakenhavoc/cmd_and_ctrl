package game

import "github.com/google/uuid"

// search_entry_keyed.go — #4253 slice 3: a library search's remaining
// work, carried as data on a fetched permanent's entry tail.
//
// A battlefield search puts its takes onto the battlefield one at a
// time, each from the previous entry's continuation, and finishes the
// search (EventSearchLibrary, the shuffle, the card's own "then") from
// the last one (searchEnterEachThenFinishLocked, #478). That
// continuation was a closure over the spec and the two lists, so a
// fetchland whose shockland asks "pay 2 life?" was not a restore point.
//
// Once the cards are taken, nothing the rest of the search does reads
// the spec's closures: the predicate and the validator judged the pick,
// which is over. So the rest of the search is the spec's data half
// (searchTailSpec), the takes not yet entered and the cards that have,
// and the card's own "then" when it is a registered tail
// (SearchLibrarySpec.ThenTail). A search whose card passes a closure
// "then" keeps the closure form, and its frame is still counted.

// searchTailSpec is a SearchLibrarySpec with its closures left out:
// what finishing the search reads after the take. Plain data, written
// with a keyed frame as it stands.
type searchTailSpec struct {
	Player               uuid.UUID          `json:"player"`
	Source               uuid.UUID          `json:"source"`
	Dest                 ZoneKind           `json:"dest,omitempty"`
	FaceDown             FaceDownKind       `json:"faceDown,omitempty"`
	LibraryOwner         uuid.UUID          `json:"libraryOwner"`
	Limit                int                `json:"limit,omitempty"`
	Unbounded            bool               `json:"unbounded,omitempty"`
	Reveal               bool               `json:"reveal,omitempty"`
	Shuffle              bool               `json:"shuffle,omitempty"`
	TappedOnEntry        bool               `json:"tappedOnEntry,omitempty"`
	EntersWithCounters   map[string]int     `json:"entersWithCounters,omitempty"`
	EntersAttachedTo     *TargetRef         `json:"entersAttachedTo,omitempty"`
	LinkTo               PermissionCardRef  `json:"linkTo"`
	Optional             bool               `json:"optional,omitempty"`
	UpTo                 bool               `json:"upTo,omitempty"`
	Reason               string             `json:"reason,omitempty"`
	ThenTail             *keyedTailSnapshot `json:"thenTail,omitempty"`
	ToTop                bool               `json:"toTop,omitempty"`
	Depth                int                `json:"depth,omitempty"`
	AlsoGraveyard        bool               `json:"alsoGraveyard,omitempty"`
	AlsoGraveyardAndHand bool               `json:"alsoGraveyardAndHand,omitempty"`
	TakeAll              bool               `json:"takeAll,omitempty"`
	LibraryUntouched     bool               `json:"libraryUntouched,omitempty"`
	ExileFoundFor        uuid.UUID          `json:"exileFoundFor"`
}

// searchEntryRun is the rest of a battlefield search, riding the entry
// tail of the take now entering: the spec's data half, the takes still
// to enter, and the permanents that have. Never written after it is
// built (each step builds the next one), so an undo snapshot may share
// it.
type searchEntryRun struct {
	Spec  searchTailSpec `json:"spec"`
	Rest  []uuid.UUID    `json:"rest,omitempty"`
	Found []uuid.UUID    `json:"found,omitempty"`
}

func searchTailSpecOf(spec SearchLibrarySpec) searchTailSpec {
	out := searchTailSpec{
		Player: spec.Player, Source: spec.Source, Dest: spec.Dest, FaceDown: spec.FaceDown,
		LibraryOwner: spec.LibraryOwner, Limit: spec.Limit, Unbounded: spec.Unbounded,
		Reveal: spec.Reveal, Shuffle: spec.Shuffle, TappedOnEntry: spec.TappedOnEntry,
		EntersWithCounters: copyCounterMap(spec.EntersWithCounters), LinkTo: spec.LinkTo,
		Optional: spec.Optional, UpTo: spec.UpTo, Reason: spec.Reason,
		ThenTail: snapshotKeyedTail(spec.ThenTail), ToTop: spec.ToTop, Depth: spec.Depth,
		AlsoGraveyard: spec.AlsoGraveyard, AlsoGraveyardAndHand: spec.AlsoGraveyardAndHand,
		TakeAll: spec.TakeAll, LibraryUntouched: spec.libraryUntouched, ExileFoundFor: spec.exileFoundFor,
	}
	if spec.EntersAttachedTo != (TargetRef{}) {
		h := spec.EntersAttachedTo
		out.EntersAttachedTo = &h
	}
	return out
}

// searchSpec is searchTailSpecOf's inverse, with no closures.
func (s searchTailSpec) searchSpec() SearchLibrarySpec {
	out := SearchLibrarySpec{
		Player: s.Player, Source: s.Source, Dest: s.Dest, FaceDown: s.FaceDown,
		LibraryOwner: s.LibraryOwner, Limit: s.Limit, Unbounded: s.Unbounded,
		Reveal: s.Reveal, Shuffle: s.Shuffle, TappedOnEntry: s.TappedOnEntry,
		EntersWithCounters: copyCounterMap(s.EntersWithCounters), LinkTo: s.LinkTo,
		Optional: s.Optional, UpTo: s.UpTo, Reason: s.Reason,
		ThenTail: restoreKeyedTail(s.ThenTail), ToTop: s.ToTop, Depth: s.Depth,
		AlsoGraveyard: s.AlsoGraveyard, AlsoGraveyardAndHand: s.AlsoGraveyardAndHand,
		TakeAll: s.TakeAll, libraryUntouched: s.LibraryUntouched, exileFoundFor: s.ExileFoundFor,
	}
	if s.EntersAttachedTo != nil {
		out.EntersAttachedTo = *s.EntersAttachedTo
	}
	return out
}

// cloneSearchEntryRun is a deep copy, for the mirror.
func cloneSearchEntryRun(r *searchEntryRun) *searchEntryRun {
	if r == nil {
		return nil
	}
	out := *r
	out.Spec.EntersWithCounters = copyCounterMap(r.Spec.EntersWithCounters)
	if r.Spec.EntersAttachedTo != nil {
		h := *r.Spec.EntersAttachedTo
		out.Spec.EntersAttachedTo = &h
	}
	if r.Spec.ThenTail != nil {
		t := *r.Spec.ThenTail
		t.Carry = copyUUIDs(r.Spec.ThenTail.Carry)
		t.Asked = copyAskedItem(r.Spec.ThenTail.Asked)
		out.Spec.ThenTail = &t
	}
	out.Rest = copyUUIDs(r.Rest)
	out.Found = copyUUIDs(r.Found)
	return &out
}

// runSearchEntryStepLocked is the keyed form of the step closure in
// searchEnterEachThenFinishLocked: count the take that just settled and
// enter the next one, or finish the search.
//
// Caller must hold g.mu.
func (g *Game) runSearchEntryStepLocked(r *searchEntryRun, entered uuid.UUID) error {
	found := r.Found
	if entered != uuid.Nil {
		found = append(append(make([]uuid.UUID, 0, len(r.Found)+1), r.Found...), entered)
	}
	return g.searchEnterEachThenFinishLocked(r.Spec.searchSpec(), copyUUIDs(r.Rest), found)
}

// unknownSearchEntryKeys is checkEffectKeys' half for a keyed search.
func unknownSearchEntryKeys(r *searchEntryRun) []string {
	if r == nil {
		return nil
	}
	return unknownKeyedTail(r.Spec.ThenTail)
}
