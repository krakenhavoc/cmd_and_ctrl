package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// attack_unless_defender_is_test.go pins the rest of #1879 (ADR 0107 §2):
// the "can't attack unless" conditions on the defending player that are
// not "controls a permanent" — poisoned (CR 122.1f), the monarch (CR
// 725.1), a graveyard of N cards, and "you control more creatures than
// defending player" — and the granted form, Veiled Serpent's
// cantAttackUnlessDefenderControls mod. Each is judged per target's
// defending player (CR 508.5, 508.5a), like DefenderMustControl.

const (
	poisonedOracle  = "test-chained-throatseeker"  // … unless defending player is poisoned
	monarchOracle   = "test-crown-hunter-hireling" // … unless defending player is the monarch
	graveyardOracle = "test-vantress-gargoyle"     // … has seven or more cards in their graveyard
	goonOracle      = "test-goblin-goon"           // … unless you control more creatures than defending player
)

func withDefenderConditionStatics(t *testing.T) {
	t.Helper()
	restrict := func(r AttackTargetRestriction) StaticAbility {
		return StaticAbility{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature()
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				r := r
				r.Source, r.SourceName = source.InstanceID, source.Name
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, r)
			},
		}
	}
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case poisonedOracle:
			return []StaticAbility{restrict(AttackTargetRestriction{DefenderMustBePoisoned: true})}
		case monarchOracle:
			return []StaticAbility{restrict(AttackTargetRestriction{DefenderMustBeMonarch: true})}
		case graveyardOracle:
			return []StaticAbility{restrict(AttackTargetRestriction{DefenderGraveyardAtLeast: 7})}
		case goonOracle:
			return []StaticAbility{restrict(AttackTargetRestriction{ControllerMustControlMore: []PermanentQuery{{Types: []string{"creature"}}}})}
		}
		return nil
	})
}

func pushBearFor(g *Game, controller uuid.UUID) uuid.UUID {
	return pushTypedTestCard(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: controller, Controller: controller})
}

// TestCantAttackUnlessDefendingPlayerIsPoisoned — Chained Throatseeker in
// four seats: only the poisoned opponent, and that opponent's planeswalker.
func TestCantAttackUnlessDefendingPlayerIsPoisoned(t *testing.T) {
	withDefenderConditionStatics(t)
	g := newFourPlayerActiveGame(t)
	me, sick, clean := g.Seats[0], g.Seats[1], g.Seats[2]
	horror := pushOwnedByAnother(g, me.ID, me.ID, "Chained Throatseeker", poisonedOracle)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(sick.ID, CounterPoison, 1); err != nil {
			t.Fatal(err)
		}
	})
	sickWalker := pushPlaneswalkerForTest(g, sick.ID, "Sick Walker", 3)
	cleanWalker := pushPlaneswalkerForTest(g, clean.ID, "Clean Walker", 3)
	advanceIntoStep(t, g, StepDeclareAttackers)

	got := attackTargetsOf(g, horror)
	if !slices.Contains(got, sick.ID) || !slices.Contains(got, sickWalker) {
		t.Errorf("targets %v miss the poisoned opponent or their planeswalker", got)
	}
	if slices.Contains(got, clean.ID) || slices.Contains(got, cleanWalker) || slices.Contains(got, g.Seats[3].ID) {
		t.Errorf("targets %v include an opponent with no poison counters", got)
	}
	re := restrictionErr(t, g.Clone().DeclareAttacker(horror, clean.ID))
	if got, want := re.Sentence(), "Chained Throatseeker can't attack P3: P3 isn't poisoned."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if err := g.Clone().DeclareAttacker(horror, sickWalker); err != nil {
		t.Errorf("declare at the poisoned player's planeswalker: %v", err)
	}
}

