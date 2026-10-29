package game

import (
	"errors"

	"github.com/google/uuid"
)

// counter_spell.go — the CR 614 window on a spell about to be COUNTERED
// (RepEventCounterSpell, #4056, ADR 0113's 2026-10-19 amendment).
//
// # What was missing
//
// Guile prints "If a spell or ability you control would counter a spell,
// instead exile that spell and you may play that card without paying its
// mana cost." That is a replacement effect on the act of countering
// (CR 614.1a, CR 701.6a). The engine had a window for the spell's MOVE
// to the graveyard and none for the counter, and replacing the move is
// a different thing: the spell would still be countered, so "can't be
// countered" and "whenever a spell is countered" would read it wrong,
// and the move does not know WHO is countering, which Guile's "you
// control" and its free play both need.
//
// # Where it opens
//
// In the three effect counter verbs (CounterTargetForEffect,
// CounterTargetToZoneForEffect, CounterSpellToLibraryKeyedForEffect), the
// ones a resolving spell or ability reaches, so a hard counter, a
// "counter unless pays" that was declined, a ward and the destination
// counters (Remand, Memory Lapse, Hinder) all pass through it. It opens
// AFTER the "can't be countered" gate: a spell that can't be countered
// is not countered, so there is nothing to replace (CR 101.2), and it
// opens BEFORE the spell leaves the stack.
//
// It does not open for the sandbox's counter action (Game.CounterSpell),
// which is not a spell or ability, nor for a spell that is "countered"
// by the rules on resolution (CR 608.2b), which no verb performs.
// Abilities are not covered: "counter a spell".
//
// # What settles
//
//	no replacement      the route the verb built, unchanged: the spell
//	                    goes where the counter sends it, countered
//	cancelled           the spell stays on the stack, and the caller's
//	                    continuation hears that nothing landed
//	counterInstead      the spell is exiled WITHOUT being countered (no
//	                    EventCounterSpell, so nothing keyed to "countered"
//	                    fires), and the player whose spell or ability
//	                    did the countering (not the spell's owner) may
//	                    cast it for free
//
// Two replacements that disagree are a CR 616.1 ordering question put to
// the countered spell's controller (the affected object's controller).
// The window can pause, with nothing moved, and the resume reaches
// finishCounterSpellLocked, the function the unpaused path runs.
//
// # The free play
//
// "You may play that card without paying its mana cost" is made while the
// countering spell or ability resolves (CR 608.2g: "If an effect
// specifically instructs or allows a player to cast a spell during
// resolution, they do so by following the steps in rules 601.2a-i, except
// no player receives priority after it's cast"). So once the card has
// landed in exile the window asks the resolution_cast prompt over it
// (OfferResolutionCastForEffect, #4286, the door every other
// resolution-time "you may cast it" takes since #4250): the countering
// player casts it there and then, before the counterspell finishes, or
// declines and the card stays in exile. The permission is
// FreeCastFromExilePermission: Cost "{0}" claimed as a granted cost,
// cast-only (a spell is never a land) and, for the pass-closed fallback,
// closing on the holder's next priority pass with the card staying in
// exile. It names the countering player, not the spell's owner, and is
// offered only if the card really landed in exile (a commander whose owner
// took the command zone, CR 903.9, was not exiled and is not offered).
//
// Every counter the engine has reaches this window inside an open
// resolution, because counterSpellByEffectLocked opens it only when a
// resolving spell or ability is the cause (resolutionCauseLocked): a hard
// counter, a declined counter-unless-pays (whose prompt holds the
// countering spell's resolution open), a ward's triggered ability, and the
// destination counters (Remand, Memory Lapse, Hinder, Devious Cover-Up).
// A CR 616.1 ordering prompt put to the countered spell's controller holds
// the resolution open too, so the resume that finishes the counter offers
// the cast inside it. The one case that is not open is a cause a caller
// carried past the end of its resolution (carriedCause, #4279: a
// non-blocking pay-unless answered after the table moved on). There the
// door answers false and the pass-closed grant is what is left.

// counterInstead is a replacement's answer to a counter: not countered,
// exiled, and a player may play the card for free. Plain data; the
// event owns it.
type counterInstead struct {
	// mayPlayFreeBy is who may play the exiled card without paying its
	// mana cost. uuid.Nil is "nobody": a bare exile.
	mayPlayFreeBy uuid.UUID
	// source and sourceName credit the grant to the permanent whose
	// replacement did it (Guile).
	source     uuid.UUID
	sourceName string
}

