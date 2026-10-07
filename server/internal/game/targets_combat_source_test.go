package game

import (
	"testing"

	"github.com/google/uuid"
)

// targets_combat_source_test.go — #1863. TargetSpec.CombatWithSource is
// judged in the one shared walk, so the legal set and the announce
// check agree, and a source that has left is read from its last-known
// record.

func combatSpec(of SourceCombatOf, role CombatRole) *TargetSpec {
	return &TargetSpec{
		Mode: "creature", Zones: []ZoneKind{ZoneBattlefield},
		CombatWithSource: SourceCombat{Of: of, Role: role},
	}
}

func setCombat(g *Game, id uuid.UUID, mutate func(c *Card)) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				mutate(&g.Battlefield.Cards[i])
			}
		}
	})
}

func legalIDs(g *Game, src TargetSource, spec *TargetSpec) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	g.WithWriteLock(func() {
		for _, id := range g.LegalTargetsForEffect(src, spec).Cards {
			out[id] = true
		}
	})
	return out
}

func sourceOf(g *Game, controller, id uuid.UUID) TargetSource {
	var c Card
	g.ReadSnapshot(func() { c = *findBattlefieldCard(g, id) })
	return SourceObject(controller, &c)
}

func TestCombatWithSourceReadsTheSourcesBlocks(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	attacker := pushScopedTestCreature(g, me, 2, 2)
	otherAttacker := pushScopedTestCreature(g, me, 2, 2)
	blocker := pushScopedTestCreature(g, opp, 1, 4)
	idle := pushScopedTestCreature(g, opp, 1, 1)
	setCombat(g, attacker, func(c *Card) { c.AttackingTarget = opp })
	setCombat(g, otherAttacker, func(c *Card) { c.AttackingTarget = opp })
	setCombat(g, blocker, func(c *Card) { c.BlockingTarget = attacker })

	// "creature it's blocking": only the attacker the source blocks.
	got := legalIDs(g, sourceOf(g, opp, blocker), combatSpec(CombatOfSource, CombatBlockedBy))
	if len(got) != 1 || !got[attacker] {
		t.Errorf("blocked-by-source = %v, want only %s", got, attacker)
	}
	// "creature blocking this creature", read off the attacker.
	got = legalIDs(g, sourceOf(g, me, attacker), combatSpec(CombatOfSource, CombatBlocking))
	if len(got) != 1 || !got[blocker] {
		t.Errorf("blocking-source = %v, want only %s", got, blocker)
	}
	if got := legalIDs(g, sourceOf(g, me, otherAttacker), combatSpec(CombatOfSource, CombatBlocking)); len(got) != 0 {
		t.Errorf("an unblocked attacker has blockers %v", got)
	}
	// A blocker is not "blocking" anything's blocker, and an idle
	// creature is neither.
	if got := legalIDs(g, sourceOf(g, opp, idle), combatSpec(CombatOfSource, CombatBlockedBy)); len(got) != 0 {
		t.Errorf("a creature that blocks nothing offers %v", got)
	}
}

func TestCombatWithSourceHostIsTheAttachedCreature(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	host := pushScopedTestCreature(g, me, 2, 2)
	blocker := pushScopedTestCreature(g, opp, 1, 4)
	equipment := pushScopedTestCreature(g, me, 0, 0)
	setCombat(g, host, func(c *Card) { c.AttackingTarget = opp })
	setCombat(g, blocker, func(c *Card) { c.BlockingTarget = host })

	spec := combatSpec(CombatOfHost, CombatBlocking)
	if got := legalIDs(g, sourceOf(g, me, equipment), spec); len(got) != 0 {
		t.Errorf("an unattached source offers %v", got)
	}
	setCombat(g, equipment, func(c *Card) { c.AttachedTo = TargetRef{Kind: TargetCard, ID: host} })
	if got := legalIDs(g, sourceOf(g, me, equipment), spec); len(got) != 1 || !got[blocker] {
		t.Errorf("blocking-host = %v, want only %s", got, blocker)
	}
	// The mirror: a host that blocks, read through the attachment.
	setCombat(g, equipment, func(c *Card) { c.AttachedTo = TargetRef{Kind: TargetCard, ID: blocker} })
	if got := legalIDs(g, sourceOf(g, opp, equipment), combatSpec(CombatOfHost, CombatBlockedBy)); len(got) != 1 || !got[host] {
		t.Errorf("blocked-by-host = %v, want only %s", got, host)
	}
}

