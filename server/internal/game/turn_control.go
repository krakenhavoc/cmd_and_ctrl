package game

import "github.com/google/uuid"

// turn_control.go — controlling another player's whole turn (CR 723.1,
// ADR 0167, #2164): Emrakul, the Promised End's "you gain control of
// target opponent during that player's next turn. After that turn, that
// player takes an extra turn."
//
// It is a PlayerControl record (player_control.go) with Until "turn":
//
//   - BeginTurnControlForEffect creates it, waiting
//     (ControlAwaitingTurn). A waiting record never applies.
//   - startTurnLocked's hook (beginTurnControlsLocked) turns every
//     waiting record over the new active player into ControlTurn as
//     their turn begins, before any step entry hook. Only a turn that
//     really begins reaches it, so a skipped turn leaves the record
//     waiting for the next one (CR 723.1b), and a record created during
//     the controlled player's own turn waits for their next.
//   - beginNextTurnLocked's hook (endTurnControlsLocked) ends it as the
//     next turn is chosen, which is CR 723.1's "the effect doesn't end
//     until the beginning of the next turn", and queues the extra turn
//     it owes (queueAnchoredExtraTurnLocked, extra_turns.go).
//
// Several records over one player all wait for the same turn. The
// newest is the one that applies (CR 723.1a, ControllerOfPlayerLocked's
// order) and each owes its own extra turn (ruling).

// TurnControl is what an effect asks for: Controller controls
// Controlled during Controlled's next turn, and (unless NoExtraTurn)
// Controlled takes an extra turn after it.
type TurnControl struct {
	Controller, Controlled uuid.UUID
	// NoExtraTurn is the plain "you control target player during that
	// player's next turn" of Mindslaver, Worst Fears and Sorin Markov's
	// -7 (ADR 0167, 2026-10-16 amendment). Emrakul, the Promised End's
	// "after that turn, that player takes an extra turn" leaves it
	// false.
	NoExtraTurn bool
	// Source is the card whose effect this is, and SourceName its name
	// for the log and the view.
	Source     uuid.UUID
	SourceName string
}

// BeginTurnControlForEffect schedules a turn-long control (ADR 0167
// §1). It refuses (false) a controller or controlled player who is not
// in the game, and a player controlling themselves. It deletes no
// record: an older one over the same player waits for the same turn and
// owes its own extra turn.
//
// The extra turn's Ref is minted now, from NextExtraRef, so the turns
// created later are told apart from it when it is queued. A NoExtraTurn
// control mints none and owes none.
//
// Caller must hold g.mu.
func (g *Game) BeginTurnControlForEffect(tc TurnControl) bool {
	if tc.Controller == tc.Controlled ||
		!g.playerInGameLocked(tc.Controller) || !g.playerInGameLocked(tc.Controlled) {
		return false
	}
	rec := PlayerControl{
		Controller: tc.Controller,
		Controlled: tc.Controlled,
		Source:     tc.Source,
		SourceName: tc.SourceName,
		Stage:      ControlAwaitingTurn,
		Until:      ControlUntilTurn,
	}
	if !tc.NoExtraTurn {
		g.NextExtraRef++
		rec.ExtraTurn = &AnchoredExtraTurn{Ref: g.NextExtraRef}
	}
	// A fresh slice: an undo entry may share the old backing array.
	next := make([]PlayerControl, len(g.PlayerControls), len(g.PlayerControls)+1)
	copy(next, g.PlayerControls)
	g.PlayerControls = append(next, rec)
	g.EmitEvent(Event{
		Kind:   EventPlayerControlScheduled,
		Actor:  tc.Controller,
		Target: tc.Controlled,
		Source: tc.Source,
	})
	return true
}

// beginTurnControlsLocked starts every waiting turn-long record over
// the player whose turn is beginning (ADR 0167 §2). startTurnLocked
// calls it once g.Turn is the new turn and before the turn's resets and
// step entry hooks, so the untap step's choices are already the
// controller's. Caller must hold g.mu.
func (g *Game) beginTurnControlsLocked() {
	if len(g.PlayerControls) == 0 {
		return
	}
	active := g.activeSeatIDLocked()
	if active == uuid.Nil {
		return
	}
	var next []PlayerControl
	for i, r := range g.PlayerControls {
		if r.UntilKind() != ControlUntilTurn || r.Stage != ControlAwaitingTurn || r.Controlled != active {
			continue
		}
		if next == nil {
			next = clonePlayerControls(g.PlayerControls)
		}
		next[i].Stage = ControlTurn
		next[i].TurnSeq = g.Turn.Seq
		if next[i].ExtraTurn != nil {
			next[i].ExtraTurn.Floor = g.NextExtraRef
		}
	}
	if next == nil {
		return
	}
	g.PlayerControls = next
	for _, r := range next {
		if r.Stage != ControlTurn || r.TurnSeq != g.Turn.Seq || r.Controlled != active {
			continue
		}
		// CR 800.4a: a controller who has left controls nobody. The
		// record still begins, so its extra turn is still owed.
		if !g.playerInGameLocked(r.Controller) {
			continue
		}
		g.EmitEvent(Event{
			Kind:   EventPlayerControlBegan,
			Actor:  r.Controller,
			Target: r.Controlled,
			Source: r.Source,
		})
	}
}

