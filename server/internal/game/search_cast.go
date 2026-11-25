package game

import (
	"slices"

	"github.com/google/uuid"
)

// search_cast.go — #6216: "While you're searching your library, you may
// cast this card from your library." (Panglacial Wurm.)
//
// CR 701.23a: "To search for a card in a zone, look at all cards in that
// zone (even if it's a hidden zone) and find a card that matches the given
// description." Every search in this engine happens while a spell or
// ability resolves, and is held open by the search_library prompt
// (search_choice.go) until the searcher answers it. So "while you're
// searching your library" is exactly: the searcher has an open, begun
// search_library prompt over their own library.
//
// CR 608.2g: "If an effect specifically instructs or allows a player to
// cast a spell during resolution, they do so by following the steps in
// rules 601.2a–i, except no player receives priority after it's cast.
// That spell becomes the topmost object on the stack, and the currently
// resolving spell or ability continues to resolve." The card's own text is
// that permission, so:
//
//   - The card's text opens its owner's library (cardOpensZoneLocked) and
//     waives its type's timing (castTimingOpenLocked) for as long as the
//     search is open, and at no other time. One predicate,
//     searchCastOpenLocked, answers both, so the view, the enumerator and
//     CastSpell cannot disagree.
//   - The cast is the ordinary one (castSpellLocked): costs, targets,
//     modes and cast prohibitions are the existing ones; payment is
//     strict, and the auto-tapper may tap any mana source (CR 608.2g's
//     "option to pay mana", CR 601.2g). No action is noted, because no
//     player receives priority after it.
//   - The search goes on: the prompt stays open with the cast card taken
//     off its candidates (it is on the stack now), and the searcher then
//     finishes it.
//
// The prompt is answered over the wire by a cast_spell naming it
// (`search_cast`), the way a resolution_cast prompt's cast names that one,
// so the searcher need not hold priority to send it.

// CatalogCastWhileSearchingLibrary reports whether a catalog entry prints
// "While you're searching your library, you may cast this card from your
// library." Wired in carddef.go; nil in a test that builds cards by hand.
var CatalogCastWhileSearchingLibrary func(key string) bool

// castsWhileSearchingLibrary reports whether `c`'s own text lets its
// owner cast it from their library while searching it.
func castsWhileSearchingLibrary(c Card) bool {
	if CatalogCastWhileSearchingLibrary == nil {
		return false
	}
	key := CatalogKey(c)
	return key != "" && CatalogCastWhileSearchingLibrary(key)
}

// openLibrarySearchLocked is the search_library prompt `player` is
// answering over their own library right now, with the search begun and
// the library among the zones searched, or nil. Only the head of the
// queue counts: a prompt queued behind another is not yet being
// answered. Caller must hold g.mu.
func (g *Game) openLibrarySearchLocked(player uuid.UUID) *PendingChoice {
	if !g.resolutionOpen || len(g.PendingChoices) == 0 {
		return nil
	}
	c := g.PendingChoices[0]
	if c == nil || c.Kind != PendingChoiceSearchLibrary || c.searchResume == nil || c.Chooser != player {
		return nil
	}
	if c.SearchIsUndecided() || c.SearchZonesPending() || c.SearchLibrarySkipped() {
		return nil
	}
	if c.searchResume.data.searchSpec().libraryOwnerID() != player {
		return nil
	}
	return c
}

// searchCastOpenLocked reports whether `player` may cast `card` from
// their library because they are searching it and its own text says so.
// Caller must hold g.mu.
func (g *Game) searchCastOpenLocked(player uuid.UUID, card Card) bool {
	if card.Owner != player || !castsWhileSearchingLibrary(card) {
		return false
	}
	if g.openLibrarySearchLocked(player) == nil {
		return false
	}
	return slices.Contains(g.SearchCastableLocked(player), card.InstanceID)
}

