package game

import "github.com/google/uuid"

// game_end_gates.go — "you can't lose the game" and "your opponents
// can't win the game" (CR 104.3, CR 614.17; ADR 0057 Decision 4).
//
// A gate is a "can't", not a replacement (CR 614.17): it has to exist
// when the loss or win would happen, it can't reach back in time, and
// it is READ where the loss or win is checked — in the SBA loss pass,
// in LoseTheGameForEffect and in WinTheGameForEffect — against the
// board as it is then. Nothing is stored on the player for it. A
// player at −5 life behind a Platinum Angel loses at the first check
// after the Angel leaves, with no window to respond (the Abyssal
// Persecutor ruling), because life is re-read at every check.
//
// TWO SOURCES, the split every player-level static in this engine
// uses (player_statics.go, life_lock.go):
//
//   - DERIVED — a permanent on the battlefield whose printed static is
//     the gate (Platinum Angel, Herald of Eternal Dawn, Abyssal
//     Persecutor). Declared as effects.Spec.GameEndGates, read through
//     CatalogGameEndGates keyed by CatalogAbilityKey, so an Angel that
//     has lost its abilities gates nothing and an Angel that changed
//     controller gates for its new controller, with no bookkeeping.
//   - GRANTED — a resolved spell, "you can't lose the game this turn
//     and your opponents can't win the game this turn" (Angel's
//     Grace). Stored as a PlayerStatic with a GameEnd payload and a
//     CR 611.2 Duration on the caster's seat, because the spell is in
//     a graveyard a moment after it resolves. ADR 0057's 2026-09-24
//     amendment moves this half from the TurnScopedGameEndGates
//     registry the ADR drew onto the PlayerStatic slice, which #755
//     (ADR 0063 durations) and #1195 / #1200 / #1316 made the one home
//     for a statement about a player with a duration.
//
// Emblems (#623) would be a third source; the gate type keeps While
// for Gideon of the Trials' emblem, and the reader gains a loop when
// that card is catalogued.

// GateScope names who a gate applies to, relative to the player it
// belongs to: the controller of the permanent that prints it, or the
// caster of the spell that granted it.
type GateScope uint8

const (
	// GateYou — the gate's own player ("you can't lose the game").
	GateYou GateScope = iota + 1
	// GateOpponents — each opponent of that player ("your opponents
	// can't win the game"). With no teams in this engine every other
	// player is an opponent.
	GateOpponents
	// GateEachPlayer — every player (Everybody Lives!). No catalogued
	// card uses it yet; the scope is read so the first one is a Spec
	// field, not an engine change.
	GateEachPlayer
)

// GameEndGate is one "can't lose" / "can't win" statement.
//
// Platinum Angel is two of them:
//
//	{Scope: GateYou, CantLose: true}
//	{Scope: GateOpponents, CantWin: true}
//
// and Abyssal Persecutor the mirror image. A gate is plain data apart
// from While, which only a battlefield static may set.
type GameEndGate struct {
	Scope    GateScope `json:"scope"`
	CantLose bool      `json:"cantLose,omitempty"`
	CantWin  bool      `json:"cantWin,omitempty"`
	// Causes narrows CantLose. Nil means every cause except
	// LossConcede. Phyrexian Unlife would be []LossCause{LossLife}; no
	// catalogued card narrows it yet.
	Causes []LossCause `json:"causes,omitempty"`
	// While is a battlefield static's condition, read at every check.
	// Nil means always. Never stored (a granted gate refuses one), so
	// it never reaches a snapshot.
	While func(g *Game, source Card) bool `json:"-"`
}

// IsZero reports whether the gate says nothing.
func (gate GameEndGate) IsZero() bool {
	return gate.Scope == 0 && !gate.CantLose && !gate.CantWin && len(gate.Causes) == 0 && gate.While == nil
}

// GameEndGrant is a GameEndGate as a PlayerStatic stores it: the same
// statement without While, because a granted gate has no permanent to
// read a condition off and a snapshot can't hold a closure. Plain
// data, carried by the snapshot on PlayerStatic.GameEnd.
type GameEndGrant struct {
	Scope    GateScope   `json:"scope"`
	CantLose bool        `json:"cantLose,omitempty"`
	CantWin  bool        `json:"cantWin,omitempty"`
	Causes   []LossCause `json:"causes,omitempty"`
}

