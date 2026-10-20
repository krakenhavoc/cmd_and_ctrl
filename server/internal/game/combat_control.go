package game

import "github.com/google/uuid"

// combat_control.go — controlling another player during one combat
// phase (CR 723.2, #3999, ADR 0167's 2026-10-19 amendment): Secret of
// Bloodbending's "You control target opponent during their next combat
// phase."
//
// It is a PlayerControl record (player_control.go) with Until "combat",
// the turn-long record's shape (turn_control.go) scoped to one phase:
//
//   - BeginCombatControlForEffect creates it, waiting
//     (ControlAwaitingCombat). A waiting record never applies.
//   - finishStepEntryLocked's hook (beginCombatControlsLocked) turns
//     every waiting record over the active player into ControlCombat as
//     a combat phase of theirs begins: the first step of a combat phase
//     instance that really begins (noteStepBegunLocked reports it). An
//     additional combat phase is a combat phase, so it counts; a combat
//     phase that never begins (skipped, or a turn with none) leaves the
//     record waiting for the next one.
//   - advanceCursorLocked's hook (endCombatControlsLocked) ends it as
//     the cursor leaves that combat phase instance, which is where
//     CR 511.3 takes creatures out of combat. A turn that ends during
//     the combat phase (CR 724.1) ends it at the rotation seam instead
//     (endTurnControlsLocked).
//
// While it is live the controller plays that combat phase exactly as a
// turn's controller plays a turn (ADR 0167 §4–§9): every rules verb
// except the two that reach past the phase (actions/controlled_player.go,
// combatRulesVerbs), the controlled seat's legal moves, the "Playing as"
// board swap. ControlUntil.PlaysForThem is the one read those callers
// share.

const (
	// ControlAwaitingCombat is a combat-phase record waiting for the
	// controlled player's next combat phase. It never applies.
	ControlAwaitingCombat ControlStage = "awaiting_combat"
	// ControlCombat is the controlled combat phase itself: live while
	// both players are in the game, the controlled player is the active
	// player, and the cursor is in a combat phase of turn TurnSeq.
	ControlCombat ControlStage = "combat"
	// ControlUntilCombat is Secret of Bloodbending's: one combat phase
	// of the controlled player's (CR 723.2).
	ControlUntilCombat ControlUntil = "combat"
)

// PlaysForThem reports whether a control of this length has the
// controller play the controlled player's game, rules verbs and all,
// rather than answer one prompt or one play: a whole turn (ADR 0167)
// or one combat phase (#3999). Routing, the cast gate, the enumerator,
// the view's cast stamps, the bots and the client's "Playing as" swap
// all read this one predicate.
func (u ControlUntil) PlaysForThem() bool {
	return u == ControlUntilTurn || u == ControlUntilCombat
}

// CombatControl is what an effect asks for: Controller controls
// Controlled during Controlled's next combat phase.
type CombatControl struct {
	Controller, Controlled uuid.UUID
	// Source is the card whose effect this is, and SourceName its name
	// for the log and the view.
	Source     uuid.UUID
	SourceName string
}

// BeginCombatControlForEffect schedules a combat-phase control. It
// refuses (false) a controller or controlled player who is not in the
// game, and a player controlling themselves. Like a turn-long control it
// deletes no record: an older one over the same player waits for the
// same phase, and the newest applies (CR 723.1a).
//
// Caller must hold g.mu.
func (g *Game) BeginCombatControlForEffect(cc CombatControl) bool {
	if cc.Controller == cc.Controlled ||
		!g.playerInGameLocked(cc.Controller) || !g.playerInGameLocked(cc.Controlled) {
		return false
	}
	rec := PlayerControl{
		Controller: cc.Controller,
		Controlled: cc.Controlled,
		Source:     cc.Source,
		SourceName: cc.SourceName,
		Stage:      ControlAwaitingCombat,
		Until:      ControlUntilCombat,
	}
	// A fresh slice: an undo entry may share the old backing array.
	next := make([]PlayerControl, len(g.PlayerControls), len(g.PlayerControls)+1)
	copy(next, g.PlayerControls)
	g.PlayerControls = append(next, rec)
	g.EmitEvent(Event{
		Kind:   EventPlayerControlScheduled,
		Actor:  cc.Controller,
		Target: cc.Controlled,
		Source: cc.Source,
		Label:  string(ControlUntilCombat),
	})
	return true
}

// combatControlLiveLocked is controlLiveLocked's ControlCombat case.
// Caller must hold g.mu.
func (g *Game) combatControlLiveLocked(r PlayerControl) bool {
	return r.UntilKind() == ControlUntilCombat &&
		g.Turn.Seq == r.TurnSeq &&
		PhaseOf(g.Turn.Step) == PhaseCombat &&
		g.activeSeatIDLocked() == r.Controlled
}

// beginCombatControlsLocked starts every waiting combat-phase record
// over the active player, now that a combat phase of theirs has begun.
// finishStepEntryLocked calls it for the first step of a phase instance
// that really begins, before the step is announced, so every trigger and
// choice of the phase is already the controller's. Caller must hold g.mu.
func (g *Game) beginCombatControlsLocked() {
	if len(g.PlayerControls) == 0 || PhaseOf(g.Turn.Step) != PhaseCombat {
		return
	}
	active := g.activeSeatIDLocked()
	if active == uuid.Nil {
		return
	}
	var next []PlayerControl
	for i, r := range g.PlayerControls {
		if r.UntilKind() != ControlUntilCombat || r.Stage != ControlAwaitingCombat || r.Controlled != active {
			continue
		}
		if next == nil {
			next = clonePlayerControls(g.PlayerControls)
		}
		next[i].Stage = ControlCombat
		next[i].TurnSeq = g.Turn.Seq
	}
	if next == nil {
		return
	}
	g.PlayerControls = next
	for _, r := range next {
		if r.Stage != ControlCombat || r.TurnSeq != g.Turn.Seq || r.Controlled != active {
			continue
		}
		// CR 800.4a: a controller who has left controls nobody. The
		// record is still spent on this phase.
		if !g.playerInGameLocked(r.Controller) {
			continue
		}
		g.EmitEvent(Event{
			Kind:   EventPlayerControlBegan,
			Actor:  r.Controller,
			Target: r.Controlled,
			Source: r.Source,
			Label:  string(ControlUntilCombat),
		})
	}
}

// endCombatControlsLocked ends every combat-phase record that has begun
// (ControlCombat): the phase it was for is over. advanceCursorLocked
// calls it as the cursor leaves a combat phase instance, and the
// rotation seam as a turn ends, so one that ended mid-combat (CR 724.1)
// ends it too. Caller must hold g.mu.
func (g *Game) endCombatControlsLocked() {
	if len(g.PlayerControls) == 0 {
		return
	}
	var ended []PlayerControl
	kept := make([]PlayerControl, 0, len(g.PlayerControls))
	for _, r := range g.PlayerControls {
		if r.UntilKind() == ControlUntilCombat && r.Stage == ControlCombat {
			ended = append(ended, r)
			continue
		}
		kept = append(kept, r)
	}
	if len(ended) == 0 {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.PlayerControls = kept
	for _, r := range ended {
		if !g.playerInGameLocked(r.Controller) || !g.playerInGameLocked(r.Controlled) {
			continue
		}
		g.EmitEvent(Event{
			Kind:   EventPlayerControlEnded,
			Actor:  r.Controller,
			Target: r.Controlled,
			Source: r.Source,
			Label:  string(ControlUntilCombat),
		})
	}
}
