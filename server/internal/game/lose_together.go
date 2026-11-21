package game

import "github.com/google/uuid"

// lose_together.go: "each player with <condition> loses the game" as ONE
// event (CR 104.3e, 800.4). LoseTheGameForEffect settles the game after
// every loser, so the first of two players who both lose would leave the
// other "the last one standing" and win by CR 104.2a before their own loss
// was applied. Here every loss in the batch is applied first and the game
// is settled once, so players who all lose at once leave a draw (CR 104.4a),
// the way a state-based batch does (checkGameOverLocked's note on CR
// 104.3f). Triskaidekaphobia, #5948.

// LoseTheGameTogetherForEffect has every player in `players` lose the game
// at once and reports how many did. A player who can't lose (Platinum Angel)
// or has left already is skipped, like LoseTheGameForEffect's. When the
// active player is among the losers, the game-over check and the rotation
// are deferred to the next state-based pass, as ADR 0057 Decision 3 does for
// a single loss.
//
// Caller must hold g.mu (write).
func (g *Game) LoseTheGameTogetherForEffect(players []uuid.UUID, source uuid.UUID) (int, error) {
	if g.State != StateActive {
		return 0, nil
	}
	active := g.activeSeatIDLocked()
	lost, activeLost := 0, false
	for _, id := range players {
		p := g.playerByIDLocked(id)
		if p == nil {
			return lost, ErrPlayerNotFound
		}
		if !g.loseGameLocked(p, LossEffect, source) {
			continue
		}
		lost++
		if p.ID == active {
			activeLost = true
		}
	}
	if lost == 0 {
		return 0, nil
	}
	if activeLost {
		g.ActiveSeatLeftPending = true
		return lost, nil
	}
	g.settleDeparturesLocked()
	return lost, nil
}
