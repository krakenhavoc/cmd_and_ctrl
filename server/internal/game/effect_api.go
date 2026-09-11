package game

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
)

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

// DiscardChoiceForEffect queues a "chooser-picked" discard into the
// existing DiscardPending map. Unlike DiscardRandomForEffect (which
// discards randomly at resolution time), this path waits for the
// target to send a discard_selection action with their chosen IDs.
// The card auto-resolves (goes to graveyard) while the choice is
// pending — the pending entry outlives the spell.
//
// Cap n to the player's current hand size per CR 701.8c ("discard
// as many as you can"). Merges additively with existing pending
// entries (e.g. cleanup-step discard stacked with a Mind Rot —
// one combined modal handles both). No-op if the player isn't
// seated or is eliminated.
//
// Used by Mind Rot et al; the UI is S13.4's DiscardPromptModal
// which already watches DiscardPending for the viewer. Added in
// S14 sub-PR 5+.
func (g *Game) DiscardChoiceForEffect(playerID uuid.UUID, n int) {
	if n <= 0 {
		return
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Eliminated {
		return
	}
	if n > p.Hand.Size() {
		n = p.Hand.Size()
	}
	if n == 0 {
		return
	}
	if g.DiscardPending == nil {
		g.DiscardPending = make(map[uuid.UUID]int)
	}
	g.DiscardPending[playerID] += n
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
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	// S22: route through the CR 614 replacement pipeline, the way
	// combat damage to a player already did. Before this, damage
	// dealt to a PLAYER by a spell or ability skipped replacements
	// entirely — so a damage doubler (Angrath's Marauders) or a
	// prevention shield could never see a Lightning Bolt, only
	// combat damage and damage marked on creatures. The three other
	// damage entry points were already routed; this was the hole.
	ev := &ReplacementEvent{
		Kind:         RepEventDamage,
		Source:       source,
		DamageSource: source,
		DamageTarget: playerID,
		DamageAmount: amount,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A replacement queued a choice (CR 616 ordering); the
		// pipeline resumes when it's answered.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled || out.DamageAmount <= 0 {
		return nil
	}
	p := g.playerByIDLocked(out.DamageTarget)
	if p == nil {
		return ErrPlayerNotFound
	}
	g.EmitEvent(Event{
		Kind:   EventDealDamage,
		Source: out.DamageSource,
		Target: out.DamageTarget,
		Amount: out.DamageAmount,
	})
	p.ChangeLife(-out.DamageAmount)
	return nil
}

// DealDamageToCreatureForEffect deals amount damage to a battlefield
// PERMANENT. Emits EventDealDamage. The SBA pass fires on the
// surrounding priority boundary (the resolution path already bookends
// with runStateChecks), so lethal damage routes the card via the
// normal SBA loop rather than a bespoke kill-now path.
//
// The name says "creature" for history's sake and is now a
// misnomer — every damage-dealing card in the catalog calls it, and
// since S27 (#406) the target may equally be a planeswalker or a
// battle. What the damage does is decided by CR 120.3 in
// applyDamageToPermanentLocked: marked on a creature, loyalty off a
// planeswalker, defense off a battle, and all of those at once for a
// permanent that is more than one of them. Before that split, a
// Lightning Bolt aimed at a planeswalker incremented a number nothing
// read — a two-mana no-op that looked like it had worked.
//
// Deathtouch is not applied here. This is the non-combat path, it has
// never applied the CR 702.2c flag, and whether a deathtouch source's
// direct damage should is a separate question from this one.
func (g *Game) DealDamageToCreatureForEffect(source, cardID uuid.UUID, amount int) error {
	if amount <= 0 {
		return nil
	}
	if !g.applyDamageToPermanentLocked(cardID, amount, false) {
		return ErrCardNotFound
	}
	g.EmitEvent(Event{
		Kind:   EventDealDamage,
		Source: source,
		Target: cardID,
		Amount: amount,
	})
	return nil
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

// SacrificePermanentForEffect sacrifices a battlefield permanent on
// behalf of its controller (CR 701.17): EventSacrifice fires while
// the card is still on the battlefield, then it takes the ordinary
// route to its owner's graveyard (emitting ZoneMove + LTB, so
// dies-triggers see it and the CR 903.9 commander-zone replacement
// still gets its say).
//
// Sacrifice is not destruction — no indestructible / regeneration
// check applies, which is why this doesn't reuse the destroy path's
// naming. A card that isn't on the battlefield returns
// ErrCardNotFound and emits nothing.
//
// Caller must hold g.mu. Added in S21 sub-PR 1.
func (g *Game) SacrificePermanentForEffect(cardID uuid.UUID) error {
	return g.sacrificePermanentLocked(cardID)
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
	if src.Kind == ZoneBattlefield {
		g.snapshotLKILocked(cardID)
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
		g.EmitEvent(Event{Kind: EventLTB, CardID: cardID, NewZone: ZoneExile})
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
	if src.Kind == ZoneBattlefield {
		g.snapshotLKILocked(cardID)
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
		g.EmitEvent(Event{Kind: EventLTB, CardID: cardID, Actor: owner.ID, NewZone: ZoneHand})
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
//
// S17 sub-PR 2: routes through the replacement pipeline so a card-
// effect-driven counter placement (Hangarback Walker's ETB,
// Tamiyo's +1 stamping loyalty, …) picks up Doubling Season /
// Hardened Scales like the public AddCounter path. Caller must
// already hold g.mu.
func (g *Game) AddCounterForEffect(cardID uuid.UUID, name string, delta int) error {
	if delta == 0 {
		return nil
	}
	if name == "" {
		return ErrInvalidParam
	}
	ev := &ReplacementEvent{
		Kind:          RepEventCounter,
		CounterTarget: cardID,
		CounterName:   name,
		CounterDelta:  delta,
	}
	out, err := g.applyReplacementsLocked(ev)
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		// errReplacementPending: the prompt queue is the caller's
		// problem; return nil so the primitive keeps moving.
		if errors.Is(err, errReplacementPending) {
			return nil
		}
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	return g.applyCounterLocked(out.CounterTarget, out.CounterName, out.CounterDelta)
}

// ReturnFromGraveyardForEffect moves a card from a player's
// graveyard to one of: owner's hand (default), battlefield (for
// Reanimate-style effects), or library top (top-of-library). The
// dest ZoneKind is one of ZoneHand, ZoneBattlefield, ZoneLibrary.
// Errors if the card isn't in a graveyard.
//
// The card lands under its OWNER's control. Reanimation that says
// "under your control" must use
// ReturnFromGraveyardUnderControlForEffect instead — see the note
// there for why the difference is not cosmetic.
func (g *Game) ReturnFromGraveyardForEffect(cardID uuid.UUID, dest ZoneKind) error {
	return g.ReturnFromGraveyardUnderControlForEffect(cardID, dest, uuid.Nil)
}

// ReturnFromGraveyardUnderControlForEffect is
// ReturnFromGraveyardForEffect with an explicit controller for the
// battlefield case: "put target creature card from A GRAVEYARD onto
// the battlefield UNDER YOUR CONTROL" (Reanimate, Portal to
// Phyrexia). uuid.Nil means "under its owner's control", which is
// the older behaviour and what Zombify-style "from your graveyard"
// text wants.
//
// This exists because the two were previously the same thing, and
// that was a real bug the moment a card reached across the table:
// the destination was picked off the graveyard's owner, so
// reanimating an opponent's creature handed the creature back to the
// opponent — the single most valuable line in a reanimator deck,
// silently inverted. Hand / library destinations are unaffected;
// those genuinely do go to the owner's zones, whoever cast the
// spell.
//
// The controller is stamped BEFORE the events fire, not after, so
// the layer listener, the trigger harvester and anything watching
// EventETB all see the permanent under the right control. Actor on
// both events is the new controller for the same reason.
//
// Caller must hold g.mu.
func (g *Game) ReturnFromGraveyardUnderControlForEffect(cardID uuid.UUID, dest ZoneKind, controller uuid.UUID) error {
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
	actor := ownerID
	var entersTapped bool
	var enterCounters map[string]int
	if destZone.Kind == ZoneBattlefield {
		newController := controller
		if newController == uuid.Nil {
			newController = ownerID
		}
		if p := g.playerByIDLocked(newController); p == nil {
			newController = ownerID
		}
		// Stamp the controller BEFORE the pipeline runs. A
		// replacement that asks "is this entering permanent mine?"
		// (Authority of the Consuls) reads Card.Controller, and while
		// the card sits in the graveyard that field still names
		// whoever controlled it when it died — which, for a
		// reanimated opponent's creature, is the wrong player.
		for i := range src.Cards {
			if src.Cards[i].InstanceID == cardID {
				src.Cards[i].Controller = newController
				break
			}
		}
		actor = newController
		// #263: a reanimated permanent enters the battlefield like
		// any other, so the CR 614 entry pipeline has to run on it.
		// Skipping it meant a reanimated shockland ignored its own
		// enters-tapped clause and a creature reanimated under an
		// opponent's nose dodged Authority of the Consuls.
		ev := &ReplacementEvent{
			Kind:    RepEventMove,
			Actor:   newController,
			CardID:  cardID,
			OldZone: ZoneGraveyard,
			NewZone: ZoneBattlefield,
		}
		out, err := g.applyReplacementsLocked(ev)
		if errors.Is(err, errReplacementPending) {
			// A CR 616 ordering prompt is open. Nothing has moved —
			// the card is still in the graveyard — so the return is
			// dropped rather than stranded mid-zone. Same posture the
			// land-play and exile-return paths take.
			return nil
		}
		if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
			g.clearReplacementEventLocked(ev.ID)
			return err
		}
		defer g.clearReplacementEventLocked(ev.ID)
		if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
			return nil
		}
		entersTapped = out.EntersTapped
		enterCounters = out.EntersWithCounters
	}
	if _, err := MoveCard(src, destZone, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(destZone, cardID)
	var oracleID string
	if destZone.Kind == ZoneBattlefield {
		for i := range destZone.Cards {
			if destZone.Cards[i].InstanceID == cardID {
				destZone.Cards[i].Controller = actor
				if entersTapped {
					destZone.Cards[i].Tapped = true
				}
				oracleID = CatalogKey(destZone.Cards[i])
				break
			}
		}
		for name, n := range enterCounters {
			_ = g.AddCounterForEffect(cardID, name, n)
		}
	}
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   actor,
		CardID:  cardID,
		OldZone: ZoneGraveyard,
		NewZone: destZone.Kind,
	})
	if destZone.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventETB, Actor: actor, CardID: cardID})
		// A reanimated permanent enters the battlefield like any
		// other, so the catalog's OnETB / StartingLoyalty hook has to
		// run — otherwise reanimating Solemn Simulacrum fetches
		// nothing and reanimating a planeswalker gives it no loyalty.
		// Every other path onto the battlefield already fires this;
		// this one was the omission. A no-op for a card with no
		// oracle ID (tokens, fixtures) or no catalog entry.
		g.fireETBHookLocked(cardID, oracleID)
	}
	return nil
}