// SearchCastableLocked lists the cards `player` may cast from their
// library right now because they are searching it (#6216), in library
// order: the searchable cards (search_top_n.go narrows them under an
// opponent's Aven Mindcensor) whose own text allows it. Nil when no
// search is open. Caller must hold g.mu.
func (g *Game) SearchCastableLocked(player uuid.UUID) []uuid.UUID {
	if CatalogCastWhileSearchingLibrary == nil || g.openLibrarySearchLocked(player) == nil {
		return nil
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.searchableLibraryCards(player, p.Library.Cards) {
		if castsWhileSearchingLibrary(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// libraryHoldsSearchCastLocked reports whether `player`'s searchable
// library holds a card they could cast while searching it, so a search
// the rules would otherwise settle unasked (a forced take-all) asks
// instead and the searcher has the window to cast it. Caller must hold
// g.mu.
func (g *Game) libraryHoldsSearchCastLocked(player uuid.UUID, p *Player) bool {
	if CatalogCastWhileSearchingLibrary == nil || p == nil {
		return false
	}
	for _, c := range g.searchableLibraryCards(player, p.Library.Cards) {
		if castsWhileSearchingLibrary(c) {
			return true
		}
	}
	return false
}

// CastWhileSearching casts `cardID` from `player`'s library during the
// open search prompt `choiceID` (#6216, CR 608.2g). Refused with nothing
// paid when the prompt is not that player's open library search or the
// card is not one its text lets them cast now. Caller must NOT hold g.mu.
func (g *Game) CastWhileSearching(player, choiceID, cardID uuid.UUID, params CastSpellParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	c := g.openLibrarySearchLocked(player)
	if c == nil || c.ID != choiceID {
		return ErrPendingChoiceNotFound
	}
	return g.castWhileSearchingLocked(player, cardID, params)
}

// castWhileSearchingLocked is the cast and the prompt's update. Caller
// must hold g.mu.
func (g *Game) castWhileSearchingLocked(player, cardID uuid.UUID, params CastSpellParams) error {
	if !slices.Contains(g.SearchCastableLocked(player), cardID) {
		return ErrCastZoneNotAllowed
	}
	params.FromZone = string(ZoneLibrary)
	params.ForceCast = false
	params.PaperPayment = nil
	params.Strict = true
	params.AutoTap = true
	if err := g.castSpellLocked(player, cardID, params); err != nil {
		return err
	}
	g.settleSearchCastLocked(player, cardID)
	return nil
}

// searchCastChoiceForCardLocked reports whether a CastSpell of `cardID`
// by `player` is a cast while searching (#6216), so CastSpell routes it
// here however it arrives. Caller must hold g.mu.
func (g *Game) searchCastChoiceForCardLocked(player, cardID uuid.UUID) bool {
	return g.openLibrarySearchLocked(player) != nil && slices.Contains(g.SearchCastableLocked(player), cardID)
}

// settleSearchCastLocked takes the cast card off the open search prompt:
// it is on the stack now, no longer in the library being searched. The
// queue entry is replaced, not written through, since an undo entry may
// share it. A floor the library can no longer meet (CR 701.23d: "or as
// many as possible") falls with it. Caller must hold g.mu.
func (g *Game) settleSearchCastLocked(player, cast uuid.UUID) {
	if len(g.PendingChoices) == 0 {
		return
	}
	c := g.PendingChoices[0]
	if c == nil || c.Kind != PendingChoiceSearchLibrary || c.Chooser != player || !slices.Contains(c.SearchCards, cast) {
		return
	}
	next := *c
	next.SearchCards = slices.DeleteFunc(slices.Clone(c.SearchCards), func(id uuid.UUID) bool { return id == cast })
	if next.SearchMin > len(next.SearchCards) {
		next.SearchMin = len(next.SearchCards)
	}
	queue := slices.Clone(g.PendingChoices)
	queue[0] = &next
	g.PendingChoices = queue
}
