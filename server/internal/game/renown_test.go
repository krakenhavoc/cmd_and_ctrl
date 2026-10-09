package game

import (
	"testing"

	"github.com/google/uuid"
)

// renown_test.go — the engine half of #2049 (CR 702.112): the numbered
// token, one trigger per instance (CR 702.112c), what the trigger
// watches (combat damage to a player, and nothing else), the
// intervening if checked on both sides (CR 603.4), and the renowned
// designation (CR 702.112b) on the gate, through a zone change, a copy,
// a clone and a snapshot. The cards are in cards/effects/renown_test.go.

func TestRenownTokenIsNumberedAndCumulative(t *testing.T) {
	for in, want := range map[string]string{
		"Renown 1":  "renown 1",
		"renown 02": "renown 2",
		" Renown 6": "renown 6",
	} {
		kws, ok := CanonicalKeywords(in)
		if !ok || len(kws) != 1 || kws[0] != want {
			t.Errorf("CanonicalKeywords(%q) = %v, %v; want [%s]", in, kws, ok, want)
		}
	}
	for _, bad := range []string{"Renown", "renown 0", "renown two", "renown  2", "renown 2 1", "renown -1", "renowned"} {
		if kws, ok := CanonicalKeywords(bad); ok {
			t.Errorf("CanonicalKeywords(%q) = %v, want refused", bad, kws)
		}
	}
	if !KeywordIsCumulative("renown 1") {
		t.Error("renown is not cumulative; CR 702.112c says each instance triggers separately")
	}
	got := AppendKeywordAbility([]string{"renown 1"}, "renown 1")
	if len(got) != 2 {
		t.Errorf("a granted renown instance was deduped: %v", got)
	}
	if _, ok := AnnihilatorValue("renown 3"); ok {
		t.Error("a renown token parsed as annihilator")
	}
}

func TestRenownTriggersOncePerInstance(t *testing.T) {
	printed := Card{Name: "Topan Freeblade", TypeLine: "Creature — Human Soldier", Keywords: []string{"vigilance", "renown 1"}}
	trigs := TriggersForCard(printed)
	if len(trigs) != 1 || trigs[0].Keyword != KeywordRenown {
		t.Fatalf("a printed renown 1 with no catalog entry has triggers %+v, want one renown trigger", trigs)
	}

	// A printed renown 2 plus a layer-6 grant of renown 1 (Aragorn,
	// Hornburg Hero) is two instances and two triggers.
	two := Card{Name: "Castellan", TypeLine: "Creature — Human Knight", Keywords: []string{"renown 2"}}
	eff := two.printedCharacteristic()
	eff.Abilities = AppendKeywordAbility(eff.Abilities, "renown 1")
	two.effective = &eff
	if got := RenownAmounts(&two); len(got) != 2 || got[0] != 2 || got[1] != 1 {
		t.Errorf("RenownAmounts = %v, want [2 1]", got)
	}
	if n := len(TriggersForCard(two)); n != 2 {
		t.Errorf("printed + granted renown = %d triggers, want 2", n)
	}

	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d renown triggers", n)
	}
}

// renownItems is every renown trigger waiting on the stack for the
// creature.
func renownItems(g *Game, creature uuid.UUID) []*StackItem {
	var out []*StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.SourceCardID == creature && it.Body == renownGrowKey {
				out = append(out, it)
			}
		}
	})
	return out
}

// renownedEvents is the Amount of every EventBecameRenowned for id.
func renownedEvents(g *Game, id uuid.UUID) []int {
	var out []int
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventBecameRenowned && ev.CardID == id {
				out = append(out, ev.Amount)
			}
		}
	})
	return out
}

// renownState reads the creature's designation and +1/+1 counters.
func renownState(g *Game, id uuid.UUID) (renowned bool, counters int) {
	g.ReadSnapshot(func() {
		if c := findBattlefieldCard(g, id); c != nil {
			renowned, counters = c.Renowned, c.Counters[CounterPlusOne]
		}
	})
	return renowned, counters
}

