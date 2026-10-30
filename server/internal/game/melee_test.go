package game

import (
	"testing"

	"github.com/google/uuid"
)

// melee_test.go — the engine half of #4722: melee (CR 702.121) as a
// canonical, cumulative keyword whose trigger is derived from the
// ability list, so a printed, granted or token melee all trigger, one
// trigger per instance (CR 702.121b), counting the opponents attacked
// this combat as it resolves (CR 702.121a). The catalog half (Titania,
// Proud Pummeler) is in cards/effects/titania_proud_pummeler_test.go.

func TestMeleeIsCumulativeAndCanonical(t *testing.T) {
	if kw, ok := CanonicalKeyword("Melee"); !ok || kw != KeywordMelee {
		t.Errorf("CanonicalKeyword(\"Melee\") = %q, %v", kw, ok)
	}
	if !KeywordIsCumulative(KeywordMelee) {
		t.Error("melee is not cumulative; CR 702.121b makes each instance trigger separately")
	}
	if got := AppendKeywordAbility([]string{KeywordMelee}, KeywordMelee); len(got) != 2 {
		t.Errorf("a second melee grant was deduped: %v", got)
	}
}

func TestMeleeTriggersOncePerInstance(t *testing.T) {
	printed := Card{Name: "Brawler", TypeLine: "Creature — Human Warrior", Keywords: []string{KeywordMelee}}
	trigs := TriggersForCard(printed)
	if len(trigs) != 1 || trigs[0].Keyword != KeywordMelee {
		t.Fatalf("a printed melee with no catalog entry has triggers %+v, want one melee trigger", trigs)
	}

	// A printed instance plus a layer-6 grant (Titania) are two
	// instances and two triggers (CR 702.121b).
	eff := printed.printedCharacteristic()
	eff.Abilities = AppendKeywordAbility(eff.Abilities, KeywordMelee)
	granted := printed
	granted.effective = &eff
	if n := MeleeCount(&granted); n != 2 {
		t.Errorf("printed + granted melee = %d instances, want 2", n)
	}
	if n := len(TriggersForCard(granted)); n != 2 {
		t.Errorf("printed + granted melee = %d triggers, want 2", n)
	}

	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d melee triggers", n)
	}
	faceDown := printed
	faceDown.FaceDown, faceDown.FaceDownKind = true, FaceDownManifested
	if n := len(TriggersForCard(faceDown)); n != 0 {
		t.Errorf("a face-down creature has %d melee triggers", n)
	}
}

// meleeAttackerForTest is a 2/2 with the given printed keywords that
// has been under owner's control since the turn began.
func meleeAttackerForTest(g *Game, owner uuid.UUID, keywords ...string) uuid.UUID {
	id := attackerForTest(g, owner)
	findBattlefieldCard(g, id).Keywords = keywords
	return id
}

// declareAt declares attacker at target (a player, or a planeswalker).
func declareAt(t *testing.T, g *Game, attacker, target uuid.UUID) {
	t.Helper()
	if err := g.DeclareAttacker(attacker, target); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

// meleeItems is every melee trigger queued or on the stack.
func meleeItems(g *Game) []*StackItem {
	var out []*StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.Body == meleePumpKey {
				out = append(out, it)
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.Body == meleePumpKey {
				out = append(out, it)
			}
		}
	})
	return out
}

// resolveMeleeItems puts the queued triggers on the stack and resolves
// n of them.
func resolveMeleeItems(t *testing.T, g *Game, n int) {
	t.Helper()
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	if c := pendingChoiceOfKind(g, PendingChoiceTriggerOrder); c != nil {
		t.Fatalf("melee triggers asked to be ordered: %+v", c)
	}
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			g.resolveTopAbilityLocked()
		}
	})
}

func TestMeleeCountsEachOpponentAttackedOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	brawler := meleeAttackerForTest(g, me.ID, KeywordMelee)
	second := meleeAttackerForTest(g, me.ID)
	third := meleeAttackerForTest(g, me.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	declareAt(t, g, brawler, o1.ID)
	// Another creature, not the melee one, attacks a second opponent,
	// and a third attacks the first again: two opponents, three attackers.
	declareAt(t, g, second, o2.ID)
	declareAt(t, g, third, o1.ID)
	lockInAttackDeclaration(t, g)

	items := meleeItems(g)
	if len(items) != 1 {
		t.Fatalf("one melee instance queued %d triggers, want 1", len(items))
	}
	if it := items[0]; it.Params.Object.ID != brawler || it.Controller != me.ID || !it.Commutes || it.Label != meleeLabel {
		t.Errorf("trigger = %+v, want the melee creature pinned, my control, commuting, labelled %q", it, meleeLabel)
	}
	resolveMeleeItems(t, g, 1)
	if ch := scopedEffectChar(t, g, brawler); ch.Power != 4 || ch.Toughness != 4 {
		t.Errorf("melee creature is %d/%d, want 4/4 (+1/+1 for each of two opponents)", ch.Power, ch.Toughness)
	}
	for _, id := range []uuid.UUID{second, third} {
		if ch := scopedEffectChar(t, g, id); ch.Power != 2 || ch.Toughness != 2 {
			t.Errorf("a creature without melee is %d/%d, want 2/2", ch.Power, ch.Toughness)
		}
	}
}

