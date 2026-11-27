package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// skip_next_combat_phase.go — "<player> skips their next combat phase"
// (#5838: Blinding Angel; CR 500.11, 614.10, 614.10a).
//
// skip_combat_phases.go's record takes every combat phase of one turn.
// This one takes ONE combat phase: the next one that player would begin,
// in whatever turn it falls, and it is used up by that one skip. A turn
// with two combat phases loses only the first.
//
// A skip is a ScopedEffect record (ScopeGame, IndefiniteDuration)
// carrying one ModSkipNextCombatPhase mod naming the player. Records,
// not a flag on the player, for turn_skip.go's reason: a binary that
// does not know the kind refuses the file instead of giving the player a
// combat their opponent's card took away.
//
// It is read at step entry (runStepEntryHooksLocked), as a combat step
// of the player's turn is about to begin and its combat phase has not
// begun yet (TurnTally.PhaseStarted is not that phase): the oldest such
// record of the active player is removed, the rest of that phase
// instance comes out of the plan, and the cursor walks on, so the phase
// is proceeded past as though it didn't exist (CR 500.11): no
// beginning-of-combat trigger, no attack, nothing counted. Nothing has to
// be remembered once the cursor has left the phase, so there is no armed
// state and a restore point holds only waiting records.
//
// A combat phase that has begun can't be skipped (CR 614.10): a record
// made during the player's own combat phase waits for their next one, an
// added combat phase later that turn or the first of a later turn. A
// combat phase another effect skips (skip_combat_phases.go, a turn
// skipped whole) does not use it up: it waits for the first one that is
// not skipped (CR 614.10a). Two records skip the player's next two
// combat phases, the oldest satisfied first (CR 614.10a).
//
// A record waiting on a player who has left the game is removed as the
// next turn begins: no turn of theirs will begin (CR 800.4k).

// ModSkipNextCombatPhase is "<Player> skips their next combat phase"
// (#5838; CR 500.11, 614.10). Scope ScopeGame, IndefiniteDuration.
// Reads Player only; Amount is always 0.
const ModSkipNextCombatPhase ModKind = "skipNextCombatPhase"

// skipNextCombatPhaseModProblem is the registration and restore check
// for ModSkipNextCombatPhase. "" when the mod is sound or is another
// kind.
func skipNextCombatPhaseModProblem(m Mod) string {
	if m.Kind != ModSkipNextCombatPhase {
		return ""
	}
	if m.Player == uuid.Nil {
		return "a skipNextCombatPhase mod names no player"
	}
	if m.Amount != 0 {
		return fmt.Sprintf("a skipNextCombatPhase mod carries an amount (%d)", m.Amount)
	}
	return ""
}

// SkipNextCombatPhaseForEffect records that `player` skips their next
// combat phase (CR 614.10): one ScopedEffect, ScopeGame,
// IndefiniteDuration, with a ModSkipNextCombatPhase mod. `label` is the
// printed clause, for the record.
//
// Returns false, writing nothing, for a player who is not seated or has
// left the game (CR 800.4a).
//
// Safe inside a resolution: it only writes a record. A combat phase in
// progress is never skipped (CR 614.10).
//
// Caller must hold g.mu (write).
func (g *Game) SkipNextCombatPhaseForEffect(player, source uuid.UUID, label string) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	m := Mod{Kind: ModSkipNextCombatPhase, Player: player}
	if problem := skipNextCombatPhaseModProblem(m); problem != "" {
		panic(fmt.Sprintf("game: %s: %s", label, problem))
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil, []Mod{m}, IndefiniteDuration(), label)
}

// SkipNextCombatPhaseThisTurnForEffect is "<player> skips their next
// combat phase this turn" (Moment of Silence): the same record as
// SkipNextCombatPhaseForEffect, but it ends with the turn, so a player
// whose combat has passed (or who is not the active player) loses
// nothing, and it never waits for a later turn (CR 614.10: a phase that
// has begun is not skipped).
//
// Returns false, writing nothing, for a player who is not seated or has
// left the game (CR 800.4a).
//
// Caller must hold g.mu (write).
func (g *Game) SkipNextCombatPhaseThisTurnForEffect(player, source uuid.UUID, label string) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	m := Mod{Kind: ModSkipNextCombatPhase, Player: player}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil, []Mod{m}, g.UntilEndOfTurnDuration(), label)
}

