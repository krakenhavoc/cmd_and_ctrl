package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// attack_requirements_test.go pins CR 508.1d on the attack declaration
// (#1571, ADR 0045 Decisions 48-52): the verbs refuse a declaration
// that makes a requirement unobeyable, the active player's pass (and
// AdvanceStep) refuses one that leaves an obeyable requirement unmet,
// and the maximum is counted over the whole declaration — against
// restrictions, limits and costs. The cards and the enumerator are
// pinned in cards/effects and legal.

// mustAttackCreature pushes a creature for `owner` carrying `n` plain
// "attacks each combat if able" requirements, as a pinned data record
// (what Legion Warboss's token gets).
func mustAttackCreature(t *testing.T, g *Game, owner *Player, n int) uuid.UUID {
	t.Helper()
	id := pushKeywordCreature(t, g, owner, 2, 2)
	mods := make([]Mod, n)
	for i := range mods {
		mods[i] = AddAttackRequirementMod(uuid.Nil)
	}
	registerScopedEffectForTest(t, g, id, mods, IndefiniteDuration())
	return id
}

func goad(t *testing.T, g *Game, id, by uuid.UUID) {
	t.Helper()
	if err := g.SetGoaded(id, by); err != nil {
		t.Fatalf("SetGoaded: %v", err)
	}
}

func requirementErr(t *testing.T, err error) *AttackRequirementError {
	t.Helper()
	if !errors.Is(err, ErrAttackRequirement) {
		t.Fatalf("err = %v, want ErrAttackRequirement", err)
	}
	var re *AttackRequirementError
	if !errors.As(err, &re) {
		t.Fatalf("refusal is %T, want *AttackRequirementError", err)
	}
	return re
}

// TestAttackRequirementRefusesThePassUntilObeyed — a single
// requirement. The pass that ends the declaration is refused, naming
// the creature; declaring it answers the requirement; and once the
// pass is accepted the step moves on as ever.
func TestAttackRequirementRefusesThePassUntilObeyed(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zurgo := mustAttackCreature(t, g, me, 1)
	bear := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	re := requirementErr(t, g.PassPriority())
	if re.Attacker != zurgo {
		t.Errorf("refusal names %s, want the creature that must attack", re.Attacker)
	}
	if got := re.Sentence(me.ID); got != "Kw Creature must attack this combat if able." {
		t.Errorf("sentence = %q", got)
	}
	requirementErr(t, func() error { _, err := g.AdvanceStep(); return err }())
	if g.Turn.Step != StepDeclareAttackers {
		t.Fatalf("a refused pass moved the cursor to %s", g.Turn.Step)
	}
	// A creature with no requirement may attack first; it does not use
	// anything the requirement needs.
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("declare the Bear: %v", err)
	}
	requirementErr(t, g.PassPriority())
	if err := g.DeclareAttacker(zurgo, opp.ID); err != nil {
		t.Fatalf("declare the creature that must attack: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the requirement obeyed: %v", err)
	}
}

// TestAttackRequirementIsJudgedOncePerCombat — the declaration is over
// once the pass is accepted: a creature goaded afterwards is not made
// to attack late, and next combat judges afresh.
func TestAttackRequirementIsJudgedOncePerCombat(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with nothing required: %v", err)
	}
	goad(t, g, bear, opp.ID)
	g.WithWriteLock(func() {
		if e := g.attackRequirementsUnmetLocked(); e != nil {
			t.Errorf("a goad after the checkpoint re-opened the declaration: %v", e)
		}
	})
	g.WithWriteLock(func() { g.clearCombatLocked() })
	if g.attacksDeclared {
		t.Error("clearing combat kept the checkpoint")
	}
}

// TestAttackRequirementWithNoPossibleAttackIsNotOwed — a requirement
// against a restriction (CR 508.1d "without disobeying any
// restrictions"): a creature that must attack but can't, or is tapped,
// or has no opponent to attack, owes nothing and the pass is accepted.
func TestAttackRequirementWithNoPossibleAttackIsNotOwed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pacified := mustAttackCreature(t, g, me, 1)
	registerScopedEffectForTest(t, g, pacified, []Mod{AddRestrictionsMod(CantAttack)}, IndefiniteDuration())
	tapped := mustAttackCreature(t, g, me, 1)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.WithWriteLock(func() { findBattlefieldCard(g, tapped).Tapped = true })
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with every requirement impossible: %v", err)
	}
}