// attackUnblockedIntoDamage declares `attacker` at `target`, blocks
// nothing and walks into the combat damage step, with any trigger the
// damage caused put on the stack.
func attackUnblockedIntoDamage(t *testing.T, g *Game, attacker, target uuid.UUID) {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, target); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
}

// CR 702.112a end to end: an unblocked renown 2 creature deals combat
// damage to a player, its trigger resolves, and it has two +1/+1
// counters and is renowned. EventBecameRenowned carries N.
func TestRenownTriggersOnCombatDamageToAPlayer(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushDamageStepCreature(g, me, "Citadel Castellan", 2, 3, "vigilance", "renown 2")
	attackUnblockedIntoDamage(t, g, knight, opp.ID)
	if n := len(renownItems(g, knight)); n != 1 {
		t.Fatalf("%d renown triggers on the stack, want 1", n)
	}
	passUntilStackEmpty(t, g)
	renowned, counters := renownState(g, knight)
	if !renowned || counters != 2 {
		t.Errorf("after renown 2 resolved: renowned %v, %d counters; want renowned with 2", renowned, counters)
	}
	if got := renownedEvents(g, knight); len(got) != 1 || got[0] != 2 {
		t.Errorf("EventBecameRenowned amounts = %v, want [2]", got)
	}
}

// Combat damage to a planeswalker is not damage to a player.
func TestRenownDoesNotTriggerOnDamageToAPlaneswalker(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushDamageStepCreature(g, me, "Topan Freeblade", 2, 2, "renown 1")
	walker := pushPlaneswalkerForTest(g, opp.ID, "Test Walker", 5)
	attackUnblockedIntoDamage(t, g, knight, walker)
	if n := len(renownItems(g, knight)); n != 0 {
		t.Errorf("combat damage to a planeswalker put %d renown triggers on the stack, want 0", n)
	}
}

// Combat damage to a blocking creature is not damage to a player.
func TestRenownDoesNotTriggerOnDamageToACreature(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushDamageStepCreature(g, me, "Topan Freeblade", 2, 2, "renown 1")
	wall := pushDamageStepCreature(g, opp, "Wall", 0, 5)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(knight, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, knight); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	if n := len(renownItems(g, knight)); n != 0 {
		t.Errorf("combat damage to a blocker put %d renown triggers on the stack, want 0", n)
	}
}

// Noncombat damage to a player is not combat damage.
func TestRenownDoesNotTriggerOnNoncombatDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushDamageStepCreature(g, me, "Topan Freeblade", 2, 2, "renown 1")
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Source: knight, Target: opp.ID, Amount: 2})
		g.drainPendingTriggersAPNAPLocked()
	})
	if n := len(renownItems(g, knight)); n != 0 {
		t.Errorf("noncombat damage put %d renown triggers on the stack, want 0", n)
	}
}

// CR 603.4, trigger side: a renowned creature's renown does not
// trigger at all.
func TestRenownDoesNotTriggerWhenAlreadyRenowned(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushDamageStepCreature(g, me, "Topan Freeblade", 2, 2, "renown 1")
	g.WithWriteLock(func() { findBattlefieldCard(g, knight).Renowned = true })
	attackUnblockedIntoDamage(t, g, knight, opp.ID)
	if n := len(renownItems(g, knight)); n != 0 {
		t.Errorf("a renowned creature put %d renown triggers on the stack, want 0", n)
	}
}

// CR 603.4, resolution side: a creature that became renowned while its
// trigger waited gets nothing from it.
func TestRenownRechecksOnResolution(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushDamageStepCreature(g, me, "Topan Freeblade", 2, 2, "renown 1")
	attackUnblockedIntoDamage(t, g, knight, opp.ID)
	if n := len(renownItems(g, knight)); n != 1 {
		t.Fatalf("%d renown triggers on the stack, want 1", n)
	}
	g.WithWriteLock(func() { findBattlefieldCard(g, knight).Renowned = true })
	passUntilStackEmpty(t, g)
	if _, counters := renownState(g, knight); counters != 0 {
		t.Errorf("the trigger put %d counters on a creature that was already renowned, want 0", counters)
	}
	if got := renownedEvents(g, knight); len(got) != 0 {
		t.Errorf("EventBecameRenowned fired %v for a creature that was already renowned", got)
	}
}

