package game

import (
	"testing"

	"github.com/google/uuid"
)

// rampage_test.go — the engine half of #5769: rampage N (CR 702.23) as
// a numbered, cumulative keyword whose trigger is derived from the
// ability list, so a printed or granted rampage triggers, one trigger
// per instance (CR 702.23c), counting the creatures blocking it beyond
// the first once, as it resolves (CR 702.23b). The catalog half (the
// Alliances and Legends cards) is in cards/effects/rampage_cards_test.go.

func TestRampageTokenIsNumberedCumulativeAndCanonical(t *testing.T) {
	for in, want := range map[string]string{"Rampage 2": "rampage 2", "rampage 01": "rampage 1", " Rampage 4 ": "rampage 4"} {
		if kw, ok := CanonicalKeyword(in); !ok || kw != want {
			t.Errorf("CanonicalKeyword(%q) = %q, %v; want %q", in, kw, ok, want)
		}
	}
	for _, in := range []string{"Rampage", "rampage 0", "rampage two", "rampage 2 and more"} {
		if kw, ok := CanonicalKeyword(in); ok {
			t.Errorf("CanonicalKeyword(%q) = %q; a rampage with no number is not a token", in, kw)
		}
	}
	if !KeywordIsCumulative("rampage 2") {
		t.Error("rampage is not cumulative; CR 702.23c makes each instance trigger separately")
	}
	if got := AppendKeywordAbility([]string{"rampage 2"}, "rampage 2"); len(got) != 2 {
		t.Errorf("a second rampage grant was deduped: %v", got)
	}
	if !lendableTriggeredKeyword("rampage 3") {
		t.Error("a borrow does not lend rampage as its token")
	}
}

func TestRampageTriggersOncePerInstance(t *testing.T) {
	printed := Card{Name: "Rampager", TypeLine: "Creature — Giant", Keywords: []string{"rampage 2"}}
	trigs := TriggersForCard(printed)
	if len(trigs) != 1 || trigs[0].Keyword != KeywordRampage {
		t.Fatalf("a printed rampage with no catalog entry has triggers %+v, want one rampage trigger", trigs)
	}

	// A printed instance plus a layer-6 grant of a different N are two
	// instances and two triggers (CR 702.23c).
	eff := printed.printedCharacteristic()
	eff.Abilities = AppendKeywordAbility(eff.Abilities, "rampage 1")
	granted := printed
	granted.effective = &eff
	if got := RampageAmounts(&granted); len(got) != 2 || got[0] != 2 || got[1] != 1 {
		t.Errorf("printed + granted rampage amounts = %v, want [2 1]", got)
	}
	if n := len(TriggersForCard(granted)); n != 2 {
		t.Errorf("printed + granted rampage = %d triggers, want 2", n)
	}

	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d rampage triggers", n)
	}
}

// rampageItems is every rampage trigger queued or on the stack.
func rampageItems(g *Game) []*StackItem {
	var out []*StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.Body == rampagePumpKey {
				out = append(out, it)
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.Body == rampagePumpKey {
				out = append(out, it)
			}
		}
	})
	return out
}

// rampageBlocked has a 2/2 with the given keywords attack seat 1 and be
// blocked by `blockers` 1/1s, then locks the blocks in, leaving the
// triggers queued.
func rampageBlocked(t *testing.T, g *Game, blockers int, keywords ...string) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	attacker := pushCombatant(t, g, g.Seats[0], "Rampager", 2, 2, keywords...)
	var bs []uuid.UUID
	for i := range blockers {
		bs = append(bs, pushCombatant(t, g, g.Seats[1], "Blocker "+string(rune('A'+i)), 1, 1))
	}
	declareAttacks(t, g, attacker)
	for _, b := range bs {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	lockInBlockDeclaration(t, g)
	return attacker, bs
}

// resolveRampageItems puts the queued triggers on the stack, checks
// that no ordering prompt was asked, and resolves n of them.
func resolveRampageItems(t *testing.T, g *Game, n int) {
	t.Helper()
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	if c := pendingChoiceOfKind(g, PendingChoiceTriggerOrder); c != nil {
		t.Fatalf("rampage triggers asked to be ordered: %+v", c)
	}
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			g.resolveTopAbilityLocked()
		}
	})
}

