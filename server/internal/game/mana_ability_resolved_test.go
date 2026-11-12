package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_ability_resolved_test.go — #5644: EventManaAbilityResolved, one
// event per mana ability that finished resolving, with the amount of
// mana that ability produced. Tyvar the Bellicose (cards/effects) is the
// card that reads it.

// watchManaResolutions puts a permanent on the battlefield whose
// triggered ability watches EventManaAbilityResolved and never applies:
// the event is written only while something watches it.
func watchManaResolutions(t *testing.T, g *Game) {
	t.Helper()
	const oracle = "test-mana-resolution-watcher"
	pushBattlefieldForTest(g, g.Seats[0].ID, "Watcher", "Enchantment", oracle)
	withCatalogTriggers(t, func(oracleID string) []TriggeredAbility {
		if oracleID != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches:   []EventKind{EventManaAbilityResolved},
			AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return false },
			Key:       "Watcher — never applies",
		}}
	})
}

// With nothing watching, a mana ability writes no resolution event: the
// log of a game without such a card is unchanged.
func TestManaAbilityResolvedIsWrittenOnlyWhileWatched(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	if err := g.ActivateManaAbility(p.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	wantManaResolutions(t, g)
}

// manaResolutions is every EventManaAbilityResolved in the log, in
// order.
func manaResolutions(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventManaAbilityResolved {
			out = append(out, ev)
		}
	}
	return out
}

// wantManaResolutions checks the log's resolutions against (source,
// amount) pairs, in order.
func wantManaResolutions(t *testing.T, g *Game, want ...Event) []Event {
	t.Helper()
	got := manaResolutions(g)
	if len(got) != len(want) {
		t.Fatalf("%d mana ability resolutions, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Source != w.Source || got[i].CardID != w.Source || got[i].Amount != w.Amount || got[i].Actor != w.Actor {
			t.Errorf("resolution %d = source %s actor %s amount %d, want source %s actor %s amount %d",
				i, got[i].Source, got[i].Actor, got[i].Amount, w.Source, w.Actor, w.Amount)
		}
	}
	return got
}

// pushReadyBirds seeds a Birds of Paradise that is not summoning sick.
func pushReadyBirds(g *Game, p *Player) uuid.UUID {
	birds := pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", "d3a0b660-358c-41bd-9cd2-41fbf3491b1a")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == birds {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	return birds
}

// The hand click: one event, after the activation's, with the ability's
// label and the mana it added.
func TestManaAbilityResolvedCarriesTheAmount(t *testing.T) {
	withCatalogHook(t, solRingHook)
	g := newActiveGame(t)
	watchManaResolutions(t, g)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	ring := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487")

	if err := g.ActivateManaAbility(p.ID, ring, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := wantManaResolutions(t, g, Event{Source: ring, Actor: p.ID, Amount: 2})
	if got[0].Label != "Add {C}{C}" {
		t.Errorf("Label = %q, want the ability's", got[0].Label)
	}
	activated, added := -1, -1
	for i, ev := range g.Events {
		switch ev.Kind {
		case EventManaAbilityActivated:
			activated = i
		case EventManaAdded:
			added = i
		case EventManaAbilityResolved:
			if activated < 0 || added < 0 || i < activated || i < added {
				t.Errorf("resolved at %d, activated at %d, last mana at %d: the resolution comes last", i, activated, added)
			}
		}
	}
}

// CR 106.12b: the amount is what the ability produced after a
// replacement, not what it prints.
func TestManaAbilityResolvedCountsReplacedMana(t *testing.T) {
	g := newActiveGame(t)
	watchManaResolutions(t, g)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 3, "triple"))
	})

	if err := g.ActivateManaAbility(p.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	wantManaResolutions(t, g, Event{Source: forest, Actor: p.ID, Amount: 3})
}

// An ability that resolves and adds nothing still resolved.
func TestManaAbilityResolvedWithNoMana(t *testing.T) {
	g := newActiveGame(t)
	watchManaResolutions(t, g)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedCancelReplacement("no mana"))
	})

	if err := g.ActivateManaAbility(p.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	wantManaResolutions(t, g, Event{Source: forest, Actor: p.ID, Amount: 0})
}

