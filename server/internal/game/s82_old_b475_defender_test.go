package game

import "testing"

// s82_old_b475_defender_test.go — S82 older-set batch 475 (#5579):
// CombatControlledByDefender, "target creature defending player controls"
// on an attacker's activated ability (Ancient Hellkite).

func TestCombatControlledByDefenderReadsThePlayerTheSourceAttacks(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, aimed, other := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	hellkite := pushScopedTestCreature(g, me, 6, 6)
	idle := pushScopedTestCreature(g, me, 2, 2)
	theirs := pushScopedTestCreature(g, aimed, 2, 2)
	bystander := pushScopedTestCreature(g, other, 2, 2)
	spec := combatSpec(CombatOfSource, CombatControlledByDefender)

	// Not attacking: nothing is a defending player's creature.
	if got := legalIDs(g, sourceOf(g, me, hellkite), spec); len(got) != 0 {
		t.Errorf("a creature that is not attacking offers %v", got)
	}
	setCombat(g, hellkite, func(c *Card) { c.AttackingTarget = aimed })
	got := legalIDs(g, sourceOf(g, me, hellkite), spec)
	if len(got) != 1 || !got[theirs] {
		t.Errorf("controlled-by-defender = %v, want only %s: the player attacked, not every opponent (%s)", got, theirs, bystander)
	}
	if got[idle] {
		t.Error("the attacker's own controller's creature is a defending player's")
	}
	// Re-aimed at the other opponent, the set follows.
	setCombat(g, hellkite, func(c *Card) { c.AttackingTarget = other })
	if got := legalIDs(g, sourceOf(g, me, hellkite), spec); len(got) != 1 || !got[bystander] {
		t.Errorf("after attacking the other player, controlled-by-defender = %v, want only %s", got, bystander)
	}
}