// IsZero reports whether the grant says nothing. The reader skips an
// entry whose GameEnd payload is zero — every PlayerStatic of another
// kind.
func (gr GameEndGrant) IsZero() bool {
	return gr.Scope == 0 && !gr.CantLose && !gr.CantWin && len(gr.Causes) == 0
}

// gate is the grant as the reader walks it.
func (gr GameEndGrant) gate() GameEndGate {
	return GameEndGate{Scope: gr.Scope, CantLose: gr.CantLose, CantWin: gr.CantWin, Causes: gr.Causes}
}

// appliesTo reports whether a gate belonging to `you` covers player p.
func (gate GameEndGate) appliesTo(you, p uuid.UUID) bool {
	switch gate.Scope {
	case GateYou:
		return p == you
	case GateOpponents:
		return p != you
	case GateEachPlayer:
		return true
	}
	return false
}

// stopsLoss reports whether the gate stops a loss for `cause`.
func (gate GameEndGate) stopsLoss(cause LossCause) bool {
	if !gate.CantLose || cause == LossConcede {
		return false
	}
	if len(gate.Causes) == 0 {
		return true
	}
	for _, c := range gate.Causes {
		if c == cause {
			return true
		}
	}
	return false
}

// CatalogGameEndGates returns the gates a battlefield permanent with
// this catalog key prints, or nil. Populated at init time by the
// cards/effects package from effects.Spec.GameEndGates. A nil hook (no
// catalog wired) means no permanent gates anything.
//
// Keyed by CatalogAbilityKey, not CatalogKey: a gate is a static
// ability of the permanent that prints it, so a Platinum Angel that
// has lost all its abilities (layer 6) stops gating.
var CatalogGameEndGates func(key string) []GameEndGate

// GrantGameEndGateForEffect registers a gate on `you` for a duration —
// Angel's Grace's "you can't lose the game this turn and your
// opponents can't win the game this turn" is two calls with the
// zero-value Duration (until end of turn). The gate outlives its
// source; the sweep and the reader both test the duration.
//
// A gate with a While refuses to register: a granted gate has no
// permanent to read a condition off, and a closure can't be written
// to a snapshot. A zero gate is a no-op.
//
// Caller must hold g.mu (write).
func (g *Game) GrantGameEndGateForEffect(you uuid.UUID, gate GameEndGate, label string, source uuid.UUID, d Duration) {
	if gate.IsZero() || gate.While != nil {
		return
	}
	p := g.playerByIDLocked(you)
	if p == nil || p.Eliminated {
		return
	}
	p.Statics = append(p.Statics, PlayerStatic{
		GameEnd: GameEndGrant{
			Scope:    gate.Scope,
			CantLose: gate.CantLose,
			CantWin:  gate.CantWin,
			Causes:   append([]LossCause(nil), gate.Causes...),
		},
		Source:   source,
		Label:    label,
		Duration: d,
	})
}

// GameEndGateSource is one gate that applies to a player, with where
// it came from — the view's `end_gates` and the model prompt read it.
type GameEndGateSource struct {
	// Source is the permanent or spell the gate came from.
	Source uuid.UUID
	// SourceName is its name: the permanent's, or a granted gate's
	// label.
	SourceName string
	// CantLose lists the causes this gate stops for the player.
	CantLose []LossCause
	// CantWin reports whether it stops the player winning.
	CantWin bool
	// ThisTurn marks a granted gate (Angel's Grace).
	ThisTurn bool
}