// SearchLibrarySpec is the full description of one "search your
// library for ..." effect. Every field except Player and Pred has a
// useful zero value, so the older positional entry points below stay
// one-line wrappers over it.
//
// The spec exists because a search is no longer a single synchronous
// mutation: when the library offers more matches than the effect may
// take, the searcher gets a real PendingChoiceSearchLibrary prompt
// and the rest of the effect has to wait for their answer. `Then`
// is that continuation.
type SearchLibrarySpec struct {
	// Player is the searcher — always the owner of the library being
	// searched. Assassin's Trophy makes the VICTIM search, so this is
	// not necessarily the resolving spell's controller.
	Player uuid.UUID

	// Source is the card that caused the search, used only as prompt
	// context on the wire. Optional.
	Source uuid.UUID

	// Pred filters the library. nil matches every card (Gamble).
	Pred func(Card) bool

	// Dest is one of ZoneHand, ZoneBattlefield, ZoneLibrary.
	Dest ZoneKind

	// Limit is the maximum number of cards taken. <= 0 means 1.
	Limit int

	// Reveal marks the TAKEN cards known to every seated player (CR
	// "reveal"). It never reveals the cards that were merely
	// considered — see queueSearchChoiceLocked for why that
	// distinction is load-bearing.
	Reveal bool

	// Shuffle shuffles the library once the search is finished, and
	// wipes per-card knowledge with it.
	Shuffle bool

	// TappedOnEntry forces a fetched permanent tapped. It expresses
	// the FETCHING effect's printed text ("put it onto the
	// battlefield tapped" — Cultivate, Solemn Simulacrum, Path to
	// Exile), NOT the fetched card's own enters-tapped clause; the
	// CR 614 pipeline owns that one. The two are OR-ed: a Darkslick
	// Shores fetched by Cultivate is tapped because Cultivate says
	// so, and one fetched by Skyshroud Claim is tapped only if its
	// own condition fires.
	TappedOnEntry bool

	// Optional marks a "you MAY search" (CR 701.19c). It forces the
	// prompt even when the pick is otherwise forced, so the searcher
	// can decline — Assassin's Trophy's victim keeps the right to
	// refuse the land and the shuffle.
	Optional bool

	// Reason is the prompt's banner copy. Defaults to a generic
	// "Search your library".
	Reason string

	// Validate is an extra legality check on the picked SET, for
	// clauses the per-card predicate cannot express. Myriad
	// Landscape's "two basic land cards that SHARE A LAND TYPE" is
	// the case: no predicate over one card can say it. Called with
	// the chosen cards (possibly zero of them, which is always
	// legal — you may always fail to find).
	Validate func([]Card) bool

	// Then is the rest of the effect, run once the search has
	// finished — immediately when no prompt was needed, and from
	// ResolveSearchLibrary when one was. It receives the instance
	// IDs actually taken, in the order they were taken, so an effect
	// like Fabled Passage's "then ... untap THAT LAND" can name what
	// it found instead of re-deriving it.
	Then func(g *Game, found []uuid.UUID) error
}

