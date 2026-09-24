package game

import (
	"errors"

	"github.com/google/uuid"
)

// game_end.go — how a game ends, and what can stop it (ADR 0057, #749).
//
// There are two ways out of an active game, and before this file there
// was one:
//
//   - A player LEAVES (CR 104.3, CR 104.5): a state-based loss, an
//     effect that says "you lose the game" (CR 104.3e), or a concession
//     (CR 104.3a). The game is over when one player or none is left.
//   - A player WINS by an effect (CR 104.2b): Felidar Sovereign,
//     Laboratory Maniac, Thassa's Oracle. The game ends at once for
//     everyone (CR 104.1), and nobody else leaves — a four-player game
//     won by effect ends with three players still seated.
//
// Everything that ends a game goes through endGameLocked, which is the
// only writer of StateEnded apart from Game.End() (an admin closing a
// table, which records no outcome), and which records the result as
// engine state (Game.Outcome) rather than leaving the client to infer
// it from who is still seated.
//
// "CAN'T LOSE" AND "CAN'T WIN" (CR 104.3, Platinum Angel) are gates read
// where the loss or the win is checked, never stored on the player —
// see game_end_gates.go. Concession is never gated (CR 104.3a).

// LossCause is why a player left the game. Every departure names one
// (ADR 0057 Decision 1): the log says it, the gates read it, and the
// bots clamp their clocks by it.
type LossCause string

const (
	// LossLife — 0 or less life (CR 704.5a / CR 104.3b).
	LossLife LossCause = "life"
	// LossEmptyDraw — attempted to draw from an empty library since
	// the last state-based action check (CR 704.5b / CR 104.3c).
	LossEmptyDraw LossCause = "empty_draw"
	// LossPoison — ten or more poison counters (CR 704.5c / CR 104.3d).
	LossPoison LossCause = "poison"
	// LossCommanderDamage — 21 or more combat damage from one commander
	// (CR 704.6c / CR 903.10a).
	LossCommanderDamage LossCause = "commander_damage"
	// LossEffect — an effect said "you lose the game" (CR 104.3e).
	LossEffect LossCause = "effect"
	// LossConcede — the player conceded (CR 104.3a). Never gated.
	LossConcede LossCause = "concede"
)

// GatableLossCauses are the five causes a "can't lose the game" gate
// can stop, in CR 704.5 order. Everything except LossConcede.
var GatableLossCauses = []LossCause{LossLife, LossEmptyDraw, LossPoison, LossCommanderDamage, LossEffect}

// Outcome kinds and causes. Strings, because they are the wire's
// vocabulary as well as the engine's (GameView.outcome).
const (
	OutcomeWin  = "win"
	OutcomeDraw = "draw"

	// OutcomeCauseLastStanding — every opponent has left (CR 104.2a).
	OutcomeCauseLastStanding = "last_standing"
	// OutcomeCauseEffect — an effect said this player wins (CR 104.2b).
	OutcomeCauseEffect = "effect"
	// OutcomeCauseAllLost — every remaining player lost at once
	// (CR 104.4a).
	OutcomeCauseAllLost = "all_lost"
)

// GameOutcome is the result of an ended game (ADR 0057 Decision 5).
// Plain data: carried by Clone, RestoreFrom and the snapshot.
type GameOutcome struct {
	// Kind is OutcomeWin or OutcomeDraw.
	Kind string `json:"kind"`
	// Winner is the winning player when Kind is OutcomeWin.
	Winner uuid.UUID `json:"winner,omitempty"`
	// Cause is OutcomeCauseLastStanding, OutcomeCauseEffect or
	// OutcomeCauseAllLost.
	Cause string `json:"cause"`
	// Source is the object whose effect won the game, when Cause is
	// OutcomeCauseEffect.
	Source uuid.UUID `json:"source,omitempty"`
}

// ErrStopResolution is returned by a catalog effect whose resolution
// cannot go on: the game has ended, or the player who controls the
// resolving spell or ability has left the game (CR 800.4a — the
// resolving object is one of theirs). It is a clean stop, not a
// failure, so it never becomes an EventEffectError — EmitEvent drops
// an effect-error event that carries it (ADR 0057 Decision 3, and the
// 2026-09-24 amendment for why the drop is central).
//
// Card files never return it by hand: effects.WinTheGame and
// effects.LoseTheGame do, and a card returns their error.
var ErrStopResolution = errors.New("game: resolution stopped (the game ended or its controller left the game)")

// IsStopResolution reports whether err is (or wraps) ErrStopResolution.
func IsStopResolution(err error) bool { return errors.Is(err, ErrStopResolution) }

// loseGameLocked is the one door a gated loss goes through: it reads
// the "can't lose" gate for `cause` and, if nothing stops the loss,
// takes the player out of the game. Reports whether the player left.
//
// It does NOT check whether the game is over and does not move the
// turn on: callers do that once per batch (Decision 2 and Decision 3).
//
// A concession must not come here — it is never gated (CR 104.3a) —
// and Concede calls leaveGameLocked directly.
//
// Caller must hold g.mu.
func (g *Game) loseGameLocked(p *Player, cause LossCause, source uuid.UUID) bool {
	if p == nil || p.Eliminated {
		return false
	}
	if cause != LossConcede && !g.canLoseLocked(p, cause) {
		return false
	}
	return g.leaveGameLocked(p, cause, source)
}