// TestAttackRequirementUnderALimitCountsTheWholeDeclaration — multiple
// requirements competing for one slot (Silent Arbiter). Either
// creature that must attack is a legal answer; a creature with no
// requirement taking the slot is refused; the one with MORE
// requirements wins the slot when they differ (CR 508.1d counts).
func TestAttackRequirementUnderALimitCountsTheWholeDeclaration(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLimitSource(t, g, opp, "Silent Arbiter")
	once := mustAttackCreature(t, g, me, 1)
	twice := mustAttackCreature(t, g, me, 2)
	bear := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	re := requirementErr(t, g.DeclareAttacker(bear, opp.ID))
	if re.Attacker != twice {
		t.Errorf("the Bear's refusal names %s, want the creature with two requirements", re.Attacker)
	}
	requirementErr(t, g.DeclareAttacker(once, opp.ID))
	if attackingCount(g) != 0 {
		t.Fatal("a refused declaration staged something")
	}
	if err := g.DeclareAttacker(twice, opp.ID); err != nil {
		t.Fatalf("the creature with the most requirements: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the maximum obeyed: %v", err)
	}
}

// TestAttackRequirementNeverForcesATax — requirement plus attack tax.
// A requirement never forces a payment (CR 508.1d), so a creature that
// could attack only through Propaganda owes nothing; paying is still
// allowed; and a goaded creature whose only free target is its goader
// must attack the goader rather than stay home.
func TestAttackRequirementNeverForcesATax(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	withCatalogAttackTaxes(t, func(key string) []AttackTax {
		if key == propagandaTestOracle {
			return []AttackTax{flatTax("{2}")}
		}
		return nil
	})
	me, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	for _, p := range []*Player{b, c, d} {
		pushTaxEnchantment(g, p, "Propaganda", propagandaTestOracle)
	}
	forced := mustAttackCreature(t, g, me, 1)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.WithWriteLock(func() {
		if e := g.attackRequirementsUnmetLocked(); e != nil {
			t.Fatalf("a requirement that needs a tax is owed: %v", e)
		}
	})
	// Paying is the player's choice, and still accepted.
	addMana(me, 2)
	if err := g.DeclareAttacker(forced, b.ID); err != nil {
		t.Fatalf("paying the tax to obey the requirement: %v", err)
	}

	// Goaded by D, with D untaxed: the free attack is at the goader,
	// and it is owed.
	g2 := newFourPlayerActiveGame(t)
	me2, b2, c2, d2 := g2.Seats[0], g2.Seats[1], g2.Seats[2], g2.Seats[3]
	pushTaxEnchantment(g2, b2, "Propaganda", propagandaTestOracle)
	pushTaxEnchantment(g2, c2, "Propaganda", propagandaTestOracle)
	goaded := pushKeywordCreature(t, g2, me2, 2, 2)
	goad(t, g2, goaded, d2.ID)
	advanceIntoStep(t, g2, StepDeclareAttackers)
	requirementErr(t, g2.PassPriority())
	if err := g2.DeclareAttacker(goaded, d2.ID); err != nil {
		t.Fatalf("the goaded creature at its goader, the only free target: %v", err)
	}
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("pass after the free attack: %v", err)
	}
}

