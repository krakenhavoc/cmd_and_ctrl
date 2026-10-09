package game

import "github.com/google/uuid"

// activation_timing_grant.go — #2797: the STORED half of CR 606.3's
// "any time you could cast an instant" clause. "Until end of turn, you
// may activate loyalty abilities of Jace planeswalkers you control on
// any player's turn any time you could cast an instant" (Jace's
// Machinations).
//
// activation_timing.go derives its statements from a permanent on the
// battlefield or an emblem, and says why it needs no duration: the
// source's presence IS the duration. A resolving instant has no
// presence to derive from — it is in a graveyard a moment later — so
// this one is stored, as cast_timing.go's granted statements are: a
// PlayerStatic entry with a CR 611.2 Duration, swept and read through
// durationExpiredLocked. It is the eleventh payload on that slice, told
// apart from the others by ActivationTimingGrant.Timing.
//
// PLAIN DATA, not a predicate, because it is snapshotted: the derived
// ActivationTiming carries a Covers closure that a restore point could
// not bring back (ADR 0041). What the printed cards narrow on is two
// facts, both data: the ability is a LOYALTY ability, and its permanent
// is a planeswalker of a named subtype that the holder controls.
//
// The three narrowings the derived Teferi helpers document hold here
// too, each a clause on the card:
//
//   - LOYALTY abilities: an equip ability on a planeswalker-turned-
//     creature stays at its printed window.
//   - Of planeswalkers with the named subtype (every planeswalker when
//     Subtype is empty). A Jace's Machinations opens a Jace and leaves
//     a Chandra at sorcery speed.
//   - YOU CONTROL: the holder controls the planeswalker and is the one
//     activating. CR 606.3's other half, one loyalty ability per
//     permanent per turn, is untouched: this opens the WINDOW, and the
//     tally in Game.LoyaltyActivatedThisTurn is read after it.

// ActivationTimingGrant is a granted "you may activate loyalty abilities
// of <subtype> planeswalkers you control any time you could cast an
// instant" statement. It lives on the HOLDER's PlayerStatic.
type ActivationTimingGrant struct {
	// Timing is what the statement says. TimingNormal — the zero
	// value — is the presence bit: the entry says nothing. Only
	// TimingFlash is declared by a card; the others are folded by
	// activationTimingVerdictLocked like a derived statement's.
	Timing GrantTiming `json:"timing"`

	// Subtype narrows the planeswalkers covered to those with this
	// planeswalker subtype ("Jace"). Empty is every planeswalker.
	Subtype string `json:"subtype,omitempty"`
}

// covers reports whether this grant is ABOUT the activation q. The
// holder is the activator: the entry is stored on q.Controller's
// PlayerStatic and read only for them.
func (t ActivationTimingGrant) covers(q ActivationQuery) bool {
	if t.Timing == TimingNormal || !q.Ability.Loyalty {
		return false
	}
	c := q.Card
	if !c.IsPlaneswalker() || c.Controller != q.Controller {
		return false
	}
	return t.Subtype == "" || c.HasSubtype(t.Subtype)
}

// GrantLoyaltyActivationTimingForEffect gives `player` the right to
// activate the loyalty abilities of their planeswalkers (those with
// `subtype`, or all of them when it is empty) at the timing `timing`
// for a duration — Jace's Machinations. The *ForEffect surface: caller
// must hold g.mu (write), which a resolution frame already does.
//
// An unstamped Duration is "until end of turn", the narrowest real
// window, as GrantCastTimingForEffect does for a card file that forgets
// to say. A no-op for a missing player or a TimingNormal statement.
func (g *Game) GrantLoyaltyActivationTimingForEffect(player uuid.UUID, timing GrantTiming, subtype, label string, source uuid.UUID, d Duration) {
	if player == uuid.Nil || timing == TimingNormal {
		return
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	p.Statics = append(p.Statics, PlayerStatic{
		ActivationTiming: ActivationTimingGrant{Timing: timing, Subtype: subtype},
		Source:           source,
		Label:            label,
		Duration:         d,
	})
}

// foldStoredActivationTimingsLocked folds the activator's live granted
// statements that speak about q into v. Caller must hold g.mu.
func (g *Game) foldStoredActivationTimingsLocked(q ActivationQuery, v *activationTimingVerdict) {
	p := g.playerByIDLocked(q.Controller)
	if p == nil {
		return
	}
	for _, st := range p.Statics {
		if st.ActivationTiming.Timing == TimingNormal {
			continue
		}
		// Tested here and not left to the sweep: the sweep is hygiene
		// run at known moments, and this read has to be right between
		// them (the argument castTimingVerdictLocked makes).
		if g.durationExpiredLocked(st.Duration, false) {
			continue
		}
		if !st.ActivationTiming.covers(q) {
			continue
		}
		switch st.ActivationTiming.Timing {
		case TimingFlash:
			v.Flash = true
		case TimingSorcery:
			v.SorceryOnly = true
		case TimingYourTurnOnly:
			v.YourTurnOnly = true
		}
	}
}
