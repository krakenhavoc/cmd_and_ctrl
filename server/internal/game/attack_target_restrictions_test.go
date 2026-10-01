package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// attack_target_restrictions_test.go pins ADR 0106 §2 (#1794): a
// creature that "can't attack its owner [or planeswalkers its owner
// controls]". Both declaration verbs refuse it, the per-attacker
// target list leaves it out, CR 508.1d's maximum is counted over the
// targets that are left, the owner is read live, and CR 508.7b's
// reselection is exempt. The enumerator half is pinned in
// internal/legal, the card-facing constructors in cards/effects.

const (
	sleeperOracle = "test-sleeper-agent"   // attacks each combat; can't attack its owner or their planeswalkers
	exileOracle   = "test-exiled-champion" // can't attack its owner (no planeswalker clause)
)

// withOwnerRestrictionStatics installs the two test cards' statics, in
// the shape the catalog constructors write them.
func withOwnerRestrictionStatics(t *testing.T) {
	t.Helper()
	self := func(target *Card, _ *Game, source *Card) bool {
		return target.InstanceID == source.InstanceID && target.IsCreature()
	}
	restrict := func(pw bool) StaticAbility {
		return StaticAbility{
			Layer:     Layer6Ability,
			AppliesTo: self,
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, AttackTargetRestriction{
					Source: source.InstanceID, SourceName: source.Name, NotOwner: true, NotOwnersPlaneswalkers: pw,
				})
			},
		}
	}
	mustAttack := StaticAbility{
		Layer:     Layer6Ability,
		AppliesTo: self,
		Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
			c.AttackRequirements = append(c.AttackRequirements, AttackRequirement{Source: source.InstanceID, SourceName: source.Name})
		},
	}
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case sleeperOracle:
			return []StaticAbility{mustAttack, restrict(true)}
		case exileOracle:
			return []StaticAbility{restrict(false)}
		}
		return nil
	})
}

// pushOwnedByAnother puts a ready creature on the battlefield under
// `controller`, owned by `owner`, as Xantcha is after its ADR 0102
// entry.
func pushOwnedByAnother(g *Game, owner, controller uuid.UUID, name, oracle string) uuid.UUID {
	id := pushTypedTestCard(g, Card{
		Name: name, TypeLine: "Legendary Creature — Phyrexian Minion", OracleID: oracle,
		Power: 5, Toughness: 5, Owner: owner, Controller: controller,
	})
	// It has been here since before the turn: not summoning sick.
	g.WithWriteLock(func() { findBattlefieldCard(g, id).SummonedThisTurn = false })
	return id
}

func attackRefIDs(refs []AttackTargetRef) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.ID)
	}
	return out
}

func attackTargetsOf(g *Game, id uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		out = attackRefIDs(g.AttackTargetsForAttackerForEffect(findBattlefieldCard(g, id)))
	})
	return out
}

func restrictionErr(t *testing.T, err error) *AttackTargetRestrictionError {
	t.Helper()
	if !errors.Is(err, ErrIllegalAttackTarget) {
		t.Fatalf("err = %v, want ErrIllegalAttackTarget", err)
	}
	var re *AttackTargetRestrictionError
	if !errors.As(err, &re) {
		t.Fatalf("refusal is %T, want *AttackTargetRestrictionError", err)
	}
	return re
}

