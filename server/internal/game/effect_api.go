package game

import "github.com/google/uuid"

// effect_api.go is the exported "locked-context" API that the
// server/internal/cards/effects package calls from inside an
// already-locked mutation frame (the resolution path holds g.mu
// write). Every method here assumes the caller holds g.mu — calling
// a public locking mutator from within a resolution would deadlock.
//
// The naming convention is `XxxForEffectLocked`. Each method is a
// thin wrapper over existing internal helpers that also emits the
// right Event so consumers of Game.Events see primitive-grained
// mutations without having to pattern-match on paired ZoneMove +
// counter events.
//
// Added in S14 sub-PR 2 as the surface the declarative card catalog
// reaches through. Not part of the external HTTP/WS API — the
// resolution path and unit tests are the only callers.

// PlayerByIDForEffect looks up a player by ID without touching the
// lock. Returns nil if no player with that ID is seated. Used by
// primitives to verify target legality and route graveyard moves.
func (g *Game) PlayerByIDForEffect(id uuid.UUID) *Player {
	return g.playerByIDLocked(id)
}

// StackItemForEffect looks up a stack item by its ID. Returns nil
// if no such item is on the stack. Used by effects that need to
// peek at a countered spell's controller / owner before
// CounterTarget deletes the StackMeta entry (Swan Song).
func (g *Game) StackItemForEffect(id uuid.UUID) *StackItem {
	if g.StackMeta == nil {
		return nil
	}
	return g.StackMeta[id]
}

// LookupCardForEffect returns a value copy of the card with the
// given instance ID from whichever zone holds it, plus ok=true.
// Empty Card and ok=false when the card isn't in any tracked zone.
// Used by effects that need to read a target's printed
// characteristics BEFORE moving it (Swords to Plowshares → read
// power before exile; Path to Exile → read controller before
// exile to drive the search clause).
func (g *Game) LookupCardForEffect(cardID uuid.UUID) (Card, bool) {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return Card{}, false
	}
	for _, c := range z.Cards {
		if c.InstanceID == cardID {
			return c, true
		}
	}
	return Card{}, false
}

// RevealHandForEffect marks every card in the named player's hand
// as known to all seated players. Used by Thoughtseize / Duress
// style "reveals hand" effects. No-op if the player isn't seated
// or is eliminated.
func (g *Game) RevealHandForEffect(playerID uuid.UUID) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return
	}
	for i := range p.Hand.Cards {
		for _, viewer := range g.Seats {
			p.Hand.Cards[i].AddKnower(viewer.ID)
		}
	}
}

// FindCardZoneForEffect returns the zone a card currently lives in,
// or nil if the card is in none of the tracked zones. Lock-free —
// caller must hold g.mu. Used by primitives for the CR 608.2b
// per-slot existence re-check.
func (g *Game) FindCardZoneForEffect(cardID uuid.UUID) *Zone {
	return g.findCardZoneLocked(cardID)
}

// BattlefieldCardsForEffect returns the current battlefield card
// slice (by value — mutating it has no effect on the game). Used
// by iterated primitives such as Wrath-of-God's "for each creature
// on battlefield: destroy." Safe because Card carries maps
// (Counters, KnownBy) by reference, but no primitive in S14
// mutates those via this slice.
func (g *Game) BattlefieldCardsForEffect() []Card {
	if g.Battlefield == nil {
		return nil
	}
	out := make([]Card, len(g.Battlefield.Cards))
	copy(out, g.Battlefield.Cards)
	return out
}

// ChangePlayerLifeForEffect adjusts a player's life by delta and
// emits EventChangeLife. Lock-free.
func (g *Game) ChangePlayerLifeForEffect(source, playerID uuid.UUID, delta int) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.ChangeLife(delta)
	g.EmitEvent(Event{
		Kind:   EventChangeLife,
		Source: source,
		Target: playerID,
		Amount: delta,
	})
	return nil
}

