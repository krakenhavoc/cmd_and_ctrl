package game

import "github.com/google/uuid"

// search_choice.go — the S22 search chooser's prompt half: the
// continuation frame a PendingChoiceSearchLibrary carries and the
// resolve entry point that answers it.
//
// Its own file rather than pending_choice.go so that concurrent
// branches adding their own PendingChoice kinds do not collide on
// one file; the engine-side search machinery itself lives in
// effect_api.go next to SearchLibraryThenForEffect.

// searchResumeFrame is the continuation for a search prompt. It
// holds the whole SearchLibrarySpec rather than a handful of copied
// fields so the resolve path and the no-prompt path run through
// exactly the same code with exactly the same settings — a fetched
// land must not enter differently depending on whether the searcher
// happened to be offered a choice.
type searchResumeFrame struct {
	spec SearchLibrarySpec

	// undecided is set on a "you may search" prompt (#4134): the
	// searcher has not yet said whether they search, so the candidates
	// are not shown, no control is recorded, and no card pick is
	// accepted. BeginSearchLibrary swaps
	// the frame for a decided one (a new pointer, never a write through
	// this one: an undo snapshot shares it).
	undecided bool
}

// ResolveSearchLibrary answers a PendingChoiceSearchLibrary (CR
// 701.23): `picks` are the cards the searcher takes, chosen from the
// candidates the prompt offered.
//
// An empty `picks` is a legal answer, not a client bug — CR 701.23b
// lets a player fail to find no matter what their library holds, when
// the search names a quality. The search still happened, so
// EventSearchLibrary still fires and the library is still shuffled.
// That is NOT what declining a "you MAY search" means: choosing not to
// search at all is DeclineSearchLibrary (#4134), which skips the
// shuffle and the event. An empty `picks` is NOT legal for an unconditional search
// (CR 701.23d, #3053): choice.SearchMin is the floor, and a short
// answer is refused with ErrInvalidParam like any other illegal set.
//
// Validation rejects before dequeuing, so a client that submits an
// illegal set (too many cards, a card that was never a candidate, a
// pair that breaks Myriad Landscape's share-a-land-type clause) gets
// an error and can try again rather than losing the prompt.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveSearchLibrary(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
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
	if choice.Kind != PendingChoiceSearchLibrary {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if choice.searchResume == nil {
		return ErrInvalidParam
	}
	spec := choice.searchResume.spec
	// #1230: the library read here is the SEARCHED one, not necessarily
	// the chooser's own — Bribery's caster picks, but the pile is
	// target opponent's. chooserID is still who must be answering
	// (checked above) and who checkSearchPicksLocked reports errors
	// against; p is which library the picks are checked and taken
	// from.
	p := g.playerByIDLocked(spec.libraryOwnerID())
	if p == nil {
		return ErrPlayerNotFound
	}
	if err := g.checkSearchPicksLocked(choice, p, picks); err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)

	// The take finishes the search itself — including the shuffle and
	// spec.Then — because a battlefield destination is an ENTRY and an
	// entry can pause on its own prompt (#478). Answering THIS prompt
	// may therefore queue the next one rather than finishing the card.
	err := g.executeSearchTakeLocked(spec, p, picks)
	if err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    chooserID,
			Source:   choice.Source,
			ErrorMsg: err.Error(),
		})
	}
	g.runStateChecksLocked()
	return nil
}