// CounterSpellStillCounters reports whether ev is a counter that no
// replacement has yet turned into something else. A replacement of the
// counter asks it in AppliesTo, so a second "instead exile that spell"
// does not apply to an event the first already changed (CR 616.1f: only
// effects that would now be applicable).
func (ev *ReplacementEvent) CounterSpellStillCounters() bool {
	return ev != nil && ev.Kind == RepEventCounterSpell && ev.counterInstead == nil
}

// ExileCounteredSpellInsteadForEffect is the answer of "If a spell or
// ability you control would counter a spell, instead exile that spell
// and you may play that card without paying its mana cost" (Guile): the
// event stops being a counter. The spell is exiled and not countered,
// and `mayPlayFreeBy` may play it for free (uuid.Nil grants nothing).
// `source` and `sourceName` credit the grant.
//
// Returns ErrInvalidParam for an event that is not a counter. A second
// call on the same event changes nothing: the first answer stands.
//
// Caller must hold g.mu (Replace callbacks always do).
func (g *Game) ExileCounteredSpellInsteadForEffect(ev *ReplacementEvent, mayPlayFreeBy, source uuid.UUID, sourceName string) error {
	if ev == nil || ev.Kind != RepEventCounterSpell {
		return ErrInvalidParam
	}
	if ev.counterInstead != nil {
		return nil
	}
	ev.counterInstead = &counterInstead{mayPlayFreeBy: mayPlayFreeBy, source: source, sourceName: sourceName}
	return nil
}

// FreeCastFromExilePermission is the permission a replaced counter puts
// over the exiled card: free, any timing (CR 608.2g), cast-only and
// pass-closed, with the card staying in exile when the window shuts.
// FreeCastFromHandPermission's sibling for a card in exile. Inside an open
// resolution it is the permission the resolution_cast prompt casts under
// (#4286); outside one it is granted alone and is the whole window.
//
// Caller must hold g.mu.
func (g *Game) FreeCastFromExilePermission(player, source uuid.UUID, sourceName string) CastPermission {
	return CastPermission{
		Player:      player,
		Zone:        ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      TimingFlash,
		CastOnly:    true,
		LapseOnPass: LapseStaysInExile,
		Source:      source,
		SourceName:  sourceName,
		Label:       sourceName + " — cast it without paying its mana cost",
	}
}

// counterSpellByEffectLocked counters a spell for a resolving spell or
// ability: the route a counter makes (spellExitRouteLocked), the CR 614
// window on the counter, and then either that route or what a
// replacement made of it. `dst` and `at` are the destination override
// and library position of the destination counters; `then` is the
// caller's continuation, a registered tail run once the spell has reached
// a terminal outcome (routed, exiled instead, or left on the stack) and
// told the spell when it landed where the route asked. It rides the route
// as a one-leg route batch (zoneRoute.batchLegs, #4361 S11c), as
// ExileSpellKeyedForEffect's does, so a move paused on a CR 616 prompt is
// a restore point. A zero tail is "no continuation".
//
// With nothing resolving (the sandbox's counter action reaches
// counterSpellLocked directly, but a test or a future caller may not) no
// spell or ability is countering, so no window opens and the route runs
// as it always did.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) counterSpellByEffectLocked(spellID uuid.UUID, dst *ZoneRef, at TuckOptions, then KeyedTail) error {
	r, err := g.spellExitRouteLocked(spellID, dst, ZoneGraveyard, true, at)
	if err != nil {
		return err
	}
	if !then.IsZero() {
		r.batchLegs = &routeBatchRun{Leg: spellID, Then: snapshotKeyedTail(then)}
	}
	cause := g.resolutionCauseLocked()
	if cause.Controller == uuid.Nil {
		_, err = g.routeCardToZoneLocked(r)
		return err
	}
	source := g.resolutionSourceLocked()
	ev := &ReplacementEvent{
		Kind:            RepEventCounterSpell,
		Actor:           cause.Controller,
		Source:          source,
		CounterSpellID:  spellID,
		CounterBy:       cause.Controller,
		CounterBySource: source,
		zoneRoute:       &r,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A CR 616.1 prompt is queued and the resume owns the counter
		// now. NOTHING has happened: the spell is still on the stack and
		// the caller's continuation is still owed.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		if tailErr := g.runRouteTailLocked(ev.zoneRoute); tailErr != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: tailErr.Error()})
		}
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// Not countered: the spell stays on the stack. The caller's
		// "if that spell is countered this way" is told it was not.
		return g.runRouteTailLocked(ev.zoneRoute)
	}
	return g.finishCounterSpellLocked(out)
}

