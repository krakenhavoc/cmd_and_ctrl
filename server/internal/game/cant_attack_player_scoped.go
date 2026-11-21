package game

import "github.com/google/uuid"

// GrantCantAttackPlayerScopedForEffect makes `attacker` unable to attack
// what `scope` covers of `protected`'s during `attacker`'s next turn:
// GrantCantAttackPlayerForEffect with the scope chosen by the caller.
// Zagorka, Mother of Sanctum's "can't attack you during their next
// turn" is PlayerOnly: a planeswalker its controller controls may still
// be attacked. Same no-ops and caller contract as that function (hold
// g.mu for writing; a missing player or attacker == protected does
// nothing).
func (g *Game) GrantCantAttackPlayerScopedForEffect(attacker, protected uuid.UUID, scope CantAttackScope, label string, source uuid.UUID) {
	g.grantCantAttackLocked(attacker, protected, scope, g.turnsBegunForLocked(attacker)+1,
		g.UntilEndOfYourNextTurnDuration(attacker), label, source)
}