// SearchLibraryForEffect scans playerID's library for cards matching
// `pred`, moves up to `limit` of them to the given destination zone,
// reveals (via KnownBy) to all seated players when `reveal` is true,
// and shuffles the library when `shuffle` is true.
//
// dest is one of ZoneHand, ZoneBattlefield, ZoneLibrary (library-
// top is not distinguishable from library via ZoneKind; a future
// "top N" primitive can extend this). Used by Demonic Tutor
// (dest=Hand, limit=1, reveal=false, shuffle=true), Vampiric Tutor
// (dest=Library, limit=1, ...).
//
// The searcher picks which matches they take (S22) — see
// SearchLibraryThenForEffect. This wrapper has no continuation, so
// it suits only effects that do nothing after the search.
//
// Caller must hold g.mu.
func (g *Game) SearchLibraryForEffect(
	playerID uuid.UUID,
	pred func(Card) bool,
	dest ZoneKind,
	limit int,
	reveal bool,
	shuffle bool,
) error {
	return g.SearchLibraryForEffectWithOptions(playerID, pred, dest, limit, reveal, shuffle, false)
}

// SearchLibraryForEffectWithOptions is SearchLibraryForEffect with
// the forced-tap flag. Retained as a positional wrapper because a
// dozen catalog cards call it; new callers should build a
// SearchLibrarySpec instead.
//
// Caller must hold g.mu.
func (g *Game) SearchLibraryForEffectWithOptions(
	playerID uuid.UUID,
	pred func(Card) bool,
	dest ZoneKind,
	limit int,
	reveal bool,
	shuffle bool,
	tappedOnEntry bool,
) error {
	return g.SearchLibraryThenForEffect(SearchLibrarySpec{
		Player:        playerID,
		Pred:          pred,
		Dest:          dest,
		Limit:         limit,
		Reveal:        reveal,
		Shuffle:       shuffle,
		TappedOnEntry: tappedOnEntry,
	})
}

