package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// damage_tail_test.go is the #694 regression suite: when two damage
// replacements apply to one damage event, CR 616.1 asks the affected
// player to order them, and the damage lands from the resume rather
// than from the entry point. The resume used to be a stale copy of the
// manual MarkDamage body, so it dropped a player's life loss (returning
// ErrCardNotFound instead), the CR 120.3 planeswalker/battle split,
// CR 702.2c deathtouch and CR 702.15 lifelink.
//
// Every test here runs the SAME scenario twice and compares the whole
// observable outcome:
//
//	paused   — two replacements, "prevent 1" then "double", answered
//	           in that order through ResolveReplacementOrder
//	unpaused — one replacement doing the same arithmetic in one step,
//	           so nothing ever prompts
//
// Both must end in exactly the same place. Asserting the specific
// numbers as well keeps the test honest if both paths break together.

// damageScenario builds a board and deals one damage event to it. The
// IDs are supplied by the caller so the paused and unpaused runs are
// comparable card-for-card.
type damageScenario struct {
	// setUp stocks the battlefield. Runs under the write lock.
	setUp func(g *Game)
	// deal fires the damage event. Runs under the write lock.
	deal func(g *Game)
}

// runDamageScenario plays sc out on a fresh two-seat game. When paused
// is true it registers two replacements so CR 616.1 prompts, then
// answers the prompt in the order offered; otherwise it registers one
// combined replacement and nothing prompts.
//
// The unpaused run gets an explicit state-check sweep at the end,
// because the paused run's answer is an action boundary and gets one
// from the resume. That is the ONLY difference the two paths are
// allowed to have, so making it explicit is what lets the rest of the
// outcome be compared directly.
func runDamageScenario(t *testing.T, sc damageScenario, paused bool) *Game {
	t.Helper()
	g := newActiveGame(t)

	g.WithWriteLock(func() {
		if paused {
			// Two applicable replacements on one event is what makes
			// CR 616.1 prompt. Registration order is gather order, so
			// the prompt offers prevent-then-double and answering it
			// as offered gives (3-1)*2 = 4.
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches:   []EventKind{EventDealDamage},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventDamage },
				Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
					ev.DamageAmount--
					return nil
				},
				Label: "Prevent 1",
			})
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches:   []EventKind{EventDealDamage},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventDamage },
				Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
					ev.DamageAmount *= 2
					return nil
				},
				Label: "Double",
			})
		} else {
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches:   []EventKind{EventDealDamage},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventDamage },
				Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
					ev.DamageAmount = (ev.DamageAmount - 1) * 2
					return nil
				},
				Label: "Prevent 1 then double",
			})
		}
		sc.setUp(g)
		sc.deal(g)
	})

	if !paused {
		if len(g.PendingChoices) != 0 {
			t.Fatalf("unpaused run queued %d choices, want 0", len(g.PendingChoices))
		}
		g.WithWriteLock(func() { g.runStateChecksLocked() })
		return g
	}

	if len(g.PendingChoices) != 1 {
		t.Fatalf("paused run queued %d choices, want 1 CR 616 ordering prompt", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]
	if prompt.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", prompt.Kind, PendingChoiceReplacementOrder)
	}
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v — the answer must not fail the action (#694)", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d choices still queued after the answer, want 0", len(g.PendingChoices))
	}
	return g
}

// damageOutcome is everything a damage event can be observed to have
// done. Compared field-for-field between the paused and unpaused runs.
type damageOutcome struct {
	Life            []int
	CommanderDamage []map[uuid.UUID]int
	Graveyard       []int
	OnBattlefield   map[uuid.UUID]bool
	DamageMarked    map[uuid.UUID]int
	Deathtouched    map[uuid.UUID]bool
	Loyalty         map[uuid.UUID]int
}

func observeDamage(g *Game) damageOutcome {
	out := damageOutcome{
		OnBattlefield: map[uuid.UUID]bool{},
		DamageMarked:  map[uuid.UUID]int{},
		Deathtouched:  map[uuid.UUID]bool{},
		Loyalty:       map[uuid.UUID]int{},
	}
	for _, p := range g.Seats {
		out.Life = append(out.Life, p.Life)
		cd := map[uuid.UUID]int{}
		for k, v := range p.CommanderDamage {
			cd[k] = v
		}
		out.CommanderDamage = append(out.CommanderDamage, cd)
		out.Graveyard = append(out.Graveyard, p.Graveyard.Size())
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		out.OnBattlefield[c.InstanceID] = true
		out.DamageMarked[c.InstanceID] = c.DamageMarked
		out.Deathtouched[c.InstanceID] = c.MarkedLethalByDeathtouch
		out.Loyalty[c.InstanceID] = c.Counters[CounterLoyalty]
	}
	return out
}

