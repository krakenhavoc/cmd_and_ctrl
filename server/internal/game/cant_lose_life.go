package game

import "github.com/google/uuid"

// cant_lose_life.go — "players can't lose life this turn" (Everybody
// Lives!, CR 119.8): the loss twin of cant_gain_life.go's stored grants.
//
// A ModCantLoseLife record is a rule gate, not a replacement: the loss does
// not happen, so there is nothing for an "if you would lose life"
// replacement to replace and the gate runs BEFORE the CR 614 window
// (lifeChangeForbiddenLocked). The DAMAGE that would have cost the life is
// still dealt (CR 120.3a is the part that does not happen: damage_tail.go
// asks playerCantLoseLifeLocked), and a cost that pays life can't be paid
// (CanPayLifeLocked, CR 119.8). Poison counters are not life.
//
// Scope ScopeGame with Player set is that one player, and with Player zero
// every player. A record is swept with its duration like every scoped rule.

// playerCantLoseLifeLocked reports whether p can't lose life right now.
//
// Caller must hold g.mu (read or write).
func (g *Game) playerCantLoseLifeLocked(p *Player) bool {
	if p == nil {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModCantLoseLife || e.Scope != ScopeGame {
				continue
			}
			if m.Player == uuid.Nil || m.Player == p.ID {
				return true
			}
		}
	}
	return false
}

// PlayerCantLoseLifeLocked is the exported surface over the reader.
//
// Caller must hold g.mu (read or write).
func (g *Game) PlayerCantLoseLifeLocked(p *Player) bool {
	return g.playerCantLoseLifeLocked(p)
}

// PlayersCantLoseLifeForEffect registers "players can't lose life" for
// duration `d` (the rest of the turn when d is the zero value), covering
// every player. Reports whether a record was written.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) PlayersCantLoseLifeForEffect(sourceID uuid.UUID, d Duration, label string) bool {
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModCantLoseLife}}, d, label)
}
