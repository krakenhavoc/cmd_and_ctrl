package game

import (
	"fmt"

	"github.com/google/uuid"
)

// turn_skip.go — CR 614.10 "skip your next turn" (ADR 0188 Decision 1,
// #3559, tracker #753).
//
// A skip debt is a ScopedEffect record carrying one ModSkipTurns mod:
// Player is who skips and Amount how many of their next turns are still
// owed. Records, not a counter on the player, for ADR 0188 Decision 11's
// reason: a binary that does not know the mod kind refuses the file,
// where a dropped counter would hand the player a turn their card made
// them pay away, which is stronger than printed.
//
// The debt is spent at the rotation seam (rotation.go, extra_turns.go),
// as a candidate turn is chosen and before it begins. That one placement
// gives every rule the ADR names:
//
//   - "next turn" is the next turn the player WOULD begin, extra or
//     normal (the Magosi ruling: the extra turn from its fourth ability
//     is the one its third ability skips);
//   - a skip made during the player's own turn waits for the next one,
//     because the turn in progress never passes through the seam again
//     (CR 614.10);
//   - two skips skip two turns (CR 614.10a), each record carrying its own
//     charge, the oldest spent first.
//
// A skipped turn is not stamped on Game.Turn: no Seq, no TurnsBegun, no
// untap, no EventTurnBegan (CR 500.11, 614.10a). Turn.Round still goes
// up when the walk wraps past the starting seat, as it does for a seat
// that has left, because Round counts table rotations.

// consumeTurnSkipLocked spends one charge of the oldest (lowest Seq)
// ModSkipTurns record of `seat`, through setShieldChargeLocked: a copy
// on write, so an undo snapshot taken before keeps the charge, and a
// record at zero is removed. source is the record's source object. ok is
// false when the seat owes none.
//
// Caller must hold g.mu (write).
func (g *Game) consumeTurnSkipLocked(seat int) (source uuid.UUID, ok bool) {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return uuid.Nil, false
	}
	id := g.Seats[seat].ID
	var (
		bestSeq int64
		bestMod int
		best    = -1
	)
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for j, m := range e.Mods {
			if m.Kind != ModSkipTurns || m.Player != id || m.Amount < 1 {
				continue
			}
			if best < 0 || e.Seq < bestSeq {
				best, bestSeq, bestMod = i, e.Seq, j
			}
			break
		}
	}
	if best < 0 {
		return uuid.Nil, false
	}
	rec := g.ScopedEffects[best]
	g.setShieldChargeLocked(rec.Seq, bestMod, rec.Mods[bestMod].Amount-1)
	return rec.Source.ID, true
}

// dropTurnSkipsOfSeatLocked removes every ModSkipTurns record that
// belongs to `seat`. A player who has left can never spend them, and a
// turn of a departed player does not begin (CR 800.4k), so the seam drops
// them as it passes the seat. Copy on write, like setShieldChargeLocked.
//
// Caller must hold g.mu (write).
func (g *Game) dropTurnSkipsOfSeatLocked(seat int) {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	id := g.Seats[seat].ID
	owned := func(e ScopedEffect) bool {
		if len(e.Mods) == 0 {
			return false
		}
		for _, m := range e.Mods {
			if m.Kind != ModSkipTurns || m.Player != id {
				return false
			}
		}
		return true
	}
	found := false
	for _, e := range g.ScopedEffects {
		if owned(e) {
			found = true
			break
		}
	}
	if !found {
		return
	}
	next := make([]ScopedEffect, 0, len(g.ScopedEffects))
	for _, e := range g.ScopedEffects {
		if !owned(e) {
			next = append(next, e)
		}
	}
	if len(next) == 0 {
		next = nil
	}
	g.ScopedEffects = next
}

// totalTurnSkipsLocked is every charge owed at the table, for the
// rotation seam's loop guard.
func (g *Game) totalTurnSkipsLocked() int {
	total := 0
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if m.Kind == ModSkipTurns && m.Amount > 0 {
				total += m.Amount
			}
		}
	}
	return total
}

