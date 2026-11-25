package game

import "github.com/google/uuid"

// cant_get_counters.go — #6241: "Players can't get counters" (Solemnity)
// and "That player can't get counters for as long as this creature
// remains on the battlefield" (Suncleanser). CR 122.1: a counter is a
// marker placed on an object or a PLAYER; these stop every kind a player
// can get (poison, energy, experience, rad, tickets). CR 101.2: the
// "can't" beats every instruction to put one, so the placement does not
// happen and no CR 614 counter replacement or "whenever you get" trigger
// sees it. Counters a player already has stay, and removing them is
// untouched.
//
// Two sources, read at the one player counter door,
// counterPlacementRefusedLocked:
//
//   - the world rule a permanent carries on itself, ExtraRestriction
//     PlayersCantGetCounters (Solemnity), live off the battlefield;
//   - a ModPlayerCantGetCounters scoped record naming one player (or
//     every player), with the duration the card states (Suncleanser's
//     "for as long as this creature remains on the battlefield").
//
// Proliferate offers no such player (ProliferateCandidatesForEffect).

// playerCantGetCountersLocked reports whether `player` can't get
// counters right now. Caller must hold g.mu with fresh layers.
func (g *Game) playerCantGetCountersLocked(player uuid.UUID) bool {
	if player == uuid.Nil {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if g.durationExpiredLocked(e.Duration, false) {
			continue
		}
		for _, m := range e.Mods {
			if m.Kind == ModPlayerCantGetCounters && (m.Player == uuid.Nil || m.Player == player) {
				return true
			}
		}
	}
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Effective().ExtraRestrictions.Has(PlayersCantGetCounters) {
			return true
		}
	}
	return false
}

// PlayerCantGetCountersForEffect is playerCantGetCountersLocked for the
// view and the catalog. Caller must hold g.mu.
func (g *Game) PlayerCantGetCountersForEffect(player uuid.UUID) bool {
	return g.playerCantGetCountersLocked(player)
}

// PlayerCantGetCountersForDurationForEffect registers "<player> can't get
// counters" for duration `d` (Suncleanser: g.PinnedTo-style "for as long
// as this creature remains on the battlefield"). Registers nothing for a
// player who is not seated; reports whether a record was written. Caller
// must hold g.mu (write).
func (g *Game) PlayerCantGetCountersForDurationForEffect(sourceID, player uuid.UUID, d Duration, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModPlayerCantGetCounters, Player: player}}, d, label)
}
