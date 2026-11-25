package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// seam6197_companions_test.go — #6197, the companion family of
// "can't attack/block alone":
//
//   - Okk: "can't attack unless a creature with greater power also
//     attacks" / "can't block unless a creature with greater power also
//     blocks" (AttackLimit.OthersPowerGreater, BlockRule.NotAloneOthersPowerGreater);
//   - Orcish Conscripts: "can't attack unless at least two other
//     creatures attack" / the block half (MinOthers);
//   - Ekundu Cyclops: "If a creature you control attacks, this creature
//     also attacks if able" (AttackRequirement.IfAnotherYouControlAttacks);
//   - Mogg Jailer: "can't attack if defending player controls an
//     untapped creature with power 2 or less"
//     (AttackTargetRestriction.DefenderMustNotControl, PermanentQuery.Untapped).
//
// CR 508.1c judges restrictions on the whole declaration, CR 508.1d
// requirements, and CR 509.1b the block side; CR 802.4b counts only the
// same defending player's blockers.

// pushSizedCreature puts an untapped, unsick creature of the given size
// on the battlefield, able to attack.
func pushSizedCreature(t *testing.T, g *Game, owner *Player, power, toughness int) uuid.UUID {
	t.Helper()
	id := pushKeywordCreature(t, g, owner, power, toughness)
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, id); c != nil {
			c.SummonedThisTurn = false
		}
	})
	return id
}

// pushCompanionRuleCreature is pushSizedCreature carrying the stubbed
// attack limits and block rules (combatLimitOracle).
func pushCompanionRuleCreature(t *testing.T, g *Game, owner *Player, power, toughness int) uuid.UUID {
	t.Helper()
	id := pushSizedCreature(t, g, owner, power, toughness)
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, id); c != nil {
			c.OracleID = combatLimitOracle
		}
	})
	return id
}

// TestOkkAttacksOnlyWithAStrongerCreature — CR 508.1c: the companion's
// power is read at declaration, so a 2-power creature is no company for
// a 3/3 Okk and a 4-power one is.
func TestOkkAttacksOnlyWithAStrongerCreature(t *testing.T) {
	t.Run("weaker companion is refused", func(t *testing.T) {
		stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitNotAlone, Max: 1, OthersPowerGreater: true}}, nil)
		g := newActiveGame(t)
		me, them := g.Seats[0], g.Seats[1]
		okk := pushCompanionRuleCreature(t, g, me, 3, 3)
		weak := pushSizedCreature(t, g, me, 2, 2)
		advanceIntoStep(t, g, StepDeclareAttackers)
		// Only a weaker creature could join it, so the declaration can
		// never become legal and the verb refuses it outright.
		if err := g.DeclareAttacker(okk, them.ID); !errors.Is(err, ErrAttackLimit) {
			t.Fatalf("Okk with only a 2-power companion = %v, want ErrAttackLimit", err)
		}
		if err := g.DeclareAttacker(weak, them.ID); err != nil {
			t.Fatalf("the plain creature was refused: %v", err)
		}
	})

	t.Run("stronger companion makes it legal", func(t *testing.T) {
		stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitNotAlone, Max: 1, OthersPowerGreater: true}}, nil)
		g := newActiveGame(t)
		me, them := g.Seats[0], g.Seats[1]
		okk := pushCompanionRuleCreature(t, g, me, 3, 3)
		strong := pushSizedCreature(t, g, me, 4, 4)
		advanceIntoStep(t, g, StepDeclareAttackers)
		if err := g.DeclareAttacker(okk, them.ID); err != nil {
			t.Fatalf("Okk with a 4-power companion available: %v", err)
		}
		// Declared alone, the checkpoint refuses the pass and names the
		// rule, because the stronger creature could still join.
		var e *AttackLimitError
		if err := g.PassPriority(); !errors.As(err, &e) {
			t.Fatalf("pass with Okk attacking alone = %v, want *AttackLimitError", err)
		}
		if got := e.Sentence(me.ID); got != "Kw Creature can't attack unless a creature with greater power also attacks." {
			t.Errorf("sentence = %q", got)
		}
		if err := g.DeclareAttacker(strong, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with both attacking: %v", err)
		}
		if c := findCard(g, okk); c.AttackingTarget == uuid.Nil {
			t.Error("Okk was taken out of a legal declaration")
		}
	})
}