// TestGoadInFourSeats — goad's two requirements (CR 701.15b) at a
// Commander table: at home refuses the pass, at the goader refuses the
// declaration while another opponent is open, at a planeswalker obeys
// only one of the two, and at another player obeys both. The bulk verb
// refuses an all-at-the-goader swing whole.
func TestGoadInFourSeats(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	goaded := pushKeywordCreature(t, g, me, 2, 2)
	other := pushKeywordCreature(t, g, me, 2, 2)
	goad(t, g, goaded, b.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)

	re := requirementErr(t, g.PassPriority())
	if re.Requirement.GoadedBy != b.ID {
		t.Errorf("refusal names %+v, want goad's requirement", re.Requirement)
	}
	re = requirementErr(t, g.DeclareAttacker(goaded, b.ID))
	if re.Requirement.OtherThan != b.ID {
		t.Errorf("at the goader: refusal names %+v, want the other-than requirement", re.Requirement)
	}
	if got := re.Sentence(b.ID); got != "Kw Creature is goaded by you and must attack a player other than you if able." {
		t.Errorf("sentence to the goader = %q", got)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{
		{Attacker: goaded, Target: b.ID}, {Attacker: other, Target: b.ID},
	}); !errors.Is(err, ErrAttackRequirement) {
		t.Fatalf("bulk swing at the goader = %v, want ErrAttackRequirement", err)
	}
	if attackingCount(g) != 0 {
		t.Fatal("a refused bulk declaration staged something")
	}
	if err := g.DeclareAttacker(goaded, c.ID); err != nil {
		t.Fatalf("the goaded creature at another opponent: %v", err)
	}
	// Re-pointing it back at the goader lowers what the declaration
	// obeys, so it is refused too.
	requirementErr(t, g.DeclareAttacker(goaded, b.ID))
	// The other creature owes nothing and may attack the goader.
	if err := g.DeclareAttacker(other, b.ID); err != nil {
		t.Fatalf("an ungoaded creature at the goader: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with goad obeyed: %v", err)
	}
}

// TestGoadedCreatureAttacksTheGoaderWhenNobodyElseCanBeAttacked — two
// seats: the goader is the only opponent, so "a player other than the
// goader" cannot be obeyed and attacking the goader is the maximum.
func TestGoadedCreatureAttacksTheGoaderWhenNobodyElseCanBeAttacked(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goaded := pushKeywordCreature(t, g, me, 2, 2)
	goad(t, g, goaded, opp.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	requirementErr(t, g.PassPriority())
	if err := g.DeclareAttacker(goaded, opp.ID); err != nil {
		t.Fatalf("the goaded creature at the only opponent: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

// TestMustAttackListsEveryAnswerUnderALimit — the view's and the
// enumerator's shared answer: both creatures that could take Silent
// Arbiter's slot are listed until one does, then nothing is.
func TestMustAttackListsEveryAnswerUnderALimit(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLimitSource(t, g, opp, "Silent Arbiter")
	a := mustAttackCreature(t, g, me, 1)
	b := mustAttackCreature(t, g, me, 1)
	advanceIntoStep(t, g, StepDeclareAttackers)
	var owed map[uuid.UUID][]uuid.UUID
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked(); owed = g.MustAttackForEffect() })
	if len(owed[a]) != 1 || len(owed[b]) != 1 {
		t.Fatalf("owed = %v, want both creatures listed at the opponent", owed)
	}
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { owed = g.MustAttackForEffect() })
	if owed != nil {
		t.Errorf("owed after the slot is used = %v, want nothing", owed)
	}
}

// TestMaxRequirementAssignmentUsesEveryCapacity is the search on its
// own: per-target rooms and a combat-wide room together.
func TestMaxRequirementAssignmentUsesEveryCapacity(t *testing.T) {
	x, y := uuid.New(), uuid.New()
	c1, c2, c3 := uuid.New(), uuid.New(), uuid.New()
	cands := []reqCandidate{
		{id: c1, pairs: []reqPair{{target: x, weight: 2}, {target: y, weight: 1}}},
		{id: c2, pairs: []reqPair{{target: x, weight: 2}}},
		{id: c3, pairs: []reqPair{{target: y, weight: 1}}},
	}
	// x takes one, the combat takes two: c2 at x (2) and c1 or c3 at y (1).
	got, witness := maxRequirementAssignment(cands, attackCapacity{global: 2, perTarget: map[uuid.UUID]int{x: 1}})
	if got != 3 {
		t.Fatalf("max = %d, want 3 (witness %v)", got, witness)
	}
	if len(witness) != 2 {
		t.Errorf("witness seats %d creatures, want 2", len(witness))
	}
	// Unlimited: everyone at their best.
	if got, _ := maxRequirementAssignment(cands, attackCapacity{global: -1}); got != 5 {
		t.Errorf("unlimited max = %d, want 5", got)
	}
}