func TestCombatWithSourceNeedsASource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	attacker := pushScopedTestCreature(g, me, 2, 2)
	blocker := pushScopedTestCreature(g, opp, 1, 4)
	setCombat(g, attacker, func(c *Card) { c.AttackingTarget = opp })
	setCombat(g, blocker, func(c *Card) { c.BlockingTarget = attacker })
	for _, of := range []SourceCombatOf{CombatOfSource, CombatOfHost} {
		if got := legalIDs(g, SourceChooser(opp), combatSpec(of, CombatBlockedBy)); len(got) != 0 {
			t.Errorf("%s with no source admits %v, want nothing", of, got)
		}
	}
}

func TestCombatWithSourceEndsWhenTheCreatureLeavesCombat(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	attacker := pushScopedTestCreature(g, me, 2, 2)
	blocker := pushScopedTestCreature(g, opp, 1, 4)
	setCombat(g, attacker, func(c *Card) { c.AttackingTarget = opp })
	setCombat(g, blocker, func(c *Card) { c.BlockingTarget = attacker })
	spec := combatSpec(CombatOfSource, CombatBlockedBy)
	if got := legalIDs(g, sourceOf(g, opp, blocker), spec); !got[attacker] {
		t.Fatalf("setup: %v", got)
	}
	// Removed from combat (CR 506.4): no longer an attacker.
	setCombat(g, attacker, func(c *Card) { c.AttackingTarget = uuid.Nil })
	if got := legalIDs(g, sourceOf(g, opp, blocker), spec); len(got) != 0 {
		t.Errorf("a creature removed from combat is still offered: %v", got)
	}
	// And the blocker itself leaving combat blocks nothing.
	setCombat(g, attacker, func(c *Card) { c.AttackingTarget = opp })
	setCombat(g, blocker, func(c *Card) { c.clearBlocking() })
	if got := legalIDs(g, sourceOf(g, opp, blocker), spec); len(got) != 0 {
		t.Errorf("a blocker that left combat still blocks: %v", got)
	}
}

// A departed source is read from its record (CR 608.2h): what it was
// blocking when it left.
func TestCombatWithSourceReadsTheDepartedSourcesRecord(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	attacker := pushScopedTestCreature(g, me, 2, 2)
	blocker := pushScopedTestCreature(g, opp, 1, 4)
	setCombat(g, attacker, func(c *Card) { c.AttackingTarget = opp })
	setCombat(g, blocker, func(c *Card) { c.BlockingTarget = attacker })

	var rec PermanentInfo
	g.WithWriteLock(func() {
		g.rememberDepartingPermanentLocked(blocker)
		c := findBattlefieldCard(g, blocker)
		rec, _ = g.lastKnownPermanentLocked(ObjectRef{ID: blocker, Epoch: c.ObjectEpoch})
	})
	if len(rec.Blocking) != 1 || rec.Blocking[0] != attacker {
		t.Fatalf("record blocking = %v, want [%s]", rec.Blocking, attacker)
	}
	src := SourceSnapshot(opp, &rec.Characteristic)
	src.Departed, src.DepartedID = &rec, blocker
	if got := legalIDs(g, src, combatSpec(CombatOfSource, CombatBlockedBy)); len(got) != 1 || !got[attacker] {
		t.Errorf("departed blocked-by = %v, want only %s", got, attacker)
	}
}