// SearchLibraryThenForEffect is the real entry point: it runs the
// search described by spec and then spec.Then.
//
// The searcher chooses. Until S22 the engine took the first matches
// in library order, which made a fetchland's entire strategic content
// — which dual do I want? — a property of deck-list order, and made
// Myriad Landscape's "share a land type" whatever the bottom-most
// basic happened to be. That simplification is gone: when the library
// offers a real choice, the searcher gets a prompt.
//
// "A real choice" means: more matches than the effect may take, or
// an Optional ("you may search") clause, or a Validate constraint the
// whole match set does not already satisfy. With nothing to decide —
// a Rampant Growth into a library holding exactly one Forest — the
// engine takes the match and moves on, because a modal offering one
// button is worse than no modal.
//
// Hidden information (CR 400.2): a library is a hidden zone, and the
// prompt does not change that for anybody but the searcher. The
// candidates are marked known to the SEARCHER ALONE, the per-viewer
// wire filter redacts them for every other seat, and the search
// prompt's option list is withheld from non-choosers entirely so the
// NUMBER of matches does not leak either. The searcher is always the
// library's owner, so nothing crosses the table.
//
// Caller must hold g.mu.
func (g *Game) SearchLibraryThenForEffect(spec SearchLibrarySpec) error {
	p := g.playerByIDLocked(spec.Player)
	if p == nil {
		return ErrPlayerNotFound
	}
	if spec.Limit <= 0 {
		spec.Limit = 1
	}
	if _, err := g.searchDestZoneLocked(p, spec.Dest); err != nil {
		return err
	}
	// Collect every match, not just the first `limit` — the whole
	// point of the chooser is that the searcher sees the full set.
	matches := make([]uuid.UUID, 0, len(p.Library.Cards))
	matched := make([]Card, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		if spec.Pred == nil || spec.Pred(c) {
			matches = append(matches, c.InstanceID)
			matched = append(matched, c)
		}
	}
	if len(matches) == 0 {
		// Nothing to find. Still a search: the event fires and the
		// shuffle happens, because the card said "then shuffle"
		// unconditionally.
		return g.finishSearchLocked(spec, p, nil)
	}
	if !spec.Optional && len(matches) <= spec.Limit &&
		(spec.Validate == nil || spec.Validate(matched)) {
		// No decision to make — take them all.
		found := g.executeSearchTakeLocked(spec, p, matches)
		return g.finishSearchLocked(spec, p, found)
	}
	g.queueSearchChoiceLocked(spec, p, matches)
	return nil
}

