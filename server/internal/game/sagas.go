package game

import "github.com/google/uuid"

// sagas.go is the S27 Saga lifecycle (CR 714): the lore counter a
// Saga enters with, the turn-based action that advances it, the
// chapter events that fire off each advance, and the state-based
// action that sacrifices a finished Saga.
//
// Everything here keys off the card's SUBTYPE, not off the catalog.
// A Saga nobody has written a Spec for still enters with a lore
// counter and still advances every precombat main phase — the
// counters are visible on the board, so a player can read the
// chapter off the card and resolve it by hand exactly as the
// pre-S14 sandbox intends. What the catalog adds is the chapter
// ABILITIES (declared with effects.ChapterTrigger) and, as a
// consequence, the final chapter number that the CR 704.5s sacrifice
// needs. A Saga with no declared chapters therefore never
// self-sacrifices, which is the conservative direction: the card
// sits on the battlefield accruing counters rather than vanishing at
// a chapter the engine guessed at.
//
// Three CR rules split across three call sites:
//
//	714.2b  "As a Saga enters ... put a lore counter on it"
//	        → sagaEntersWithLoreCounterLocked, from fireETBHookLocked
//	714.2b  "... and after your draw step, put a lore counter on it"
//	        → advanceSagasForActiveSeatLocked, from the
//	          StepPrecombatMain entry hook
//	704.5s  a Saga at or past its final chapter, with no chapter
//	        ability of its own still on the stack, is sacrificed
//	        → sagasReadyToSacrificeLocked, from stateBasedActionsLocked
//
// The "still on the stack" clause in the last one is the whole
// reason the rule is worded that way: the final chapter has to
// RESOLVE before the Saga leaves, or Elspeth Conquers Death's third
// chapter would return nothing.

// SagaSubtype is the subtype that makes a permanent a Saga.
const SagaSubtype = "saga"

// IsSaga reports whether the card is a Saga right now, after
// continuous effects. Reads the effective subtypes, so a permanent
// that was turned into a Saga (nothing prints that today) would
// count and a Saga whose types were overwritten would not.
func IsSaga(c Card) bool { return c.HasSubtype(SagaSubtype) }

// SagaFinalChapter returns the highest chapter number the card's
// catalog entry declares, or 0 when the card has no catalog entry,
// no chapter abilities, or no catalog is wired at all.
//
// Derived from the declared triggers rather than stored as its own
// field: the final chapter IS "the biggest number printed on the
// card", so deriving it means a card file that adds a chapter IV can
// never forget to bump a second declaration out of sync with it.
func SagaFinalChapter(c Card) int {
	if CatalogTriggers == nil {
		return 0
	}
	key := CatalogKey(c)
	if key == "" {
		return 0
	}
	final := 0
	for _, t := range CatalogTriggers(key) {
		if t.Chapter > final {
			final = t.Chapter
		}
	}
	return final
}

// sagaEntersWithLoreCounterLocked puts the CR 714.2b entry lore
// counter on a Saga that has just crossed onto the battlefield, and
// fires the chapter events for whatever chapters that counter
// reached.
//
// Routed through AddCounterForEffect — the CR 614 replacement
// pipeline — for the same reason the starting-loyalty stamp is: the
// counter is put on "as it enters", which is a counter-placement
// event Doubling Season replaces. A Saga entering under a Doubling
// Season really does enter on chapter II, and a direct map write
// would silently skip that.
//
// The already-has-counters guard makes the call idempotent for the
// entry paths that fire the ETB hook more than once, and keeps a
// Saga that somehow entered with counters from a different effect
// from double-stamping.
//
// Caller must hold g.mu.
func (g *Game) sagaEntersWithLoreCounterLocked(cardID uuid.UUID) {
	card := findBattlefieldCard(g, cardID)
	if card == nil || !IsSaga(*card) {
		return
	}
	if card.Counters[CounterLore] > 0 {
		return
	}
	if err := g.AddCounterForEffect(cardID, CounterLore, 1); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
		return
	}
	// Re-read: AddCounterForEffect may have gone through a
	// replacement that changed the delta, and the battlefield slice
	// may have been reallocated underneath the old pointer.
	after := findBattlefieldCard(g, cardID)
	if after == nil {
		return
	}
	g.fireSagaChaptersLocked(*after, 0, after.Counters[CounterLore])
}

