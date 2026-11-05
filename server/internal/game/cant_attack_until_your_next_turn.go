package game

import "github.com/google/uuid"

// cant_attack_until_your_next_turn.go — the third window of a
// CantAttackGrant (cant_attack_player.go): "Until your next turn,
// creatures can't attack you" (Chronomantic Escape, Star Trek
// Commander batch 48, #4974).
//
// The restriction binds `attacker` from now (its floor is the
// attacker's CURRENT seat-turn count, so a creature given haste later
// this turn or on the attacker's own next turn is still refused, as the
// card's 2013-04-15 ruling says) until `protected`'s next turn begins
// (CR 611.2: UntilYourNextTurnDuration, counted on the protected
// player's seat). Same grant, same reader, same refusal: nothing new is
// stored, it is the existing PlayerStatic payload with another floor
// and duration.

// GrantCantAttackPlayerUntilYourNextTurnForEffect makes `attacker`
// unable to attack what `scope` covers of `protected`'s until
// `protected`'s next turn. Same no-ops and caller contract as
// GrantCantAttackPlayerForEffect.
func (g *Game) GrantCantAttackPlayerUntilYourNextTurnForEffect(attacker, protected uuid.UUID, scope CantAttackScope, label string, source uuid.UUID) {
	g.grantCantAttackLocked(attacker, protected, scope, g.turnsBegunForLocked(attacker),
		g.UntilYourNextTurnDuration(protected), label, source)
}
