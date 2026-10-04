package game

import (
	"testing"

	"github.com/google/uuid"
)

// annihilator_test.go — the engine half of #2073 (ADR 0113 §2): the
// numbered token, one trigger per instance (CR 702.86b), and the two
// CR 508.5 cases the catalog package cannot reach — an attacker removed
// from combat, and one that has left the battlefield, before its
// trigger resolves. The rest of the rule end to end (grants, several
// attackers, planeswalkers, battles, a defender who has lost, the
// titans) is in cards/effects/annihilator_test.go.

func TestAnnihilatorTokenIsNumberedAndCumulative(t *testing.T) {
	for in, want := range map[string]string{
		"Annihilator 4":  "annihilator 4",
		"annihilator 02": "annihilator 2",
		" Annihilator 1": "annihilator 1",
	} {
		kws, ok := CanonicalKeywords(in)
		if !ok || len(kws) != 1 || kws[0] != want {
			t.Errorf("CanonicalKeywords(%q) = %v, %v; want [%s]", in, kws, ok, want)
		}
	}
	for _, bad := range []string{"Annihilator", "annihilator 0", "annihilator two", "annihilator  2", "annihilator 2 1", "annihilator -1"} {
		if kws, ok := CanonicalKeywords(bad); ok {
			t.Errorf("CanonicalKeywords(%q) = %v, want refused", bad, kws)
		}
	}
	if !KeywordIsCumulative("annihilator 2") {
		t.Error("annihilator is not cumulative; CR 702.86b says each instance triggers separately")
	}
	got := AppendKeywordAbility([]string{"annihilator 4"}, "annihilator 2")
	got = AppendKeywordAbility(got, "annihilator 2")
	if len(got) != 3 {
		t.Errorf("granted annihilator instances were deduped: %v", got)
	}
	// Toxic shares the parser and must still read the same.
	if n, ok := ToxicValue("Toxic 3"); !ok || n != 3 {
		t.Errorf("ToxicValue(\"Toxic 3\") = %d, %v", n, ok)
	}
	if _, ok := ToxicValue("annihilator 3"); ok {
		t.Error("an annihilator token parsed as toxic")
	}
}

func TestAnnihilatorTriggersOncePerInstance(t *testing.T) {
	printed := Card{Name: "Pathrazer", TypeLine: "Creature — Eldrazi", Keywords: []string{"annihilator 3"}}
	trigs := TriggersForCard(printed)
	if len(trigs) != 1 || trigs[0].Keyword != KeywordAnnihilator {
		t.Fatalf("a printed annihilator 3 with no catalog entry has triggers %+v, want one annihilator trigger", trigs)
	}

	// Printed 4 plus a layer-6 grant of 2 (Eldrazi Conscription) is two
	// instances and two triggers, each with its own number.
	big := Card{Name: "Titan", TypeLine: "Creature — Eldrazi", Keywords: []string{"annihilator 4"}}
	eff := big.printedCharacteristic()
	eff.Abilities = AppendKeywordAbility(eff.Abilities, "annihilator 2")
	big.effective = &eff
	if got := AnnihilatorAmounts(&big); len(got) != 2 || got[0] != 4 || got[1] != 2 {
		t.Errorf("AnnihilatorAmounts = %v, want [4 2]", got)
	}
	if n := len(TriggersForCard(big)); n != 2 {
		t.Errorf("printed + granted annihilator = %d triggers, want 2", n)
	}

	// Losing all abilities, or being face down, leaves none.
	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d annihilator triggers", n)
	}
	faceDown := printed
	faceDown.FaceDown, faceDown.FaceDownKind = true, FaceDownManifested
	if n := len(TriggersForCard(faceDown)); n != 0 {
		t.Errorf("a face-down creature has %d annihilator triggers", n)
	}
}

// pushAnnihilatorForTest puts a ready creature with the given keyword
// tokens printed on it onto the battlefield.
func pushAnnihilatorForTest(g *Game, owner uuid.UUID, keywords ...string) uuid.UUID {
	c := NewCard("Eldrazi", owner)
	c.TypeLine = "Creature — Eldrazi"
	c.Power, c.Toughness = 5, 5
	c.Keywords = keywords
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// pushFodderForTest puts n plain permanents under owner's control.
func pushFodderForTest(g *Game, owner uuid.UUID, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		c := NewCard("Fodder", owner)
		c.TypeLine = "Artifact"
		g.Battlefield.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}

// attackWithAnnihilator declares `attacker` at `defender`, locks the
// declaration in and puts the trigger on the stack.
func attackWithAnnihilator(t *testing.T, g *Game, attacker, defender uuid.UUID) *StackItem {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == attacker && it.Body == annihilatorSacrificeKey {
			return it
		}
	}
	t.Fatalf("no annihilator trigger on the stack: %+v", g.StackMeta)
	return nil
}

// TestAnnihilatorAfterTheAttackerIsRemovedFromCombat — CR 508.5: once
// the creature is no longer attacking, the defending player is the one
// it was attacking before it was removed from combat.
func TestAnnihilatorAfterTheAttackerIsRemovedFromCombat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushAnnihilatorForTest(g, me.ID, "annihilator 2")
	fodder := pushFodderForTest(g, opp.ID, 3)
	item := attackWithAnnihilator(t, g, attacker, opp.ID)
	if item.Params.Player != opp.ID || item.Params.Amount != 2 {
		t.Fatalf("the trigger recorded %+v, want the defender and N = 2", item.Params)
	}

	g.WithWriteLock(func() { g.removeFromCombatLocked(findBattlefieldCard(g, attacker)) })
	passUntilPrompt(t, g)
	c := pendingChoiceOfKind(g, PendingChoiceOwnPermanents)
	if c == nil || c.Chooser != opp.ID || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("prompt = %+v, want the old defender choosing two", c)
	}
	if err := g.ResolveOwnPermanents(c.ID, opp.ID, fodder[:2]); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	for _, id := range fodder[:2] {
		if findBattlefieldCard(g, id) != nil {
			t.Errorf("%s was not sacrificed", id)
		}
	}
}

// TestAnnihilatorAfterTheAttackerLeavesTheBattlefield — the ability has
// no "if" clause, so it resolves without its source, against the
// player the creature was attacking.
func TestAnnihilatorAfterTheAttackerLeavesTheBattlefield(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushAnnihilatorForTest(g, me.ID, "annihilator 1")
	pushFodderForTest(g, opp.ID, 2)
	attackWithAnnihilator(t, g, attacker, opp.ID)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(attacker); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passUntilPrompt(t, g)
	c := pendingChoiceOfKind(g, PendingChoiceOwnPermanents)
	if c == nil || c.Chooser != opp.ID || c.ChooseMax != 1 {
		t.Fatalf("prompt = %+v, want the defender choosing one", c)
	}
}
