package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// skip_combat_phases.go — "<player> skips all combat phases of their
// next turn" (S79, #5075: Empty City Ruse, False Peace; CR 500.11,
// 614.10, 614.10a).
//
// A skip is a ScopedEffect record (ScopeGame, IndefiniteDuration)
// carrying one ModSkipCombatPhases mod: Player is who skips, and Amount
// is the Turn.Seq of the turn whose combat phases it skips, or 0 while
// it waits for that player's next turn. Records, not a flag on the
// player, for turn_skip.go's reason: a binary that does not know the
// kind refuses the file instead of giving the player a combat their
// opponent's card took away.
//
// The life of one record, in three places:
//
//   - written by SkipCombatPhasesOfNextTurnForEffect, waiting (Amount 0);
//   - ARMED as the player's next turn really begins (startTurnLocked,
//     armCombatSkipsLocked): a turn skipped whole never begins, so the
//     record waits for the next one that does (CR 614.10a), and a record
//     made during the player's own turn waits for the next one, because
//     that turn has already begun (CR 614.10). One record arms per turn,
//     the oldest first: two such effects skip the combat phases of the
//     player's next two turns (CR 614.10a, "one effect will be satisfied
//     in skipping the first occurrence, while the other will remain").
//     The armed turn's combat phases are taken out of the plan then, so
//     every reader of "is there a combat still to come" sees none (the
//     before-attackers window, CR 506.8e);
//   - READ at step entry (runStepEntryHooksLocked): a step of a combat
//     phase of the armed turn never begins, so a combat phase added later
//     in that turn (Relentless Assault) is skipped too ("all combat
//     phases") and a plan rebuilt from the template after a restore
//     cannot bring one back. A skipped phase is proceeded past as though
//     it didn't exist (CR 500.11): no beginning-of-combat trigger, no
//     attack, nothing counted.
//
// The armed record is removed as the next turn begins. A record waiting
// on a player who has left the game is removed then too: no turn of
// theirs will begin (CR 800.4k).

// ModSkipCombatPhases is "<Player> skips all combat phases of their next
// turn" (S79, #5075; CR 500.11, 614.10). Scope ScopeGame,
// IndefiniteDuration. Amount is 0 while it waits for Player's next turn
// and that turn's Turn.Seq once armed. Reads Player and Amount.
const ModSkipCombatPhases ModKind = "skipCombatPhases"

// skipCombatPhasesModProblem is the registration and restore check for
// ModSkipCombatPhases. "" when the mod is sound or is another kind.
func skipCombatPhasesModProblem(m Mod) string {
	if m.Kind != ModSkipCombatPhases {
		return ""
	}
	if m.Player == uuid.Nil {
		return "a skipCombatPhases mod names no player"
	}
	if m.Amount < 0 {
		return fmt.Sprintf("a skipCombatPhases mod is armed for turn %d", m.Amount)
	}
	return ""
}

// SkipCombatPhasesOfNextTurnForEffect records that `player` skips all
// combat phases of their next turn (CR 614.10): one ScopedEffect,
// ScopeGame, IndefiniteDuration, with a waiting ModSkipCombatPhases mod.
// `label` is the printed clause, for the record.
//
// Returns false, writing nothing, for a player who is not seated or has
// left the game (CR 800.4a).
//
// Safe inside a resolution: it only writes a record. The phases are
// skipped in the player's next turn, never in the turn in progress.
//
// Caller must hold g.mu (write).
func (g *Game) SkipCombatPhasesOfNextTurnForEffect(player, source uuid.UUID, label string) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	m := Mod{Kind: ModSkipCombatPhases, Player: player}
	if problem := skipCombatPhasesModProblem(m); problem != "" {
		panic(fmt.Sprintf("game: %s: %s", label, problem))
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil, []Mod{m}, IndefiniteDuration(), label)
}

// combatSkipRecord reports whether e is a combat-phase skip record, and
// its one mod.
func combatSkipRecord(e ScopedEffect) (Mod, bool) {
	if len(e.Mods) != 1 || e.Mods[0].Kind != ModSkipCombatPhases {
		return Mod{}, false
	}
	return e.Mods[0], true
}

// armCombatSkipsLocked runs as a turn really begins, after the plan is
// reset: it drops every record armed for an earlier turn and every
// waiting record of a player who has left, arms the oldest waiting
// record of the active player for this turn, and takes this turn's
// combat phases out of the plan if one is armed. Copy on write, so an
// undo snapshot taken before keeps the records as they were.
//
// Caller must hold g.mu (write).
func (g *Game) armCombatSkipsLocked() {
	if !slices.ContainsFunc(g.ScopedEffects, func(e ScopedEffect) bool {
		_, ok := combatSkipRecord(e)
		return ok
	}) {
		return
	}
	active := g.activePlayerIDLocked()
	seq := g.Turn.Seq
	armed := false
	next := make([]ScopedEffect, 0, len(g.ScopedEffects))
	for _, e := range g.ScopedEffects {
		m, ok := combatSkipRecord(e)
		if !ok {
			next = append(next, e)
			continue
		}
		if m.Amount != 0 {
			// Armed for a turn that has ended: spent.
			continue
		}
		if p := g.playerByIDLocked(m.Player); p == nil || p.Eliminated {
			continue
		}
		if m.Player == active && !armed {
			// Records are in registration order, so this is the oldest.
			m.Amount = seq
			e.Mods = []Mod{m}
			armed = true
		}
		next = append(next, e)
	}
	if len(next) == 0 {
		next = nil
	}
	g.ScopedEffects = next
	if armed {
		g.dropCombatPhasesFromPlanLocked()
	}
}

// dropCombatPhasesFromPlanLocked removes every combat step still to come
// from the plan. A fresh slice, as every plan edit builds.
//
// Caller must hold g.mu (write).
func (g *Game) dropCombatPhasesFromPlanLocked() {
	out := make([]PlannedStep, 0, len(g.TurnPlan))
	for _, s := range g.TurnPlan {
		if PhaseOf(s.Step) != PhaseCombat {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		out = nil
	}
	g.TurnPlan = out
}

// combatPhasesSkippedThisTurnLocked reports whether a record armed for
// the turn in progress skips its combat phases.
//
// Caller holds g.mu.
func (g *Game) combatPhasesSkippedThisTurnLocked() bool {
	active := g.activePlayerIDLocked()
	for _, e := range g.ScopedEffects {
		if m, ok := combatSkipRecord(e); ok && m.Player == active && m.Amount == g.Turn.Seq {
			return true
		}
	}
	return false
}

// CombatPhaseSkipsForLocked is how many of `player`'s next turns will
// have their combat phases skipped, counting the turn in progress when
// it is theirs and a record is armed for it: the view and the bots read
// it. A player who has left has none.
//
// Caller must hold g.mu (read or write).
func (g *Game) CombatPhaseSkipsForLocked(player uuid.UUID) int {
	if p := g.playerByIDLocked(player); p == nil || p.Eliminated {
		return 0
	}
	n := 0
	for _, e := range g.ScopedEffects {
		m, ok := combatSkipRecord(e)
		if !ok || m.Player != player {
			continue
		}
		if m.Amount == 0 || m.Amount == g.Turn.Seq {
			n++
		}
	}
	return n
}