// TestCantAttackItsOwnerOrTheirPlaneswalkers — Xantcha's whole
// sentence in four seats. The owner and the owner's planeswalker are
// refused by both verbs and left out of the per-attacker list; another
// opponent, their planeswalker and a battle the owner PROTECTS are all
// legal.
func TestCantAttackItsOwnerOrTheirPlaneswalkers(t *testing.T) {
	withOwnerRestrictionStatics(t)
	g := newFourPlayerActiveGame(t)
	me, owner, other, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	sleeper := pushOwnedByAnother(g, owner.ID, me.ID, "Sleeper Agent", sleeperOracle)
	ownersWalker := pushPlaneswalkerForTest(g, owner.ID, "Owner's Walker", 3)
	othersWalker := pushPlaneswalkerForTest(g, other.ID, "Other Walker", 3)
	ownersBattle := pushBattleForTest(g, other.ID, owner.ID, "Owner's Siege", 3)
	advanceIntoStep(t, g, StepDeclareAttackers)

	got := attackTargetsOf(g, sleeper)
	for _, want := range []uuid.UUID{other.ID, third.ID, othersWalker, ownersBattle} {
		if !slices.Contains(got, want) {
			t.Errorf("per-attacker targets %v miss %s", got, want)
		}
	}
	for _, refused := range []uuid.UUID{owner.ID, ownersWalker, me.ID} {
		if slices.Contains(got, refused) {
			t.Errorf("per-attacker targets %v include the forbidden %s", got, refused)
		}
	}
	// The per-seat list is unchanged: it is the view's, for prices.
	var perSeat []uuid.UUID
	g.WithWriteLock(func() { perSeat = attackRefIDs(g.AttackTargetsForEffect(me.ID)) })
	if !slices.Contains(perSeat, owner.ID) || !slices.Contains(perSeat, ownersWalker) {
		t.Errorf("per-seat targets %v lost the owner: they are not about one creature", perSeat)
	}

	re := restrictionErr(t, g.Clone().DeclareAttacker(sleeper, owner.ID))
	if re.Attacker != sleeper || !re.Restriction.NotOwnersPlaneswalkers {
		t.Errorf("refusal = %+v", re)
	}
	if got, want := re.Sentence(), "Sleeper Agent can't attack its owner or planeswalkers its owner controls."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	restrictionErr(t, g.Clone().DeclareAttacker(sleeper, ownersWalker))
	if _, err := g.Clone().DeclareAttackers([]AttackDeclaration{{Attacker: sleeper, Target: owner.ID}}); !errors.Is(err, ErrNoLegalAttackers) {
		t.Errorf("bulk declaration at the owner: err = %v, want ErrNoLegalAttackers", err)
	}
	for _, ok := range []uuid.UUID{other.ID, othersWalker, ownersBattle} {
		if err := g.Clone().DeclareAttacker(sleeper, ok); err != nil {
			t.Errorf("declare at %s: %v", ok, err)
		}
	}
}

// TestCantAttackItsOwnerAloneLeavesTheirPlaneswalkers — Alexios's
// shorter sentence: the owner is refused, a planeswalker the owner
// controls is not.
func TestCantAttackItsOwnerAloneLeavesTheirPlaneswalkers(t *testing.T) {
	withOwnerRestrictionStatics(t)
	g := newFourPlayerActiveGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	champ := pushOwnedByAnother(g, owner.ID, me.ID, "Exiled Champion", exileOracle)
	walker := pushPlaneswalkerForTest(g, owner.ID, "Owner's Walker", 3)
	advanceIntoStep(t, g, StepDeclareAttackers)

	got := attackTargetsOf(g, champ)
	if slices.Contains(got, owner.ID) || !slices.Contains(got, walker) {
		t.Fatalf("targets = %v: want the walker and not the owner", got)
	}
	re := restrictionErr(t, g.Clone().DeclareAttacker(champ, owner.ID))
	if got, want := re.Sentence(), "Exiled Champion can't attack its owner."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if err := g.DeclareAttacker(champ, walker); err != nil {
		t.Fatalf("declare at the owner's planeswalker: %v", err)
	}
}

// TestCantAttackOwnerIsReadOffTheCreaturesOwnOwner — the owner is read
// live (CR 108.3). Back under its owner's control, the restriction
// names its own controller and narrows nothing; a different owner moves
// the refusal with it, as a Clone's would.
func TestCantAttackOwnerIsReadOffTheCreaturesOwnOwner(t *testing.T) {
	withOwnerRestrictionStatics(t)
	g := newFourPlayerActiveGame(t)
	me, other := g.Seats[0], g.Seats[2]
	home := pushOwnedByAnother(g, me.ID, me.ID, "Sleeper At Home", exileOracle)
	clone := pushOwnedByAnother(g, other.ID, me.ID, "Sleeper Copy", exileOracle)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, home); len(got) != 3 {
		t.Errorf("owned by its controller: targets = %v, want all three opponents", got)
	}
	got := attackTargetsOf(g, clone)
	if slices.Contains(got, other.ID) || len(got) != 2 {
		t.Errorf("owned by %s: targets = %v, want the two others", other.ID, got)
	}
}

