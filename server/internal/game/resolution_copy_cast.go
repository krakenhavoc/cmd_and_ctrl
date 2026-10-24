package game

import (
	"slices"

	"github.com/google/uuid"
)

// resolution_copy_cast.go — casting a copy of a card while the effect
// that made it is still resolving (CR 707.12, CR 608.2g; ADR 0066's
// 2026-10-22 amendment, #2947).
//
// CR 707.12: an effect that instructs a player to cast a copy of an
// object creates the copy and "then cast[s it] while another spell or
// ability is resolving". CR 608.2g: such a spell is cast by following
// CR 601.2a–i, "except no player receives priority after it's cast",
// it becomes the topmost object on the stack, and the resolving spell
// or ability continues to resolve.
//
// card_copy.go makes the copies and their offers (one stored ScopeCards
// permission per copy, or one over the set for a mana value budget).
// Until this file the cast was taken from that permission after the
// resolution had finished, the grant model's stand-in. An offer made
// with CardCopyOffer.DuringResolution now also queues a resolution_cast
// prompt over the copies, inside the open resolution:
//
//   - The prompt is mid-resolution, so it blocks the table and holds the
//     CR 704.3 boundary (resolution_pause.go): nothing else is cast, no
//     state-based action is performed and no trigger is put on the
//     stack until the caster has finished with the copies.
//   - It is answered by casting a candidate: a cast_spell naming the
//     prompt, or any CastSpell of a candidate by its chooser. The cast is
//     the ordinary one (castSpellLocked) through the copy's own
//     permission, so the price, X, targets, modes, additional costs,
//     cast prohibitions and the "if you do" follow-up are all the
//     existing ones. Payment is strict and the auto-tapper may use every
//     ordinary mana source: CR 608.2g's "option to pay mana" lets the
//     caster activate mana abilities. No action is noted, because no
//     player receives priority after it.
//   - Each cast spends only its own candidate. The prompt stays open
//     while another candidate can still be cast (CR 707.12a: each copy
//     is its own "may"), and goes once none can.
//   - The other answer is the empty resolve_choice, "cast no more of
//     them", always legal. It withdraws the offers over the candidates
//     left, so the CR 704.5e sweep at the boundary removes them.
//
// The prompt is plain data — the chooser, the source and the candidate
// list in DiscardOptions — and the permissions, the copies and
// Game.resolutionOpen already persist, so a table paused on it is a
// restore point.
//
// #4243: the same prompt is cascade's, discover's, cipher's and
// paradigm's cast (cascade.go, discover.go, cipher.go, paradigm.go). For
// cascade and discover the candidate is a real card in exile rather than
// a copy, under the pass-closed permission those keywords already
// stamped (so a restore point written before this change keeps working
// through the pass-closed window), and the prompt also carries the cards
// the keyword passed over (PendingChoice.CastPile). The decline lapses
// the permission as a pass would have, sending the card where its keyword
// says (the bottom of the library, or the hand), and the pile and any
// card that goes to the bottom are put there together, in one random
// order, once the prompt is answered (CR 702.85a, CR 701.57a).

// PendingChoiceResolutionCast is the open "you may cast the copy"
// inside a resolution (CR 707.12, CR 608.2g). Chooser casts, Source is
// the card whose effect made the copies, Reason its name, and
// DiscardOptions the copies still on offer.
const PendingChoiceResolutionCast PendingChoiceKind = "resolution_cast"

// queueResolutionCastLocked raises the resolution_cast prompt over
// copies, the copies CopyCardsAndOfferCastsForEffect has just made and
// offered. Nothing is queued outside a resolution, where the offer's
// permissions alone are the window. Caller must hold g.mu.
func (g *Game) queueResolutionCastLocked(offer CardCopyOffer, copies []uuid.UUID) uuid.UUID {
	return g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:    offer.Player,
		Source:     offer.Source,
		Reason:     offer.SourceName,
		Candidates: copies,
	})
}

// resolutionCastOffer is what a resolution_cast prompt asks: who may
// cast which cards, under what name, and what else rides on the answer.
type resolutionCastOffer struct {
	Chooser, Source uuid.UUID
	// Reason names the effect (the prompt's header).
	Reason string
	// Candidates are the cards on offer. Each one is named by a stored
	// ScopeCards permission of Chooser's.
	Candidates []uuid.UUID
	// Keyword, AcceptLabel and DeclineLabel word the prompt for a
	// keyword that has its own rule ("cascade", "discover"). Empty for
	// the copies.
	Keyword, AcceptLabel, DeclineLabel string
	// Pile is cascade's or discover's passed-over cards, bottomed when
	// the prompt is answered.
	Pile []uuid.UUID
}