// searchDestZoneLocked resolves a search's destination ZoneKind to
// the concrete zone. Caller must hold g.mu.
func (g *Game) searchDestZoneLocked(p *Player, dest ZoneKind) (*Zone, error) {
	switch dest {
	case ZoneHand:
		return p.Hand, nil
	case ZoneBattlefield:
		return g.Battlefield, nil
	case ZoneLibrary:
		return p.Library, nil
	case ZoneGraveyard:
		// Entomb, Buried Alive, Gamble's discard half — "search your
		// library for a card, put that card into your GRAVEYARD".
		// The generic MoveCard branch below handles it unchanged;
		// only this lookup was missing, which made those cards fail
		// with ErrZoneNotFound and silently find nothing. Added with
		// the roadmap's batch 02 (#295).
		return p.Graveyard, nil
	}
	return nil, ErrZoneNotFound
}

// queueSearchChoiceLocked opens the search prompt. The candidates
// are marked known to the searcher only: they are looking through
// their own library, which is exactly the knowledge the rules grant
// them and no more. `Reveal` is deliberately NOT applied here —
// "reveal those cards" on Cultivate names the cards you take, not
// every basic you flipped past on the way, and revealing the whole
// match set would hand the table a census of the library.
//
// Caller must hold g.mu.
func (g *Game) queueSearchChoiceLocked(spec SearchLibrarySpec, p *Player, matches []uuid.UUID) {
	known := make(map[uuid.UUID]bool, len(matches))
	for _, id := range matches {
		known[id] = true
	}
	for i := range p.Library.Cards {
		if known[p.Library.Cards[i].InstanceID] {
			p.Library.Cards[i].AddKnower(spec.Player)
		}
	}
	reason := spec.Reason
	if reason == "" {
		reason = "Search your library"
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:         PendingChoiceSearchLibrary,
		Chooser:      spec.Player,
		FromPlayer:   spec.Player,
		Count:        spec.Limit,
		Source:       spec.Source,
		Reason:       reason,
		SearchCards:  append([]uuid.UUID(nil), matches...),
		SearchMax:    spec.Limit,
		searchResume: &searchResumeFrame{spec: spec},
	})
}

// executeSearchTakeLocked moves the chosen cards out of the library
// and returns the IDs that actually made it. Shared by the
// no-decision path and by ResolveSearchLibrary, so a fetched
// permanent enters identically either way.
//
// Caller must hold g.mu.
func (g *Game) executeSearchTakeLocked(spec SearchLibrarySpec, p *Player, ids []uuid.UUID) []uuid.UUID {
	destZone, err := g.searchDestZoneLocked(p, spec.Dest)
	if err != nil || destZone == p.Library {
		// ZoneLibrary is "leave it where it is" (Vampiric Tutor's
		// put-on-top is not modelled yet); only the reveal applies.
		if err == nil && spec.Reveal {
			g.revealLibraryCardsLocked(p, ids)
		}
		return nil
	}
	found := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if !p.Library.Contains(id) {
			continue
		}
		if spec.Reveal {
			g.revealLibraryCardsLocked(p, []uuid.UUID{id})
		}
		if destZone.Kind == ZoneBattlefield {
			moved, ok := g.searchEnterBattlefieldLocked(spec, p, id)
			if !ok {
				continue
			}
			found = append(found, moved)
			continue
		}
		if _, err := MoveCard(p.Library, destZone, id); err != nil {
			continue
		}
		g.markCardKnownInZoneLocked(destZone, id)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   spec.Player,
			CardID:  id,
			OldZone: ZoneLibrary,
			NewZone: destZone.Kind,
		})
		found = append(found, id)
	}
	return found
}