// DealDamageToPlayerForEffect writes amount damage to a player's
// life (via ChangeLife -amount) and emits an EventDealDamage. The
// damage is recorded before the life change so the event log reads
// "damage dealt → life changed." SBA check fires via the caller
// (effects run inside resolveTopOfStackLocked, which pairs with
// runStateChecks on the surrounding priority boundary).
func (g *Game) DealDamageToPlayerForEffect(source, playerID uuid.UUID, amount int) error {
	if amount <= 0 {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	g.EmitEvent(Event{
		Kind:   EventDealDamage,
		Source: source,
		Target: playerID,
		Amount: amount,
	})
	p.ChangeLife(-amount)
	return nil
}

// DealDamageToCreatureForEffect marks amount damage on a
// battlefield creature. Emits EventDealDamage. The SBA pass fires
// on the surrounding priority boundary (resolution path already
// bookends with runStateChecks), so lethal damage routes the card
// via the normal SBA loop rather than a bespoke kill-now path.
func (g *Game) DealDamageToCreatureForEffect(source, cardID uuid.UUID, amount int) error {
	if amount <= 0 {
		return nil
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].DamageMarked += amount
			g.EmitEvent(Event{
				Kind:   EventDealDamage,
				Source: source,
				Target: cardID,
				Amount: amount,
			})
			return nil
		}
	}
	return ErrCardNotFound
}

// DrawNForEffect draws n cards for the given player, emitting one
// EventDrawCard per card (drawCardLocked already emits). Returns
// an error only on the first failure — partial draws are allowed
// (ErrZoneEmpty on the Nth card flags LosesAtNextSBA for the loss
// on the next SBA pass, which is already the drawCardLocked
// behaviour).
func (g *Game) DrawNForEffect(playerID uuid.UUID, n int) error {
	for i := 0; i < n; i++ {
		if err := g.drawCardLocked(playerID); err != nil {
			if err == ErrZoneEmpty {
				// Flag set, stop drawing. SBA loop will handle the loss.
				return nil
			}
			return err
		}
	}
	return nil
}

// DiscardRandomForEffect discards n cards from playerID's hand at
// random order (top of the stack — hand isn't visibly ordered to
// opponents, so the RNG choice isn't observable). Emits
// EventDiscardCard + EventZoneMove per card. If the hand has fewer
// than n cards, discards all of them.
func (g *Game) DiscardRandomForEffect(playerID uuid.UUID, n int) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	for i := 0; i < n; i++ {
		if p.Hand.Size() == 0 {
			return nil
		}
		// Pop a random index. The RNG is the captured per-game source
		// so deterministic tests stay deterministic.
		idx := 0
		if p.Hand.Size() > 1 && g.rng != nil {
			idx = g.rng.IntN(p.Hand.Size())
		}
		cardID := p.Hand.Cards[idx].InstanceID
		if _, err := MoveCard(p.Hand, p.Graveyard, cardID); err != nil {
			return err
		}
		g.markCardKnownInZoneLocked(p.Graveyard, cardID)
		g.EmitEvent(Event{
			Kind:    EventDiscardCard,
			Actor:   playerID,
			CardID:  cardID,
			OldZone: ZoneHand,
			NewZone: ZoneGraveyard,
		})
	}
	return nil
}

// MillNForEffect moves n cards from the top of playerID's library
// to their graveyard. Emits EventMill per card. An empty library
// during the mill sets LosesAtNextSBA (CR 704.5b-equivalent read
// from the top of an empty library) via the same path drawCardLocked
// uses.
func (g *Game) MillNForEffect(playerID uuid.UUID, n int) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	for i := 0; i < n; i++ {
		if p.Library.Size() == 0 {
			p.LosesAtNextSBA = true
			return nil
		}
		c, err := p.Library.PopTop()
		if err != nil {
			return err
		}
		p.Graveyard.PushTop(c)
		g.markCardKnownInZoneLocked(p.Graveyard, c.InstanceID)
		g.EmitEvent(Event{
			Kind:    EventMill,
			Actor:   playerID,
			CardID:  c.InstanceID,
			OldZone: ZoneLibrary,
			NewZone: ZoneGraveyard,
		})
	}
	return nil
}

