package game

import (
	"github.com/google/uuid"
)

// library_order.go — "put them on top / on the bottom of your library
// in any order" (CR 401.4, #996, ADR 0088).
//
// The placement half of the library-top row. The LOOK half already had
// a prompt family (scry, surveil, look_at_top, all on
// lookAtTopForEffect); what it did not have was a way to order a pile
// headed to the bottom, to order cards put on top from somewhere other
// than the top of the library (Brainstorm's hand), or to let a player
// choose top OR bottom without borrowing scry (Aetherspouts). One
// prompt kind covers all three, PendingChoicePutInLibrary, and it is a
// fourth member of the scry family so the view, the dispatcher and the
// client dialog serve it unchanged.

// LibraryPlacement is where a PendingChoicePutInLibrary's cards may go:
// which lanes of the answer are open.
type LibraryPlacement string

const (
	// LibraryPlaceTop — "put them on top of your library in any
	// order". Answered with top_order alone.
	LibraryPlaceTop LibraryPlacement = "top"
	// LibraryPlaceBottom — "put the rest on the bottom of your library
	// in any order". Answered with bottom alone, TOP-FIRST.
	LibraryPlaceBottom LibraryPlacement = "bottom"
	// LibraryPlaceTopOrBottom — "put it on your choice of the top or
	// bottom of its owner's library". Answered with both lanes, each
	// top-first, every card in exactly one.
	LibraryPlaceTopOrBottom LibraryPlacement = "top_or_bottom"
)

// allowsTop / allowsBottom report which answer lanes a placement opens.
func (p LibraryPlacement) allowsTop() bool {
	return p == LibraryPlaceTop || p == LibraryPlaceTopOrBottom
}

func (p LibraryPlacement) allowsBottom() bool {
	return p == LibraryPlaceBottom || p == LibraryPlaceTopOrBottom
}

// valid reports whether p is one of the three placements.
func (p LibraryPlacement) valid() bool {
	return p.allowsTop() || p.allowsBottom()
}

// PutInLibrarySpec is one "put these cards into their owners' libraries
// in the order you choose" instruction.
type PutInLibrarySpec struct {
	// Chooser orders the pile — "you" on every printed card except
	// Aetherspouts, whose owners each choose for their own cards (the
	// card raises one prompt per owner).
	Chooser uuid.UUID

	// Source is the card whose instruction this is — prompt context.
	Source uuid.UUID

	// Cards are the cards to place, in the order the chooser should see
	// them first (top-first for cards already in a library). Each goes
	// to its OWNER's library. A repeated ID is placed once.
	Cards []uuid.UUID

	// From is the zone kind the cards are in now: ZoneLibrary for "the
	// rest" of a look or a reveal, ZoneHand for Brainstorm. A card that
	// is no longer in a zone of that kind — when the prompt is raised
	// or when the answer arrives — is skipped: it is a new object the
	// instruction no longer refers to (CR 400.7). Empty accepts any
	// zone.
	From ZoneKind

	// Placement is which lanes the answer may use.
	Placement LibraryPlacement

	// Reason is the prompt banner — "<card> — <the printed clause>".
	Reason string

	// Then is the rest of the effect. It runs once every card has been
	// placed (or skipped), including immediately when there was nothing
	// to ask. Runs with g.mu held; must capture only scalars.
	Then func(g *Game) error
}

