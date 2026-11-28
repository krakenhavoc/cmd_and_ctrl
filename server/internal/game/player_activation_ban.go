package game

import "github.com/google/uuid"

// player_activation_ban.go — #5922: a granted "<player> can't activate
// abilities" for a duration (Sen Triplets: "This turn, that player can't
// cast spells or activate abilities").
//
// The activation twin of a granted cast ban (cast_ban.go): a
// PlayerStatic on the restricted player, CantActivateAbilities, with a
// CR 611.2 Duration, read by ActivationGateLocked before any printed
// restriction (CR 101.2: "can't" beats "can"). Every activation path
// asks that gate, the auto-tapper and the bot enumerator included, so
// a banned player taps nothing for mana either: the clause names no
// exception, and mana abilities are activated abilities (CR 605.1a).
// A player mana ability granted for a duration (Channel) is refused
// at its own activation and offered as not payable.

// GrantCantActivateAbilitiesForEffect makes `player` unable to activate
// abilities for duration d. The zero Duration is "until end of turn",
// stamped against this turn, as GrantCastBanForEffect stamps its own.
//
// The *ForEffect surface: caller must hold g.mu (write).
func (g *Game) GrantCantActivateAbilitiesForEffect(player uuid.UUID, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	p.Statics = append(p.Statics, PlayerStatic{
		CantActivateAbilities: true,
		Source:                source,
		Label:                 label,
		Duration:              d,
	})
}

// playerActivationBanLocked reports whether a live grant stops `player`
// from activating abilities, with the grant's label and source for the
// refusal. The duration is tested here as well as in the sweep, for the
// reason castBanForbidsLocked gives.
//
// Caller must hold g.mu (read or write).
func (g *Game) playerActivationBanLocked(player uuid.UUID) (label string, source uuid.UUID, banned bool) {
	p := g.playerByIDLocked(player)
	if p == nil {
		return "", uuid.Nil, false
	}
	for _, s := range p.Statics {
		if !s.CantActivateAbilities || s.NonManaOnly || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		return s.Label, s.Source, true
	}
	return "", uuid.Nil, false
}

// GrantCantActivateNonManaAbilitiesForEffect makes `player` unable to
// activate abilities that aren't mana abilities for duration d (Abeyance).
// Mana abilities (CR 605.1a) stay available. The zero Duration is "until end
// of turn".
//
// The *ForEffect surface: caller must hold g.mu (write).
func (g *Game) GrantCantActivateNonManaAbilitiesForEffect(player uuid.UUID, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	p.Statics = append(p.Statics, PlayerStatic{
		CantActivateAbilities: true,
		NonManaOnly:           true,
		Source:                source,
		Label:                 label,
		Duration:              d,
	})
}

// playerNonManaActivationBanLocked is playerActivationBanLocked's twin for a
// grant that spares mana abilities: the gate asks it only of an ability that
// isn't one.
//
// Caller must hold g.mu (read or write).
func (g *Game) playerNonManaActivationBanLocked(player uuid.UUID) (label string, source uuid.UUID, banned bool) {
	p := g.playerByIDLocked(player)
	if p == nil {
		return "", uuid.Nil, false
	}
	for _, s := range p.Statics {
		if !s.CantActivateAbilities || !s.NonManaOnly || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		return s.Label, s.Source, true
	}
	return "", uuid.Nil, false
}

// anyPlayerActivationBanLocked reports whether any player is under a
// live activation ban: AnyActivationRestrictionsForEffect's fast
// negative.
//
// Caller must hold g.mu (read or write).
func (g *Game) anyPlayerActivationBanLocked() bool {
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		if _, _, ok := g.playerActivationBanLocked(p.ID); ok {
			return true
		}
		if _, _, ok := g.playerNonManaActivationBanLocked(p.ID); ok {
			return true
		}
		if _, _, ok := g.playerLoyaltyBanLocked(p.ID); ok {
			return true
		}
	}
	return false
}

// playerActivationBanned is playerActivationBanLocked's bit alone.
//
// Caller must hold g.mu (read or write).
func (g *Game) playerActivationBanned(player uuid.UUID) bool {
	_, _, ok := g.playerActivationBanLocked(player)
	return ok
}

// GrantCantActivateLoyaltyAbilitiesForEffect makes `player` unable to
// activate planeswalkers' loyalty abilities for duration d (Revel in
// Silence: "Your opponents can't cast spells or activate planeswalkers'
// loyalty abilities this turn"). The zero Duration is "until end of
// turn". Other abilities, mana abilities included, stay available.
//
// The *ForEffect surface: caller must hold g.mu (write).
func (g *Game) GrantCantActivateLoyaltyAbilitiesForEffect(player uuid.UUID, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	p.Statics = append(p.Statics, PlayerStatic{
		CantActivateLoyaltyAbilities: true,
		Source:                       source,
		Label:                        label,
		Duration:                     d,
	})
}

// playerLoyaltyBanLocked reports whether a live grant stops `player`
// from activating planeswalkers' loyalty abilities.
//
// Caller must hold g.mu (read or write).
func (g *Game) playerLoyaltyBanLocked(player uuid.UUID) (label string, source uuid.UUID, banned bool) {
	p := g.playerByIDLocked(player)
	if p == nil {
		return "", uuid.Nil, false
	}
	for _, s := range p.Statics {
		if !s.CantActivateLoyaltyAbilities || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		return s.Label, s.Source, true
	}
	return "", uuid.Nil, false
}