// queueResolutionCastOfferLocked raises the prompt for o, or answers
// uuid.Nil when there is nothing to hold open: no resolution is open, no
// card is on offer, or the chooser is not in the game. The caller then
// settles the offer without a prompt. Caller must hold g.mu.
func (g *Game) queueResolutionCastOfferLocked(o resolutionCastOffer) uuid.UUID {
	if len(o.Candidates) == 0 || !g.resolutionOpen || !g.playerInGameLocked(o.Chooser) {
		return uuid.Nil
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:           PendingChoiceResolutionCast,
		Chooser:        o.Chooser,
		Count:          1,
		Source:         o.Source,
		Reason:         o.Reason,
		DiscardOptions: slices.Clone(o.Candidates),
		CastPile:       slices.Clone(o.Pile),
		MayCastKeyword: o.Keyword,
		AcceptLabel:    o.AcceptLabel,
		DeclineLabel:   o.DeclineLabel,
	})
}

// ResolutionCastCandidateLocked reports whether id is a card chooser
// may still cast through a resolution_cast prompt: a copy of a card, or
// the card a cascade or discover hit (#4243), still in exile or a
// graveyard, that one of chooser's stored permissions names. It answers
// the zone it sits in. Caller must hold g.mu.
func (g *Game) ResolutionCastCandidateLocked(chooser, id uuid.UUID) (Card, ZoneKind, bool) {
	z := g.findCardZoneLocked(id)
	if z == nil || (z.Kind != ZoneExile && z.Kind != ZoneGraveyard) {
		return Card{}, "", false
	}
	c, ok := g.cardInZoneLocked(z, id)
	if !ok {
		return Card{}, "", false
	}
	p := g.playerByIDLocked(chooser)
	if p == nil {
		return Card{}, "", false
	}
	for i := range p.CastPermissions {
		if p.CastPermissions[i].Scope == ScopeCards && p.CastPermissions[i].NamesCard(c) {
			return c, z.Kind, true
		}
	}
	return Card{}, "", false
}

// resolutionCastChoiceForCardLocked is the index of the open
// resolution_cast prompt of chooser's that offers card, or -1. Caller
// must hold g.mu.
func (g *Game) resolutionCastChoiceForCardLocked(chooser, card uuid.UUID) int {
	for i, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceResolutionCast && c.Chooser == chooser &&
			slices.Contains(c.DiscardOptions, card) {
			return i
		}
	}
	return -1
}

// CastForResolutionChoice is the resolution_cast prompt's "cast it"
// answer: caller casts the candidate cardID during the resolution that
// made it (CR 707.12, CR 608.2g), with every choice of the cast in
// params. Refused, with the prompt left open for another try or a
// decline:
//
//   - a prompt that is not caller's resolution_cast
//     (ErrPendingChoiceNotFound, ErrNotTheChooser) or a card it does
//     not offer (ErrInvalidParam);
//   - a cast the engine refuses (whatever castSpellLocked returns: a
//     cost that can't be paid, an illegal target).
//
// Caller must NOT hold g.mu.
func (g *Game) CastForResolutionChoice(caller, choiceID, cardID uuid.UUID, params CastSpellParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := g.choiceIndexLocked(choiceID)
	if idx < 0 || g.PendingChoices[idx].Kind != PendingChoiceResolutionCast {
		return ErrPendingChoiceNotFound
	}
	c := g.PendingChoices[idx]
	if c.Chooser != caller {
		return ErrNotTheChooser
	}
	if !slices.Contains(c.DiscardOptions, cardID) {
		return ErrInvalidParam
	}
	return g.castForResolutionChoiceLocked(c.ID, caller, cardID, params)
}

// castForResolutionChoiceLocked casts cardID for the prompt choiceID
// and settles the prompt. The cast is paid in full ("cast anyway" is
// not offered: the copy costs what its offer says), with the
// auto-tapper free to use any of the caster's mana sources but the
// ones locked out. Caller must hold g.mu.
func (g *Game) castForResolutionChoiceLocked(choiceID, caster, cardID uuid.UUID, params CastSpellParams) error {
	params.ForceCast = false
	params.PaperPayment = nil
	params.Strict = true
	params.AutoTap = true
	if err := g.castSpellLocked(caster, cardID, params); err != nil {
		return err
	}
	g.settleResolutionCastLocked(choiceID, cardID)
	// The last answer lets the resolution finish; its boundary runs
	// now, as an answered prompt's does (resolution_pause.go). With a
	// candidate left it holds.
	g.runStateChecksLocked()
	return nil
}

