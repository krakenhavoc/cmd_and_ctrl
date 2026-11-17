package game

import "github.com/google/uuid"

// cant_get_poison.go — "you can't get additional poison counters this
// turn" (Melira, the Living Cure, ONE).
//
// Melira's replacement turns any number of poison counters into one, and
// then says the player can't get more this turn. The second half is a
// continuing effect that outlives Melira (CR 611.2a, 611.2b), so it is a
// stored ModCantGetPoisonCounters ScopedEffect naming the player, swept at
// cleanup (CR 514.2). It is a rule, not a replacement: CR 101.2 lets a
// "can't" beat anything that says a thing can happen, so every placement
// door asks it (counterPlacementRefusedLocked) before the CR 614 window
// opens, and no replacement sees a placement that can't happen.

// playerCantGetPoisonLocked reports whether `player` can't get poison
// counters right now. Caller must hold g.mu.
func (g *Game) playerCantGetPoisonLocked(player uuid.UUID) bool {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind == ModCantGetPoisonCounters && (m.Player == uuid.Nil || m.Player == player) {
				return true
			}
		}
	}
	return false
}

// PlayerCantGetPoisonThisTurnForEffect registers "<player> can't get
// additional poison counters this turn". Registers nothing for a player
// who is not seated; reports whether a record was written. Caller must
// hold g.mu (write).
func (g *Game) PlayerCantGetPoisonThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModCantGetPoisonCounters, Player: player}}, g.UntilEndOfTurnDuration(), label)
}
