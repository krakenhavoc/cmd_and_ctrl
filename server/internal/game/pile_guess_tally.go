package game

import (
	"slices"

	"github.com/google/uuid"
)

// pile_guess_tally.go — the turn's tally of spells and abilities that
// caused one or more players to guess or to group cards or permanents
// into a pile (#6175, CR 700.3; The Celestial Toymaker: "each opponent
// loses 2 life for each spell or ability that caused one or more
// players to guess or to group cards or permanents into a pile this
// turn").
//
// The unit is the spell or ability, so a stack item is recorded once
// however many piles or guesses it made. Every pile and guess effect in
// the catalog notes itself here (effects.NotePileOrGuess, called by the
// shared pile helpers and by each card that builds its piles or its
// guess by hand); the effects package's audit test holds every catalog
// card whose text says "pile" or "guess" to that.

// NotePileOrGuessForEffect records that the spell or ability `item`
// made a player guess or group objects into a pile this turn. A Nil id
// and a second note of the same item change nothing.
//
// Caller must hold g.mu (write).
func (g *Game) NotePileOrGuessForEffect(item uuid.UUID) {
	if item == uuid.Nil || slices.Contains(g.PilesOrGuessesThisTurn, item) {
		return
	}
	g.PilesOrGuessesThisTurn = append(g.PilesOrGuessesThisTurn, item)
}

// PilesOrGuessesThisTurnForEffect is how many spells and abilities made
// a player guess or group objects into a pile this turn.
//
// Caller must hold g.mu (read or write).
func (g *Game) PilesOrGuessesThisTurnForEffect() int {
	return len(g.PilesOrGuessesThisTurn)
}
