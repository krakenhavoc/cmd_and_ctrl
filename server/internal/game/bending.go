package game

import "github.com/google/uuid"

// bending.go — the four Avatar bending actions as events (#5276): "whenever
// you waterbend, earthbend, firebend, or airbend" (Avatar Aang) and "if
// you've done all four this turn".
//
// Each is its own EventKind, raised where the CR says the action happens:
//
//   - EventWaterbend: a player paid a waterbend cost, however they paid it
//     (CR 701.67c), as part of a spell's total cost (a waterbend additional
//     cost, an optional one, Hama's waterbend-priced cast), an activated
//     ability's cost, or a resolving "unless they pay" (a ward). One event
//     per waterbend cost paid. The cast path raises it once its taps are
//     paid, the activation once its waterbend taps are, the pay-unless once
//     the payment is made.
//   - EventEarthbend: the earthbend's return trigger was created (CR
//     701.66b), raised by applyEarthbendLocked for a land that was on the
//     battlefield to animate. An earthbend whose target is gone does
//     nothing (the ruling) and raises nothing.
//   - EventFirebend: a firebending ability the player controls resolved
//     (CR 702.189b), raised by the ability's own effect
//     (effects.firebending) at the end of its resolution, also when it
//     added no mana (X = 0).
//   - EventAirbend: a player exiled one or more objects as the result of an
//     airbend instruction (CR 701.65b), raised once per instruction, once
//     the cards have landed in exile (effects.AirbendAll). An airbend that
//     exiled nothing raises nothing.
//
// Actor is the bending player, Source the card whose cost, ability or
// effect it was. Each also counts on the player's PlayerTurnTally (a
// count per kind, reset with the turn), which "if you've done all four this
// turn" reads through AllFourBendsThisTurn.

// BendKinds is the four bending events, in the order Avatar Aang prints
// them.
var BendKinds = []EventKind{EventWaterbend, EventEarthbend, EventFirebend, EventAirbend}

// IsBendKind reports whether k is one of the four bending events.
func IsBendKind(k EventKind) bool {
	for _, b := range BendKinds {
		if b == k {
			return true
		}
	}
	return false
}

// noteBendLocked announces that `player` performed the bending action
// `kind` for `source`. A kind that is not a bending event, or no player,
// emits nothing. Caller must hold g.mu in write mode.
func (g *Game) noteBendLocked(kind EventKind, player, source uuid.UUID) {
	if player == uuid.Nil || !IsBendKind(kind) {
		return
	}
	g.EmitEvent(Event{Kind: kind, Actor: player, Source: source})
}

// NoteBendForEffect is noteBendLocked for a card-side action: the
// firebending ability as it resolves, an airbend once its cards have
// landed. Caller must hold g.mu in write mode.
func (g *Game) NoteBendForEffect(kind EventKind, player, source uuid.UUID) {
	g.noteBendLocked(kind, player, source)
}

// NoteAirbendForEffect announces an airbend instruction of `player`'s
// that exiled `cards` (CR 701.65b): one EventAirbend carrying the cards
// that landed. No cards, or no player, emits nothing.
//
// Caller must hold g.mu in write mode.
func (g *Game) NoteAirbendForEffect(player, source uuid.UUID, cards []uuid.UUID) {
	if player == uuid.Nil || len(cards) == 0 {
		return
	}
	g.EmitEvent(Event{Kind: EventAirbend, Actor: player, Source: source, CardIDs: append([]uuid.UUID(nil), cards...)})
}

// noteWaterbendsOfCastLocked raises one EventWaterbend for each waterbend
// cost the cast of `card` paid as part of its total cost: its own
// waterbend additional cost, a chosen optional one, a permission's
// waterbend price (CR 701.67b, 701.67c). Called once the cast's taps are
// paid. Caller must hold g.mu in write mode.
func (g *Game) noteWaterbendsOfCastLocked(player, cardID uuid.UUID, card Card, params CastSpellParams) {
	for _, l := range g.tapCostForCastParamsLocked(player, card, params).links() {
		if l.Key == "waterbend" {
			g.noteBendLocked(EventWaterbend, player, cardID)
		}
	}
}

// recordBendLocked counts one bending action on its player's turn tally.
// Called from the turn tally's event listener. Caller must hold g.mu.
func (g *Game) recordBendLocked(ev Event) {
	g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) {
		switch ev.Kind {
		case EventWaterbend:
			p.Waterbends++
		case EventEarthbend:
			p.Earthbends++
		case EventFirebend:
			p.Firebends++
		case EventAirbend:
			p.Airbends++
		}
	})
}

// BendsThisTurn is how many times `player` has performed the bending
// action `kind` this turn. Zero for a kind that is not a bending event.
//
// Caller must hold g.mu.
func (g *Game) BendsThisTurn(player uuid.UUID, kind EventKind) int {
	t := g.TurnTallyFor(player)
	switch kind {
	case EventWaterbend:
		return t.Waterbends
	case EventEarthbend:
		return t.Earthbends
	case EventFirebend:
		return t.Firebends
	case EventAirbend:
		return t.Airbends
	}
	return 0
}

// AllFourBendsThisTurn reports whether `player` has waterbent, earthbent,
// firebent and airbent this turn (Avatar Aang's "if you've done all four
// this turn").
//
// Caller must hold g.mu.
func (g *Game) AllFourBendsThisTurn(player uuid.UUID) bool {
	for _, k := range BendKinds {
		if g.BendsThisTurn(player, k) == 0 {
			return false
		}
	}
	return true
}
