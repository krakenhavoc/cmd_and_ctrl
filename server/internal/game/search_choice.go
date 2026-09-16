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
}

// ResolveSearchLibrary answers a PendingChoiceSearchLibrary (CR
// 701.19): `picks` are the cards the searcher takes, chosen from the
// candidates the prompt offered.
//
// An empty `picks` is a legal answer, not a client bug — CR 701.23b
// lets a player fail to find no matter what their library holds, and
// declining is the whole content of a "you MAY search" clause. The
// search still happened, so EventSearchLibrary still fires and the
// library is still shuffled; that shuffle is the drawback the victim
// of an Assassin's Trophy is accepting either way.
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
	p := g.playerByIDLocked(chooserID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if err := g.checkSearchPicksLocked(choice, p, picks); err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)

	found := g.executeSearchTakeLocked(spec, p, picks)
	err := g.finishSearchLocked(spec, p, found)
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
	if len(picks) > choice.SearchMax {
		return ErrInvalidParam
	}
	candidates := make(map[uuid.UUID]bool, len(choice.SearchCards))
	for _, id := range choice.SearchCards {
		candidates[id] = true
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
		if !ok || !p.Library.Contains(id) {
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

// SearchPickLegalLocked reports whether `picks` is an answer
// ResolveSearchLibrary would accept for `choice` right now. It is
// the enumerator's read-only window onto the search's Validate hook,
// which rides on the unexported continuation frame and must stay
// there: the hook is a closure over the effect, not wire state, and
// handing internal/legal the frame itself would make every future
// continuation field part of the enumerator's contract.
//
// An empty `picks` is always legal — CR 701.23b lets a player fail
// to find — which is what gives a stuck seat an answer to fall back
// on (see legal.Move.AlwaysLegal).
//
// Caller must hold g's read lock; internal/legal calls it from
// inside ReadSnapshot, like every other *ForEffect / *Locked surface
// the enumerator uses.
func (g *Game) SearchPickLegalLocked(choice *PendingChoice, picks []uuid.UUID) bool {
	if choice == nil || choice.Kind != PendingChoiceSearchLibrary {
		return false
	}
	p := g.playerByIDLocked(choice.Chooser)
	if p == nil {
		return false
	}
	return g.checkSearchPicksLocked(choice, p, picks) == nil
}