// endTurnControlsLocked ends every turn-long record of the turn that is
// ending (ADR 0167 §2) and queues the extra turn each owes (§3). It runs
// first thing in beginNextTurnLocked, while g.Turn is still the
// controlled turn and before the next turn is chosen, so the cleanup
// step and a turn ended early (CR 724.1) are still controlled, and the
// extra turn is in the queue when popExtraTurnLocked looks.
//
// The extra turn is queued when the controller has left (CR 800.4a ends
// only the control) and not when the controlled player has (CR 800.4k:
// "the extra turn never begins"). A ControlTurn record of any other turn
// cannot be live again and is dropped. Caller must hold g.mu.
func (g *Game) endTurnControlsLocked() {
	if len(g.PlayerControls) == 0 {
		return
	}
	var ended []PlayerControl
	kept := make([]PlayerControl, 0, len(g.PlayerControls))
	for _, r := range g.PlayerControls {
		if r.UntilKind() == ControlUntilTurn && r.Stage == ControlTurn {
			if r.TurnSeq == g.Turn.Seq {
				ended = append(ended, r)
			}
			continue
		}
		kept = append(kept, r)
	}
	if len(kept) == len(g.PlayerControls) {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.PlayerControls = kept
	for _, r := range ended {
		inGame := g.playerInGameLocked(r.Controlled)
		if inGame && g.playerInGameLocked(r.Controller) {
			g.EmitEvent(Event{
				Kind:   EventPlayerControlEnded,
				Actor:  r.Controller,
				Target: r.Controlled,
				Source: r.Source,
			})
		}
		if !inGame || r.ExtraTurn == nil {
			continue
		}
		if p := g.playerByIDLocked(r.Controlled); p != nil {
			g.queueAnchoredExtraTurnLocked(p.Seat, r.Source, r.ExtraTurn.Ref, r.ExtraTurn.Floor)
		}
	}
}

// ControlUntilOfPlayerLocked is how long the control that applies to p
// right now lasts: the Until of the newest live record over p (the one
// ControllerOfPlayerLocked reads), or "" when nobody controls p. Caller
// must hold g.mu (read or write).
func (g *Game) ControlUntilOfPlayerLocked(p uuid.UUID) ControlUntil {
	for i := len(g.PlayerControls) - 1; i >= 0; i-- {
		r := g.PlayerControls[i]
		if r.Controlled == p && g.controlLiveLocked(r) {
			return r.UntilKind()
		}
	}
	return ""
}

// ControlUntilOfPlayer is ControlUntilOfPlayerLocked under the read
// lock. Caller must NOT hold g.mu.
func (g *Game) ControlUntilOfPlayer(p uuid.UUID) ControlUntil {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.ControlUntilOfPlayerLocked(p)
}

// ControlStageOfPlayerLocked is the stage of the newest live record over
// p (the one ControllerOfPlayerLocked reads), or "" when nobody controls
// p. The actions layer reads it to decide which verbs the controller may
// send as p (CR 723.5b, ADR 0166 amendment of 2026-10-19). Caller must
// hold g.mu (read or write).
func (g *Game) ControlStageOfPlayerLocked(p uuid.UUID) ControlStage {
	for i := len(g.PlayerControls) - 1; i >= 0; i-- {
		r := g.PlayerControls[i]
		if r.Controlled == p && g.controlLiveLocked(r) {
			return r.Stage
		}
	}
	return ""
}

// ControlStageOfPlayer is ControlStageOfPlayerLocked under the read
// lock. Caller must NOT hold g.mu.
func (g *Game) ControlStageOfPlayer(p uuid.UUID) ControlStage {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.ControlStageOfPlayerLocked(p)
}

// TurnControlledByLocked lists the players whose turn controller is
// playing right now (ADR 0167 §7): each seat whose live control is
// controller's and plays for them (ControlUntil.PlaysForThem: a whole
// turn, or one combat phase, #3999). Caller must hold g.mu.
func (g *Game) TurnControlledByLocked(controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.ID == controller || p.Eliminated {
			continue
		}
		if g.ControllerOfPlayerLocked(p.ID) == controller && g.ControlUntilOfPlayerLocked(p.ID).PlaysForThem() {
			out = append(out, p.ID)
		}
	}
	return out
}

// ControlsATurn reports whether controller is playing another player's
// turn, or one combat phase of it (#3999), right now. Caller must NOT
// hold g.mu.
func (g *Game) ControlsATurn(controller uuid.UUID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.TurnControlledByLocked(controller)) > 0
}

// ScheduledPlayerControlsLocked is every turn-long record still waiting
// for its turn, and every combat-phase record still waiting for its
// combat phase (#3999), whose players are both in the game, oldest
// first: the view's scheduled_player_controls (ADR 0167 §8). Caller must
// hold g.mu.
func (g *Game) ScheduledPlayerControlsLocked() []PlayerControl {
	var out []PlayerControl
	for _, r := range g.PlayerControls {
		if r.Stage != ControlAwaitingTurn && r.Stage != ControlAwaitingCombat {
			continue
		}
		if !g.playerInGameLocked(r.Controller) || !g.playerInGameLocked(r.Controlled) {
			continue
		}
		out = append(out, r)
	}
	return out
}

// LivePlayerControlsLocked is the record that applies to each
// controlled player right now (the newest live one, CR 723.1a), in
// record order: the view's player_controls. An older live record over
// the same player (a Word of Command's under it, or a second Emrakul's)
// is not listed. Caller must hold g.mu.
func (g *Game) LivePlayerControlsLocked() []PlayerControl {
	var out []PlayerControl
	for i, r := range g.PlayerControls {
		if !g.controlLiveLocked(r) {
			continue
		}
		newer := false
		for _, o := range g.PlayerControls[i+1:] {
			if o.Controlled == r.Controlled && g.controlLiveLocked(o) {
				newer = true
				break
			}
		}
		if !newer {
			out = append(out, r)
		}
	}
	return out
}
