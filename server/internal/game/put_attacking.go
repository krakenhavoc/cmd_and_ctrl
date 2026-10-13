package game

import "github.com/google/uuid"

// put_attacking.go — a permanent that is already on the battlefield is
// made an attacker without being declared one (#3143, CR 506.3c).

// PutOnBattlefieldAttackingForEffect makes a permanent that is ALREADY on
// the battlefield an attacking creature without declaring it one: "gain
// control of target creature … tap it, and it's attacking that player"
// (Midnight Crusader Shuttle). It is the second door onto
// stampEntryAttackerLocked, the one a permanent that did not just enter
// uses, and what it does is the same CR 506.3c rule: the creature was
// never DECLARED as an attacker, so it fires no "whenever ~ attacks"
// trigger and nothing that watches declarations sees it.
//
// The layer pass runs FIRST. A control change registered by the same
// effect has not been applied yet, and applying it later takes the
// creature out of combat (a creature that changes controller stops
// attacking, CR 506.4); stamping before that would be undone by it.
// Tapping is the caller's own step, since a creature put into combat
// is not always tapped.
//
// Returns false, changing nothing, when the permanent is not a creature
// on the battlefield, when it is already attacking something, or when
// `defender` is not a player, planeswalker or battle that can be
// attacked.
//
// Caller must hold g.mu.
func (g *Game) PutOnBattlefieldAttackingForEffect(id, defender uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	c := findBattlefieldCard(g, id)
	if c == nil || !c.IsCreature() || c.AttackingTarget != uuid.Nil {
		return false
	}
	if defender == uuid.Nil || g.classifyAttackTargetLocked(defender) == AttackTargetNone {
		return false
	}
	g.stampEntryAttackerLocked(id, defender)
	return true
}