func TestMeleeCountsOnlyTheOpponentsAttackedNotThePlaneswalkersTheyControl(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	pw := pushPlaneswalkerForTest(g, o1.ID, "Their Walker", 5)
	brawler := meleeAttackerForTest(g, me.ID, KeywordMelee)
	other := meleeAttackerForTest(g, me.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	declareAt(t, g, brawler, pw)
	declareAt(t, g, other, o2.ID)
	lockInAttackDeclaration(t, g)
	resolveMeleeItems(t, g, 1)
	// The planeswalker is attacked, its controller is not (CR 506.3):
	// only the player attacked directly counts.
	if ch := scopedEffectChar(t, g, brawler); ch.Power != 3 || ch.Toughness != 3 {
		t.Errorf("melee creature is %d/%d, want 3/3 (one opponent attacked, not the walker's controller)", ch.Power, ch.Toughness)
	}
}

func TestMeleeInstancesTriggerSeparately(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	// A printed melee and a second instance: two triggers, each +1/+1
	// per opponent attacked (CR 702.121b).
	brawler := meleeAttackerForTest(g, me.ID, KeywordMelee, KeywordMelee)
	other := meleeAttackerForTest(g, me.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	declareAt(t, g, brawler, o1.ID)
	declareAt(t, g, other, o2.ID)
	lockInAttackDeclaration(t, g)
	if n := len(meleeItems(g)); n != 2 {
		t.Fatalf("two melee instances queued %d triggers, want 2", n)
	}
	resolveMeleeItems(t, g, 2)
	if ch := scopedEffectChar(t, g, brawler); ch.Power != 6 || ch.Toughness != 6 {
		t.Errorf("melee creature is %d/%d, want 6/6 (two instances, two opponents each)", ch.Power, ch.Toughness)
	}
}

func TestMeleeNeedsTheCreatureToAttack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	meleeAttackerForTest(g, me.ID, KeywordMelee) // stays home
	other := meleeAttackerForTest(g, me.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	declareAt(t, g, other, o1.ID)
	lockInAttackDeclaration(t, g)
	if n := len(meleeItems(g)); n != 0 {
		t.Errorf("a melee creature that did not attack triggered %d times", n)
	}
}

func TestMeleeGivesNothingToAnObjectThatLeftAndCameBack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	brawler := meleeAttackerForTest(g, me.ID, KeywordMelee)
	advanceIntoStep(t, g, StepDeclareAttackers)
	declareAt(t, g, brawler, o1.ID)
	lockInAttackDeclaration(t, g)
	// The creature becomes a new object before the trigger resolves
	// (CR 400.7).
	g.WithWriteLock(func() { findBattlefieldCard(g, brawler).ObjectEpoch++ })
	resolveMeleeItems(t, g, 1)
	if ch := scopedEffectChar(t, g, brawler); ch.Power != 2 || ch.Toughness != 2 {
		t.Errorf("a new object got the pump: %d/%d, want 2/2", ch.Power, ch.Toughness)
	}
}

func TestOpponentsAttackedThisCombatIsZeroForAPlayerWhoDidNotDeclare(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	a := meleeAttackerForTest(g, me.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	declareAt(t, g, a, o1.ID)
	lockInAttackDeclaration(t, g)
	g.ReadSnapshot(func() {
		if n := g.OpponentsAttackedThisCombat(me.ID); n != 1 {
			t.Errorf("active player attacked %d opponents, want 1", n)
		}
		if n := g.OpponentsAttackedThisCombat(o1.ID); n != 0 {
			t.Errorf("a defender attacked %d opponents, want 0", n)
		}
		if n := g.OpponentsAttackedThisCombat(uuid.Nil); n != 0 {
			t.Errorf("no player attacked %d opponents, want 0", n)
		}
	})
}
