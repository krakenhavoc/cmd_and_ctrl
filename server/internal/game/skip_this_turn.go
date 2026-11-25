package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// skip_this_turn.go — "The player skips each instance of the chosen
// step or phase this turn" (Fatespinner, #6226; CR 500.11, 614.10).
//
// A ScopedEffect record (ScopeGame, IndefiniteDuration) with one
// ModSkipThisTurn mod: Player is who skips, Amount is the Turn.Seq of
// the turn it is for, and Text is what is skipped — "draw" (the draw
// step), "main" (every main phase, an added one included) or "combat"
// (every combat phase). It only ever binds the turn it was made in, and
// only while Player is the active player: it is read at step entry
// beside the S79 combat skip, so a step or phase of that kind that would
// begin is proceeded past as though it didn't exist (CR 500.11). A step
// or phase already begun is not skipped (CR 614.10). The record goes as
// the next turn begins. A combat skip also takes the turn's remaining
// combat phases out of the plan, as the S79 skip does, so every reader
// of "is there a combat still to come" sees none.

// ModSkipThisTurn is "<Player> skips each <draw step | main phase |
// combat phase> this turn" (#6226). Reads Player, Amount (the Turn.Seq)
// and Text.
const ModSkipThisTurn ModKind = "skipThisTurn"

// The parts of a turn a ModSkipThisTurn names (Mod.Text).
const (
	SkipThisTurnDraw   = "draw"
	SkipThisTurnMain   = "main"
	SkipThisTurnCombat = "combat"
)

var skipThisTurnParts = []string{SkipThisTurnDraw, SkipThisTurnMain, SkipThisTurnCombat}

// skipThisTurnModProblem is the registration and restore check. "" when
// the mod is sound or is another kind.
func skipThisTurnModProblem(m Mod) string {
	if m.Kind != ModSkipThisTurn {
		return ""
	}
	if m.Player == uuid.Nil {
		return "a skipThisTurn mod names no player"
	}
	if m.Amount < 1 {
		return fmt.Sprintf("a skipThisTurn mod names turn %d", m.Amount)
	}
	if !slices.Contains(skipThisTurnParts, m.Text) {
		return fmt.Sprintf("a skipThisTurn mod names %q, not draw, main or combat", m.Text)
	}
	return ""
}

// SkipThisTurnForEffect records that `player` skips each instance of
// `part` (SkipThisTurnDraw, SkipThisTurnMain or SkipThisTurnCombat)
// this turn. Returns false, writing nothing, for a player who is not
// seated or has left. Panics on another part.
//
// Safe inside a resolution: it writes a record (and, for combat, trims
// the plan).
//
// Caller must hold g.mu (write).
func (g *Game) SkipThisTurnForEffect(player, source uuid.UUID, part, label string) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	m := Mod{Kind: ModSkipThisTurn, Player: player, Amount: g.Turn.Seq, Text: part}
	if problem := skipThisTurnModProblem(m); problem != "" {
		panic(fmt.Sprintf("game: %s: %s", label, problem))
	}
	if !g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil, []Mod{m}, IndefiniteDuration(), label) {
		return false
	}
	if part == SkipThisTurnCombat && player == g.activePlayerIDLocked() && PhaseOf(g.Turn.Step) != PhaseCombat {
		g.dropCombatPhasesFromPlanLocked()
	}
	return true
}

func skipThisTurnRecord(e ScopedEffect) (Mod, bool) {
	if len(e.Mods) != 1 || e.Mods[0].Kind != ModSkipThisTurn {
		return Mod{}, false
	}
	return e.Mods[0], true
}

// skipThisTurnPart is the part of a turn `step` belongs to, "" for one
// no record can name.
func skipThisTurnPart(step Step) string {
	switch PhaseOf(step) {
	case PhaseCombat:
		return SkipThisTurnCombat
	case PhasePrecombatMain, PhasePostcombatMain:
		return SkipThisTurnMain
	}
	if step == StepDraw {
		return SkipThisTurnDraw
	}
	return ""
}

// stepSkippedThisTurnLocked reports whether a record skips the step
// about to begin.
//
// Caller holds g.mu.
func (g *Game) stepSkippedThisTurnLocked() bool {
	part := skipThisTurnPart(g.Turn.Step)
	if part == "" {
		return false
	}
	active := g.activePlayerIDLocked()
	return slices.ContainsFunc(g.ScopedEffects, func(e ScopedEffect) bool {
		m, ok := skipThisTurnRecord(e)
		return ok && m.Player == active && m.Amount == g.Turn.Seq && m.Text == part
	})
}

// sweepSkipThisTurnLocked drops the records of earlier turns, as a turn
// begins.
//
// Caller must hold g.mu (write).
func (g *Game) sweepSkipThisTurnLocked() {
	g.dropScopedEffectsLocked(func(_ int, e ScopedEffect) bool {
		m, ok := skipThisTurnRecord(e)
		return ok && m.Amount != g.Turn.Seq
	})
}