// PutInLibraryInChosenOrderThenForEffect raises the ordering prompt for
// spec, or places the cards at once when there is nothing to choose.
//
// No prompt when:
//   - no card of spec.Cards is still in a spec.From zone: Then runs;
//   - one card remains and the placement opens one lane: it is placed
//     where the card says, since a pile of one has no order.
//
// A top_or_bottom placement always asks, even about one card — "top or
// bottom" is the choice.
//
// The chooser is made a knower of every card in the prompt, because
// ordering cards you cannot see is not a choice; nobody else is. See
// ADR 0088 Decision 3 for what they know afterwards.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PutInLibraryInChosenOrderThenForEffect(spec PutInLibrarySpec) error {
	if !spec.Placement.valid() {
		return ErrInvalidParam
	}
	cards := g.libraryOrderLiveCardsLocked(spec.From, spec.Cards)
	then := spec.Then
	finish := func(g *Game) error {
		if then == nil {
			return nil
		}
		return then(g)
	}
	if len(cards) == 0 {
		return finish(g)
	}
	chooser, from, placement := spec.Chooser, spec.From, spec.Placement
	if len(cards) == 1 && placement != LibraryPlaceTopOrBottom {
		var top, bottom []uuid.UUID
		if placement == LibraryPlaceTop {
			top = cards
		} else {
			bottom = cards
		}
		return g.placeInLibraryInOrderLocked(chooser, from, top, bottom, finish)
	}
	for _, id := range cards {
		if c := g.findCardByIDLocked(id); c != nil {
			c.AddKnower(chooser)
		}
	}
	reason := spec.Reason
	if reason == "" {
		reason = defaultPutInLibraryReason(placement)
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoicePutInLibrary,
		Chooser:          chooser,
		FromPlayer:       chooser,
		Count:            len(cards),
		Source:           spec.Source,
		Reason:           reason,
		ScryCards:        cards,
		LibraryPlacement: placement,
		libraryOrderResume: func(g *Game, top, bottom []uuid.UUID) error {
			return g.placeInLibraryInOrderLocked(chooser, from, top, bottom, finish)
		},
	})
	if id == uuid.Nil {
		// The chooser is not seated (#864). Nobody can answer, so
		// nothing is placed — but the rest of the effect still runs.
		return finish(g)
	}
	return nil
}

// defaultPutInLibraryReason is the banner when the card gave none.
func defaultPutInLibraryReason(p LibraryPlacement) string {
	switch p {
	case LibraryPlaceTop:
		return "Put these cards on top of the library in any order"
	case LibraryPlaceBottom:
		return "Put these cards on the bottom of the library in any order"
	}
	return "Put each card on the top or the bottom of its owner's library"
}

// libraryOrderLiveCardsLocked is `ids` with repeats dropped and every
// card no longer in a zone of kind `from` removed, in the order given.
//
// Caller must hold g.mu.
func (g *Game) libraryOrderLiveCardsLocked(from ZoneKind, ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		if g.libraryOrderCardLiveLocked(from, id) {
			out = append(out, id)
		}
	}
	return out
}

// libraryOrderCardLiveLocked reports whether card `id` is still in a
// zone of kind `from` (any zone, when `from` is empty).
//
// Caller must hold g.mu.
func (g *Game) libraryOrderCardLiveLocked(from ZoneKind, id uuid.UUID) bool {
	z := g.findCardZoneLocked(id)
	if z == nil {
		return false
	}
	return from == "" || z.Kind == from
}

// ResolvePutInLibrary answers a PendingChoicePutInLibrary: `topOrder`
// are the cards going on top of their owners' libraries and `bottom`
// the ones going under them, BOTH listed top-first.
//
// Every card the prompt named must appear in exactly one list, and a
// lane the placement does not open must be empty — a card put in the
// wrong lane is refused rather than moved, because "the rest on the
// bottom" put on top is a different (and much better) card.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolvePutInLibrary(choiceID, chooserID uuid.UUID, bottom, topOrder []uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoicePutInLibrary {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if err := checkPutInLibraryAnswer(choice.LibraryPlacement, choice.ScryCards, bottom, topOrder); err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)
	if resume := choice.libraryOrderResume; resume != nil {
		// Copies, so a replay of the same continuation after an undo
		// never sees this call's slices.
		top := append([]uuid.UUID(nil), topOrder...)
		under := append([]uuid.UUID(nil), bottom...)
		if err := resume(g, top, under); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    chooserID,
				Source:   choice.Source,
				ErrorMsg: err.Error(),
			})
		}
	}
	g.runStateChecksLocked()
	return nil
}

// checkPutInLibraryAnswer validates an answer's shape: the lanes the
// placement opens, and every card exactly once.
func checkPutInLibraryAnswer(p LibraryPlacement, cards, bottom, topOrder []uuid.UUID) error {
	if len(topOrder) > 0 && !p.allowsTop() {
		return ErrInvalidParam
	}
	if len(bottom) > 0 && !p.allowsBottom() {
		return ErrInvalidParam
	}
	_, err := partitionChoiceCards(cards, bottom, topOrder)
	return err
}