// assertPausedMatchesUnpaused is the heart of #694: a damage event that
// paused for a CR 616 prompt has to end exactly where one that never
// paused would.
func assertPausedMatchesUnpaused(t *testing.T, sc damageScenario) *Game {
	t.Helper()
	pausedGame := runDamageScenario(t, sc, true)
	unpausedGame := runDamageScenario(t, sc, false)
	got, want := observeDamage(pausedGame), observeDamage(unpausedGame)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("paused damage ended somewhere else than unpaused damage:\n paused: %+v\n direct: %+v", got, want)
	}
	return pausedGame
}

// pushTestCreature puts a creature on the battlefield with a caller-
// supplied instance ID (so both runs of a scenario are comparable) and
// the given effective keywords.
func pushTestCreature(g *Game, id uuid.UUID, owner *Player, power, toughness int, keywords ...string) {
	c := NewCard("Test Creature", owner.ID)
	c.InstanceID = id
	c.TypeLine = "Creature — Test"
	c.Power = power
	c.Toughness = toughness
	c.Controller = owner.ID
	g.Battlefield.PushTop(c)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].effective = printedEffectiveWith(c, keywords...)
		}
	}
}

// TestPausedDamageToPlayerStillChangesLife — the headline symptom.
// DealDamageToPlayerForEffect's tail is a life change; the old resume
// looked for a battlefield card with the player's ID, found none, and
// returned ErrCardNotFound with the prompt already dequeued. Life never
// moved and the action failed.
func TestPausedDamageToPlayerStillChangesLife(t *testing.T) {
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(*Game) {},
		deal: func(g *Game) {
			_ = g.DealDamageToPlayerForEffect(uuid.Nil, g.Seats[1].ID, 3)
		},
	})
	if got := g.Seats[1].Life; got != StartingLife-4 {
		t.Errorf("target life = %d, want %d — (3 prevented by 1, then doubled) = 4",
			got, StartingLife-4)
	}
}

// TestPausedDamageToPlaneswalkerTakesLoyalty — CR 120.3c. The old
// resume incremented DamageMarked on whatever it found, a number the
// CR 704.5i loyalty SBA never reads, so an ordered pair of damage
// replacements made a planeswalker unkillable again (the #406 bug,
// reintroduced through the back door).
func TestPausedDamageToPlaneswalkerTakesLoyalty(t *testing.T) {
	pwID := uuid.New()
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(g *Game) {
			owner := g.Seats[1]
			c := NewCard("Test Walker", owner.ID)
			c.InstanceID = pwID
			c.TypeLine = "Legendary Planeswalker — Test"
			c.Controller = owner.ID
			c.Counters = map[string]int{CounterLoyalty: 5}
			g.Battlefield.PushTop(c)
		},
		deal: func(g *Game) {
			_ = g.DealDamageToCreatureForEffect(uuid.Nil, pwID, 3)
		},
	})
	pw := findBattlefieldCard(g, pwID)
	if pw == nil {
		t.Fatal("the planeswalker left the battlefield; it should be at 1 loyalty")
	}
	if got := pw.Counters[CounterLoyalty]; got != 1 {
		t.Errorf("loyalty = %d, want 1 — 4 damage removes 4 loyalty counters (CR 120.3c)", got)
	}
	if got := pw.DamageMarked; got != 0 {
		t.Errorf("DamageMarked = %d, want 0 — damage to a planeswalker is not marked on it", got)
	}
}

// TestPausedCombatDamageKeepsDeathtouchAndLifelink — CR 702.2c and
// CR 702.15 ride along with combat damage. The old resume applied
// neither, so an ordered pair of replacements turned a lifelinking
// deathtoucher into a creature that marked a number and nothing else.
func TestPausedCombatDamageKeepsDeathtouchAndLifelink(t *testing.T) {
	atkID, blkID := uuid.New(), uuid.New()
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(g *Game) {
			pushTestCreature(g, atkID, g.Seats[0], 3, 3, "deathtouch", "lifelink")
			pushTestCreature(g, blkID, g.Seats[1], 1, 10)
		},
		deal: func(g *Game) {
			g.markCombatDamageOnCardLocked(blkID, 3, atkID, "")
		},
	})
	if got := g.Seats[0].Life; got != StartingLife+4 {
		t.Errorf("attacker's controller life = %d, want %d — lifelink gains the damage dealt (CR 702.15)",
			got, StartingLife+4)
	}
	if findBattlefieldCard(g, blkID) != nil {
		t.Error("the 1/10 blocker survived 4 deathtouch damage — CR 702.2c makes any nonzero " +
			"damage from a deathtouch source lethal (CR 704.5h)")
	}
	if got := g.Seats[1].Graveyard.Size(); got != 1 {
		t.Errorf("defender's graveyard holds %d cards, want 1 (the dead blocker)", got)
	}
}