// TestOrcishConscriptsNeedsTwoOtherAttackers — MinOthers 2: one other
// attacker is not enough, two are (CR 508.1c).
func TestOrcishConscriptsNeedsTwoOtherAttackers(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitNotAlone, Max: 1, MinOthers: 2}}, nil)
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	conscripts := pushCompanionRuleCreature(t, g, me, 2, 2)
	a := pushSizedCreature(t, g, me, 1, 1)
	b := pushSizedCreature(t, g, me, 1, 1)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, id := range []uuid.UUID{conscripts, a} {
		if err := g.DeclareAttacker(id, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	var e *AttackLimitError
	if err := g.PassPriority(); !errors.As(err, &e) {
		t.Fatalf("pass with one other attacker = %v, want *AttackLimitError", err)
	}
	if got := e.Sentence(me.ID); got != "Kw Creature can't attack unless at least two other creatures attack." {
		t.Errorf("sentence = %q", got)
	}
	if err := g.DeclareAttacker(b, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with two other attackers: %v", err)
	}
}

// TestOrcishConscriptsIsTakenOutWhenNoCompanionIsLeft — the checkpoint's
// other half: with only one other creature that can attack, the
// declaration can never be legal, so the Conscripts are taken out of it
// rather than wedging the table (CR 508.1c).
func TestOrcishConscriptsIsTakenOutWhenNoCompanionIsLeft(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitNotAlone, Max: 1, MinOthers: 2}}, nil)
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	conscripts := pushCompanionRuleCreature(t, g, me, 2, 2)
	a := pushSizedCreature(t, g, me, 1, 1)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(a, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// The Conscripts can't be declared at all: only one other creature
	// could ever attack beside them.
	if err := g.DeclareAttacker(conscripts, them.ID); !errors.Is(err, ErrAttackLimit) {
		t.Fatalf("Conscripts with one possible companion = %v, want ErrAttackLimit", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

// TestOkkBlocksOnlyWithAStrongerCreature — the block half (CR 509.1b):
// a weaker co-blocker leaves the declaration illegal, a stronger one
// makes it legal, and the refusal names the clause.
func TestOkkBlocksOnlyWithAStrongerCreature(t *testing.T) {
	stubCombatLimits(t, nil, []BlockRule{{NotAlone: true, NotAloneOthersPowerGreater: true}})
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	atk := pushSizedCreature(t, g, me, 1, 1)
	other := pushSizedCreature(t, g, me, 1, 1)
	okk := pushCompanionRuleCreature(t, g, them, 3, 3)
	weak := pushSizedCreature(t, g, them, 2, 2)
	strong := pushSizedCreature(t, g, them, 4, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range []uuid.UUID{atk, other} {
		if err := g.DeclareAttacker(a, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	var br *BlockRefusedError
	if e := g.DeclareBlocker(okk, atk); !errors.As(e, &br) {
		t.Fatalf("Okk blocking alone = %v, want *BlockRefusedError", e)
	}
	if got := br.Sentence(them.ID); got != "Kw Creature can't block unless a creature with greater power also blocks." {
		t.Errorf("sentence = %q", got)
	}
	// A weaker co-blocker is no company; a stronger one is.
	if e := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: okk, Attacker: atk}, {Blocker: weak, Attacker: other},
	}); !errors.Is(e, ErrIllegalBlock) {
		t.Fatalf("Okk with a 2-power co-blocker = %v, want ErrIllegalBlock", e)
	}
	if e := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: okk, Attacker: atk}, {Blocker: strong, Attacker: other},
	}); e != nil {
		t.Fatalf("Okk with a 4-power co-blocker: %v", e)
	}
}

// TestEkunduCyclopsAttacksOnceAnotherCreatureDoes — CR 508.1d: the
// requirement exists only in a declaration that has another creature of
// its controller in it, so declaring nothing owes nothing and declaring
// one other creature owes the Cyclops' attack.
func TestEkunduCyclopsAttacksOnceAnotherCreatureDoes(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	cyclops := pushSizedCreature(t, g, me, 3, 3)
	other := pushSizedCreature(t, g, me, 1, 1)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, cyclops)
		eff := c.Effective()
		eff.AttackRequirements = append(eff.AttackRequirements, AttackRequirement{
			Source: cyclops, SourceName: "Ekundu Cyclops", IfAnotherYouControlAttacks: true,
		})
		c.effective = &eff
	})
	advanceIntoStep(t, g, StepDeclareAttackers)

	// Nothing declared: the requirement does not exist, so the pass is
	// free.
	var owed map[uuid.UUID][]uuid.UUID
	g.WithWriteLock(func() { owed = g.MustAttackForEffect() })
	if owed != nil {
		t.Errorf("owed with an empty declaration = %v, want nothing", owed)
	}

	if err := g.DeclareAttacker(other, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	g.WithWriteLock(func() { owed = g.MustAttackForEffect() })
	if len(owed[cyclops]) == 0 {
		t.Fatalf("owed after another creature attacks = %v, want the Cyclops listed", owed)
	}
	var e *AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &e) {
		t.Fatalf("pass with the Cyclops home = %v, want *AttackRequirementError", err)
	}
	if err := g.DeclareAttacker(cyclops, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both attacking: %v", err)
	}
}

// TestMoggJailerReadsAnUntappedWeakCreature — the DefenderMustNotControl
// clause with PermanentQuery.Untapped (CR 508.1c): an untapped 2-power
// creature under the defender refuses the attack, the same creature
// tapped does not, and a 3-power one never did.
func TestMoggJailerReadsAnUntappedWeakCreature(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	jailer := pushSizedCreature(t, g, me, 1, 1)
	guard := pushSizedCreature(t, g, them, 2, 2)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, jailer)
		eff := c.Effective()
		eff.AttackTargetRestrictions = append(eff.AttackTargetRestrictions, AttackTargetRestriction{
			Source: jailer, SourceName: "Mogg Jailer",
			DefenderMustNotControl: []PermanentQuery{
				PermanentQuery{Types: []string{"creature"}, Untapped: true}.WithPowerAtMost(2),
			},
		})
		c.effective = &eff
	})
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(jailer, them.ID); !errors.Is(err, ErrIllegalAttackTarget) {
		t.Fatalf("an untapped 2-power guard = %v, want ErrIllegalAttackTarget", err)
	}
	g.WithWriteLock(func() { findBattlefieldCard(g, guard).Tapped = true })
	if err := g.DeclareAttacker(jailer, them.ID); err != nil {
		t.Fatalf("a tapped guard: %v", err)
	}
}
