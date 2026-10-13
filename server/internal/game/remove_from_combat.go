package game

import (
	"errors"

	"github.com/google/uuid"
)

// ErrNotInCombat is returned by RemoveFromCombat when the named
// permanent is on the battlefield but is neither attacking nor
// blocking, so there is nothing to take it out of. A refusal rather
// than a silent no-op for the sandbox verb: the menu only offers the
// row on a creature in combat, so reaching this means a stale click,
// and an undo entry for a change that changed nothing would be a
// press the player has to spend to discover that.
var ErrNotInCombat = errors.New("game: that permanent is not attacking or blocking")

// RemoveFromCombat takes one permanent out of combat (CR 506.4). The
// sandbox verb behind the context menu's "Remove from combat" row
// (#672), for a player fixing a misclick, and the same state change
// an effect that says "remove target creature from combat" makes
// (RemoveFromCombatForEffect): both end in removeFromCombatLocked,
// the primitive regeneration (CR 701.19a, #667) and a control change
// already use, so there is one definition of what leaving combat
// clears.
//
// What that means for the other side of the fight, decided here
// because CR 506.4 leaves it to the surrounding rules:
//
//   - A removed ATTACKER stops attacking and drops its blocked
//     record. The creatures that were blocking it stay blocking
//     creatures (CR 506.4: removing the attacker does not remove its blockers; one
//     remains a blocking creature) and keep their
//     BlockingTarget, so "can't block again" and "whenever ~ blocks"
//     bookkeeping is undisturbed; they simply block nothing. Combat
//     damage is read off the attackers, so they deal and take none.
//   - A removed BLOCKER stops blocking. The attacker it blocked stays
//     blocked (CR 509.1h): "blocked" is the lock-in's record,
//     Game.blockedAttackers, keyed by the attacker, which this does
//     not touch.
//   - Either way the creature deals and receives no combat damage
//     (it is not in combat); it stays on the battlefield, tapped or not as it
//     was.
//
// Not step-gated, like clear_combat: it is a correction, and combat
// can be wedged in any step. Returns ErrGameNotActive outside an
// active game, ErrCardNotFound when the permanent is not on the
// battlefield, ErrNotInCombat when it is not in combat.
func (g *Game) RemoveFromCombat(cardID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	return g.RemoveFromCombatForEffect(cardID)
}

// RemoveFromCombatForEffect is RemoveFromCombat for a resolution that
// already holds g.mu: "remove target creature from combat" (CR 506.4)
// calls it. Same errors, minus the active-game check the caller's own
// resolution implies.
//
// Caller must hold g.mu.
func (g *Game) RemoveFromCombatForEffect(cardID uuid.UUID) error {
	i := findCardOnBattlefield(g, cardID)
	if i < 0 {
		return ErrCardNotFound
	}
	c := &g.Battlefield.Cards[i]
	if c.AttackingTarget == uuid.Nil && c.BlockingTarget == uuid.Nil {
		return ErrNotInCombat
	}
	g.removeFromCombatLocked(c)
	return nil
}