// DestroyPermanentForEffect routes a battlefield permanent to its
// owner's graveyard (or exile if the owner is no longer seated).
// Wrapper around the existing internal helper — exposed so effect
// primitives can call it from an already-locked context.
func (g *Game) DestroyPermanentForEffect(cardID uuid.UUID) error {
	return g.routeBattlefieldCardToOwnerGraveyardLocked(cardID)
}

// ExileCardForEffect moves a card from whatever zone it's in to
// the shared exile zone. The source zone is found by scanning; if
// the card is already in exile, the call is a no-op.
func (g *Game) ExileCardForEffect(cardID uuid.UUID) error {
	src := g.findCardZoneLocked(cardID)
	if src == nil {
		return ErrCardNotFound
	}
	if src == g.Exile {
		return nil
	}
	if _, err := MoveCard(src, g.Exile, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(g.Exile, cardID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		CardID:  cardID,
		OldZone: src.Kind,
		NewZone: ZoneExile,
	})
	if src.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventLTB, CardID: cardID})
	}
	return nil
}

// BounceToHandForEffect moves a card from wherever it is to its
// owner's hand. Used by Unsummon and similar. If the owner is no
// longer seated, the call returns ErrPlayerNotFound without moving
// the card.
func (g *Game) BounceToHandForEffect(cardID uuid.UUID) error {
	src := g.findCardZoneLocked(cardID)
	if src == nil {
		return ErrCardNotFound
	}
	var ownerID uuid.UUID
	for _, c := range src.Cards {
		if c.InstanceID == cardID {
			ownerID = c.Owner
			break
		}
	}
	owner := g.playerByIDLocked(ownerID)
	if owner == nil {
		return ErrPlayerNotFound
	}
	if src == owner.Hand {
		return nil
	}
	if _, err := MoveCard(src, owner.Hand, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(owner.Hand, cardID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   owner.ID,
		CardID:  cardID,
		OldZone: src.Kind,
		NewZone: ZoneHand,
	})
	if src.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventLTB, CardID: cardID, Actor: owner.ID})
	}
	return nil
}

// TapTargetForEffect taps a battlefield card. Wrapper around the
// internal path; emits EventTapCard.
func (g *Game) TapTargetForEffect(cardID uuid.UUID) error {
	return g.setTapStateLocked(cardID, true)
}

// UntapTargetForEffect untaps a battlefield card.
func (g *Game) UntapTargetForEffect(cardID uuid.UUID) error {
	return g.setTapStateLocked(cardID, false)
}

// setTapStateLocked is the internal helper behind TapCard and the
// TapTarget / UntapTarget effect primitives. Caller must hold g.mu.
func (g *Game) setTapStateLocked(cardID uuid.UUID, tapped bool) error {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].Tapped = tapped
			if tapped {
				g.EmitEvent(Event{Kind: EventTapCard, CardID: cardID})
			} else {
				g.EmitEvent(Event{Kind: EventUntapCard, CardID: cardID})
			}
			return nil
		}
	}
	return ErrCardNotFound
}

// CounterTargetForEffect counters a stack item. Discriminates
// between spell and ability via the stored StackItem.Kind and
// delegates to the appropriate internal helper. Spell counters
// route to the item's owner's graveyard by default.
func (g *Game) CounterTargetForEffect(stackID uuid.UUID) error {
	item, ok := g.StackMeta[stackID]
	if !ok || item == nil {
		return ErrCardNotOnStack
	}
	switch item.Kind {
	case StackItemSpell:
		return g.counterSpellLocked(stackID, nil)
	case StackItemActivated, StackItemTriggered:
		return g.counterAbilityLocked(stackID)
	default:
		return ErrCardNotOnStack
	}
}

