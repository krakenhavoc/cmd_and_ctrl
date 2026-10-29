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
// holds the whole SearchLibrarySpec's data (spec() rebuilds it) rather
// than a handful of copied fields so the resolve path and the no-prompt
// path run through exactly the same code with exactly the same
// settings — a fetched land must not enter differently depending on
// whether the searcher happened to be offered a choice.
//
// #4361 S4: data only, the very fields a restore point writes
// (searchPromptSnapshot), so the live table and a restored one answer
// through the same frame. The predicate is not held: it listed the
// candidates and is not read again (search_prompt_keyed.go).
type searchResumeFrame struct {
	// data is the spec's data half, its continuation included.
	data searchTailSpec
	// quality is SearchLibrarySpec.statedQuality, read before the
	// predicate was dropped (CR 701.23b, the floor).
	quality bool
	// rule is the spec's registered set rule.
	rule SearchRule
	// graveyardFinds is a TakeAll search's graveyard matches (#4033).
	graveyardFinds []uuid.UUID

	// undecided is set on a "you may search" prompt (#4134): the
	// searcher has not yet said whether they search, so the candidates
	// are not shown, no control is recorded, and no card pick is
	// accepted. BeginSearchLibrary swaps
	// the frame for a decided one (a new pointer, never a write through
	// this one: an undo snapshot shares it).
	undecided bool

	// zones is the "which zones do you search?" step of an and/or search
	// (#4519, CR 701.23): see searchZones. A frame is replaced, never
	// written through, when the answer arrives (an undo snapshot shares it).
	zones searchZones
}

// searchZones is where an "and/or" search ("your library and/or graveyard",
// "your hand and/or library") stands on the question of which zones the
// searcher looks in (#4519). The printed text shuffles only "if you search
// your library this way" (CR 701.23), and looking through the library is
// searching it (CR 701.23a), so the question is asked BEFORE any library
// card is shown: the choice is what shows the library's candidates, and a
// library choice is what shuffles.
type searchZones uint8

const (
	// searchZonesUnasked is every search that does not ask: a plain
	// library search, "graveyard, hand, and library" (all three named),
	// and a prompt written before #4519. The take decides the shuffle
	// (executeSearchTakeLocked).
	searchZonesUnasked searchZones = iota
	// searchZonesAsking: the question is open. No candidate is shown, no
	// knowledge granted, no control recorded, and no pick is accepted.
	searchZonesAsking
	// searchZonesLibrary: the library was chosen (with or without the
	// other zone). It is shuffled, and the search announced, whatever is
	// taken.
	searchZonesLibrary
	// searchZonesOther: only the other zone (the graveyard or the hand)
	// was chosen. The library is never shown, shuffled or announced.
	searchZonesOther
)

// The zone names a search_zones answer uses (the wire's strings).
const (
	SearchZoneLibrary   = "library"
	SearchZoneGraveyard = "graveyard"
	SearchZoneHand      = "hand"
)