// nextCombatSkipRecord reports whether e is a next-combat-phase skip
// record, and its one mod.
func nextCombatSkipRecord(e ScopedEffect) (Mod, bool) {
	if len(e.Mods) != 1 || e.Mods[0].Kind != ModSkipNextCombatPhase {
		return Mod{}, false
	}
	return e.Mods[0], true
}

// spendNextCombatSkipLocked runs at step entry, after the whole-turn
// combat skip: when the cursor stands on a combat step whose phase has
// not begun and the active player has a waiting record, it removes the
// oldest one and takes the rest of that phase out of the plan, and
// reports true so the caller walks the cursor on (CR 500.11). Copy on
// write, so an undo snapshot taken before keeps the record.
//
// Caller must hold g.mu (write).
func (g *Game) spendNextCombatSkipLocked() bool {
	if PhaseOf(g.Turn.Step) != PhaseCombat {
		return false
	}
	if g.Turn.PhaseID == 0 || g.TurnTally.PhaseStarted == g.Turn.PhaseID {
		// This combat phase has begun: it can't be skipped (CR 614.10).
		return false
	}
	active := g.activePlayerIDLocked()
	idx := slices.IndexFunc(g.ScopedEffects, func(e ScopedEffect) bool {
		m, ok := nextCombatSkipRecord(e)
		return ok && m.Player == active
	})
	if idx < 0 {
		return false
	}
	// Records are in registration order, so this is the oldest.
	next := make([]ScopedEffect, 0, len(g.ScopedEffects)-1)
	next = append(next, g.ScopedEffects[:idx]...)
	next = append(next, g.ScopedEffects[idx+1:]...)
	if len(next) == 0 {
		next = nil
	}
	g.ScopedEffects = next
	g.dropPlannedPhaseLocked(g.Turn.PhaseID)
	return true
}

// dropPlannedPhaseLocked removes every step of phase instance `id` still
// to come from the plan. A fresh slice, as every plan edit builds.
//
// Caller must hold g.mu (write).
func (g *Game) dropPlannedPhaseLocked(id int) {
	g.ensureTurnPlanLocked()
	out := make([]PlannedStep, 0, len(g.TurnPlan))
	for _, s := range g.TurnPlan {
		if s.PhaseID != id {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		out = nil
	}
	g.TurnPlan = out
}

// sweepLeftPlayersNextCombatSkipsLocked removes the waiting records of
// players who have left the game, as a turn begins (CR 800.4k). Copy on
// write.
//
// Caller must hold g.mu (write).
func (g *Game) sweepLeftPlayersNextCombatSkipsLocked() {
	gone := func(e ScopedEffect) bool {
		m, ok := nextCombatSkipRecord(e)
		if !ok {
			return false
		}
		p := g.playerByIDLocked(m.Player)
		return p == nil || p.Eliminated
	}
	if !slices.ContainsFunc(g.ScopedEffects, gone) {
		return
	}
	next := make([]ScopedEffect, 0, len(g.ScopedEffects))
	for _, e := range g.ScopedEffects {
		if !gone(e) {
			next = append(next, e)
		}
	}
	if len(next) == 0 {
		next = nil
	}
	g.ScopedEffects = next
}

// NextCombatPhaseSkipsForLocked is how many of `player`'s upcoming
// combat phases will be skipped by "skips their next combat phase"
// records. A player who has left has none.
//
// Caller must hold g.mu (read or write).
func (g *Game) NextCombatPhaseSkipsForLocked(player uuid.UUID) int {
	if p := g.playerByIDLocked(player); p == nil || p.Eliminated {
		return 0
	}
	n := 0
	for _, e := range g.ScopedEffects {
		if m, ok := nextCombatSkipRecord(e); ok && m.Player == player {
			n++
		}
	}
	return n
}
