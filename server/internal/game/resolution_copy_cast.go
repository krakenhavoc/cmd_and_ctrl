package game

import (
	"log/slog"
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
//
// #4250, part 2: the same prompt is the cast an effect's resolution
// offers over cards it names, wherever they are
// (OfferResolutionCastForEffect, the door the catalog uses): a card in
// the caster's hand (Buster Sword, the Expertises, Aetherflux Conduit,
// The Great Synthesis), the top of a library (Aetherworks Marvel,
// Sunbird's Invocation), a graveyard (Torrential Gearhulk, Memory
// Plunder) or exile (Etali, Kefka, the Sieges). Two things are new for
// them:
//
//   - A HAND card also has its ordinary cast (CR 601.2), and a hand
//     permission prices only the cast that claims its offer
//     (CastPermission.ForClaim). So a hand candidate's cast is held to
//     the permission's claim: castForResolutionChoiceLocked names it
//     whatever the cast sent, and the enumerator offers only the casts
//     that make it (ResolutionCastClaimLocked).
//   - What the resolution does AFTER the casts ("then exile this Saga",
//     "then each player who owns a spell you cast this way loses life",
//     "put the rest on the bottom") is the prompt's own continuation, a
//     registered ChooseCardsThen key in PendingChoice.PickThen with the
//     caller's cards in OptionCarry. It runs once the prompt is answered,
//     with the cards of the carry that were cast (now on the stack) as
//     its picks. Both are plain data, so the prompt stays a restore point.

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
	// Then is the rest of the resolution after the casts, run once the
	// prompt is answered (#4250), and Carry the cards it is told about.
	// Zero for none.
	Then  ChooseCardsThen
	Carry []uuid.UUID
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
		PickThen:       o.Then.Key(),
		OptionCarry:    slices.Clone(o.Carry),
	})
}

// ResolutionCastOffer is OfferResolutionCastForEffect's question: who may
// cast which cards during the resolution that is offering them, under
// what name, and what the resolution does once the casts are made.
type ResolutionCastOffer struct {
	// Chooser casts; Source is the card whose effect offers the casts.
	Chooser, Source uuid.UUID
	// Reason names the effect: the prompt's header and every move's
	// label prefix.
	Reason string
	// Cards are the cards on offer. Each one must be named by a stored
	// ScopeCards permission of Chooser's, or by Permission.
	Cards []uuid.UUID
	// Permission, when set, is stamped over Cards as ONE permission, and
	// only once the prompt is sure to be asked. A pass-closed one is
	// spent whole by the first cast it is the reason for
	// (consumePassClosedGrantLocked), which is "cast a spell from among
	// them": one cast. Its Zone is filled in from where the first card
	// is, and a card in any other zone is left out. A hand permission
	// with a Cost and no AltCostKey is given AltCostKeyGrantedCost, as
	// GrantCastPermissionOverCardForEffect does.
	Permission *CastPermission
	// Keyword, AcceptLabel and DeclineLabel word the prompt
	// (MayCastKeywordHandCast, MayCastKeywordFreeCast, ...).
	Keyword, AcceptLabel, DeclineLabel string
	// Then and Carry are the rest of the resolution after the casts:
	// Then runs once the prompt is answered, with the cards of Carry
	// that were cast as its Picked and Carry as its Carry.
	Then  ChooseCardsThen
	Carry []uuid.UUID
}

// Keywords a ResolutionCastOffer can name (#4250). The client words each
// (mayCast.ts keywordCastWords). On-disk values, never renamed.
const (
	// MayCastKeywordHandCast is "you may cast a spell from your hand
	// without paying its mana cost": the cards are in the caster's hand
	// and stay there if they are not cast.
	MayCastKeywordHandCast = "hand_cast"
	// MayCastKeywordFreeCast is "you may cast it without paying its mana
	// cost" over cards in exile, a graveyard or a library. The decline
	// label says where an uncast card ends up.
	MayCastKeywordFreeCast = "free_cast"
	// MayCastKeywordPaidCast is the same offer at a price: the card's own
	// mana cost, or the alternative the effect names (energy, a
	// discard).
	MayCastKeywordPaidCast = "paid_cast"
)

// OfferResolutionCastForEffect asks o.Chooser to cast o.Cards during the
// resolution that is offering them (CR 608.2g), through the
// resolution_cast prompt. It answers false, and changes nothing, when no
// resolution is open or the chooser has left the game: the caller then
// keeps its own fallback (a pass-closed permission, or a keyed pick).
//
// Otherwise it answers true. o.Permission is stamped, the prompt is
// raised over the cards a permission really opens, and when none does
// (every card has moved, or none was a spell) o.Then runs at once with
// nothing cast.
//
// Caller must hold g.mu (write).
func (g *Game) OfferResolutionCastForEffect(o ResolutionCastOffer) bool {
	if !g.resolutionOpen || !g.playerInGameLocked(o.Chooser) {
		return false
	}
	if o.Permission != nil {
		g.stampResolutionCastPermissionLocked(o.Chooser, *o.Permission, o.Cards)
	}
	// A land is played, never cast (CR 116.2a, CR 305.9), so it is never
	// on a cast offer, whatever the permission names.
	var candidates []uuid.UUID
	for _, id := range o.Cards {
		if slices.Contains(candidates, id) {
			continue
		}
		if c, _, ok := g.ResolutionCastCandidateLocked(o.Chooser, id); ok && !c.IsLand() {
			candidates = append(candidates, id)
		}
	}
	queued := g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      o.Chooser,
		Source:       o.Source,
		Reason:       o.Reason,
		Candidates:   candidates,
		Keyword:      o.Keyword,
		AcceptLabel:  o.AcceptLabel,
		DeclineLabel: o.DeclineLabel,
		Then:         o.Then,
		Carry:        o.Carry,
	})
	if queued == uuid.Nil {
		g.runResolutionCastThenLocked(o.Then.Key(), o.Chooser, o.Source, o.Carry)
	}
	return true
}

