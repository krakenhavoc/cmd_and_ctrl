package game

import (
	"testing"

	"github.com/google/uuid"
)

// block_requirement_two_or_more_test.go — #6243, Gorm the Great: "must be
// blocked if able, and must be blocked by two or more creatures if able".
// Two requirements on the attacker (BlockRequirementMustBeBlocked and
// BlockRequirementTwoOrMore), and CR 509.1c asks the defending player to
// obey as many as a legal declaration can.

func gormAttacking(t *testing.T, blockers int) (g *Game, gorm uuid.UUID, walls []uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gorm = pushCombatant(t, g, me, "Gorm the Great", 2, 7)
	withBlockRequirement(t, g, gorm, BlockRequirementMustBeBlocked, BlockRequirementTwoOrMore)
	for i := 0; i < blockers; i++ {
		walls = append(walls, pushCombatant(t, g, opp, "Wall", 0, 4))
	}
	declareAttacks(t, g, gorm)
	return g, gorm, walls
}

// With two creatures able to block, one blocker alone obeys one of the two
// requirements: a legal step, but refused at the checkpoint with the
// two-or-more sentence. Both blocking obeys both.
func TestTwoOrMoreBlockersAreRequiredWhenTwoCanBlock(t *testing.T) {
	g, gorm, walls := gormAttacking(t, 2)
	opp := g.Seats[1]
	if err := block(t, g, walls[0], gorm); err != nil {
		t.Fatalf("one blocker as a step on the way: %v", err)
	}
	passToDefender(t, g)
	br := blockRequirementErr(t, g.PassPriority())
	if br.Requirement.Kind != BlockRequirementTwoOrMore {
		t.Errorf("the refusal gives up %q, want twoOrMore", br.Requirement.Kind)
	}
	if got := br.Sentence(opp.ID); got != "Gorm the Great must be blocked by two or more creatures if able." {
		t.Errorf("sentence = %q", got)
	}
	blockRequirementErr(t, g.FinishBlocks(opp.ID))
	if err := block(t, g, walls[1], gorm); err != nil {
		t.Fatalf("the second blocker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both requirements obeyed: %v", err)
	}
}

// With one creature able to block, the most a declaration can obey is
// "must be blocked": that creature must block, and blocking with it is
// enough.
func TestTwoOrMoreWithOneCreatureStillRequiresTheOneBlock(t *testing.T) {
	g, gorm, walls := gormAttacking(t, 1)
	opp := g.Seats[1]
	passToDefender(t, g)
	br := blockRequirementErr(t, g.PassPriority())
	if br.Requirement.Kind != BlockRequirementMustBeBlocked {
		t.Errorf("an empty declaration gives up %q, want mustBeBlocked", br.Requirement.Kind)
	}
	if err := block(t, g, walls[0], gorm); err != nil {
		t.Fatalf("the lone blocker: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("finish with the one creature blocking: %v", err)
	}
}

// With three, two blockers obey both requirements; the third is free.
func TestTwoOrMoreIsObeyedByAnyTwo(t *testing.T) {
	g, gorm, walls := gormAttacking(t, 3)
	opp := g.Seats[1]
	if err := block(t, g, walls[0], gorm, walls[2], gorm); err != nil {
		t.Fatalf("two of three: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("finish with two blocking: %v", err)
	}
}

// The requirement alone (without "must be blocked") asks nothing of a
// defender with one creature: no declaration it can make obeys it.
func TestTwoOrMoreAloneAsksNothingOfOneCreature(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushCombatant(t, g, me, "Pack Leader", 2, 2)
	withBlockRequirement(t, g, atk, BlockRequirementTwoOrMore)
	pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, atk)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("no blocks with one creature: %v", err)
	}
}

func TestTwoOrMoreIsAKnownKind(t *testing.T) {
	if !KnownBlockRequirementKind(BlockRequirementTwoOrMore) {
		t.Error("twoOrMore is not a known block requirement kind")
	}
}