// searchEnterBattlefieldLocked puts one fetched card onto the
// battlefield through the CR 614 entry pipeline (#263).
//
// Before this, the search path pushed the card straight onto the
// battlefield and emitted EventZoneMove / EventETB without ever
// calling applyReplacementsLocked — so a fetched Darkslick Shores
// ignored its own "enters tapped unless" clause, every conditional
// dual cycle came in strictly better than printed, and a creature an
// opponent fetched walked past Authority of the Consuls. The ad-hoc
// TappedOnEntry flag hid it for the always-tapped cases and could not
// express a conditional one.
//
// The pipeline is consulted BEFORE the card leaves the library, for
// the same reason the exile-return path does it: a CR 616 ordering
// prompt means bailing, and bailing with the card already lifted out
// of its zone would strand it. On that bail this card is simply not
// found — the same posture the land-play path takes when its own
// pipeline call pauses.
//
// DECLARED SIMPLIFICATION — this event is NOT entryResumable, so a
// fetched permanent never gets an entry prompt.
//
// The concrete case is a fetchland cracking for a shockland. On the
// play path the controller is asked "pay 2 life so it enters
// untapped?"; fetched, they are not asked, and the land enters
// tapped. Weaker than printed, never stronger — which is exactly the
// posture ReplacementEvent.entryResumable exists to enforce, and a
// strict improvement on the old behaviour, where a fetched shockland
// ignored its entry clause altogether and arrived untapped for free.
//
// Setting entryResumable here would be actively worse, not better.
// executeEntryToBattlefieldLocked can finish the MOVE, but it knows
// nothing about the search that started it: the library would never
// be shuffled, EventSearchLibrary would never fire, and the Then
// continuation — Fabled Passage's "untap that land", Gamble's random
// discard — would never run. A missing shuffle is worse than a
// missing prompt, because it silently leaks library order.
//
// A faithful version needs the search's own continuation to survive
// the entry prompt: a second frame stacked under the replacement
// one, plus a hook in the entry resume to run it. That is its own
// change, not a rider on this one.
//
// Returns the moved card's ID and whether it moved.
//
// Caller must hold g.mu.
func (g *Game) searchEnterBattlefieldLocked(spec SearchLibrarySpec, p *Player, id uuid.UUID) (uuid.UUID, bool) {
	// The controller has to be stamped before the pipeline runs:
	// Authority of the Consuls asks whose permanent is entering, and
	// a self-replacement's condition ("unless you control two or
	// fewer other lands") counts the lands of whoever that is.
	for i := range p.Library.Cards {
		if p.Library.Cards[i].InstanceID == id {
			p.Library.Cards[i].Controller = spec.Player
			break
		}
	}
	ev := &ReplacementEvent{
		Kind:    RepEventMove,
		Actor:   spec.Player,
		CardID:  id,
		OldZone: ZoneLibrary,
		NewZone: ZoneBattlefield,
		// The fetching effect's own "put it onto the battlefield
		// TAPPED" clause is seeded onto the event rather than OR-ed
		// in after the pipeline. Same result for the inline path, and
		// it is the difference between right and wrong for any resume
		// path: a resume reads ev.EntersTapped and has no idea what
		// spell sent the card, so a Cultivate-fetched land that
		// paused for a prompt would otherwise come back untapped.
		EntersTapped: spec.TappedOnEntry,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return uuid.Nil, false
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return uuid.Nil, false
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
		return uuid.Nil, false
	}
	moved, err := MoveCard(p.Library, g.Battlefield, id)
	if err != nil {
		return uuid.Nil, false
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID != moved.InstanceID {
			continue
		}
		g.Battlefield.Cards[i].Controller = spec.Player
		// out.EntersTapped carries both inputs: the fetching effect's
		// printed "tapped" clause, seeded onto the event above, and
		// whatever the CR 614 pipeline added on top.
		if out.EntersTapped {
			g.Battlefield.Cards[i].Tapped = true
		}
		break
	}
	g.markCardKnownInZoneLocked(g.Battlefield, moved.InstanceID)
	for name, n := range out.EntersWithCounters {
		_ = g.AddCounterForEffect(moved.InstanceID, name, n)
	}
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   spec.Player,
		CardID:  moved.InstanceID,
		OldZone: ZoneLibrary,
		NewZone: ZoneBattlefield,
	})
	g.EmitEvent(Event{Kind: EventETB, Actor: spec.Player, CardID: moved.InstanceID})
	// Same omission as the reanimation path had: a fetched permanent
	// enters like any other, so its catalog OnETB hook runs.
	g.fireETBHookLocked(moved.InstanceID, CatalogKey(moved))
	return moved.InstanceID, true
}

// revealLibraryCardsLocked marks the named library cards known to
// every seated player — the CR sense of "reveal". Caller must hold
// g.mu.
func (g *Game) revealLibraryCardsLocked(p *Player, ids []uuid.UUID) {
	want := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	for i := range p.Library.Cards {
		if !want[p.Library.Cards[i].InstanceID] {
			continue
		}
		for _, seat := range g.Seats {
			p.Library.Cards[i].AddKnower(seat.ID)
		}
	}
}

