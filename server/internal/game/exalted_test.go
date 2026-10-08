package game

import (
	"testing"

	"github.com/google/uuid"
)

// exalted_test.go — the engine half of #2538 (ADR 0101 amendment
// 2026-10-08): the canonical token, one trigger per instance (CR
// 113.2c), one instance per exalted counter (the Emissary of Soulfire
// ruling of 2024-06-07), "attacks alone" (CR 506.5), and the restore
// alias for a Hierarch trigger written before the change (owner answer
// 3). The catalog half (the Hierarchs, grants, Emissary of Soulfire) is
// in cards/effects/exalted_test.go.

func TestExaltedIsCumulativeAndCanonical(t *testing.T) {
	if kw, ok := CanonicalKeyword("Exalted"); !ok || kw != KeywordExalted {
		t.Errorf("CanonicalKeyword(\"Exalted\") = %q, %v", kw, ok)
	}
	if !KeywordIsCumulative(KeywordExalted) {
		t.Error("exalted is not cumulative; CR 113.2c makes each instance its own trigger")
	}
	if got := AppendKeywordAbility([]string{KeywordExalted}, KeywordExalted); len(got) != 2 {
		t.Errorf("a second exalted grant was deduped: %v", got)
	}
	if !IsKeywordCounter(CounterExalted) || CounterExalted != KeywordExalted {
		t.Error("the exalted counter is not a keyword counter for the exalted token")
	}
}

func TestExaltedTriggersOncePerInstance(t *testing.T) {
	printed := Card{Name: "Hierarch", TypeLine: "Creature — Human Druid", Keywords: []string{KeywordExalted}}
	trigs := TriggersForCard(printed)
	if len(trigs) != 1 || trigs[0].Keyword != KeywordExalted {
		t.Fatalf("a printed exalted with no catalog entry has triggers %+v, want one exalted trigger", trigs)
	}

	// A printed instance plus a layer-6 grant (Sublime Archangel) are
	// two instances and two triggers.
	eff := printed.printedCharacteristic()
	eff.Abilities = AppendKeywordAbility(eff.Abilities, KeywordExalted)
	granted := printed
	granted.effective = &eff
	if n := ExaltedCount(&granted); n != 2 {
		t.Errorf("printed + granted exalted = %d instances, want 2", n)
	}
	if n := len(TriggersForCard(granted)); n != 2 {
		t.Errorf("printed + granted exalted = %d triggers, want 2", n)
	}

	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d exalted triggers", n)
	}
	faceDown := printed
	faceDown.FaceDown, faceDown.FaceDownKind = true, FaceDownManifested
	if n := len(TriggersForCard(faceDown)); n != 0 {
		t.Errorf("a face-down creature has %d exalted triggers", n)
	}
}

// The Emissary of Soulfire ruling: "A creature with multiple exalted
// counters will have that many instances of exalted." Other keyword
// counters stay one keyword however many there are.
func TestEachExaltedCounterIsOneInstance(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 1, 4)
	addCounterForTest(t, g, id, CounterExalted, 1)
	addCounterForTest(t, g, id, CounterExalted, 1)
	addCounterForTest(t, g, id, CounterFlying, 2)
	ch := scopedEffectChar(t, g, id)
	if n := countOf(ch.Abilities, KeywordExalted); n != 2 {
		t.Errorf("two exalted counters give %d instances, want 2 (%v)", n, ch.Abilities)
	}
	if n := countOf(ch.Abilities, "flying"); n != 1 {
		t.Errorf("two flying counters give %d flying, want 1", n)
	}
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		if n := len(TriggersForCard(*c)); n != 2 {
			t.Errorf("two exalted counters are %d triggers, want 2", n)
		}
	})

	// One counter on a creature that prints exalted is two instances.
	printed := NewCard("Hierarch", g.Seats[0].ID)
	printed.TypeLine = "Creature — Human Druid"
	printed.Keywords = []string{KeywordExalted}
	g.Battlefield.PushTop(printed)
	addCounterForTest(t, g, printed.InstanceID, CounterExalted, 1)
	if n := countOf(scopedEffectChar(t, g, printed.InstanceID).Abilities, KeywordExalted); n != 2 {
		t.Errorf("a printed exalted and an exalted counter are %d instances, want 2", n)
	}

	// Off the battlefield the counters' own walk agrees with the layer
	// pass, so the badge and the rule match.
	card := Card{Counters: map[string]int{CounterExalted: 3, CounterFlying: 2}}
	if got := KeywordCounterTokens(card); countOf(got, KeywordExalted) != 3 || countOf(got, "flying") != 1 {
		t.Errorf("KeywordCounterTokens = %v, want three exalted and one flying", got)
	}
}

