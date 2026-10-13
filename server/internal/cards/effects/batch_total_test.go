package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch_total_test.go — #3172 slice cav-d, ADR 0108 amendment
// 2026-10-13: "that much" / "that many" triggers read the TOTAL of what
// happened at the same time (CR 120.3, CR 510.2, CR 603.2c) instead of
// firing once per damage source, card or permanent.

// triggersFrom counts the triggered items waiting or on the stack from
// one source.
func triggerAmountsFrom(g *game.Game, source uuid.UUID) []int {
	var out []int
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source && item.Trigger != nil {
			out = append(out, item.Trigger.Event.Amount)
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source && item.Trigger != nil {
			out = append(out, item.Trigger.Event.Amount)
		}
	}
	return out
}

func dealDamageEventsTo(g *game.Game, target uuid.UUID) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventDealDamage && ev.Target == target {
			out = append(out, ev)
		}
	}
	return out
}

// Two sources damaging one creature in ONE damage instruction are one
// trigger carrying the sum.
func TestRepercussionReadsTheTotalOfSimultaneousDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rep := pushCatalogPermanent(g, me.ID, "Repercussion", "Enchantment", b28RepercussionOracle, false)
	victim := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 9, 9)
	srcA := b12Creature(g, me.ID, "Source A", "Creature — Bear", 2, 2)
	srcB := b12Creature(g, me.ID, "Source B", "Creature — Bear", 3, 3)

	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(srcA, victim, 2)
			return g.DealDamageToCreatureForEffect(srcB, victim, 3)
		})
	})
	g.RunStateChecksForTest()
	got := triggerAmountsFrom(g, rep)
	if len(got) != 1 || got[0] != 5 {
		t.Fatalf("one trigger for the 2+3 dealt together, want [5], got %v", got)
	}
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("the controller takes the total in one go: %d -> %d, want -5", life, opp.Life)
	}
}

// Two SEPARATE damage instructions on one creature are two events, so
// they stay two triggers even inside one batch.
func TestRepercussionSeparateInstructionsAreSeparateTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rep := pushCatalogPermanent(g, me.ID, "Repercussion", "Enchantment", b28RepercussionOracle, false)
	victim := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 9, 9)
	src := b12Creature(g, me.ID, "Source", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(src, victim, 2)
		_ = g.DealDamageToCreatureForEffect(src, victim, 3)
	})
	g.RunStateChecksForTest()
	got := triggerAmountsFrom(g, rep)
	if len(got) != 2 {
		t.Fatalf("two instructions are two triggers, got %v", got)
	}
	sum := got[0] + got[1]
	if sum != 5 {
		t.Errorf("amounts %v should add to 5", got)
	}
}

// Two different creatures dealt damage together are two triggers (each
// is its own object): Blasphemous Act-shaped.
func TestRepercussionTriggersOncePerDamagedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rep := pushCatalogPermanent(g, me.ID, "Repercussion", "Enchantment", b28RepercussionOracle, false)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 9, 9)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 9, 9)
	src := b12Creature(g, me.ID, "Source", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(src, a, 4)
			return g.DealDamageToCreatureForEffect(src, b, 4)
		})
	})
	g.RunStateChecksForTest()
	if got := triggerAmountsFrom(g, rep); len(got) != 2 || got[0] != 4 || got[1] != 4 {
		t.Fatalf("one trigger of 4 per creature, got %v", got)
	}
}

// The headline case: a creature double-blocked in combat is dealt
// damage by both blockers in one combat damage step, and reflects the
// total with ONE target pick.
func TestBrashTaunterDoubleBlockedReflectsTheTotalOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	taunter := b12Push(g, me.ID, "Brash Taunter", "Creature — Goblin", s58BrashTaunter, 1, 1)
	bearA := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	bearB := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	declareAttack(t, g, opp.ID, taunter)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bearA, taunter); err != nil {
		t.Fatalf("DeclareBlocker A: %v", err)
	}
	if err := g.DeclareBlocker(bearB, taunter); err != nil {
		t.Fatalf("DeclareBlocker B: %v", err)
	}
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	b26AnswerDamageAssignments(t, g) // the Taunter's own damage among its two blockers
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a second target pick was queued: the two blockers' damage should be one trigger")
	}
	if opp.Life != life-5 {
		t.Errorf("the Taunter was dealt 2+3 and deals 5 once: %d -> %d", life, opp.Life)
	}
	events := dealDamageEventsTo(g, taunter)
	if len(events) != 2 || events[0].DamageInstance == 0 || events[0].DamageInstance != events[1].DamageInstance {
		t.Errorf("both blockers' damage is one combat-step instance: %+v", events)
	}
}