// partitionChoiceCards checks that `away` and `topOrder` together name
// every card of `cards` exactly once, and returns `cards` as a lookup.
// The shape rule every member of the scry family shares; the
// library-membership re-check is partitionLookedAtCards' own.
func partitionChoiceCards(cards, away, topOrder []uuid.UUID) (map[uuid.UUID]bool, error) {
	want := make(map[uuid.UUID]bool, len(cards))
	for _, id := range cards {
		want[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(want))
	for _, list := range [][]uuid.UUID{away, topOrder} {
		for _, id := range list {
			if !want[id] || seen[id] {
				return nil, ErrInvalidParam
			}
			seen[id] = true
		}
	}
	if len(seen) != len(want) {
		return nil, ErrInvalidParam
	}
	return want, nil
}

// libraryOrderLeg is one card of a placement: where it goes, and how
// many cards share its lane (which decides who knows it afterwards).
type libraryOrderLeg struct {
	id       uuid.UUID
	toBottom bool
	laneSize int
}

// placeInLibraryInOrderLocked puts `top` on top of their owners'
// libraries and `bottom` under them, each lane top-first, then runs
// `then`.
//
// The legs go in PILE ORDER so the positions come out right: the
// bottom lane first card first (each PushBottom goes under the one
// before, so the first entry ends nearest the top of the pile), then
// the top lane last card first (each PushTop lands above, so the first
// entry ends on top).
//
// A card already in its owner's library is lifted and re-pushed — a
// reorder, not a zone change, so no window opens and no event fires. A
// card anywhere else goes through the tuck route, so the CR 614 window
// opens and a commander is offered the command zone (CR 903.9); the
// next leg runs from that leg's continuation, so a paused leg pauses
// the pile rather than letting the rest overtake it.
//
// Caller must hold g.mu in write mode.
func (g *Game) placeInLibraryInOrderLocked(chooser uuid.UUID, from ZoneKind, top, bottom []uuid.UUID, then func(g *Game) error) error {
	legs := make([]libraryOrderLeg, 0, len(top)+len(bottom))
	for _, id := range bottom {
		legs = append(legs, libraryOrderLeg{id: id, toBottom: true, laneSize: len(bottom)})
	}
	for i := len(top) - 1; i >= 0; i-- {
		legs = append(legs, libraryOrderLeg{id: top[i], laneSize: len(top)})
	}
	return g.placeLibraryOrderLegsLocked(chooser, from, legs, then)
}

// placeLibraryOrderLegsLocked places the head of `legs` and continues
// with the tail from that leg's continuation. The empty list is the
// base case.
//
// Caller must hold g.mu in write mode.
func (g *Game) placeLibraryOrderLegsLocked(chooser uuid.UUID, from ZoneKind, legs []libraryOrderLeg, then func(g *Game) error) error {
	for len(legs) > 0 && !g.libraryOrderCardLiveLocked(from, legs[0].id) {
		legs = legs[1:]
	}
	if len(legs) == 0 {
		if then == nil {
			return nil
		}
		return then(g)
	}
	leg, rest := legs[0], legs[1:]
	next := func(g *Game) error {
		return g.placeLibraryOrderLegsLocked(chooser, from, rest, then)
	}
	src := g.findCardZoneLocked(leg.id)
	card, ok := g.cardInZoneLocked(src, leg.id)
	if !ok {
		return next(g)
	}
	knowers := libraryOrderKnowers(card, chooser, leg.laneSize)
	if src.Kind == ZoneLibrary && src.Owner == card.Owner {
		c, err := src.Remove(leg.id)
		if err != nil {
			return next(g)
		}
		c.KnownBy = knowers
		if leg.toBottom {
			src.PushBottom(c)
		} else {
			src.PushTop(c)
		}
		return next(g)
	}
	err := g.TuckToLibraryThenForEffect(leg.id, TuckOptions{ToBottom: leg.toBottom}, func(g *Game, tucked bool) error {
		if tucked {
			if c := g.findCardByIDLocked(leg.id); c != nil {
				c.KnownBy = copyKnowers(knowers)
			}
		}
		return next(g)
	})
	return err
}

// libraryOrderKnowers is who knows a card after it is placed (ADR 0088
// Decision 3, CR 401.4): the chooser always; everyone who knew it
// before only when it is ALONE in its lane, because the order of two
// or more is not revealed and a knower mark is per card at a position.
func libraryOrderKnowers(c Card, chooser uuid.UUID, laneSize int) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	if laneSize <= 1 {
		for id, known := range c.KnownBy {
			if known {
				out[id] = true
			}
		}
	}
	if chooser != uuid.Nil {
		out[chooser] = true
	}
	return out
}

// copyKnowers gives a knower set its own map, so two cards (or a card
// and the continuation that set it) never share one.
func copyKnowers(in map[uuid.UUID]bool) map[uuid.UUID]bool {
	out := make(map[uuid.UUID]bool, len(in))
	for id, v := range in {
		out[id] = v
	}
	return out
}