// finishCounterSpellLocked performs a settled RepEventCounterSpell: the
// route the counter would have made, or the exile a replacement put in
// its place. The unpaused path and the CR 616 resume both end here, so a
// counter that paused and one that did not cannot drift apart.
//
// Caller must hold g.mu.
func (g *Game) finishCounterSpellLocked(ev *ReplacementEvent) error {
	route := ev.zoneRoute
	if route == nil {
		return nil
	}
	// The continuation moves off the event's route, so an abandoned or
	// re-entered event cannot run it a second time. (A copy of the
	// event for an undo snapshot carries its own copy of the route.)
	r := route.detachTail()
	in := ev.counterInstead
	if in == nil {
		_, err := g.routeCardToZoneLocked(r)
		return err
	}
	spellID := ev.CounterSpellID
	if item, ok := g.StackMeta[spellID]; !ok || item == nil || item.Kind != StackItemSpell ||
		g.Stack == nil || !g.Stack.Contains(spellID) {
		// The spell left the stack while the ordering prompt was open
		// (it cannot, with the table blocked, but a restore can): there
		// is nothing to exile, and the caller is still owed its answer.
		return g.runRouteTailLocked(&r)
	}
	// #4361 S11c: the exile and the free cast that follows it are data
	// (counterExiledInstead), with the counter's own continuation as its
	// Next, so the exile paused on a CR 616 prompt is a restore point.
	after := KeyedTail{
		Ref: counterExiledInstead, Player: in.mayPlayFreeBy, Source: in.source, Word: in.sourceName,
	}
	if legs := r.batchLegs; legs != nil && legs.Then != nil {
		next := restoreKeyedTail(legs.Then)
		after.Next = &next
	}
	exile, err := g.spellExitRouteLocked(spellID, &ZoneRef{Kind: ZoneExile}, ZoneExile, false, TuckOptions{})
	if err != nil {
		return err
	}
	exile.batchLegs = &routeBatchRun{Leg: spellID, Then: snapshotKeyedTail(after)}
	_, err = g.routeCardToZoneLocked(exile)
	return err
}

// counterExiledInstead finishes a counter a replacement turned into an
// exile (ExileCounteredSpellInsteadForEffect, Guile): the countering
// player (Player; uuid.Nil for none) may cast the spell free once it is in
// exile (IDs), during the countering spell's resolution (#4286, CR
// 608.2g), and then the counter's own continuation (Next) is told the
// spell did not land where the counter would have put it. Source and Word
// are the replacement's source and its name. An on-disk key: never
// renamed, never reused.
var counterExiledInstead = RegisterTail("engine/counter-exiled-instead", func(g *Game, o TailOutcome) error {
	t := o.Tail
	if t.Player != uuid.Nil && len(o.IDs) == 1 {
		spellID := o.IDs[0]
		if c, ok := g.cardInZoneLocked(g.Exile, spellID); ok {
			perm := g.FreeCastFromExilePermission(t.Player, t.Source, t.Word)
			// #4286, CR 608.2g: the play is made during the
			// resolution of the countering spell or ability.
			asked := g.OfferResolutionCastForEffect(ResolutionCastOffer{
				Chooser:      t.Player,
				Source:       t.Source,
				Reason:       t.Word,
				Cards:        []uuid.UUID{spellID},
				Permission:   &perm,
				Keyword:      MayCastKeywordFreeCast,
				AcceptLabel:  "Cast it free",
				DeclineLabel: "Leave it in exile",
			})
			if !asked {
				g.GrantCastPermissionToCardsForEffect(perm, []Card{c})
			}
		}
	}
	return g.runKeyedTailLocked(nextTailValue(t.Next), nil, 0)
})