// TurnSkipsForLocked is the total charge owed by `player`: how many of
// their next turns will be skipped. The view and the bots read it. A
// player who has left owes none: the seam drops their records as it
// passes the seat, and the view must not show a skip nobody can take.
//
// Caller must hold g.mu (read or write).
func (g *Game) TurnSkipsForLocked(player uuid.UUID) int {
	if p := g.playerByIDLocked(player); p == nil || p.Eliminated {
		return 0
	}
	total := 0
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if m.Kind == ModSkipTurns && m.Player == player && m.Amount > 0 {
				total += m.Amount
			}
		}
	}
	return total
}

// TurnSkipsFor is the locking form of TurnSkipsForLocked.
func (g *Game) TurnSkipsFor(player uuid.UUID) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.TurnSkipsForLocked(player)
}

// skipTurnsModProblem is the registration and restore check for
// ModSkipTurns. "" when the mod is sound or is another kind.
func skipTurnsModProblem(m Mod) string {
	if m.Kind != ModSkipTurns {
		return ""
	}
	if m.Player == uuid.Nil {
		return "a skipTurns mod names no player"
	}
	if m.Amount < 1 {
		return fmt.Sprintf("a skipTurns mod owes %d turns, want at least 1", m.Amount)
	}
	return ""
}

// SkipNextTurnsForEffect records that `player` skips their next n turns
// (CR 614.10, 614.10a): one ScopedEffect, ScopeGame, IndefiniteDuration,
// with a ModSkipTurns mod of Amount n. Emits EventTurnSkipAdded (Actor
// player, Source source, Amount n). `label` is the printed clause, for
// the record.
//
// Returns false, writing nothing, for n < 1 and for a player who is not
// seated or has left the game (CR 800.4a).
//
// Safe inside a resolution: it only writes a record. The turn is skipped
// at the rotation seam when it would begin, never here, so a skip made
// during the player's own turn does not end that turn.
//
// Caller must hold g.mu (write).
func (g *Game) SkipNextTurnsForEffect(player, source uuid.UUID, n int, label string) bool {
	if n < 1 {
		return false
	}
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	if !g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModSkipTurns, Player: player, Amount: n}}, IndefiniteDuration(), label) {
		return false
	}
	g.EmitEvent(Event{
		Kind:   EventTurnSkipAdded,
		Actor:  player,
		Source: source,
		Amount: n,
	})
	return true
}

// emitTurnSkippedLocked announces one spent charge. extraRef is the
// skipped extra turn's ExtraTurn.Ref, or 0 for a normal turn.
func (g *Game) emitTurnSkippedLocked(seat int, source uuid.UUID, extraRef int) {
	label := ""
	if extraRef != 0 {
		label = "extra"
	}
	g.EmitEvent(Event{
		Kind:   EventTurnSkipped,
		Actor:  g.Seats[seat].ID,
		Source: source,
		Amount: extraRef,
		Label:  label,
	})
}

// --- a step of one extra turn (ADR 0188 Decision 3a, #3563) -----------

// SkipStepOfExtraTurnForEffect is "skip the <step> of that turn" (CR
// 614.10; Savor the Moment's "Skip the untap step of that turn"): one
// ScopedEffect, ScopeGame, IndefiniteDuration, with a
// ModSkipExtraTurnStep mod naming the queued turn `ref` and the step.
//
// The record is spent at step entry (spendSkippedStepLocked), by the
// first matching step of that turn, so an untap step in a beginning
// phase added later in the same turn (Sphinx of the Second Sun) still
// happens (owner answer 2). It is swept, with the delayed triggers bound
// to the same ref, when the turn is skipped or ends without reaching the
// step (sweepUnreachableBoundTriggersLocked), because CR 614.10a says
// anything scheduled for a skipped turn won't happen.
//
// Returns false, writing nothing, when ref names no queued extra turn or
// step is not a step that can be skipped.
//
// Caller must hold g.mu (write).
func (g *Game) SkipStepOfExtraTurnForEffect(ref int, step Step, source uuid.UUID, label string) bool {
	if !g.extraTurnQueuedLocked(ref) {
		return false
	}
	m := Mod{Kind: ModSkipExtraTurnStep, Step: step, ExtraTurn: ref}
	if skipExtraTurnStepModProblem(m) != "" {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil,
		[]Mod{m}, IndefiniteDuration(), label)
}