// counterSpellLocked is the lock-free body of CounterSpell. Caller
// must hold g.mu.
func (g *Game) counterSpellLocked(spellID uuid.UUID, dst *ZoneRef) error {
	item, ok := g.StackMeta[spellID]
	if !ok || item == nil || item.Kind != StackItemSpell {
		return ErrCardNotOnStack
	}
	if g.Stack == nil || !g.Stack.Contains(spellID) {
		return ErrCardNotOnStack
	}
	var destZone *Zone
	if dst == nil {
		owner := g.playerByIDLocked(item.Owner)
		if owner == nil {
			destZone = g.Exile
		} else {
			destZone = owner.Graveyard
		}
	} else {
		if dst.Kind == ZoneBattlefield || dst.Kind == ZoneStack {
			return ErrInvalidStackDestination
		}
		destZone = g.zoneFromRefLocked(*dst)
		if destZone == nil {
			return ErrZoneNotFound
		}
	}
	if _, err := MoveCard(g.Stack, destZone, spellID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(destZone, spellID)
	delete(g.StackMeta, spellID)
	g.recomputeSplitSecondLocked()
	g.EmitEvent(Event{
		Kind:   EventCounterSpell,
		Target: spellID,
		CardID: spellID,
	})
	return nil
}

// counterAbilityLocked is the lock-free body of CounterAbility.
// Caller must hold g.mu.
func (g *Game) counterAbilityLocked(abilityID uuid.UUID) error {
	item, ok := g.StackMeta[abilityID]
	if !ok || item == nil {
		return ErrCardNotOnStack
	}
	if item.Kind != StackItemActivated && item.Kind != StackItemTriggered {
		return ErrCardNotOnStack
	}
	source := item.SourceCardID
	delete(g.StackMeta, abilityID)
	g.recomputeSplitSecondLocked()
	g.EmitEvent(Event{
		Kind:   EventCounterSpell,
		Source: source,
		Target: abilityID,
	})
	return nil
}

// AddCounterForEffect adds (or removes, via negative delta) the
// named counter on a card. Zero deltas are no-ops. Emits
// EventCounterPlaced with the post-change count.
func (g *Game) AddCounterForEffect(cardID uuid.UUID, name string, delta int) error {
	if delta == 0 {
		return nil
	}
	if name == "" {
		return ErrInvalidParam
	}
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return ErrCardNotFound
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID != cardID {
			continue
		}
		if z.Cards[i].Counters == nil {
			z.Cards[i].Counters = make(map[string]int)
		}
		z.Cards[i].Counters[name] += delta
		newAmount := z.Cards[i].Counters[name]
		if z.Cards[i].Counters[name] <= 0 {
			delete(z.Cards[i].Counters, name)
			if len(z.Cards[i].Counters) == 0 {
				z.Cards[i].Counters = nil
			}
		}
		g.EmitEvent(Event{
			Kind:   EventCounterPlaced,
			Target: cardID,
			Label:  name,
			Amount: newAmount,
		})
		return nil
	}
	return ErrCardNotFound
}

// ReturnFromGraveyardForEffect moves a card from a player's
// graveyard to one of: owner's hand (default), battlefield (for
// Reanimate-style effects — not used in S14 catalog), or library
// top (top-of-library). The dest ZoneKind is one of ZoneHand,
// ZoneBattlefield, ZoneLibrary. Errors if the card isn't in a
// graveyard.
func (g *Game) ReturnFromGraveyardForEffect(cardID uuid.UUID, dest ZoneKind) error {
	src := g.findCardZoneLocked(cardID)
	if src == nil || src.Kind != ZoneGraveyard {
		return ErrCardNotFound
	}
	// Identify the graveyard's owner so we can route to the same
	// player's hand / library. (Graveyards are per-player; Zone.Owner
	// holds the ID.)
	ownerID := src.Owner
	owner := g.playerByIDLocked(ownerID)
	if owner == nil {
		return ErrPlayerNotFound
	}
	var destZone *Zone
	switch dest {
	case ZoneHand:
		destZone = owner.Hand
	case ZoneLibrary:
		destZone = owner.Library
	case ZoneBattlefield:
		destZone = g.Battlefield
	default:
		return ErrZoneNotFound
	}
	if _, err := MoveCard(src, destZone, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(destZone, cardID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   ownerID,
		CardID:  cardID,
		OldZone: ZoneGraveyard,
		NewZone: destZone.Kind,
	})
	if destZone.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventETB, Actor: ownerID, CardID: cardID})
	}
	return nil
}