// ResolveSearchLibrary answers a PendingChoiceSearchLibrary (CR
// 701.23): `picks` are the cards the searcher takes, chosen from the
// candidates the prompt offered.
//
// An empty `picks` is a legal answer, not a client bug — CR 701.23b
// lets a player fail to find no matter what their library holds, when
// the search names a quality. A plain library search still happened, so
// EventSearchLibrary still fires and the library is still shuffled.
// That is NOT what declining a "you MAY search" means: choosing not to
// search at all is DeclineSearchLibrary (#4134), which skips the
// shuffle and the event. An "and/or" search is different (#4516,
// #4519): the zones were chosen first (ChooseSearchZones), so an empty
// answer after choosing only the graveyard (or the hand) does not
// shuffle, and after choosing the library does. An empty `picks` is NOT legal for an unconditional search
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
	spec := choice.searchResume.spec()
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
	// spec.ThenTail — because a battlefield destination is an ENTRY and an
	// entry can pause on its own prompt (#478). Answering THIS prompt
	// may therefore queue the next one rather than finishing the card.
	err := g.executeSearchTakeLocked(spec, p, picks, choice.searchResume.zones)
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
	if choice.searchResume.undecided || choice.searchResume.zones == searchZonesAsking {
		// #4134, #4519: the searcher has not begun to search, or has not
		// said which zones; there is nothing to take yet.
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
	// The registered set rule (#4253 slice 5). The per-card predicate is
	// not read here: it listed the candidates, and the picks were checked
	// against that list above.
	if len(chosen) > 0 && !choice.searchResume.spec().setAllows(chosen) {
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
		return choice.searchResume.spec().floorZones(inLibrary, inGraveyard)
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
	return c != nil && c.searchResume != nil && c.searchResume.data.AlsoGraveyardAndHand
}

// SearchZonesPending reports whether a search prompt is still asking
// which zones the searcher looks in (#4519): "search your library and/or
// graveyard" and "search your hand and/or library". No candidate is shown
// until it is answered (ChooseSearchZones), because looking through the
// library is searching it (CR 701.23a). The all-zones search names the
// library, and a plain one has only the library, so neither asks.
func (c *PendingChoice) SearchZonesPending() bool {
	return c != nil && c.Kind == PendingChoiceSearchLibrary && c.searchResume != nil &&
		c.searchResume.zones == searchZonesAsking
}

// SearchOtherZone is the zone an and/or search offers beside the library,
// "graveyard" or "hand" (SearchZoneGraveyard, SearchZoneHand), or "" for a
// prompt that has no such zone.
func (c *PendingChoice) SearchOtherZone() string {
	if c == nil || c.Kind != PendingChoiceSearchLibrary || c.searchResume == nil {
		return ""
	}
	return c.searchResume.data.searchSpec().otherZoneName()
}

// SearchLibrarySkipped reports whether the searcher answered an and/or
// search's zone question with the other zone alone (#4519): the library is
// not searched, so it is neither shown nor shuffled.
func (c *PendingChoice) SearchLibrarySkipped() bool {
	return c != nil && c.Kind == PendingChoiceSearchLibrary && c.searchResume != nil &&
		c.searchResume.zones == searchZonesOther
}

// ChooseSearchZones answers an and/or search's zone question (#4519, CR
// 701.23): `zones` names SearchZoneLibrary, the prompt's other zone
// (SearchOtherZone) or both, at least one. It is what lets the searcher
// see the candidates: the library's are shown and known to them only if
// they chose the library, and that choice is what shuffles it afterwards,
// whatever they then take. A search of the other zone alone never looks at
// the library, so no library card is offered, none is made known, an
// Opposition Agent never takes control, and the library is neither
// shuffled nor announced.
//
// The prompt stays the same prompt (same ID): its candidates are trimmed to
// the chosen zones and its floor recomputed, and the next answer is the
// ordinary pick. It is asked even when the chosen zones hold no match, so
// that the table cannot tell, from the prompt closing, whether the library
// had one (#4140). Refused with ErrInvalidParam on a prompt that is not
// asking, on a zone name the prompt does not have, and on an empty answer.
//
// Caller must NOT hold g.mu.
func (g *Game) ChooseSearchZones(choiceID, chooserID uuid.UUID, zones []string) error {
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
	// An optional search asks "do you search?" first (#4134), and an
	// unasked or already-answered prompt has no zone question.
	if choice.searchResume.zones != searchZonesAsking || choice.searchResume.undecided {
		return ErrInvalidParam
	}
	spec := choice.searchResume.spec()
	other := spec.otherZoneName()
	var library, elsewhere bool
	for _, z := range zones {
		switch {
		case z == SearchZoneLibrary:
			library = true
		case z != "" && z == other:
			elsewhere = true
		default:
			return ErrInvalidParam
		}
	}
	if !library && !elsewhere {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(spec.libraryOwnerID())
	if p == nil {
		return ErrPlayerNotFound
	}
	outside := make(map[uuid.UUID]bool, len(choice.SearchGraveyardCards)+len(choice.SearchHandCards))
	for _, id := range choice.SearchGraveyardCards {
		outside[id] = true
	}
	for _, id := range choice.SearchHandCards {
		outside[id] = true
	}
	var cards, libraryCards []uuid.UUID
	for _, id := range choice.SearchCards {
		switch {
		case outside[id]:
			if elsewhere {
				cards = append(cards, id)
			}
		case library:
			cards = append(cards, id)
			libraryCards = append(libraryCards, id)
		}
	}
	// Fresh slices, never a write through the old ones, which an undo
	// snapshot may share.
	var gy, hand []uuid.UUID
	if elsewhere {
		gy = append(gy, choice.SearchGraveyardCards...)
		hand = append(hand, choice.SearchHandCards...)
	}
	choice.SearchCards = cards
	choice.SearchGraveyardCards = gy
	choice.SearchHandCards = hand
	choice.SearchMin = spec.floorZones(len(libraryCards), len(gy)+len(hand))
	answered := *choice.searchResume
	answered.zones = searchZonesOther
	if library {
		answered.zones = searchZonesLibrary
	}
	choice.searchResume = &answered
	if library {
		g.markSearchCandidatesKnownLocked(spec.Player, p, libraryCards)
		if spec.libraryOwnerID() == spec.Player {
			g.beginSearchControlsLocked(spec.Player, choice.ID)
		}
	}
	return nil
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
	p := g.playerByIDLocked(choice.searchResume.spec().libraryOwnerID())
	if p == nil {
		return 0
	}
	return g.searchFloorLocked(choice, p)
}

// SearchPickLegalLocked reports whether `picks` is an answer
// ResolveSearchLibrary would accept for `choice` right now. It is
// the enumerator's read-only window onto the search's set rule,
// which rides on the unexported continuation frame and must stay
// there: handing internal/legal the frame itself would make every
// future continuation field part of the enumerator's contract.
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
	p := g.playerByIDLocked(choice.searchResume.spec().libraryOwnerID())
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
		c.searchResume.data.Optional
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
// begun. Looking through the library is searching (CR 701.23a), so once
// the searcher has said yes, and so seen the candidates, there is no
// way back: declining would be a free look that never shuffles. For the
// same reason an Opposition Agent's controller, who takes over only
// after the search begins, can never decline for the searcher.
//
// Caller must hold g's lock (read or write).
func (g *Game) SearchDeclineLegalLocked(choice *PendingChoice) bool {
	return choice.SearchIsUndecided() && choice.searchResume.data.Optional
}

// DeclineSearchLibrary answers a "you may search" prompt with "I don't
// search" (#4134, CR 701.23a). The search never happens: the library is
// not looked at and not shuffled, no EventSearchLibrary fires (so an
// Archivist of Oghma stays quiet), and an Opposition Agent never gets
// control. Only the rest of the effect runs (spec.ThenTail, with no cards),
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
	spec := choice.searchResume.spec()
	g.dequeueChoiceLocked(idx)
	if !spec.ThenTail.IsZero() {
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
	spec := choice.searchResume.spec()
	p := g.playerByIDLocked(spec.libraryOwnerID())
	if p == nil {
		return ErrPlayerNotFound
	}
	begun := *choice.searchResume
	begun.undecided = false
	choice.searchResume = &begun
	if begun.zones == searchZonesAsking {
		// #4519: "do you search?" is answered, but which zones is not:
		// nothing is looked at, and no control begins, until the library
		// is chosen (ChooseSearchZones).
		return nil
	}
	g.markSearchCandidatesKnownLocked(spec.Player, p, choice.SearchCards)
	if spec.libraryOwnerID() == spec.Player {
		g.beginSearchControlsLocked(spec.Player, choice.ID)
	}
	return nil
}
