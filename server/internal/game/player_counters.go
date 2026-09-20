package game

import (
	"errors"

	"github.com/google/uuid"
)

// player_counters.go is ADR 0056 Decision 5: counters on PLAYERS go
// through the same CR 614 replacement window counters on permanents
// already went through, and every counter placement — on a player or
// on a permanent — now says WHO PUT IT.
//
// Before this file, a player counter was a map write. AddPlayerCounter
// and AddPlayerCounterForEffect adjusted Player.Counters, mirrored the
// legacy Poison / Energy ints and returned; no event was emitted, so
// nothing could replace the placement and nothing could trigger on it,
// and — the reason it finally mattered — the layer listener never
// bumped, so a static that reads an opponent's poison count
// ("corrupted", Vishgraz's "+1/+1 for each poison counter your
// opponents have") went stale until some unrelated permanent moved.
//
// Three facts change hands here:
//
//   - A player counter is a RepEventCounter with CounterPlayer set
//     instead of CounterTarget. One event kind for both targets, not a
//     new one, because the cards that replace counters on players
//     (Vorinclex, Lae'zel, Solemnity, Melira) are the same cards that
//     replace counters on permanents — see ADR 0056 Decision 5 for why
//     every EXISTING counter replacement is safe against a player
//     event: each one opens with g.LookupCardForEffect(ev.CounterTarget),
//     which fails for uuid.Nil. player_counters_test.go holds every
//     registered replacement to that.
//   - The placer rides the event (CounterPlacer) and lands on the
//     emitted Event's Actor. CR 120.3b and 120.3d say the SOURCE'S
//     CONTROLLER puts the counters, and until now nothing could say
//     that: Vorinclex keyed "you put" on the TARGET's controller and
//     shipped a caveat about it, Lae'zel read the last resolution and
//     counted uuid.Nil as "you", and Nest of Scarabs credited whoever
//     resolved anything most recently.
//   - The placement emits EventPlayerCounterPlaced, which is a
//     DIFFERENT kind from EventCounterPlaced (see its doc in events.go)
//     and carries the DELTA THAT LANDED rather than the new total.
//
// The counters that the CR 120.3 damage results put on a creature and
// on a player — infect, wither and toxic — are ADR 0056's PR 2. They
// land through exactly this window, which is why it arrives first.

// AddPlayerCounterForEffect is the effect-time counterpart of
// AddPlayerCounter: it adjusts a player-level counter (poison, energy,
// experience, rad) from inside a resolving effect, where the caller
// already holds the write lock.
//
// The placer is unknown, so the placement carries uuid.Nil and a
// "you put" replacement falls back to its own heuristic. A caller that
// KNOWS whose effect is placing the counters — a proliferate, a damage
// result — calls AddPlayerCounterByForEffect instead.
//
// Caller must hold g.mu.
func (g *Game) AddPlayerCounterForEffect(playerID uuid.UUID, name string, delta int) error {
	return g.AddPlayerCounterByForEffect(uuid.Nil, playerID, name, delta)
}