// An ability whose colour is picked has not resolved until the answer,
// and the pick carries that across a restore.
func TestManaAbilityResolvedWaitsForThePickAcrossARestore(t *testing.T) {
	withCatalogHook(t, birdsHook)
	g := newActiveGame(t)
	watchManaResolutions(t, g)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	birds := pushReadyBirds(g, p)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})

	if err := g.ActivateManaAbility(p.ID, birds, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	wantManaResolutions(t, g)
	pick := choiceByKind(g, PendingChoiceMana)
	if pick == nil || pick.ManaAbilityRun == "" {
		t.Fatalf("the Birds pick does not carry its mana ability: %+v", pick)
	}

	// A replacement registered for the test is not catalog state, so the
	// restored game is asked without it: one mana.
	_, restored := roundTrip(t, g)
	rp := restored.Seats[0]
	rpick := choiceByKind(restored, PendingChoiceMana)
	if rpick == nil || rpick.ManaAbilityRun != pick.ManaAbilityRun {
		t.Fatalf("the restored pick lost its mana ability: %+v", rpick)
	}
	if err := restored.ResolveManaChoice(rpick.ID, rp.ID, "U"); err != nil {
		t.Fatalf("restored ResolveManaChoice: %v", err)
	}
	wantManaResolutions(t, restored, Event{Source: birds, Actor: rp.ID, Amount: rp.ManaPool.Total()})

	if err := g.ResolveManaChoice(pick.ID, p.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	got := wantManaResolutions(t, g, Event{Source: birds, Actor: p.ID, Amount: 2})
	if got[0].Label != "Add one mana of any color" {
		t.Errorf("Label = %q, want the ability's", got[0].Label)
	}
}

// A pick a resolving spell queues is no mana ability and announces
// nothing.
func TestManaAbilityResolvedIsNotASpellsMana(t *testing.T) {
	g := newActiveGame(t)
	watchManaResolutions(t, g)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(p.ID, uuid.New(), "{B}{B}{W|U}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	})
	pick := choiceByKind(g, PendingChoiceMana)
	if pick == nil {
		t.Fatal("no pick for the spell's colour choice")
	}
	if err := g.ResolveManaChoice(pick.ID, p.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	wantManaResolutions(t, g)
}

// CR 605.1b: a triggered mana ability resolves on its own, after the
// ability that triggered it, and counts only its own mana.
func TestManaAbilityResolvedForATriggeredManaAbility(t *testing.T) {
	g := newActiveGame(t)
	watchManaResolutions(t, g)
	me := g.Seats[0]
	forest := pushForest(g, me)
	aura := pushAuraOn(g, me, "Overgrowth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}{G}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	wantManaResolutions(t, g,
		Event{Source: forest, Actor: me.ID, Amount: 1},
		Event{Source: aura, Actor: me.ID, Amount: 2})
}

// A triggered mana ability with a colour to pick resolves when the last
// of its picks is answered, with the mana it added before them counted.
func TestManaAbilityResolvedForATriggeredPick(t *testing.T) {
	g := newActiveGame(t)
	watchManaResolutions(t, g)
	me := g.Seats[0]
	forest := pushForest(g, me)
	aura := pushAuraOn(g, me, "Fertile Ground", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}{W|U}{B|R}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	wantManaResolutions(t, g, Event{Source: forest, Actor: me.ID, Amount: 1})
	if len(g.PendingChoices) != 2 {
		t.Fatalf("PendingChoices = %d, want the trigger's two picks", len(g.PendingChoices))
	}
	// Answered out of order: the count follows whichever pick is left.
	second, first := g.PendingChoices[1].ID, g.PendingChoices[0].ID
	if err := g.ResolveManaChoice(second, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	wantManaResolutions(t, g, Event{Source: forest, Actor: me.ID, Amount: 1})
	if err := g.ResolveManaChoice(first, me.ID, "W"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	wantManaResolutions(t, g,
		Event{Source: forest, Actor: me.ID, Amount: 1},
		Event{Source: aura, Actor: me.ID, Amount: 3})
	if me.ManaPool.Total() != 4 {
		t.Errorf("pool = %v, want four mana", me.ManaPool)
	}
}

// The auto-tapper activates the ability too, and announces it the same
// way, triggered mana abilities included.
func TestManaAbilityResolvedFromTheAutoTapper(t *testing.T) {
	g := newActiveGame(t)
	watchManaResolutions(t, g)
	me := g.Seats[0]
	forest := pushForest(g, me)
	aura := pushAuraOn(g, me, "Fertile Ground", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{W|U|B|R|G}")}
		}
		return nil
	})

	cost, _ := ParseCost("{G}")
	g.WithWriteLock(func() {
		plan, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok {
			t.Fatalf("no plan for {G} off a Forest")
		}
		g.materializePlanLocked(me, plan, cost)
	})
	wantManaResolutions(t, g,
		Event{Source: forest, Actor: me.ID, Amount: 1},
		Event{Source: aura, Actor: me.ID, Amount: 1})
}
