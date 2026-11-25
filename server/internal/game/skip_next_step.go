package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// skip_next_step.go — "you skip your next untap step" (Avizoa) and "you
// skip your next draw step" (Ivory Gargoyle): a one-shot skip of one
// step of one player (#6144; CR 500.11, 614.10, 614.10a).
//
// The step-shaped sibling of skip_next_combat_phase.go. A skip is a
// ScopedEffect record (ScopeGame, IndefiniteDuration) carrying one
// ModSkipNextStep mod naming the player and the step. It is not a
// standing replacement in the CR 614 step-transition window (that is
// Stasis's and Necropotence's "skip your untap/draw steps"): it is read
// once the window has let the step through, as the step is about to
// begin (finishStepEntryLocked), so a step another effect already skips
// does not use it up and it waits for the first occurrence that isn't
// skipped (CR 614.10a). The oldest record of the active player for that
// step is removed and the cursor walks on, as for any skipped step
// (CR 500.11): no untap, no draw, no announcement, no priority.
//
// A step that has begun can't be skipped (CR 614.10): a record made
// during that player's draw step waits for their next draw step. Two
// records skip the next two occurrences, the oldest satisfied first
// (CR 614.10a). A record waiting on a player who has left the game goes
// as the next turn begins (CR 800.4k), as the combat sibling's does.

// ModSkipNextStep is "<Player> skips their next <step>" (#6144; CR
// 614.10). Scope ScopeGame, IndefiniteDuration. Reads Player and Step;
// Step is StepUntap or StepDraw, the two any card prints. Amount is
// always 0.
const ModSkipNextStep ModKind = "skipNextStep"

// skipNextStepModProblem is the registration and restore check for
// ModSkipNextStep. "" when the mod is sound or is another kind.
func skipNextStepModProblem(m Mod) string {
	if m.Kind != ModSkipNextStep {
		return ""
	}
	if m.Player == uuid.Nil {
		return "a skipNextStep mod names no player"
	}
	if m.Step != StepUntap && m.Step != StepDraw {
		return fmt.Sprintf("a skipNextStep mod names step %q, not untap or draw", m.Step)
	}
	if m.Amount != 0 {
		return fmt.Sprintf("a skipNextStep mod carries an amount (%d)", m.Amount)
	}
	return ""
}

// SkipNextStepForEffect records that `player` skips their next `step`
// (CR 614.10): "you skip your next untap step" with StepUntap, "you skip
// your next draw step" with StepDraw. One ScopedEffect, ScopeGame,
// IndefiniteDuration. `label` is the printed clause, for the record.
//
// Returns false, writing nothing, for a player who is not seated or has
// left the game (CR 800.4a). Panics on a step other than untap or draw.
//
// Safe inside a resolution: it only writes a record.
//
// Caller must hold g.mu (write).
func (g *Game) SkipNextStepForEffect(player uuid.UUID, step Step, source uuid.UUID, label string) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	m := Mod{Kind: ModSkipNextStep, Player: player, Step: step}
	if problem := skipNextStepModProblem(m); problem != "" {
		panic(fmt.Sprintf("game: %s: %s", label, problem))
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil, []Mod{m}, IndefiniteDuration(), label)
}

// nextStepSkipRecord reports whether e is a next-step skip record, and
// its one mod.
func nextStepSkipRecord(e ScopedEffect) (Mod, bool) {
	if len(e.Mods) != 1 || e.Mods[0].Kind != ModSkipNextStep {
		return Mod{}, false
	}
	return e.Mods[0], true
}

// spendNextStepSkipLocked runs as a step that exists and that the CR 614
// window did not cancel is about to begin: when the active player has a
// waiting record for this step, it removes the oldest one and reports
// true so the caller walks the cursor on (CR 500.11). Copy on write, so
// an undo snapshot taken before keeps the record.
//
// Caller must hold g.mu (write).
func (g *Game) spendNextStepSkipLocked() bool {
	if g.Turn.Step != StepUntap && g.Turn.Step != StepDraw {
		return false
	}
	if len(g.ScopedEffects) == 0 {
		return false
	}
	active := g.activePlayerIDLocked()
	idx := slices.IndexFunc(g.ScopedEffects, func(e ScopedEffect) bool {
		m, ok := nextStepSkipRecord(e)
		return ok && m.Player == active && m.Step == g.Turn.Step
	})
	if idx < 0 {
		return false
	}
	// Records are in registration order, so this is the oldest.
	g.dropScopedEffectsLocked(func(i int, _ ScopedEffect) bool { return i == idx })
	return true
}

// sweepLeftPlayersNextStepSkipsLocked removes the waiting records of
// players who have left the game, as a turn begins (CR 800.4k).
//
// Caller must hold g.mu (write).
func (g *Game) sweepLeftPlayersNextStepSkipsLocked() {
	g.dropScopedEffectsLocked(func(_ int, e ScopedEffect) bool {
		m, ok := nextStepSkipRecord(e)
		if !ok {
			return false
		}
		p := g.playerByIDLocked(m.Player)
		return p == nil || p.Eliminated
	})
}

// NextStepSkipsForLocked is how many of `player`'s upcoming `step`s will
// be skipped by "skip your next <step>" records. A player who has left
// has none.
//
// Caller must hold g.mu (read or write).
func (g *Game) NextStepSkipsForLocked(player uuid.UUID, step Step) int {
	if p := g.playerByIDLocked(player); p == nil || p.Eliminated {
		return 0
	}
	n := 0
	for _, e := range g.ScopedEffects {
		if m, ok := nextStepSkipRecord(e); ok && m.Player == player && m.Step == step {
			n++
		}
	}
	return n
}
