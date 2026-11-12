package game

import "github.com/google/uuid"

// since_last_turn.go — what has happened to a player SINCE the last turn
// they took (#5652, Marchesa, Resolute Monarch).
//
// "At the beginning of your upkeep, if you haven't been dealt combat
// damage since your last turn, you draw a card and you lose 1 life."
// The per-turn tally (PlayerTurnTally.CombatDamageTaken) is replaced as
// each turn begins, and the span the words ask about crosses every
// opponent's turn, so the answer is kept on the player:
// Player.CombatDamagedSinceLastTurn.
//
// It is set as combat damage is dealt to the player (the same event
// that counts CombatDamageTaken: damage that was prevented was not
// dealt, and damage to a planeswalker or battle the player controls was
// not dealt to the player), and cleared as each of that player's turns
// ends, beside Player.LastTurnAttacks (ADR 0108 §6) and with the same
// meaning of "your last turn":
//   - an extra turn is a turn (CR 500.7) and clears it;
//   - a turn that is not taken clears nothing;
//   - a player who has taken no turn yet is asked of the whole game so
//     far: nothing has cleared it.
//
// Combat damage dealt to a player during their own turn is cleared with
// that turn: it was dealt during their last turn, not since it. An
// upkeep reads the flag before any combat of the turn it begins.

// noteCombatDamagedSinceLastTurnLocked records that `player` was dealt
// combat damage. Caller must hold g.mu.
func (g *Game) noteCombatDamagedSinceLastTurnLocked(player uuid.UUID) {
	if p := g.playerByIDLocked(player); p != nil {
		p.CombatDamagedSinceLastTurn = true
	}
}

// resetSinceLastTurnLocked starts "since your last turn" over for the
// player whose turn is ending. It runs as the turn ends, beside
// recordLastTurnAttacksLocked. Caller must hold g.mu.
func (g *Game) resetSinceLastTurnLocked() {
	seat := g.Turn.ActiveSeat
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	g.Seats[seat].CombatDamagedSinceLastTurn = false
}

// DealtCombatDamageSinceLastTurnForEffect reports whether `player` has
// been dealt combat damage since their last turn: Marchesa, Resolute
// Monarch's "if you haven't been dealt combat damage since your last
// turn" is its negation. False for a player who is not at the table.
//
// Caller must hold g.mu.
func (g *Game) DealtCombatDamageSinceLastTurnForEffect(player uuid.UUID) bool {
	p := g.playerByIDLocked(player)
	return p != nil && p.CombatDamagedSinceLastTurn
}