// TestCantAttackUnlessDefendingPlayerIsTheMonarch — Crown-Hunter Hireling:
// nobody while there is no monarch, only the monarch once there is one,
// and the crown moving moves the target.
func TestCantAttackUnlessDefendingPlayerIsTheMonarch(t *testing.T) {
	withDefenderConditionStatics(t)
	g := newFourPlayerActiveGame(t)
	me, crowned, other := g.Seats[0], g.Seats[1], g.Seats[2]
	ogre := pushOwnedByAnother(g, me.ID, me.ID, "Crown-Hunter Hireling", monarchOracle)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, ogre); len(got) != 0 {
		t.Fatalf("targets = %v with no monarch, want none", got)
	}
	g.WithWriteLock(func() { g.Monarch = crowned.ID })
	if got := attackTargetsOf(g, ogre); !slices.Equal(got, []uuid.UUID{crowned.ID}) {
		t.Fatalf("targets = %v, want only the monarch", got)
	}
	re := restrictionErr(t, g.Clone().DeclareAttacker(ogre, other.ID))
	if got, want := re.Sentence(), "Crown-Hunter Hireling can't attack P3: P3 isn't the monarch."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	// I am the monarch: I am not a defending player, so nobody.
	g.WithWriteLock(func() { g.Monarch = me.ID })
	if got := attackTargetsOf(g, ogre); len(got) != 0 {
		t.Errorf("targets = %v while I am the monarch, want none", got)
	}
}

// TestCantAttackUnlessDefendingPlayerHasSevenCardsInTheirGraveyard —
// Vantress Gargoyle: six is not enough, seven is.
func TestCantAttackUnlessDefendingPlayerHasSevenCardsInTheirGraveyard(t *testing.T) {
	withDefenderConditionStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gargoyle := pushOwnedByAnother(g, me.ID, me.ID, "Vantress Gargoyle", graveyardOracle)
	for i := 0; i < 6; i++ {
		opp.Graveyard.PushTop(Card{InstanceID: uuid.New(), Name: "Card", Owner: opp.ID})
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	re := restrictionErr(t, g.Clone().DeclareAttacker(gargoyle, opp.ID))
	if got, want := re.Sentence(), "Vantress Gargoyle can't attack P2: P2 has fewer than seven cards in their graveyard."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	opp.Graveyard.PushTop(Card{InstanceID: uuid.New(), Name: "Card", Owner: opp.ID})
	if err := g.DeclareAttacker(gargoyle, opp.ID); err != nil {
		t.Errorf("declare at a seven-card graveyard: %v", err)
	}
}

// TestCantAttackUnlessYouControlMoreCreaturesThanDefendingPlayer — Goblin
// Goon in four seats. I control the Goon and a Bear: two. It may attack
// the opponent with one creature, not the one with two or three. The
// count is of effective creatures, so an animated land counts.
func TestCantAttackUnlessYouControlMoreCreaturesThanDefendingPlayer(t *testing.T) {
	withDefenderConditionStatics(t)
	g := newFourPlayerActiveGame(t)
	me, one, two, three := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	goon := pushOwnedByAnother(g, me.ID, me.ID, "Goblin Goon", goonOracle)
	pushBearFor(g, me.ID)
	pushBearFor(g, one.ID)
	pushBearFor(g, two.ID)
	pushBearFor(g, two.ID)
	for i := 0; i < 3; i++ {
		pushBearFor(g, three.ID)
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, goon); !slices.Equal(got, []uuid.UUID{one.ID}) {
		t.Fatalf("targets = %v, want only the opponent with fewer creatures", got)
	}
	re := restrictionErr(t, g.Clone().DeclareAttacker(goon, two.ID))
	if got, want := re.Sentence(), "Goblin Goon can't attack P3: P1 doesn't control more creatures than P3."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	// A land of mine that becomes a creature makes three: now P3 too.
	land := pushLandForTest(g, me.ID, "Forest", "Basic Land — Forest")
	registerScopedEffectForTest(t, g, land, []Mod{AddTypesMod("Creature"), SetBasePowerMod(1), SetBaseToughnessMod(1)}, IndefiniteDuration())
	if got := attackTargetsOf(g, goon); !slices.Contains(got, two.ID) || slices.Contains(got, three.ID) {
		t.Errorf("targets = %v, want P3 (two creatures) but not P4 (three)", got)
	}
}

// TestDefenderConditionRefusalsSayWhy — the chip's data names each
// opponent the creature can't attack and the clause they don't meet.
func TestDefenderConditionRefusalsSayWhy(t *testing.T) {
	withDefenderConditionStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horror := pushOwnedByAnother(g, me.ID, me.ID, "Chained Throatseeker", poisonedOracle)
	var got []DefenderRefusal
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		got = g.DefenderRefusalsForEffect(findBattlefieldCard(g, horror))
	})
	if len(got) != 1 || got[0].Player != opp.ID || got[0].Why != "P2 isn't poisoned" {
		t.Fatalf("refusals = %+v, want P2, not poisoned", got)
	}
}