// finishSearchLocked emits EventSearchLibrary, shuffles if the card
// said to, and runs the continuation. Caller must hold g.mu.
func (g *Game) finishSearchLocked(spec SearchLibrarySpec, p *Player, found []uuid.UUID) error {
	g.EmitEvent(Event{
		Kind:   EventSearchLibrary,
		Actor:  spec.Player,
		Amount: len(found),
	})
	if spec.Shuffle {
		p.Library.Shuffle(g.rng)
		// The shuffle is what un-knows the library again: whatever
		// the searcher saw while looking, they no longer know where
		// any of it is.
		clearKnownInZoneLocked(p.Library)
	}
	if spec.Then != nil {
		return spec.Then(g, found)
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

// ScryForEffect performs a scry N (CR 701.18): the player looks at the
// top N cards of their library and then decides which go to the bottom
// and in what order the rest go back on top.
//
// Returns the number of cards actually looked at, which is fewer than n
// when the library is short and zero when it is empty — a scry with an
// empty library is not an error, it simply does nothing, and no choice
// is queued.
//
// Scry is LOOK AT, not reveal. Only the scrying player becomes a knower
// of the cards, so the per-viewer wire filter redacts them for everyone
// else. Marking every seat a knower — which is what "reveal" does, and
// what a copy-paste from SearchLibraryForEffect would give you — would
// hand the whole table the top of a library, which is a real
// information advantage rather than a cosmetic slip.
//
// Nothing moves here. The cards stay on top until the choice is
// answered, so a scry left unanswered leaves the library exactly as it
// was rather than in a half-applied order.
//
// Caller must hold g.mu.
func (g *Game) ScryForEffect(playerID, source uuid.UUID, n int) int {
	return g.ScryThenForEffect(playerID, source, n, nil)
}

// ScryThenForEffect is scry N with a continuation: `after` runs once the
// player has finished putting the cards back, with the library in the
// order they chose.
//
// This exists because of the word "then". Preordain is "Scry 2, THEN
// draw a card" — the card left on top is the card drawn, so the draw
// cannot happen in the same step that queues the choice. Running it
// eagerly would draw one of the very cards the player is still deciding
// about, and would then leave the choice unanswerable, because that
// card is no longer in the library for ResolveScry to put back.
//
// `after` still runs when the scry looked at nothing (an empty
// library): the instruction after "then" is not conditional on the
// scry having had cards to look at.
//
// Caller must hold g.mu.
func (g *Game) ScryThenForEffect(playerID, source uuid.UUID, n int, after func(g *Game) error) int {
	runAfter := func() {
		if after != nil {
			if err := after(g); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					Actor:    playerID,
					Source:   source,
					ErrorMsg: err.Error(),
				})
			}
		}
	}
	if n <= 0 {
		runAfter()
		return 0
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Library == nil {
		runAfter()
		return 0
	}
	size := p.Library.Size()
	if size == 0 {
		runAfter()
		return 0
	}
	if n > size {
		n = size
	}
	// Library top is the LAST element, so the top n cards are the tail
	// — collected top-first so the chooser sees them in draw order.
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		idx := size - 1 - i
		p.Library.Cards[idx].AddKnower(playerID)
		ids = append(ids, p.Library.Cards[idx].InstanceID)
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:       PendingChoiceScry,
		Chooser:    playerID,
		FromPlayer: playerID,
		Count:      len(ids),
		Source:     source,
		Reason:     scryReason(len(ids)),
		ScryCards:  ids,
		scryResume: after,
	})
	return len(ids)
}

// scryReason is the picker's banner copy, phrased as the card prints
// it.
func scryReason(n int) string {
	switch n {
	case 1:
		return "Scry 1"
	case 2:
		return "Scry 2"
	case 3:
		return "Scry 3"
	}
	return "Scry " + strconv.Itoa(n)
}