// checkGameOverLocked runs after every batch of departures — an SBA
// pass, an effect loss, a concession — and ends the game when one
// player or none is left. Reports whether the game ended; the turn
// rotation runs only when it didn't.
//
//  1. No player left: a draw (CR 104.4a).
//  2. One player left: that player wins (CR 104.2a). NO GATE IS READ
//     HERE: CR 104.2a "overrides all effects that would preclude that
//     player from winning the game", so Abyssal Persecutor's
//     controller wins when the opponent it kept alive concedes.
//
// CR 104.3f's order is the batch's: every loss in the batch is applied
// before this runs, so a player who would be the last one standing in
// the same pass that makes them lose is a loser — a draw when everyone
// left loses at once, never a win.
//
// A game that has ended keeps its cursor where it was: its caller does
// not move play on. Ending the last turn would sweep marked damage and
// pull attackers out of combat on the board the game ended with, and
// begin a turn nobody takes.
//
// Caller must hold g.mu.
func (g *Game) checkGameOverLocked() bool {
	if g.State != StateActive {
		return g.State == StateEnded
	}
	var last *Player
	left := 0
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			left++
			last = p
		}
	}
	switch left {
	case 0:
		g.endGameLocked(GameOutcome{Kind: OutcomeDraw, Cause: OutcomeCauseAllLost})
		return true
	case 1:
		g.endGameLocked(GameOutcome{Kind: OutcomeWin, Winner: last.ID, Cause: OutcomeCauseLastStanding})
		return true
	}
	return false
}

// endGameLocked ends the game with outcome `o`: the state flips to
// StateEnded, the outcome is recorded, and every open prompt is
// dropped so nothing is left asking a question over the game-over
// banner (ADR 0057 Decision 3). Emits EventGameOver.
//
// Idempotent on an ended game (the first outcome stands).
//
// Caller must hold g.mu.
func (g *Game) endGameLocked(o GameOutcome) {
	if g.State == StateEnded {
		return
	}
	// CR 702.143f, #658: all face-down foretold cards are revealed as
	// the game ends. Before the state flips, so the reveal listeners
	// still see an active game.
	g.revealForetoldAtGameEndLocked()
	g.State = StateEnded
	out := o
	g.Outcome = &out
	g.PendingChoices = nil
	g.PendingTriggers = nil
	g.DiscardPending = nil
	g.ActiveSeatLeftPending = false
	g.EmitEvent(Event{
		Kind:   EventGameOver,
		Actor:  o.Winner,
		Source: o.Source,
		CardID: o.Source,
		Label:  o.Cause,
	})
}

// WinTheGameForEffect is CR 104.2b: an effect says `player` wins the
// game. It happens at once, during the resolution (the Jace, Wielder
// of Mysteries ruling: the −8 wins "before state-based actions would
// cause you to lose the game for trying to draw from an empty
// library"). Reports whether the game ended with this player as the
// winner.
//
// A player who can't win (an opponent's Platinum Angel) doesn't, and
// nothing is remembered for later (CR 614.17a): EventWinPrevented is
// emitted — once per prevented win — and the game goes on.
//
// No-op, (false, nil), on a game that is not active and for a player
// who has already left: a departed player can't win (CR 104.5), and a
// second "win" in a game that already ended changes nothing.
//
// Card files call effects.WinTheGame, which turns a win into
// ErrStopResolution. Caller must hold g.mu (write).
func (g *Game) WinTheGameForEffect(playerID, source uuid.UUID) (bool, error) {
	if g.State != StateActive {
		return false, nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return false, ErrPlayerNotFound
	}
	if p.Eliminated {
		return false, nil
	}
	if gate, blocked := g.winGateLocked(p); blocked {
		g.EmitEvent(Event{
			Kind:   EventWinPrevented,
			Actor:  p.ID,
			Source: source,
			CardID: source,
			Target: gate,
		})
		return false, nil
	}
	g.endGameLocked(GameOutcome{Kind: OutcomeWin, Winner: p.ID, Cause: OutcomeCauseEffect, Source: source})
	return true, nil
}

// LoseTheGameForEffect is CR 104.3e: an effect says `player` loses the
// game — the Pact cycle's unpaid debt, Strixhaven Stadium, Final
// Fortune's end step. Reports whether the player lost.
//
// IMMEDIATE, not deferred to the next state-based action check: CR
// 104.3e has no SBA clause, unlike the four losses in CR 104.3b–d/j.
// The player is Eliminated, and their stack items, prompts and objects
// are gone (leaveGameLocked, CR 800.4a), before this returns.
//
// A player who can't lose (Platinum Angel) doesn't, and nothing is
// remembered for later (CR 614.17a) — the gate is read now, while the
// effect happens, not at a later check it might no longer exist for.
//
// What happens next depends on who left (ADR 0057 Decision 3):
//
//   - Not the active player: the game-over check and the rotation's
//     priority branch run now.
//   - The active player: both are DEFERRED to the next SBA loss pass
//     (Game.ActiveSeatLeftPending), because with the active seat gone
//     the rotation ends the turn and begins the next one, and none of
//     that may run inside a callback that is still resolving (ADR 0059
//     Decision 6).
//
// Card files call effects.LoseTheGame, which turns a loss by the
// resolving item's own controller, or a game that ended, into
// ErrStopResolution. Caller must hold g.mu (write).
func (g *Game) LoseTheGameForEffect(playerID, source uuid.UUID) (bool, error) {
	if g.State != StateActive {
		return false, nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return false, ErrPlayerNotFound
	}
	if !g.loseGameLocked(p, LossEffect, source) {
		return false, nil
	}
	if g.activeSeatIDLocked() == p.ID {
		g.ActiveSeatLeftPending = true
		return true, nil
	}
	g.settleDeparturesLocked()
	return true, nil
}

// Result returns the game's outcome under the read lock: nil while the
// game is active, and nil for a game ended by End() with no result (an
// abandoned table). The copy is the caller's.
func (g *Game) Result() *GameOutcome {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.Outcome == nil {
		return nil
	}
	out := *g.Outcome
	return &out
}