// TestGrantedDefenderControlsRestrictionIsItsOwnAbility — Veiled Serpent's
// "it becomes a 4/4 Serpent creature with 'This creature can't attack
// unless defending player controls an Island.'" as a scoped record whose
// source is the Serpent itself. It restricts like the printed static; a
// later "loses all abilities" takes it (CR 613.1f, 613.7), an earlier one
// does not, and a restriction another permanent imposes survives the
// removal.
func TestGrantedDefenderControlsRestrictionIsItsOwnAbility(t *testing.T) {
	newTable := func(t *testing.T) (*Game, uuid.UUID, uuid.UUID) {
		g := newActiveGame(t)
		me := g.Seats[0]
		serpent := pushOwnedByAnother(g, me.ID, me.ID, "Veiled Serpent", "")
		advanceIntoStep(t, g, StepDeclareAttackers)
		return g, serpent, g.Seats[1].ID
	}
	grant := func(t *testing.T, g *Game, source, target uuid.UUID) {
		t.Helper()
		g.WithWriteLock(func() {
			if !g.RegisterScopedEffectForEffect(source, g.PinnedObjectsLocked(target),
				[]Mod{CantAttackUnlessDefenderControlsMod(PermanentQuery{Subtypes: []string{"Island"}})},
				IndefiniteDuration(), "Veiled Serpent") {
				t.Fatal("registered nothing")
			}
		})
	}

	t.Run("restricts", func(t *testing.T) {
		g, serpent, opp := newTable(t)
		grant(t, g, serpent, serpent)
		re := restrictionErr(t, g.Clone().DeclareAttacker(serpent, opp))
		if got, want := re.Sentence(), "Veiled Serpent can't attack P2: P2 controls no Island."; got != want {
			t.Errorf("sentence = %q, want %q", got, want)
		}
		pushLandForTest(g, opp, "Island", "Basic Land — Island")
		if err := g.DeclareAttacker(serpent, opp); err != nil {
			t.Errorf("declare once P2 controls an Island: %v", err)
		}
	})
	t.Run("a later removal takes it", func(t *testing.T) {
		g, serpent, opp := newTable(t)
		grant(t, g, serpent, serpent)
		registerScopedEffectForTest(t, g, serpent, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
		if err := g.DeclareAttacker(serpent, opp); err != nil {
			t.Errorf("declare after a later loses-all-abilities: %v", err)
		}
	})
	t.Run("an earlier removal does not", func(t *testing.T) {
		g, serpent, opp := newTable(t)
		registerScopedEffectForTest(t, g, serpent, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
		grant(t, g, serpent, serpent)
		restrictionErr(t, g.Clone().DeclareAttacker(serpent, opp))
	})
	t.Run("another permanent's restriction stays", func(t *testing.T) {
		g, serpent, opp := newTable(t)
		other := pushBearFor(g, g.Seats[0].ID)
		grant(t, g, other, serpent)
		registerScopedEffectForTest(t, g, serpent, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
		restrictionErr(t, g.Clone().DeclareAttacker(serpent, opp))
	})
}

// TestCantAttackUnlessDefenderControlsModNeedsAQuery — a granted
// restriction that asks for nothing would refuse every target, so
// registration refuses it.
func TestCantAttackUnlessDefenderControlsModNeedsAQuery(t *testing.T) {
	if defenderControlsModProblem(CantAttackUnlessDefenderControlsMod()) == "" {
		t.Error("a mod with no query passed the check")
	}
	if p := defenderControlsModProblem(CantAttackUnlessDefenderControlsMod(PermanentQuery{Subtypes: []string{"Island"}})); p != "" {
		t.Errorf("an Island mod failed the check: %s", p)
	}
	if p := nextFromSourceModProblem(CantAttackUnlessDefenderControlsMod(PermanentQuery{Subtypes: []string{"Island"}})); p != "" {
		t.Errorf("the next-damage check refuses the mod's queries: %s", p)
	}
}