// SearchLibraryForEffect scans playerID's library for the first
// (up to `limit`) cards matching `pred`, moves them to the given
// destination zone, reveals (via KnownBy) to all seated players
// when `reveal` is true, and shuffles the library when `shuffle`
// is true.
//
// dest is one of ZoneHand, ZoneBattlefield, ZoneLibrary (library-
// top is not distinguishable from library via ZoneKind; a future
// "top N" primitive can extend this). Used by Cultivate (dest=Hand,
// limit=2, reveal=true, shuffle=true), Demonic Tutor (dest=Hand,
// limit=1, reveal=false, shuffle=true), Vampiric Tutor (dest=
// Library, limit=1, ...).
//
// Sandbox simplification: the picker is deterministic — first match
// wins. A real "you choose" UI is deferred to S22. Effects that
// need a specific card pass a tight predicate (e.g. basic land type).
func (g *Game) SearchLibraryForEffect(
	playerID uuid.UUID,
	pred func(Card) bool,
	dest ZoneKind,
	limit int,
	reveal bool,
	shuffle bool,
) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if limit <= 0 {
		limit = 1
	}
	var destZone *Zone
	switch dest {
	case ZoneHand:
		destZone = p.Hand
	case ZoneBattlefield:
		destZone = g.Battlefield
	case ZoneLibrary:
		destZone = p.Library
	default:
		return ErrZoneNotFound
	}
	// Collect the matching card IDs up to limit. Walk the library
	// slice without mutating, then move by ID.
	matchIDs := make([]uuid.UUID, 0, limit)
	for _, c := range p.Library.Cards {
		if len(matchIDs) >= limit {
			break
		}
		if pred == nil || pred(c) {
			matchIDs = append(matchIDs, c.InstanceID)
		}
	}
	for _, id := range matchIDs {
		if reveal {
			// All seated players see the card identity (CR: "reveal").
			for i := range p.Library.Cards {
				if p.Library.Cards[i].InstanceID != id {
					continue
				}
				for _, seat := range g.Seats {
					p.Library.Cards[i].AddKnower(seat.ID)
				}
				break
			}
		}
		if destZone != p.Library {
			if _, err := MoveCard(p.Library, destZone, id); err != nil {
				return err
			}
			g.markCardKnownInZoneLocked(destZone, id)
			g.EmitEvent(Event{
				Kind:    EventZoneMove,
				Actor:   playerID,
				CardID:  id,
				OldZone: ZoneLibrary,
				NewZone: destZone.Kind,
			})
			if destZone.Kind == ZoneBattlefield {
				g.EmitEvent(Event{Kind: EventETB, Actor: playerID, CardID: id})
			}
		}
	}
	g.EmitEvent(Event{
		Kind:   EventSearchLibrary,
		Actor:  playerID,
		Amount: len(matchIDs),
	})
	if shuffle {
		p.Library.Shuffle(g.rng)
		clearKnownInZoneLocked(p.Library)
	}
	return nil
}

// CreateTokenForEffect puts n freshly-minted tokens onto the
// battlefield under `controller`'s control. Each token has a new
// InstanceID, empty ScryfallID (tokens aren't in the Scryfall-
// printing index), and KnownBy pre-populated with every seated
// player (tokens are always public). The emitted EventTokenCreated
// references the first token ID — callers loop for the others via
// the event stream.
func (g *Game) CreateTokenForEffect(controller uuid.UUID, template Card, n int) error {
	if n <= 0 {
		return nil
	}
	for i := 0; i < n; i++ {
		tok := template
		tok.InstanceID = uuid.New()
		tok.Owner = controller
		tok.Controller = controller
		tok.Counters = nil
		tok.KnownBy = nil
		for _, seat := range g.Seats {
			tok.AddKnower(seat.ID)
		}
		g.Battlefield.PushTop(tok)
		g.EmitEvent(Event{
			Kind:   EventTokenCreated,
			Actor:  controller,
			CardID: tok.InstanceID,
		})
		g.EmitEvent(Event{
			Kind:   EventETB,
			Actor:  controller,
			CardID: tok.InstanceID,
		})
	}
	return nil
}