// forEachGameEndGateLocked walks every gate that applies to player p,
// battlefield statics first and granted gates after, until fn returns
// false. `source` is the gate's object; `name` its attribution;
// `granted` is true for a stored gate.
//
// Departed players contribute nothing: their permanents left with them
// (CR 800.4a), and a granted gate on a seat that left is ended by the
// same rule.
//
// Caller must hold g.mu (read or write).
func (g *Game) forEachGameEndGateLocked(p *Player, fn func(gate GameEndGate, source uuid.UUID, name string, granted bool) bool) {
	if p == nil {
		return
	}
	if CatalogGameEndGates != nil && g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			key := CatalogAbilityKey(*c)
			if key == "" {
				continue
			}
			gates := CatalogGameEndGates(key)
			if len(gates) == 0 {
				continue
			}
			if owner := g.playerByIDLocked(c.Controller); owner == nil || owner.Eliminated {
				continue
			}
			for _, gate := range gates {
				if !gate.appliesTo(c.Controller, p.ID) {
					continue
				}
				if gate.While != nil && !gate.While(g, *c) {
					continue
				}
				if !fn(gate, c.InstanceID, c.Name, false) {
					return
				}
			}
		}
	}
	for _, q := range g.Seats {
		if q == nil || q.Eliminated {
			continue
		}
		for _, s := range q.Statics {
			if s.GameEnd.IsZero() {
				continue
			}
			if g.durationExpiredLocked(s.Duration, false) {
				continue
			}
			gate := s.GameEnd.gate()
			if !gate.appliesTo(q.ID, p.ID) {
				continue
			}
			if !fn(gate, s.Source, s.Label, true) {
				return
			}
		}
	}
}

// canLoseLocked reports whether player p can lose the game for
// `cause` right now. Concession is never gated. Caller must hold g.mu.
func (g *Game) canLoseLocked(p *Player, cause LossCause) bool {
	if cause == LossConcede {
		return true
	}
	can := true
	g.forEachGameEndGateLocked(p, func(gate GameEndGate, _ uuid.UUID, _ string, _ bool) bool {
		if gate.stopsLoss(cause) {
			can = false
			return false
		}
		return true
	})
	return can
}

// winGateLocked reports whether something stops player p winning the
// game, and the first gate's source for the log. Caller must hold
// g.mu.
func (g *Game) winGateLocked(p *Player) (source uuid.UUID, blocked bool) {
	g.forEachGameEndGateLocked(p, func(gate GameEndGate, src uuid.UUID, _ string, _ bool) bool {
		if gate.CantWin {
			source, blocked = src, true
			return false
		}
		return true
	})
	return source, blocked
}

// canWinLocked reports whether player p can win the game by an effect
// right now. CR 104.2a (the last player standing) is not asked this:
// it overrides every "can't win". Caller must hold g.mu.
func (g *Game) canWinLocked(p *Player) bool {
	_, blocked := g.winGateLocked(p)
	return !blocked
}

// CantLoseCausesForEffect lists the causes that can't make this player
// lose right now, in GatableLossCauses order; nil when nothing gates
// them. The view's `cant_lose` and the bots read it.
//
// Caller must hold g.mu (read or write).
func (g *Game) CantLoseCausesForEffect(p *Player) []LossCause {
	if p == nil || p.Eliminated {
		return nil
	}
	var out []LossCause
	for _, cause := range GatableLossCauses {
		if !g.canLoseLocked(p, cause) {
			out = append(out, cause)
		}
	}
	return out
}

// CantWinForEffect reports whether something stops this player winning
// the game by an effect right now. Caller must hold g.mu.
func (g *Game) CantWinForEffect(p *Player) bool {
	if p == nil || p.Eliminated {
		return false
	}
	return !g.canWinLocked(p)
}

// GameEndGatesForEffect lists every gate that applies to this player,
// with its source, for the seat tooltip and the model prompt. Caller
// must hold g.mu.
func (g *Game) GameEndGatesForEffect(p *Player) []GameEndGateSource {
	if p == nil || p.Eliminated {
		return nil
	}
	var out []GameEndGateSource
	g.forEachGameEndGateLocked(p, func(gate GameEndGate, src uuid.UUID, name string, granted bool) bool {
		entry := GameEndGateSource{Source: src, SourceName: name, CantWin: gate.CantWin, ThisTurn: granted}
		if gate.CantLose {
			for _, cause := range GatableLossCauses {
				if gate.stopsLoss(cause) {
					entry.CantLose = append(entry.CantLose, cause)
				}
			}
		}
		if len(entry.CantLose) == 0 && !entry.CantWin {
			return true
		}
		out = append(out, entry)
		return true
	})
	return out
}