// CR 702.23a: +N/+N for each blocker beyond the first. Unblocked, it
// does not trigger; one blocker triggers for +0/+0; three are +2N/+2N.
func TestRampageCountsBlockersBeyondTheFirst(t *testing.T) {
	for _, tc := range []struct {
		blockers, triggers, power, toughness int
	}{
		{0, 0, 2, 2},
		{1, 1, 2, 2},
		{3, 1, 6, 6},
	} {
		g := newActiveGame(t)
		attacker, _ := rampageBlocked(t, g, tc.blockers, "rampage 2")
		items := rampageItems(g)
		if len(items) != tc.triggers {
			t.Fatalf("%d blockers queued %d rampage triggers, want %d", tc.blockers, len(items), tc.triggers)
		}
		if tc.triggers > 0 {
			it := items[0]
			if it.Params.Object.ID != attacker || it.Params.Amount != 2 || it.Controller != g.Seats[0].ID || !it.Commutes {
				t.Errorf("trigger = %+v, want the attacker pinned, N 2, its controller's, commuting", it)
			}
			resolveRampageItems(t, g, tc.triggers)
		}
		if ch := scopedEffectChar(t, g, attacker); ch.Power != tc.power || ch.Toughness != tc.toughness {
			t.Errorf("%d blockers: the rampage 2 creature is %d/%d, want %d/%d", tc.blockers, ch.Power, ch.Toughness, tc.power, tc.toughness)
		}
	}
}

// CR 702.23b: the count is taken as the trigger resolves. A blocker
// that left the battlefield, or was removed from combat, before then
// does not count, and one that leaves afterwards changes nothing.
func TestRampageCountsAsItResolves(t *testing.T) {
	g := newActiveGame(t)
	attacker, bs := rampageBlocked(t, g, 3, "rampage 2")
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bs[0]); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if findBattlefieldCard(g, bs[0]) != nil {
		t.Fatal("the destroyed blocker is still on the battlefield")
	}
	resolveRampageItems(t, g, 1)
	// Two blockers left: one beyond the first, +2/+2.
	if ch := scopedEffectChar(t, g, attacker); ch.Power != 4 || ch.Toughness != 4 {
		t.Errorf("after a blocker died in response the creature is %d/%d, want 4/4", ch.Power, ch.Toughness)
	}
	// Taking another blocker out of combat now changes nothing.
	g.WithWriteLock(func() { g.removeFromCombatLocked(findBattlefieldCard(g, bs[1])) })
	if ch := scopedEffectChar(t, g, attacker); ch.Power != 4 || ch.Toughness != 4 {
		t.Errorf("a blocker removed after resolution changed the bonus: %d/%d, want 4/4", ch.Power, ch.Toughness)
	}

	// Removed from combat (not from the battlefield) before resolution:
	// two of three blockers gone is one blocker, and +0/+0.
	g = newActiveGame(t)
	attacker, bs = rampageBlocked(t, g, 3, "rampage 2")
	g.WithWriteLock(func() {
		g.drainPendingTriggersAPNAPLocked()
		g.removeFromCombatLocked(findBattlefieldCard(g, bs[0]))
		g.removeFromCombatLocked(findBattlefieldCard(g, bs[1]))
	})
	resolveRampageItems(t, g, 1)
	if ch := scopedEffectChar(t, g, attacker); ch.Power != 2 || ch.Toughness != 2 {
		t.Errorf("with one blocker left the creature is %d/%d, want 2/2", ch.Power, ch.Toughness)
	}
}

// CR 702.23c: two instances trigger separately, each with its own N.
func TestRampageTwoInstancesTriggerSeparately(t *testing.T) {
	g := newActiveGame(t)
	attacker, _ := rampageBlocked(t, g, 3, "rampage 2", "rampage 1")
	items := rampageItems(g)
	if len(items) != 2 {
		t.Fatalf("two rampage instances queued %d triggers, want 2", len(items))
	}
	resolveRampageItems(t, g, 2)
	// Two blockers beyond the first: +4/+4 and +2/+2.
	if ch := scopedEffectChar(t, g, attacker); ch.Power != 8 || ch.Toughness != 8 {
		t.Errorf("rampage 2 and rampage 1 with three blockers: %d/%d, want 8/8", ch.Power, ch.Toughness)
	}
}

// A rampage given by an effect triggers like a printed one.
func TestRampageGrantedTriggers(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	registerScopedEffectForTest(t, g, attacker, []Mod{AddKeywordsMod("rampage 3")}, g.UntilEndOfTurnDuration())
	scopedEffectChar(t, g, attacker) // recompute the layers
	if got := RampageAmounts(findBattlefieldCard(g, attacker)); len(got) != 1 || got[0] != 3 {
		t.Fatalf("granted rampage amounts = %v, want [3]", got)
	}
	b1 := pushCombatant(t, g, g.Seats[1], "Blocker A", 1, 1)
	b2 := pushCombatant(t, g, g.Seats[1], "Blocker B", 1, 1)
	declareAttacks(t, g, attacker)
	for _, b := range []uuid.UUID{b1, b2} {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	lockInBlockDeclaration(t, g)
	if n := len(rampageItems(g)); n != 1 {
		t.Fatalf("a granted rampage queued %d triggers, want 1", n)
	}
	resolveRampageItems(t, g, 1)
	if ch := scopedEffectChar(t, g, attacker); ch.Power != 5 || ch.Toughness != 5 {
		t.Errorf("granted rampage 3 with two blockers: %d/%d, want 5/5", ch.Power, ch.Toughness)
	}
}