// settleResolutionCastLocked takes cast, and every candidate that can
// no longer be cast, off the prompt, and answers the prompt once none is
// left. The queue entry is replaced, not written through, since an undo
// entry may share it. Caller must hold g.mu.
func (g *Game) settleResolutionCastLocked(choiceID, cast uuid.UUID) {
	i := g.choiceIndexLocked(choiceID)
	if i < 0 {
		return
	}
	c := g.PendingChoices[i]
	var left []uuid.UUID
	for _, id := range c.DiscardOptions {
		if id == cast {
			continue
		}
		if _, _, ok := g.ResolutionCastCandidateLocked(c.Chooser, id); ok {
			left = append(left, id)
		}
	}
	if len(left) == 0 {
		g.dequeueChoiceLocked(i)
		g.finishResolutionCastLocked(c, nil)
		return
	}
	next := *c
	next.DiscardOptions = left
	queue := slices.Clone(g.PendingChoices)
	queue[i] = &next
	g.PendingChoices = queue
}

// resolveResolutionCastLocked answers a resolution_cast prompt with
// resolve_choice. The only answer it takes is the empty one, "cast no
// more of them", which is always legal; the casts are
// CastForResolutionChoice. The offers over the cards left are withdrawn:
// a copy ceases to exist at the boundary (CR 704.5e), and a card a
// keyword exiled goes where not casting it sends it (CR 702.85a, CR
// 701.57a). Caller must hold g.mu.
func (g *Game) resolveResolutionCastLocked(idx int, c *PendingChoice, picks []uuid.UUID) error {
	if len(picks) != 0 {
		return ErrInvalidParam
	}
	g.dequeueChoiceLocked(idx)
	g.finishResolutionCastLocked(c, c.DiscardOptions)
	g.runStateChecksLocked()
	return nil
}

// finishResolutionCastLocked ends a resolution_cast prompt that has
// just left the queue. declined are the cards the chooser did not cast:
// their offers are withdrawn, each lapsing as a pass would have sent it
// (a copy only stops being offered, a cascade hit goes to the bottom, a
// discovered card to the hand, and an "if you don't" follow-up runs).
// The pile the prompt carries and any card that goes to the bottom are
// put there together in a random order. Caller must hold g.mu.
func (g *Game) finishResolutionCastLocked(c *PendingChoice, declined []uuid.UUID) {
	bottom := slices.Clone(c.CastPile)
	g.withdrawCardCopyOffersLocked(c.Chooser, declined, &bottom)
	if len(bottom) > 0 {
		_ = g.PutOnBottomInRandomOrderForEffect(c.Chooser, ZoneExile, bottom)
	}
}

// withdrawCardCopyOffersLocked drops every ScopeCards permission holder
// holds that names one of the cards in ids, whatever its epoch. The
// cards were offered for one prompt, so nothing else names them. A fresh
// slice, for the reason lapsePassClosedGrantsLocked gives.
//
// A permission that is pass-closed lapses as the pass it would have
// closed on would (lapseCastPermissionLocked); a card that would go to
// the bottom of the library is appended to bottom instead of being put
// there alone, so it lands in the same random order as the rest of its
// cascade (CR 702.85a). Caller must hold g.mu.
func (g *Game) withdrawCardCopyOffersLocked(holder uuid.UUID, ids []uuid.UUID, bottom *[]uuid.UUID) {
	p := g.playerByIDLocked(holder)
	if p == nil || len(p.CastPermissions) == 0 || len(ids) == 0 {
		return
	}
	names := func(perm CastPermission) bool {
		if perm.Scope != ScopeCards {
			return false
		}
		for _, ref := range perm.Cards {
			if slices.Contains(ids, ref.ID) {
				return true
			}
		}
		return false
	}
	kept := p.CastPermissions[:0:0]
	var gone []CastPermission
	for _, perm := range p.CastPermissions {
		if names(perm) {
			gone = append(gone, perm)
			continue
		}
		kept = append(kept, perm)
	}
	if len(gone) == 0 {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	p.CastPermissions = kept
	for _, perm := range gone {
		if perm.LapseOnPass != "" {
			g.lapseCastPermissionCollectingLocked(holder, perm, bottom)
		}
	}
}