// skipExtraTurnStepModProblem is the registration and restore check for
// ModSkipExtraTurnStep: it names a real, skippable step and a real turn,
// and the two fields it adds appear on no other kind. "" when the mod
// is sound or is another kind with neither field set.
func skipExtraTurnStepModProblem(m Mod) string {
	if m.Kind != ModSkipExtraTurnStep {
		if m.Step != "" || m.ExtraTurn != 0 {
			return fmt.Sprintf("a step or extra-turn ref on a %s mod", m.Kind)
		}
		return ""
	}
	if m.ExtraTurn < 1 {
		return "a skipExtraTurnStep mod names no extra turn"
	}
	known := false
	for _, s := range turnSequence {
		if s == m.Step {
			known = true
			break
		}
	}
	if !known {
		return fmt.Sprintf("a skipExtraTurnStep mod names step %q", m.Step)
	}
	if m.Step == StepCleanup {
		// The cleanup step carries the turn's end-of-turn sweep. No card
		// skips it, and a skip that did would leave "until end of turn"
		// effects running into the next turn.
		return "a skipExtraTurnStep mod names the cleanup step"
	}
	return ""
}

// spendSkippedStepLocked reports whether the step now being entered is
// skipped by a ModSkipExtraTurnStep record bound to the extra turn in
// progress, and if so removes the record (CR 614.10a: a skip is
// satisfied by the first occurrence). Copy on write, so an undo snapshot
// taken before keeps the record.
//
// Called from runStepEntryHooksLocked only, after the step is known to
// exist and before the CR 614 replacement window opens: a skipped step
// is never offered to Stasis.
//
// Caller must hold g.mu (write).
func (g *Game) spendSkippedStepLocked() bool {
	if !g.Turn.Extra || g.Turn.ExtraRef == 0 || len(g.ScopedEffects) == 0 {
		return false
	}
	at := -1
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if m.Kind == ModSkipExtraTurnStep && m.ExtraTurn == g.Turn.ExtraRef && m.Step == g.Turn.Step {
				at = i
				break
			}
		}
		if at >= 0 {
			break
		}
	}
	if at < 0 {
		return false
	}
	g.dropScopedEffectsLocked(func(i int, _ ScopedEffect) bool { return i == at })
	return true
}

// sweepSkippedStepsLocked removes every ModSkipExtraTurnStep record
// whose extra turn is neither the turn now beginning nor still queued:
// the turn was skipped (Trouble in Pairs, Ugin's Nexus, a skip debt) or
// ended before the step came (CR 614.10a). Called from
// sweepUnreachableBoundTriggersLocked. Copy on write.
//
// Caller must hold g.mu (write).
func (g *Game) sweepSkippedStepsLocked() {
	g.dropScopedEffectsLocked(func(_ int, e ScopedEffect) bool {
		for _, m := range e.Mods {
			if m.Kind == ModSkipExtraTurnStep && !g.extraTurnPendingLocked(m.ExtraTurn) {
				return true
			}
		}
		return false
	})
}

// dropScopedEffectsLocked rebuilds the registry without the records
// `drop` selects, into a fresh slice (the backing array is shared with
// every undo snapshot). Nothing is written when nothing is dropped.
//
// Caller must hold g.mu (write).
func (g *Game) dropScopedEffectsLocked(drop func(i int, e ScopedEffect) bool) {
	var next []ScopedEffect
	dropped := false
	for i, e := range g.ScopedEffects {
		if drop(i, e) {
			dropped = true
			continue
		}
		next = append(next, e)
	}
	if !dropped {
		return
	}
	g.ScopedEffects = next
	g.layerVersion.Add(1)
}