// checkSearchPicksLocked is the ONE copy of "would this answer be
// accepted for this search prompt". ResolveSearchLibrary runs it
// before it dequeues; internal/legal runs it — through
// SearchPickLegalLocked — before it OFFERS a combination, so the
// enumerator cannot advertise a pick the resolver will refuse.
//
// It is one function rather than two deliberately. The predicate a
// search enforces is card text (Myriad Landscape's "share a land
// type"), and a second copy of card text drifts: Game.CounterSpell
// and counterSpellLocked drifted on flashback exactly that way. The
// enumerator therefore does not re-derive the rule, it asks.
//
// Caller must hold g.mu.
func (g *Game) checkSearchPicksLocked(choice *PendingChoice, p *Player, picks []uuid.UUID) error {
	if choice == nil || choice.searchResume == nil {
		return ErrInvalidParam
	}
	if choice.searchResume.undecided {
		// #4134: the searcher has not begun to search; there is nothing
		// to take yet.
		return ErrInvalidParam
	}
	if len(picks) > choice.SearchMax {
		return ErrInvalidParam
	}
	// The floor (#3053, CR 701.23d). It is lowered by candidates that
	// have left the library since the prompt opened (another seat's
	// trigger milled one): the rule says "as many as possible", and
	// refusing an answer nobody can give would wedge the table.
	if len(picks) < g.searchFloorLocked(choice, p) {
		return ErrInvalidParam
	}
	candidates := make(map[uuid.UUID]bool, len(choice.SearchCards))
	for _, id := range choice.SearchCards {
		candidates[id] = true
	}
	// #3217: which of the candidates were offered from the graveyard.
	// A candidate is checked against the zone it was offered from: a
	// library card that was milled while the prompt was open is no
	// longer a legal pick, and must not become one by turning up in
	// the other pile.
	fromGraveyard := make(map[uuid.UUID]bool, len(choice.SearchGraveyardCards))
	for _, id := range choice.SearchGraveyardCards {
		fromGraveyard[id] = true
	}
	// #3988: and which from the searched hand.
	fromHand := make(map[uuid.UUID]bool, len(choice.SearchHandCards))
	for _, id := range choice.SearchHandCards {
		fromHand[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(picks))
	chosen := make([]Card, 0, len(picks))
	for _, id := range picks {
		if !candidates[id] || seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		// The board can move between prompt and answer (another
		// player's trigger milling the library, say), so every pick
		// is re-checked against the live library rather than trusted
		// from the frozen candidate list.
		card, ok := g.LookupCardForEffect(id)
		if !ok {
			return ErrCardNotFound
		}
		switch {
		case fromGraveyard[id]:
			if !p.Graveyard.Contains(id) {
				return ErrCardNotFound
			}
		case fromHand[id]:
			if !p.Hand.Contains(id) {
				return ErrCardNotFound
			}
		case !p.Library.Contains(id):
			return ErrCardNotFound
		}
		chosen = append(chosen, card)
	}
	spec := choice.searchResume.spec
	if spec.Validate != nil && len(chosen) > 0 && !spec.Validate(chosen) {
		return ErrInvalidParam
	}
	return nil
}

// searchFloorLocked is the fewest cards `choice` will accept right now:
// its SearchMin, capped by how many of the candidates are still in the
// searched library. Zero for every search for a stated quality.
//
// Caller must hold g.mu.
func (g *Game) searchFloorLocked(choice *PendingChoice, p *Player) int {
	if choice == nil {
		return 0
	}
	inLibrary, inGraveyard := 0, 0
	inGY := make(map[uuid.UUID]bool, len(choice.SearchGraveyardCards)+len(choice.SearchHandCards))
	for _, id := range choice.SearchGraveyardCards {
		inGY[id] = true
		if p.Graveyard.Contains(id) {
			inGraveyard++
		}
	}
	// #3988: a hand candidate counts with the graveyard's, as a card
	// outside the library.
	for _, id := range choice.SearchHandCards {
		inGY[id] = true
		if p.Hand.Contains(id) {
			inGraveyard++
		}
	}
	if len(choice.SearchGraveyardCards)+len(choice.SearchHandCards) > 0 && choice.searchResume != nil {
		// #3217: an "and/or graveyard" search recomputes its floor from
		// the live zones: an unconditional one must find min(limit,
		// cards in BOTH zones), and those cards may have left.
		for _, id := range choice.SearchCards {
			if !inGY[id] && p.Library.Contains(id) {
				inLibrary++
			}
		}
		return choice.searchResume.spec.floorZones(inLibrary, inGraveyard)
	}
	if choice.SearchMin <= 0 {
		return 0
	}
	for _, id := range choice.SearchCards {
		if p.Library.Contains(id) {
			inLibrary++
		}
	}
	return min(choice.SearchMin, inLibrary)
}

// SearchesGraveyardHandAndLibrary reports whether a search prompt is
// "search that player's graveyard, hand, and library" (#3988): all three
// zones are searched, so the library is shuffled whatever is taken, and
// the client says so rather than "and/or graveyard"'s hint.
func (c *PendingChoice) SearchesGraveyardHandAndLibrary() bool {
	return c != nil && c.searchResume != nil && c.searchResume.spec.AlsoGraveyardAndHand
}

// SearchFloorLocked is searchFloorLocked for the enumerator: how few
// cards an answer to `choice` may name. Zero means "find nothing" is a
// legal answer.
//
// Caller must hold g.mu.
func (g *Game) SearchFloorLocked(choice *PendingChoice) int {
	if choice == nil || choice.searchResume == nil {
		return 0
	}
	p := g.playerByIDLocked(choice.searchResume.spec.libraryOwnerID())
	if p == nil {
		return 0
	}
	return g.searchFloorLocked(choice, p)
}

// SearchPickLegalLocked reports whether `picks` is an answer
// ResolveSearchLibrary would accept for `choice` right now. It is
// the enumerator's read-only window onto the search's Validate hook,
// which rides on the unexported continuation frame and must stay
// there: the hook is a closure over the effect, not wire state, and
// handing internal/legal the frame itself would make every future
// continuation field part of the enumerator's contract.
//
// An empty `picks` is legal for a search for a stated quality — CR
// 701.23b lets a player fail to find — which is what gives a stuck seat
// an answer to fall back on (see legal.Move.AlwaysLegal). An
// unconditional search has a floor (CR 701.23d, #3053) and refuses it.
//
// Caller must hold g's read lock; internal/legal calls it from
// inside ReadSnapshot, like every other *ForEffect / *Locked surface
// the enumerator uses.
func (g *Game) SearchPickLegalLocked(choice *PendingChoice, picks []uuid.UUID) bool {
	if choice == nil || choice.Kind != PendingChoiceSearchLibrary || choice.searchResume == nil {
		return false
	}
	// Same read as ResolveSearchLibrary: the library checked is the
	// spec's, which for a foreign search (Bribery) is not the
	// chooser's own.
	p := g.playerByIDLocked(choice.searchResume.spec.libraryOwnerID())
	if p == nil {
		return false
	}
	return g.checkSearchPicksLocked(choice, p, picks) == nil
}

// SearchIsOptional reports whether the prompt is a "you may search"
// (#4134): the searcher may decline the search outright, which is not
// the same answer as finding nothing.
func (c *PendingChoice) SearchIsOptional() bool {
	return c != nil && c.Kind == PendingChoiceSearchLibrary && c.searchResume != nil &&
		c.searchResume.spec.Optional
}

// SearchIsUndecided reports whether the prompt is still asking "do you
// search?" (#4134): an Optional search, before the searcher has said
// yes. No candidates are shown yet, and "don't search" is still on the
// table; neither is true afterwards.
func (c *PendingChoice) SearchIsUndecided() bool {
	return c != nil && c.Kind == PendingChoiceSearchLibrary && c.searchResume != nil &&
		c.searchResume.undecided
}

// SearchDeclineLegalLocked reports whether "don't search" is an answer
// to `choice` right now (#4134): the search is a "you may" and has not
// begun. Looking through the library is searching (CR 701.19), so once
// the searcher has said yes, and so seen the candidates, there is no
// way back: declining would be a free look that never shuffles. For the
// same reason an Opposition Agent's controller, who takes over only
// after the search begins, can never decline for the searcher.
//
// Caller must hold g's lock (read or write).
func (g *Game) SearchDeclineLegalLocked(choice *PendingChoice) bool {
	return choice.SearchIsUndecided() && choice.searchResume.spec.Optional
}

// DeclineSearchLibrary answers a "you may search" prompt with "I don't
// search" (#4134, CR 701.19). The search never happens: the library is
// not looked at and not shuffled, no EventSearchLibrary fires (so an
// Archivist of Oghma stays quiet), and an Opposition Agent never gets
// control. Only the rest of the effect runs (spec.Then, with no cards),
// the way every declined "if you do" lead-in leaves the instructions
// after it standing: Volatile Fault's Treasure is created either way.
//
// This is NOT the empty answer to ResolveSearchLibrary, which is a
// search that finds nothing (CR 701.23b) and still shuffles. Refused
// with ErrInvalidParam for a prompt that is not Optional, and for one
// whose search has already begun (the candidates were seen).
//
// Caller must NOT hold g.mu.
func (g *Game) DeclineSearchLibrary(choiceID, chooserID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := g.choiceIndexLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceSearchLibrary || choice.searchResume == nil {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if !g.SearchDeclineLegalLocked(choice) {
		return ErrInvalidParam
	}
	spec := choice.searchResume.spec
	g.dequeueChoiceLocked(idx)
	if spec.Then != nil || !spec.ThenTail.IsZero() {
		if err := g.runSearchThenLocked(spec, nil); err != nil {
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

// BeginSearchLibrary answers an undecided "you may search" prompt with
// "yes, I search" (#4134). Only now does the search start: the
// candidates become known to the searcher and an Opposition Agent's
// control begins, so the next answer is the Agent controller's. Refused
// with ErrInvalidParam on a prompt that is not undecided.
//
// Caller must NOT hold g.mu.
func (g *Game) BeginSearchLibrary(choiceID, chooserID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := g.choiceIndexLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceSearchLibrary || choice.searchResume == nil {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if !choice.searchResume.undecided {
		return ErrInvalidParam
	}
	spec := choice.searchResume.spec
	p := g.playerByIDLocked(spec.libraryOwnerID())
	if p == nil {
		return ErrPlayerNotFound
	}
	g.markSearchCandidatesKnownLocked(spec.Player, p, choice.SearchCards)
	choice.searchResume = &searchResumeFrame{spec: spec}
	if spec.libraryOwnerID() == spec.Player {
		g.beginSearchControlsLocked(spec.Player, choice.ID)
	}
	return nil
}