// stampResolutionCastPermissionLocked stores perm over the cards of ids
// that sit in its zone (the first card's, when it names none), as one
// ScopeCards permission. Caller must hold g.mu (write).
func (g *Game) stampResolutionCastPermissionLocked(chooser uuid.UUID, perm CastPermission, ids []uuid.UUID) {
	var cards []Card
	for _, id := range ids {
		z := g.findCardZoneLocked(id)
		if z == nil {
			continue
		}
		if perm.Zone == "" {
			perm.Zone = z.Kind
		}
		if z.Kind != perm.Zone {
			continue
		}
		if c, ok := g.cardInZoneLocked(z, id); ok {
			cards = append(cards, c)
		}
	}
	if perm.Player == uuid.Nil {
		perm.Player = chooser
	}
	if perm.Zone == ZoneHand && perm.AltCostKey == "" && perm.Cost != "" {
		perm.AltCostKey = AltCostKeyGrantedCost
	}
	g.GrantCastPermissionToCardsForEffect(perm, cards)
}

// ResolutionOpenForEffect reports whether a resolution is open, so a
// resolution_cast offer will be asked (OfferResolutionCastForEffect).
// A card whose fallback differs (Kefka's per-cast life loss) asks it
// before it stamps its permissions. Caller must hold g.mu.
func (g *Game) ResolutionOpenForEffect() bool { return g.resolutionOpen }

// RunResolutionCastThenForEffect runs then as a resolution_cast prompt's
// continuation would, for an offer OfferResolutionCastForEffect did not
// take (no resolution open): the caller's fallback for "and then".
//
// Caller must hold g.mu (write).
func (g *Game) RunResolutionCastThenForEffect(then ChooseCardsThen, chooser, source uuid.UUID, carry []uuid.UUID) {
	g.runResolutionCastThenLocked(then.Key(), chooser, source, carry)
}

// runResolutionCastThenLocked runs a resolution_cast prompt's
// continuation, if it has one: the cards of carry now on the stack are
// the ones cast through the prompt (nobody received priority since, so
// nothing else put them there or took them off). An unknown key is
// logged, not fatal: the casts are made. Caller must hold g.mu (write).
func (g *Game) runResolutionCastThenLocked(key string, chooser, source uuid.UUID, carry []uuid.UUID) {
	if key == "" {
		return
	}
	fn, ok := lookupChooseCardsThen(key)
	if !ok {
		slog.Warn("resolution_cast: unknown continuation", "key", key)
		return
	}
	var cast []uuid.UUID
	for _, id := range carry {
		if g.Stack != nil && g.Stack.Contains(id) {
			cast = append(cast, id)
		}
	}
	if err := fn(g, CardsChosen{
		Chooser:  chooser,
		Source:   source,
		Picked:   cast,
		Answered: true,
		Carry:    slices.Clone(carry),
	}); err != nil {
		slog.Warn("resolution_cast: continuation failed", "key", key, "err", err)
	}
}

// ResolutionCastCandidateLocked reports whether id is a card chooser
// may still cast through a resolution_cast prompt: a copy of a card, or
// the card a keyword or an effect put on offer (#4243, #4250), still in
// exile, a graveyard, a library or chooser's hand, that one of chooser's
// stored permissions names. It answers the zone it sits in. A library
// card is a top card an effect let its owner look at (Ziatora's Envoy,
// Aetherworks Marvel), cast from where it sits (#4250); its permission
// opens the library, so nothing else in a library is ever a candidate.
// A hand card is one an effect lets its owner cast for free (#4250, part
// 2); its cast is held to the permission's claim
// (ResolutionCastClaimLocked). Caller must hold g.mu.
func (g *Game) ResolutionCastCandidateLocked(chooser, id uuid.UUID) (Card, ZoneKind, bool) {
	z := g.findCardZoneLocked(id)
	if z == nil || (z.Kind != ZoneExile && z.Kind != ZoneGraveyard && z.Kind != ZoneLibrary && z.Kind != ZoneHand) {
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

// ResolutionCastClaimLocked is the alternative-cost claim a cast of the
// candidate id through a resolution_cast prompt must make, and whether it
// must make one: a HAND card's (#4250). A hand card also has its ordinary
// cast, at its printed cost and timing, and the permission prices only
// the cast that claims its offer (CastPermission.ForClaim), so the
// prompt's cast is the claimed one and no other. Every other zone answers
// "", false: the permission is the reason the cast is legal at all,
// whatever it claims. Caller must hold g.mu.
func (g *Game) ResolutionCastClaimLocked(chooser, id uuid.UUID) (string, bool) {
	c, zone, ok := g.ResolutionCastCandidateLocked(chooser, id)
	if !ok || zone != ZoneHand {
		return "", false
	}
	p := g.playerByIDLocked(chooser)
	for i := range p.CastPermissions {
		perm := &p.CastPermissions[i]
		if perm.Scope == ScopeCards && perm.Zone == ZoneHand && perm.NamesCard(c) && perm.AltCostKey != "" {
			return perm.AltCostKey, true
		}
	}
	return "", false
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
	// #4250: a hand card is cast through the offer and no other way.
	if claim, ok := g.ResolutionCastClaimLocked(caster, cardID); ok && AltCostKeyword(params.AlternativeCost) != claim {
		params.AlternativeCost = claim
	}
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
	// #4250: the rest of the resolution after the casts.
	g.runResolutionCastThenLocked(c.PickThen, c.Chooser, c.Source, c.OptionCarry)
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