// advanceSagasForActiveSeatLocked is the CR 714.2b turn-based action
// at the start of the active player's precombat main phase: one lore
// counter onto each Saga that player controls.
//
// NOT routed through the replacement pipeline, unlike the entry
// counter above. This one is a turn-based action, not an effect, and
// CR 614.1 replacements only replace events an EFFECT would cause —
// which is exactly why a Doubling Season doubles a Saga's entry
// counter and does nothing at all on subsequent turns.
//
// The set of Sagas is collected before any counter is placed:
// placing one emits a chapter event, a chapter ability can queue a
// trigger, and the trigger harvester runs synchronously inside
// EmitEvent, so holding a pointer into g.Battlefield.Cards across
// the loop would dangle.
//
// Caller must hold g.mu.
func (g *Game) advanceSagasForActiveSeatLocked() {
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return
	}
	active := g.Seats[g.Turn.ActiveSeat]
	if active == nil || active.Eliminated {
		return
	}
	var sagas []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == active.ID && IsSaga(c) {
			sagas = append(sagas, c.InstanceID)
		}
	}
	for _, id := range sagas {
		g.advanceSagaLocked(id)
	}
}

// advanceSagaLocked puts one lore counter on the named Saga and
// fires the chapter events it crossed. Caller must hold g.mu.
func (g *Game) advanceSagaLocked(cardID uuid.UUID) {
	card := findBattlefieldCard(g, cardID)
	if card == nil || !IsSaga(*card) {
		return
	}
	before := card.Counters[CounterLore]
	if err := g.applyCounterLocked(cardID, CounterLore, 1); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
		return
	}
	after := findBattlefieldCard(g, cardID)
	if after == nil {
		return
	}
	g.fireSagaChaptersLocked(*after, before, after.Counters[CounterLore])
}

// fireSagaChaptersLocked emits one EventSagaChapter per chapter
// number in (from, to], in ascending order (CR 714.2c — a chapter
// ability triggers when the counter count becomes greater than or
// equal to its number and was previously less). Chapters beyond the
// card's final one are not emitted: CR 714.4 caps the Saga at its
// last chapter, and a card that gained an extra lore counter from
// somewhere has nothing left to trigger.
//
// Caller must hold g.mu.
func (g *Game) fireSagaChaptersLocked(saga Card, from, to int) {
	if to <= from {
		return
	}
	final := SagaFinalChapter(saga)
	for n := from + 1; n <= to; n++ {
		if final > 0 && n > final {
			return
		}
		g.EmitEvent(Event{
			Kind:   EventSagaChapter,
			Actor:  saga.Controller,
			Source: saga.InstanceID,
			Target: saga.InstanceID,
			CardID: saga.InstanceID,
			Amount: n,
		})
	}
}

// sagasReadyToSacrificeLocked returns the battlefield Sagas the
// CR 704.5s state-based action sacrifices right now: lore counters
// at or past the final chapter, and no chapter ability of that Saga
// waiting on the stack or in the trigger queue.
//
// A Saga with no declared chapters (final == 0) is never returned —
// see the file header.
//
// Caller must hold g.mu.
func (g *Game) sagasReadyToSacrificeLocked() []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if !IsSaga(c) {
			continue
		}
		final := SagaFinalChapter(c)
		if final <= 0 || c.Counters[CounterLore] < final {
			continue
		}
		if g.sagaHasChapterOnStackLocked(c.InstanceID) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// sagaHasChapterOnStackLocked reports whether a triggered ability
// sourced from this Saga is still on the stack or still queued to go
// there. Both halves matter: a chapter harvested during the
// precombat-main turn-based action sits on PendingTriggers until the
// next APNAP drain, and the SBA loop runs between those two points.
// Sacrificing there would lose the final chapter entirely, which is
// the exact thing CR 704.5s's "and isn't the source of a chapter
// ability that has triggered but not yet left the stack" clause
// exists to prevent.
//
// Any triggered ability from the Saga counts, not only a chapter
// one: nothing else on a printed Saga triggers, and distinguishing
// them would need a marker on the stack item that carried no other
// information.
//
// Caller must hold g.mu.
func (g *Game) sagaHasChapterOnStackLocked(sagaID uuid.UUID) bool {
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == StackItemTriggered && item.SourceCardID == sagaID {
			return true
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == sagaID {
			return true
		}
	}
	// A chapter that is waiting on a player's answer — an optional
	// "you may" prompt, or a target pick — has triggered and has not
	// yet left the stack either. The PendingChoice holds the trigger
	// OFF both queues above until it is answered, so without this the
	// Saga would be sacrificed out from under a prompt nobody could
	// usefully answer.
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Source == sagaID {
			return true
		}
	}
	return false
}
