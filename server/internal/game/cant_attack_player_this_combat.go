package game

import "github.com/google/uuid"

// cant_attack_player_this_combat.go — the fourth window of a
// CantAttackGrant (cant_attack_player.go): "they can't attack you this
// combat" (Champions of Minas Tirith, S82 GAP batch 118).
//
// The restriction binds `attacker` from now (its floor is the attacker's
// CURRENT seat-turn count) until the combat phase in progress ends
// (CR 511.3, UntilEndOfCombatDuration). Same grant, same reader, same
// refusal: nothing new is stored, it is the existing PlayerStatic payload
// with another duration. Outside a combat phase there is no combat to last
// through, so nothing is registered (CR 611.2b).

// GrantCantAttackPlayerThisCombatForEffect makes `attacker` unable to
// attack what `scope` covers of `protected`'s for the rest of the combat
// phase in progress. Same no-ops and caller contract as
// GrantCantAttackPlayerForEffect (hold g.mu for writing; a missing player
// or attacker == protected does nothing).
func (g *Game) GrantCantAttackPlayerThisCombatForEffect(attacker, protected uuid.UUID, scope CantAttackScope, label string, source uuid.UUID) {
	d, ok := g.UntilEndOfCombatDuration()
	if !ok {
		return
	}
	g.grantCantAttackLocked(attacker, protected, scope, g.turnsBegunForLocked(attacker), d, label, source)
}