// ReturnFromExileToBattlefieldForEffect is the other half of a
// flicker: it takes a card sitting in exile and puts it back onto the
// battlefield. ExileCardForEffect could already send a permanent
// there; until S22 nothing could bring one back, which is why
// "exile it, then return it" had no expression at all.
//
// `controller` is who it returns under — uuid.Nil (or a player who
// has left) means "under its owner's control", which is what almost
// every card says; Thassa's "under your control" passes the
// controller explicitly. `tapped` covers "return that card to the
// battlefield tapped".
//
// The returned permanent is a NEW OBJECT (CR 400.7): it gets a fresh
// InstanceID and none of the old object's battlefield state — no
// counters, no damage, no combat declarations, no summoning-sickness
// or timestamp history, and no lingering exile-play permission. That
// is the rules-correct behaviour and it is what makes blink a removal
// answer (counters fall off, a stolen creature goes home) as well as
// an ETB engine. Returns the new instance ID; a caller that needs to
// keep referring to the permanent must use it, because the pre-exile
// ID now names nothing.
//
// The entry runs through the CR 614 replacement pipeline exactly as
// an ordinary battlefield entry does, so Authority of the Consuls
// taps the blinked creature and a self "enters tapped" replacement
// still applies. The pipeline is consulted BEFORE the card is lifted
// out of exile: if it queues a CR 616 ordering prompt the card has
// not moved yet, so the return simply doesn't happen rather than
// stranding the card between zones. (That path needs two competing
// replacements on one entry; no card in the catalog produces it
// today.)
//
// Both EventETB and the catalog's OnETB hook fire, so the permanent
// re-triggers everything a fresh entry would.
//
// Caller must hold g.mu. Added in S22.
func (g *Game) ReturnFromExileToBattlefieldForEffect(cardID, controller uuid.UUID, tapped bool) (uuid.UUID, error) {
	if g.Exile == nil || !g.Exile.Contains(cardID) {
		return uuid.Nil, ErrCardNotFound
	}
	// Resolve the destination controller before the pipeline runs: a
	// replacement that asks "is the entering permanent mine?"
	// (Authority of the Consuls) reads Card.Controller, and while the
	// card sits in exile that field still names whoever controlled it
	// before it left.
	newController := controller
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID != cardID {
			continue
		}
		if newController == uuid.Nil || g.playerByIDLocked(newController) == nil {
			newController = g.Exile.Cards[i].Owner
		}
		g.Exile.Cards[i].Controller = newController
		break
	}
	ev := &ReplacementEvent{
		Kind:    RepEventMove,
		Actor:   newController,
		CardID:  cardID,
		OldZone: ZoneExile,
		NewZone: ZoneBattlefield,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A CR 616 ordering prompt is open. Nothing has moved; the
		// card stays in exile and the return is dropped.
		return uuid.Nil, nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return uuid.Nil, err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
		// Canceled, or redirected elsewhere by a replacement. There
		// is no generic "put it wherever the pipeline said" helper
		// for an exile source, so a redirect is treated as a cancel
		// rather than guessed at.
		return uuid.Nil, nil
	}
	card, err := g.Exile.Remove(cardID)
	if err != nil {
		return uuid.Nil, err
	}
	newID := uuid.New()
	card.InstanceID = newID
	card.Controller = newController
	card.Tapped = tapped || out.EntersTapped
	card.Counters = nil
	card.KnownBy = nil
	card.ExilePlay = ExilePlayPermission{}
	card.DamageMarked = 0
	card.MarkedLethalByDeathtouch = false
	card.AttackingTarget = uuid.Nil
	card.BlockingTarget = uuid.Nil
	card.GoadedBy = uuid.Nil
	card.FaceDown = false
	card.BattleX = 0
	card.BattleY = 0
	card.EnteredBattlefieldAt = 0
	card.SummonedThisTurn = false
	card.effective = nil
	g.Battlefield.PushTop(card)
	g.markCardKnownInZoneLocked(g.Battlefield, newID)
	for name, n := range out.EntersWithCounters {
		_ = g.AddCounterForEffect(newID, name, n)
	}
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   newController,
		CardID:  newID,
		OldZone: ZoneExile,
		NewZone: ZoneBattlefield,
	})
	g.EmitEvent(Event{Kind: EventETB, Actor: newController, CardID: newID})
	g.fireETBHookLocked(newID, CatalogKey(card))
	return newID, nil
}

// AddManaForEffect adds the mana a SPELL or a non-mana ability
// produces to playerID's pool — Dark Ritual's "Add {B}{B}{B}", Mana
// Drain's delayed "add an amount of {C}", Jeska's Will. Every other
// mana in the engine arrives through ActivateManaAbility (CR 605); a
// spell that adds mana resolves off the stack like any other spell
// and lands its mana here, in the same resolution frame, with the
// same EventManaAdded per unit the mana-ability path emits.
//
// `produced` is the Scryfall brace grammar ParseProducedMana reads,
// pipe syntax included: a single-colour slot goes straight into the
// pool, a multi-option slot queues the same PendingChoiceMana pick a
// Birds of Paradise activation does, narrowed to the controller's
// commander identity exactly as Treasure and Phyrexian Altar are.
// `source` is the card the mana is attributed to (the spell itself
// for Dark Ritual); it rides on each ManaToken.
//
// The pool still empties at the end of the step (CR 106.4), so mana
// added by a spell has to be spent in the step it resolved in — the
// printed behaviour, and the reason Dark Ritual is cast in a main
// phase.
//
// An eliminated or unseated player gets nothing and no error: the
// spell resolved, there was just nobody to give the mana to.
//
// Caller must hold g.mu.
func (g *Game) AddManaForEffect(playerID, source uuid.UUID, produced string) error {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Eliminated {
		return nil
	}
	slots, err := ParseProducedMana(produced)
	if err != nil {
		return err
	}
	for _, slot := range slots {
		options := slot.Options
		if len(options) == 0 {
			continue
		}
		if len(options) == 1 {
			p.ManaPool.AddMana(ManaToken{Color: options[0], Source: source})
			g.EmitEvent(Event{Kind: EventManaAdded, Actor: playerID, Source: source})
			continue
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:         PendingChoiceMana,
			Chooser:      playerID,
			FromPlayer:   playerID,
			Count:        1,
			Source:       source,
			Reason:       "Add one mana of any color",
			ColorOptions: filterPipeByCommanderIdentity(options, p),
		})
	}
	return nil
}