// Ingenious Artillerist: two artifacts entering together are ONE
// 2-damage hit, so a per-instance booster adds its bonus once.
func TestIngeniousArtilleristDealsOneHitForTheWholeBatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	art := pushCatalogPermanent(g, me.ID, "Ingenious Artillerist", "Creature — Human Artificer", ingeniousArtilleristOracle, false)
	before := lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, ClueToken(), 2) })
	g.RunStateChecksForTest()
	if got := triggerAmountsFrom(g, art); len(got) != 1 || got[0] != 2 {
		t.Fatalf("one trigger for two artifacts, want [2], got %v", got)
	}
	passPriorityAroundTable(t, g)
	for i, b := range before {
		opp := g.Seats[i+1]
		if opp.Life != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2", i+1, b, opp.Life)
		}
		hits := 0
		for _, ev := range dealDamageEventsTo(g, opp.ID) {
			if ev.Source == art {
				hits++
				if ev.Amount != 2 {
					t.Errorf("the hit is 2 in one event, got %d", ev.Amount)
				}
			}
		}
		if hits != 1 {
			t.Errorf("opponent %d was hit %d times, want one combined hit", i+1, hits)
		}
	}
}

// Magmakin Artillerist: a three-card discard is one 3-damage hit.
func TestMagmakinArtilleristDealsOneHitForAMultiCardDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mag := pushCatalogPermanent(g, me.ID, "Magmakin Artillerist", "Creature — Elemental Pirate", magmakinOracle, false)
	for i := 0; i < 3; i++ {
		handCard(me, "Fodder", "Sorcery")
	}
	before := lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 3) })
	g.RunStateChecksForTest()
	if got := triggerAmountsFrom(g, mag); len(got) != 1 || got[0] != 3 {
		t.Fatalf("one trigger for three discards, want [3], got %v", got)
	}
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-3 {
			t.Errorf("opponent %d: %d -> %d, want -3", i+1, b, got)
		}
	}
}

// Cryptcaller Chariot: a three-card discard is one trigger making three
// tapped Zombies.
func TestCryptcallerChariotMakesThatManyZombiesFromOneTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chariot := pushCatalogPermanent(g, me.ID, "Cryptcaller Chariot", "Artifact — Vehicle", b43CryptcallerChariotOracle, false)
	for i := 0; i < 3; i++ {
		handCard(me, "Fodder", "Sorcery")
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 3) })
	g.RunStateChecksForTest()
	if got := triggerAmountsFrom(g, chariot); len(got) != 1 || got[0] != 3 {
		t.Fatalf("one trigger for three discards, want [3], got %v", got)
	}
	passPriorityAroundTable(t, g)
	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Zombie" {
			zombies++
			if !c.Tapped {
				t.Error("the Zombies enter tapped")
			}
		}
	}
	if zombies != 3 {
		t.Errorf("Zombies = %d, want 3", zombies)
	}
}

// Widespread Brutality: the Army's damage to every non-Army creature is
// ONE damage instruction (the caveat that it hit "one after another" was
// stale), so no creature has died before the next is dealt to.
func TestWidespreadBrutalityDamageIsOneSimultaneousInstance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	castCatalogSpell(t, g, "Widespread Brutality", "Sorcery", widespreadBrutalityOracl, nil)
	passPriorityAroundTable(t, g)
	army := armyOf(t, g, me.ID)

	var inst game.DamageInstance
	seen := 0
	for _, ev := range g.Events {
		if ev.Kind != game.EventDealDamage || ev.Source != army.InstanceID {
			continue
		}
		seen++
		if ev.DamageInstance == 0 {
			t.Fatal("damage event carries no damage instance")
		}
		if inst == 0 {
			inst = ev.DamageInstance
		}
		if ev.DamageInstance != inst {
			t.Errorf("the Army's damage is split across instances: %d vs %d", inst, ev.DamageInstance)
		}
	}
	if seen != len(bears) {
		t.Errorf("damage events = %d, want %d", seen, len(bears))
	}
	if spec, _ := Lookup(widespreadBrutalityOracl); spec.Completeness != CompletenessFull {
		t.Error("Widespread Brutality has no gap left")
	}
}
