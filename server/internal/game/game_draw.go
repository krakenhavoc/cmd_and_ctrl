package game

import "github.com/google/uuid"

// game_draw.go — "the game is a draw" from an effect (#6198; CR 104.4c:
// "An effect may state that the game is a draw"), and "the player with
// the highest life total wins the game. If two or more players are tied
// for highest life total, the game is a draw" (Celestial Convergence).

// DrawTheGameForEffect is CR 104.4c: an effect states that the game is
// a draw. The game ends at once for every player still in it, as a draw
// (OutcomeCauseDrawEffect, Source the effect's source). A draw is no
// player's win and no player's loss, so neither a "can't win" nor a
// "can't lose" gate (Platinum Angel) stops it. Reports whether the game
// ended; a game that is not active is left alone.
//
// Card files call effects.DrawTheGame, which turns the end into
// ErrStopResolution. Caller must hold g.mu (write).
func (g *Game) DrawTheGameForEffect(source uuid.UUID) bool {
	if g.State != StateActive {
		return false
	}
	g.endGameLocked(GameOutcome{Kind: OutcomeDraw, Cause: OutcomeCauseDrawEffect, Source: source})
	return true
}

// HighestLifeWinsOrDrawForEffect is Celestial Convergence's "the player
// with the highest life total wins the game. If two or more players are
// tied for highest life total, the game is a draw." Only players still
// in the game are compared (CR 800.4a). A single leader wins through
// WinTheGameForEffect, so a "can't win" gate holds and the game goes on
// (CR 614.17a); a tie draws the game. Reports whether the game ended.
//
// Caller must hold g.mu (write).
func (g *Game) HighestLifeWinsOrDrawForEffect(source uuid.UUID) (bool, error) {
	if g.State != StateActive {
		return false, nil
	}
	var leader *Player
	tied := false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		switch {
		case leader == nil || p.Life > leader.Life:
			leader, tied = p, false
		case p.Life == leader.Life:
			tied = true
		}
	}
	if leader == nil {
		return false, nil
	}
	if tied {
		return g.DrawTheGameForEffect(source), nil
	}
	return g.WinTheGameForEffect(leader.ID, source)
}