// TestPausedCombatDamageToPlayerKeepsCommanderDamage — CR 903.10a. A
// commander's combat damage feeds the 21-damage clock, and the old
// resume never even changed the player's life, let alone the tally.
func TestPausedCombatDamageToPlayerKeepsCommanderDamage(t *testing.T) {
	atkID := uuid.New()
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(g *Game) {
			pushTestCreature(g, atkID, g.Seats[0], 3, 3, "lifelink")
			findBattlefieldCard(g, atkID).IsCommander = true
		},
		deal: func(g *Game) {
			g.markCombatDamageToPlayerLocked(g.Seats[1].ID, atkID, 3, "")
		},
	})
	if got := g.Seats[1].Life; got != StartingLife-4 {
		t.Errorf("defender life = %d, want %d", got, StartingLife-4)
	}
	if got := g.Seats[1].CommanderDamage[atkID]; got != 4 {
		t.Errorf("commander damage = %d, want 4 (CR 903.10a, keyed on the commander's instance ID)", got)
	}
	if got := g.Seats[0].Life; got != StartingLife+4 {
		t.Errorf("attacker's controller life = %d, want %d — lifelink on combat damage to a player too",
			got, StartingLife+4)
	}
}

// dealDamageRecorder captures every EventDealDamage the engine emits,
// so a test can assert on the Actor / Combat fields a trigger would
// key off rather than only on the resulting board state.
type dealDamageRecorder struct{ events []Event }

func (r *dealDamageRecorder) OnEvent(_ *Game, ev Event) {
	if ev.Kind == EventDealDamage {
		r.events = append(r.events, ev)
	}
}

// TestPausedDamageEmitsACombatFlaggedEvent — "whenever ~ deals combat
// damage" triggers key off Event.Combat and Event.Actor. The old resume
// emitted a bare EventDealDamage with neither, so a resumed combat
// damage event was invisible to them.
func TestPausedDamageEmitsACombatFlaggedEvent(t *testing.T) {
	atkID, blkID := uuid.New(), uuid.New()
	seen := &dealDamageRecorder{}
	sc := damageScenario{
		setUp: func(g *Game) {
			pushTestCreature(g, atkID, g.Seats[0], 3, 3)
			pushTestCreature(g, blkID, g.Seats[1], 1, 10)
			g.Listeners = append(g.Listeners, seen)
		},
		deal: func(g *Game) {
			g.markCombatDamageOnCardLocked(blkID, 3, atkID, "")
		},
	}
	_ = runDamageScenario(t, sc, true)
	if len(seen.events) != 1 {
		t.Fatalf("%d EventDealDamage emitted, want exactly 1", len(seen.events))
	}
	ev := seen.events[0]
	if !ev.Combat {
		t.Error("Combat is false on a resumed combat damage event")
	}
	if ev.Amount != 4 {
		t.Errorf("Amount = %d, want 4 (the post-replacement amount)", ev.Amount)
	}
	if ev.Source != atkID || ev.Target != blkID {
		t.Errorf("Source/Target = %s/%s, want %s/%s", ev.Source, ev.Target, atkID, blkID)
	}
}

// TestPausedDamageWhoseTargetLeftDoesNotFailTheAction — the last
// bullet of #694. ResolveReplacementOrder dequeues the prompt before it
// dispatches, so an error from the tail takes the player's prompt away
// AND fails their action. A target that left between the prompt and the
// answer is not an error: the damage simply does not happen.
func TestPausedDamageWhoseTargetLeftDoesNotFailTheAction(t *testing.T) {
	g := newActiveGame(t)
	victimID := uuid.New()
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches:   []EventKind{EventDealDamage},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventDamage },
				Replace:   func(*ReplacementEvent, *Game, *Card) error { return nil },
				Label:     "no-op",
			})
		}
		pushTestCreature(g, victimID, g.Seats[1], 2, 2)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, victimID, 3)
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]

	// The creature leaves before the chooser answers.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victimID {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				break
			}
		}
	})

	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Errorf("ResolveReplacementOrder = %v, want nil — the prompt is already dequeued, so "+
			"a vanished target must not fail the action too", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d choices still queued, want 0", len(g.PendingChoices))
	}
}

// TestMarkDamageSandboxVerbKeepsItsOwnTail — the manual MarkDamage
// action is a signed delta on DamageMarked with no CR 120.3 split and
// no combat riders, and #694 must not quietly promote it to a rules
// path. A negative delta (undoing a mark) still works, and still
// clamps at zero.
func TestMarkDamageSandboxVerbKeepsItsOwnTail(t *testing.T) {
	g := newActiveGame(t)
	id := uuid.New()
	g.WithWriteLock(func() { pushTestCreature(g, id, g.Seats[0], 4, 4) })

	if err := g.MarkDamage(id, 2); err != nil {
		t.Fatalf("MarkDamage +2: %v", err)
	}
	if got := findBattlefieldCard(g, id).DamageMarked; got != 2 {
		t.Errorf("DamageMarked = %d, want 2", got)
	}
	if err := g.MarkDamage(id, -5); err != nil {
		t.Fatalf("MarkDamage -5: %v", err)
	}
	if got := findBattlefieldCard(g, id).DamageMarked; got != 0 {
		t.Errorf("DamageMarked = %d, want 0 — a negative delta clamps at zero", got)
	}
	if err := g.MarkDamage(uuid.New(), 1); err != ErrCardNotFound {
		t.Errorf("MarkDamage on an unknown card = %v, want ErrCardNotFound", err)
	}
}