// TestCantAttackOwnerYieldsToTheRequirement — CR 508.1d counts
// requirements "without disobeying any restrictions". In four seats the
// creature must attack, and only a non-owner answers; in two, its only
// opponent is its owner, so nothing is owed and the pass is accepted.
func TestCantAttackOwnerYieldsToTheRequirement(t *testing.T) {
	withOwnerRestrictionStatics(t)

	t.Run("four seats", func(t *testing.T) {
		g := newFourPlayerActiveGame(t)
		me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]
		sleeper := pushOwnedByAnother(g, owner.ID, me.ID, "Sleeper Agent", sleeperOracle)
		advanceIntoStep(t, g, StepDeclareAttackers)

		requirementErr(t, g.Clone().PassPriority())
		var owed map[uuid.UUID][]uuid.UUID
		g.WithWriteLock(func() { owed = g.MustAttackForEffect() })
		if len(owed[sleeper]) != 2 || slices.Contains(owed[sleeper], owner.ID) {
			t.Fatalf("owed answers = %v, want the two opponents who are not the owner", owed[sleeper])
		}
		if err := g.DeclareAttacker(sleeper, other.ID); err != nil {
			t.Fatal(err)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with the requirement obeyed: %v", err)
		}
	})

	t.Run("two seats", func(t *testing.T) {
		g := newActiveGame(t)
		me, owner := g.Seats[0], g.Seats[1]
		sleeper := pushOwnedByAnother(g, owner.ID, me.ID, "Sleeper Agent", sleeperOracle)
		advanceIntoStep(t, g, StepDeclareAttackers)

		if got := attackTargetsOf(g, sleeper); len(got) != 0 {
			t.Fatalf("targets = %v, want none: the only opponent is the owner", got)
		}
		var owed map[uuid.UUID][]uuid.UUID
		g.WithWriteLock(func() { owed = g.MustAttackForEffect() })
		if len(owed) != 0 {
			t.Errorf("owed = %v, want nothing: the requirement can't be obeyed", owed)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with only the owner to attack: %v", err)
		}
	})
}

// TestCantAttackOwnerGoesWithItsAbilities — CR 613.1f: a creature that
// loses all abilities loses its own restriction and may attack its
// owner.
func TestCantAttackOwnerGoesWithItsAbilities(t *testing.T) {
	withOwnerRestrictionStatics(t)
	g := newFourPlayerActiveGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	sleeper := pushOwnedByAnother(g, owner.ID, me.ID, "Sleeper Agent", sleeperOracle)
	registerScopedEffectForTest(t, g, sleeper, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, sleeper); !slices.Contains(got, owner.ID) {
		t.Fatalf("targets = %v: a creature with no abilities may attack its owner", got)
	}
	if err := g.DeclareAttacker(sleeper, owner.ID); err != nil {
		t.Fatalf("declare at the owner after losing all abilities: %v", err)
	}
}

// TestCantAttackOwnerDoesNotStopAReselection — CR 508.7b: while
// reselecting what a creature attacks, it "isn't affected by
// requirements or restrictions that apply to the declaration of
// attackers". The reselection may point it at its owner.
func TestCantAttackOwnerDoesNotStopAReselection(t *testing.T) {
	withOwnerRestrictionStatics(t)
	g := newFourPlayerActiveGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]
	sleeper := pushOwnedByAnother(g, owner.ID, me.ID, "Sleeper Agent", sleeperOracle)
	declaredAttackerAt(t, g, sleeper, other.ID)
	var err error
	var refs []uuid.UUID
	g.WithWriteLock(func() {
		refs = attackRefIDs(g.ReselectAttackTargetsForEffect(sleeper))
		err = g.ReselectAttackTargetForEffect(sleeper, owner.ID)
	})
	if !slices.Contains(refs, owner.ID) {
		t.Errorf("reselect offer %v leaves out the owner", refs)
	}
	if err != nil {
		t.Fatalf("reselect onto the owner: %v", err)
	}
	if c := findBattlefieldCard(g, sleeper); c.AttackingTarget != owner.ID {
		t.Errorf("attacking %s after the reselection, want the owner", c.AttackingTarget)
	}
}