// CR 702.112c: two instances both trigger; the first to resolve makes
// the creature renowned and the second does nothing. The creature has
// the first's counters, and one EventBecameRenowned.
func TestTwoRenownInstancesTriggerAndOnlyTheFirstResolves(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushDamageStepCreature(g, me, "Doubly Renowned", 2, 2, "renown 1", "renown 1")
	attackUnblockedIntoDamage(t, g, knight, opp.ID)
	if n := len(renownItems(g, knight)); n != 2 {
		t.Fatalf("%d renown triggers on the stack, want 2 (CR 702.112c)", n)
	}
	passUntilStackEmpty(t, g)
	renowned, counters := renownState(g, knight)
	if !renowned || counters != 1 {
		t.Errorf("after both resolved: renowned %v, %d counters; want renowned with 1", renowned, counters)
	}
	if got := renownedEvents(g, knight); len(got) != 1 {
		t.Errorf("EventBecameRenowned fired %d times, want 1", len(got))
	}
}

// A creature that leaves the battlefield before its trigger resolves
// gets nothing, and the new object is not renowned (CR 400.7).
func TestRenownSourceThatLeftGetsNothing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushDamageStepCreature(g, me, "Topan Freeblade", 2, 2, "renown 1")
	attackUnblockedIntoDamage(t, g, knight, opp.ID)
	var newID uuid.UUID
	g.WithWriteLock(func() { newID = g.resetAsNewObjectLocked(knight) })
	passUntilStackEmpty(t, g)
	if renowned, counters := renownState(g, newID); renowned || counters != 0 {
		t.Errorf("the new object: renowned %v, %d counters; want neither", renowned, counters)
	}
}

// CR 702.112b: the designation is kept until the permanent leaves the
// battlefield, and is gone when it does — at both CR 400.7 sites.
func TestRenownedIsLostWhenThePermanentLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	knight := pushTypedTestCard(g, Card{
		Name: "Topan Freeblade", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Renowned: true, Keywords: []string{"renown 1"},
	})
	if !RenownedGate().Active(*findBattlefieldCard(g, knight)) {
		t.Fatal("the renowned gate is off on a renowned permanent")
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneHand, Owner: me.ID}, knight); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	var inHand Card
	g.ReadSnapshot(func() {
		for _, c := range me.Hand.Cards {
			if c.Name == "Topan Freeblade" {
				inHand = c
			}
		}
	})
	if inHand.Renowned {
		t.Error("a bounced renowned creature is still renowned in its owner's hand")
	}

	other := pushTypedTestCard(g, Card{
		Name: "Other", TypeLine: "Creature — Test", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Renowned: true,
	})
	var newID uuid.UUID
	g.WithWriteLock(func() { newID = g.resetAsNewObjectLocked(other) })
	var renowned bool
	g.ReadSnapshot(func() { renowned = g.IsRenowned(newID) })
	if renowned {
		t.Error("the new-object reset left the designation on")
	}
}

// CR 702.112b: renowned is not a copiable value, and it rides the
// clone (undo) and the snapshot.
func TestRenownedIsNotCopiedAndSurvivesCloneAndSnapshot(t *testing.T) {
	source := Card{InstanceID: uuid.New(), Name: "Topan Freeblade", TypeLine: "Creature — Human Soldier", Renowned: true}
	clone := Card{InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter"}
	clone.applyCopy(CopiableValuesOf(source), source)
	if clone.Renowned {
		t.Error("a copy of a renowned creature is renowned; CR 702.112b says the designation is not copiable")
	}

	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTypedTestCard(g, Card{
		Name: "Topan Freeblade", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Renowned: true,
	})
	find := func(cards []Card) Card {
		for _, c := range cards {
			if c.InstanceID == id {
				return c
			}
		}
		return Card{}
	}
	if !find(g.Clone().Battlefield.Cards).Renowned {
		t.Error("the clone lost the designation")
	}
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	var back Card
	restored.ReadSnapshot(func() { back = find(restored.Battlefield.Cards) })
	if !back.Renowned {
		t.Error("the snapshot round-trip lost the designation")
	}
}
