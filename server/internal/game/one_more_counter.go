package game

import "github.com/google/uuid"

// one_more_counter.go — a resolved "until end of turn, if one or more
// counters would be put on a permanent you control, that many plus one
// of each of those kinds of counters are put on that permanent instead"
// (Brad Boimler, Eager Ensign; Star Trek, #4973).
//
// It is Pir, Imaginative Rascal's replacement (CR 614.1a) made by a
// resolving ability rather than printed on a permanent, so it is an
// ADR 0041 P8 ScopedEffect record like Flames of the Blood Hand's
// ModGainNoLife: data, swept at cleanup, a restore point while it lives.
//
// Each counter kind opens its own RepEventCounter window, so adding one
// to the window's delta is "plus one of each of those kinds". A counter
// on a PLAYER is not on a permanent and is left alone.

// ModOneMoreCounter is "until end of turn, if one or more counters would
// be put on a permanent <Player> controls, that many plus one of each of
// those kinds are put on it instead". Reads Player; the record's
// controller is "you". Scope ScopeGame. An on-disk identity: never
// renamed, never reused.
const ModOneMoreCounter ModKind = "oneMoreCounter"

// watchCounterPlaced is a counter placement's replacement window.
var watchCounterPlaced = []EventKind{EventCounterPlaced}

// oneMoreCounterAppliesLocked is ModOneMoreCounter's AppliesTo: counters
// about to be put on a permanent the record's player controls. Passive,
// like Pir's: whoever puts the counters, the permanent's controller is
// what is asked. Caller must hold g.mu.
func oneMoreCounterAppliesLocked(g *Game, m Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventCounter || ev.CounterDelta <= 0 || ev.CounterPlayer != uuid.Nil || m.Player == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CounterTarget)
	return ok && c.Controller == m.Player
}

// ModOneMorePlusOneCounter is the narrower form (Prairie Dog, Outlaws of
// Thunder Junction, #5235): "until end of turn, if you would put one or
// more +1/+1 counters on a creature you control, put that many plus one
// +1/+1 counters on it instead". Reads Player, who must be both the
// player putting the counters (ReplacementEvent.CounterPlacer) and the
// creature's controller; another kind of counter, a noncreature
// permanent and counters an opponent puts are left alone. Scope
// ScopeGame. An on-disk identity: never renamed, never reused.
const ModOneMorePlusOneCounter ModKind = "oneMorePlusOneCounter"

// oneMorePlusOneCounterAppliesLocked is ModOneMorePlusOneCounter's
// AppliesTo: +1/+1 counters the record's player is about to put on a
// creature they control, read through the layers as the counters
// arrive. Caller must hold g.mu.
func oneMorePlusOneCounterAppliesLocked(g *Game, m Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventCounter || ev.CounterDelta <= 0 || ev.CounterPlayer != uuid.Nil ||
		ev.CounterName != CounterPlusOne || m.Player == uuid.Nil || ev.CounterPlacer != m.Player {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CounterTarget)
	return ok && c.Controller == m.Player && c.IsCreature()
}

// OneMorePlusOneCounterThisTurnForEffect registers a
// ModOneMorePlusOneCounter record for `player` until end of turn.
// Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) OneMorePlusOneCounterThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, player,
		[]Mod{{Kind: ModOneMorePlusOneCounter, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// OneMoreCounterThisTurnForEffect registers the record for `player`
// until end of turn. Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) OneMoreCounterThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, player,
		[]Mod{{Kind: ModOneMoreCounter, Player: player}}, g.UntilEndOfTurnDuration(), label)
}
