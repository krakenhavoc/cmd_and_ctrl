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