// AddPlayerCounterByForEffect is AddPlayerCounterForEffect with the
// CR 120.3b / 120.3d placer named: `placer` is the player who PUTS or
// GIVES the counters, which is the source's controller for a damage
// result and the proliferating player for CR 701.34.
//
// Two deliberate properties, both inherited from the card-counter path
// this now mirrors:
//
//   - It does not run state checks. Effects resolve inside
//     resolveTopOfStackLocked, which pairs with runStateChecks at the
//     surrounding priority boundary — the same contract
//     DealDamageToPlayerForEffect documents. A player proliferated to
//     ten poison therefore loses at that boundary, not mid-effect. The
//     one place that DOES sweep is the CR 616 resume, because
//     answering a prompt is an action boundary (pending_choice.go).
//   - It CAN PAUSE. A window with two applicable replacements in it
//     queues a CR 616 ordering prompt and returns nil with NO counters
//     placed; the resume lands them. errReplacementPending is the
//     prompt queue's business, not the caller's, exactly as
//     AddCounterForEffect treats it.
//
// Caller must hold g.mu.
func (g *Game) AddPlayerCounterByForEffect(placer, playerID uuid.UUID, name string, delta int) error {
	if name == "" {
		return ErrInvalidParam
	}
	if delta == 0 {
		return nil
	}
	// Checked BEFORE the window opens, so a player who is not there is
	// the same error it has always been rather than a CR 616 prompt put
	// to nobody. An eliminated seat is "not there" (CR 800.4a) even
	// though g.Seats keeps the row.
	switch p := g.playerByIDLocked(playerID); {
	case p == nil:
		return ErrPlayerNotFound
	case p.Eliminated:
		return ErrPlayerEliminated
	}
	ev := &ReplacementEvent{
		Kind:          RepEventCounter,
		CounterPlayer: playerID,
		CounterPlacer: placer,
		CounterName:   name,
		CounterDelta:  delta,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	return g.applyPlayerCounterLocked(out)
}

// applyPlayerCounterLocked is the landed body of a settled player
// counter placement: the map write, the legacy int mirror and the
// event. The counterpart of applyCounterLocked, and reached from the
// same two places — the unpaused exit of the window above and the
// CR 616 resume — so a paused placement and an unpaused one cannot
// drift apart.
//
// The event carries the delta THAT LANDED, which is not ev.CounterDelta
// when a removal would have driven the count below zero: "remove two
// poison counters" from a player with one is a change of one, and a
// trigger counting counters removed must not be told two. A placement
// that moves nothing emits nothing, so a removal from zero neither logs
// nor bumps the layer version.
//
// Caller must hold g.mu.
func (g *Game) applyPlayerCounterLocked(ev *ReplacementEvent) error {
	if ev == nil || ev.CounterName == "" {
		return ErrInvalidParam
	}
	if ev.CounterDelta == 0 {
		return nil
	}
	// Read the player back OFF THE EVENT rather than off the argument
	// the caller started with: a replacement is free to have redirected
	// the counters to somebody else, the same way the life pipeline
	// re-reads ev.LifePlayer.
	//
	// An ELIMINATED seat is gone as far as the rules are concerned
	// (CR 800.4a) and is never removed from g.Seats, so the nil check
	// alone would let a player who conceded between a CR 616 prompt and
	// its answer take poison counters — #808's bug on the life side,
	// and the same one here. Both spellings of "the player is gone" are
	// terminal, and the resume treats them as "the placement simply
	// does not happen" rather than failing an action whose prompt is
	// already dequeued.
	p := g.playerByIDLocked(ev.CounterPlayer)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	before := p.Counters[ev.CounterName]
	next := before + ev.CounterDelta
	if next < 0 {
		next = 0
	}
	setPlayerCounterLocked(p, ev.CounterName, next)
	mirrorLegacyPlayerCounterLocked(p, ev.CounterName, next)
	g.emitPlayerCounterDeltaLocked(ev.CounterPlayer, ev.CounterName, before, next, ev.CounterPlacer, ev.Source)
	return nil
}

// applyResolvedCounterLocked is the ONE settled exit of a
// RepEventCounter window, for both halves of the kind: a player
// placement when CounterPlayer is set, a card placement otherwise.
//
// Every entry point that opens the window — AddCounter, the two
// *ForEffect helpers, and the CR 616 resume — lands through here, so
// "which half of the kind is this" is decided in one place. Reading
// CounterPlayer rather than CounterTarget is deliberate: a card
// placement leaves CounterPlayer at uuid.Nil, and a player placement
// leaves CounterTarget at uuid.Nil, so the two are never ambiguous and
// a malformed event with both set is treated as the player placement
// it names rather than silently writing a counter onto a card nobody
// asked about.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedCounterLocked(ev *ReplacementEvent) error {
	if ev == nil {
		return ErrInvalidParam
	}
	if ev.CounterPlayer != uuid.Nil {
		return g.applyPlayerCounterLocked(ev)
	}
	return g.applyCounterByLocked(ev.CounterTarget, ev.CounterName, ev.CounterDelta, ev.CounterPlacer, ev.Source)
}

// mirrorLegacyPlayerCounterLocked keeps the two player-counter ints
// that predate Player.Counters in step with the map. Poison and energy
// are the only two, and both are read by code older than S13.2 (the
// SBA loop's fast read, the set_poison / set_energy actions, the wire's
// PlayerView).
//
// Caller must hold g.mu.
func mirrorLegacyPlayerCounterLocked(p *Player, name string, value int) {
	switch name {
	case CounterPoison:
		p.Poison = value
	case CounterEnergy:
		p.Energy = value
	}
}

// emitPlayerCounterDeltaLocked emits EventPlayerCounterPlaced for a
// change from `before` to `after`, and nothing at all when the two are
// equal.
//
// The shared emit of every player-counter path: the replacement-window
// body above and the two sandbox verbs, which skip the window
// (set_poison is SET semantics and there is nothing for a replacement
// to say about "make the total 7") but still owe the table the event
// and the layer bump that rides it.
//
// Caller must hold g.mu.
func (g *Game) emitPlayerCounterDeltaLocked(playerID uuid.UUID, name string, before, after int, placer, source uuid.UUID) {
	if after == before {
		return
	}
	g.EmitEvent(Event{
		Kind:   EventPlayerCounterPlaced,
		Target: playerID,
		Label:  name,
		Amount: after - before,
		Actor:  placer,
		Source: source,
	})
}