// exaltedSourceForTest puts a 0/1 with the given printed keywords
// under owner's control.
func exaltedSourceForTest(g *Game, owner uuid.UUID, keywords ...string) uuid.UUID {
	c := NewCard("Exalted Probe", owner)
	c.TypeLine = "Creature — Human Druid"
	c.Power, c.Toughness = 0, 1
	c.Keywords = keywords
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// attackerForTest is a 2/2 that has been under owner's control since
// the turn began, so it may attack.
func attackerForTest(g *Game, owner uuid.UUID) uuid.UUID {
	id := pushScopedTestCreature(g, owner, 2, 2)
	findBattlefieldCard(g, id).SummonedThisTurn = false
	return id
}

// declareAndLock declares each attacker at defender and locks the
// declaration in, leaving the triggers queued.
func declareAndLock(t *testing.T, g *Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttackDeclaration(t, g)
}

// exaltedItems is every exalted trigger queued or on the stack.
func exaltedItems(g *Game) []*StackItem {
	var out []*StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.Body == exaltedPumpKey {
				out = append(out, it)
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.Body == exaltedPumpKey {
				out = append(out, it)
			}
		}
	})
	return out
}

func TestExaltedPumpsTheLoneAttackerOncePerInstance(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exaltedSourceForTest(g, me.ID, KeywordExalted, KeywordExalted)
	attacker := attackerForTest(g, me.ID)
	declareAndLock(t, g, opp.ID, attacker)

	items := exaltedItems(g)
	if len(items) != 2 {
		t.Fatalf("two exalted instances queued %d triggers, want 2", len(items))
	}
	for _, it := range items {
		if it.Params.Object.ID != attacker || it.Controller != me.ID || !it.Commutes || it.Label != exaltedLabel {
			t.Errorf("trigger = %+v, want the attacker pinned, my control, commuting, labelled %q", it, exaltedLabel)
		}
	}
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	if c := pendingChoiceOfKind(g, PendingChoiceTriggerOrder); c != nil {
		t.Fatalf("exalted triggers asked to be ordered: %+v", c)
	}
	g.WithWriteLock(func() {
		g.resolveTopAbilityLocked()
		g.resolveTopAbilityLocked()
	})
	if ch := scopedEffectChar(t, g, attacker); ch.Power != 4 || ch.Toughness != 4 {
		t.Errorf("attacker is %d/%d, want 4/4 (two +1/+1s)", ch.Power, ch.Toughness)
	}
}

func TestExaltedNeedsALoneAttackerYouControl(t *testing.T) {
	// Two attackers: neither is alone (CR 506.5).
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exaltedSourceForTest(g, me.ID, KeywordExalted)
	a := attackerForTest(g, me.ID)
	b := attackerForTest(g, me.ID)
	declareAndLock(t, g, opp.ID, a, b)
	if n := len(exaltedItems(g)); n != 0 {
		t.Errorf("two attackers triggered exalted %d times", n)
	}

	// An opponent's exalted does not trigger for my lone attacker.
	g = newActiveGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	exaltedSourceForTest(g, opp.ID, KeywordExalted)
	lone := attackerForTest(g, me.ID)
	declareAndLock(t, g, opp.ID, lone)
	if n := len(exaltedItems(g)); n != 0 {
		t.Errorf("an opponent's exalted triggered %d times for my attacker", n)
	}
}

// CR 400.7: an attacker that left and came back is a new object and
// gets nothing.
func TestExaltedSkipsANewObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exaltedSourceForTest(g, me.ID, KeywordExalted)
	attacker := attackerForTest(g, me.ID)
	declareAndLock(t, g, opp.ID, attacker)
	g.WithWriteLock(func() {
		g.drainPendingTriggersAPNAPLocked()
		findBattlefieldCard(g, attacker).ObjectEpoch++
		g.resolveTopAbilityLocked()
	})
	if ch := scopedEffectChar(t, g, attacker); ch.Power != 2 {
		t.Errorf("a new object got the pump: power %d, want 2", ch.Power)
	}
}

const retiredExaltedOracle = "fixture-retired-exalted"

// oldExaltedRow is effects.Exalted() as it was before #2538: a catalog
// row whose declared name is the exalted label.
func oldExaltedRow() TriggeredAbility {
	return TriggeredAbility{
		Watches: []EventKind{EventAttack},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
			return ev.Actor == source.Controller && AttackedAlone(g)
		},
		Key: exaltedLabel,
		Effect: func(g *Game, item *StackItem) error {
			return nil
		},
	}
}

// TestRetiredExaltedRowRestoresAsTheKeyword — owner answer 3: a
// Hierarch's exalted trigger captured by a binary from before #2538
// restores as the keyword's own body, resolves, and leaves the Hierarch
// unflagged; neither boot report lists it.
func TestRetiredExaltedRowRestoresAsTheKeyword(t *testing.T) {
	old := &CardDef{Triggered: []TriggeredAbility{oldExaltedRow()}}
	IdentifyCatalogRows(retiredExaltedOracle, old)
	withCatalog(t, retiredExaltedOracle, old)

	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hierarch := pushTypedTestCard(g, Card{
		Name: "Noble Hierarch", TypeLine: "Creature — Human Druid", OracleID: retiredExaltedOracle,
		Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	attacker := attackerForTest(g, me.ID)
	declareAndLock(t, g, opp.ID, attacker)
	var itemID uuid.UUID
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it.SourceCardID == hierarch && it.Body == CatalogTriggeredBodyKey {
				itemID = it.ID
			}
		}
	})
	if itemID == uuid.Nil {
		t.Fatal("the old exalted row queued no catalog trigger")
	}
	snap := throughJSON(t, g.CaptureSnapshot())

	// The deploy: the row is gone and the card prints the keyword.
	withCatalog(t, retiredExaltedOracle, &CardDef{PrintedKeywords: []string{KeywordExalted}})
	if lost := snap.LostStackAbilities(); len(lost) != 0 {
		t.Errorf("LostStackAbilities = %+v, want none: the alias restores it", lost)
	}
	if short := snap.AbilityShortfalls(); len(short) != 0 {
		t.Errorf("AbilityShortfalls = %+v, want none: the row became the keyword", short)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	back := restoredPendingTrigger(t, restored)
	if back.ID != itemID || back.Body != exaltedPumpKey || back.Params.Ability != nil ||
		back.Params.Object.ID != attacker || back.Effect == nil {
		t.Fatalf("the old trigger came back as %+v, want the exalted body pinned to the attacker", back)
	}
	restored.ReadSnapshot(func() {
		if c := restored.findCardByIDLocked(hierarch); c == nil || c.AbilitiesLostOnRestore {
			t.Error("the Hierarch is flagged AbilitiesLostOnRestore")
		}
	})
	resolveQueuedTrigger(restored)
	if ch := scopedEffectChar(t, restored, attacker); ch.Power != 3 || ch.Toughness != 3 {
		t.Errorf("attacker is %d/%d after the restored trigger, want 3/3", ch.Power, ch.Toughness)
	}

	// Any other lost row still restores as a manual item.
	other := &StackItem{Body: CatalogTriggeredBodyKey, Params: EffectParams{Ability: &AbilityRef{
		Key: retiredExaltedOracle, Slot: AbilitySlotTriggered, Ref: "own:0", Name: "Something else",
	}}, Trigger: back.Trigger}
	if restoreRetiredExaltedRow(other) {
		t.Error("the alias took a row that is not the retired exalted row")
	}
}
