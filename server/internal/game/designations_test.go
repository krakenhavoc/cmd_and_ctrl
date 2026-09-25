package game

import (
	"testing"

	"github.com/google/uuid"
)

// designations_test.go is the engine-side suite for ADR 0071: one
// gate on printed abilities, evaluated at the one place an object
// becomes abilities.
//
// Like layer4_authoritative_test.go and layer6_ability_removal_test.go
// it drives stubbed catalog hooks, so it pins the ENGINE contract
// without depending on a particular card. The catalog side — Wizard
// Class, Fortune Teller's Talent, Case of the Shattered Pact, The
// Seriema — is pinned in cards/effects/designations_test.go.

const (
	gatedOracle = "designation-gated"
	plainOracle = "designation-plain"
)

// --- the predicate ---------------------------------------------------

// TestDesignationActiveReadsTheObject pins Active, which is the only
// place any of the four rules is read. Everything else in this file
// is a consequence of it.
func TestDesignationActiveReadsTheObject(t *testing.T) {
	cases := []struct {
		name string
		gate Designation
		card Card
		want bool
	}{
		// CR 716.2b: a Class permanent with no level designation is
		// level 1, so the zero value satisfies a level-1 gate.
		{"zero value is no gate", Designation{}, Card{}, true},
		{"level 1 gate on an unlevelled Class", ClassLevel(1), Card{}, true},
		{"level 2 gate on an unlevelled Class", ClassLevel(2), Card{}, false},
		{"level 2 gate at level 2", ClassLevel(2), Card{ClassLevel: 2}, true},
		// CR 716.2: "level N or greater", never "exactly N" — a
		// level-3 Class keeps its level-2 line.
		{"level 2 gate at level 3", ClassLevel(2), Card{ClassLevel: 3}, true},
		{"solved gate on an unsolved Case", CaseSolved(), Card{}, false},
		{"solved gate on a solved Case", CaseSolved(), Card{Solved: true}, true},
		{"harnessed gate on an unharnessed permanent", Harnessed(), Card{}, false},
		{"harnessed gate on a harnessed permanent", Harnessed(), Card{Harnessed: true}, true},
		{"7+ with no counters", ChargeCounters(7), Card{}, false},
		{"7+ with six", ChargeCounters(7), Card{Counters: map[string]int{CounterCharge: 6}}, false},
		{"7+ with seven", ChargeCounters(7), Card{Counters: map[string]int{CounterCharge: 7}}, true},
		{"7+ with nine", ChargeCounters(7), Card{Counters: map[string]int{CounterCharge: 9}}, true},
		// The wrong kind of counter is not the threshold's counter.
		{"7+ with seven lore counters", ChargeCounters(7), Card{Counters: map[string]int{CounterLore: 7}}, false},
		// ADR 0071 decision 3: reserved, and false until #886.
		{"door gate is never satisfied", Designation{Kind: DesignationDoorUnlocked, Door: DoorLeft}, Card{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.gate.Active(tc.card); got != tc.want {
				t.Errorf("Active = %v, want %v", got, tc.want)
			}
		})
	}
}

// --- the filter ------------------------------------------------------

// levelledStatics is a Class-shaped static list: one ability per
// level, each a self-only layer-6 keyword grant so the test can read
// which are live off the layered card.
func levelledStatics() []StaticAbility {
	grant := func(level int, kw string) StaticAbility {
		return StaticAbility{
			Layer:      Layer6Ability,
			ActiveWhen: ClassLevel(level),
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = append(c.Abilities, kw)
			},
		}
	}
	return []StaticAbility{grant(1, "vigilance"), grant(2, "flying"), grant(3, "trample")}
}

// TestGatedStaticsAreNotGatheredBelowTheirLevel is the property the
// whole ADR turns on: an inactive ability is ABSENT, not filtered
// later. If it were gathered, the layer pass would apply it.
func TestGatedStaticsAreNotGatheredBelowTheirLevel(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != gatedOracle {
			return nil
		}
		return levelledStatics()
	})
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Class", TypeLine: "Enchantment — Class", OracleID: gatedOracle,
		Owner: seat, Controller: seat,
	})

	assertLevels := func(level int, wantFlying, wantTrample bool) {
		t.Helper()
		c := layeredBattlefieldCard(t, g, id)
		if !HasKeyword(&c, "vigilance") {
			t.Errorf("level %d: the level-1 ability must always be live", level)
		}
		if got := HasKeyword(&c, "flying"); got != wantFlying {
			t.Errorf("level %d: level-2 ability live = %v, want %v", level, got, wantFlying)
		}
		if got := HasKeyword(&c, "trample"); got != wantTrample {
			t.Errorf("level %d: level-3 ability live = %v, want %v", level, got, wantTrample)
		}
	}

	assertLevels(1, false, false)

	g.WithWriteLock(func() {
		if err := g.SetClassLevelForEffect(id, 2); err != nil {
			t.Fatalf("SetClassLevelForEffect(2): %v", err)
		}
	})
	assertLevels(2, true, false)

	g.WithWriteLock(func() {
		if err := g.SetClassLevelForEffect(id, 3); err != nil {
			t.Fatalf("SetClassLevelForEffect(3): %v", err)
		}
	})
	assertLevels(3, true, true)
}

// TestGatedStaticsReadChargeCountersLive is CR 721.2a's "as long as",
// which is the one behavioural difference from a Saga's lore ratchet:
// take the counters away and the abilities go with them.
func TestGatedStaticsReadChargeCountersLive(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != gatedOracle {
			return nil
		}
		return []StaticAbility{{
			Layer:      Layer6Ability,
			ActiveWhen: ChargeCounters(4),
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = append(c.Abilities, "flying")
			},
		}}
	})
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Spacecraft", TypeLine: "Artifact — Spacecraft", OracleID: gatedOracle,
		Owner: seat, Controller: seat,
	})

	if c := layeredBattlefieldCard(t, g, id); HasKeyword(&c, "flying") {
		t.Error("no charge counters: the 4+ ability must not be live")
	}
	if err := g.AddCounter(id, CounterCharge, 4); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if c := layeredBattlefieldCard(t, g, id); !HasKeyword(&c, "flying") {
		t.Error("four charge counters: the 4+ ability must be live")
	}
	if err := g.AddCounter(id, CounterCharge, -1); err != nil {
		t.Fatalf("AddCounter(-1): %v", err)
	}
	if c := layeredBattlefieldCard(t, g, id); HasKeyword(&c, "flying") {
		t.Error("three charge counters: the 4+ ability must go off again (CR 721.2a is an 'as long as')")
	}
}

// TestGatedTriggersDoNotFire — the harvester's half. A trigger below
// its gate is not matched, not prompted and not queued.
func TestGatedTriggersDoNotFire(t *testing.T) {
	trigger := func(gate Designation) TriggeredAbility {
		return TriggeredAbility{
			ActiveWhen: gate,
			Watches:    []EventKind{EventTapCard},
			AppliesTo:  func(Event, *Card, Characteristic, *Game) bool { return true },
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "gated", func(*Game, *StackItem) error { return nil })
			},
		}
	}
	prev := CatalogTriggers
	CatalogTriggers = func(key string) []TriggeredAbility {
		if key != gatedOracle {
			return nil
		}
		return []TriggeredAbility{trigger(CaseSolved())}
	}
	t.Cleanup(func() { CatalogTriggers = prev })

	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Case", TypeLine: "Enchantment — Case", OracleID: gatedOracle,
		Owner: seat, Controller: seat,
	})

	queued := func() int {
		n := 0
		g.ReadSnapshot(func() { n = len(g.PendingTriggers) })
		return n
	}
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventTapCard, CardID: id}) })
	if queued() != 0 {
		t.Fatal("an unsolved Case's solved trigger must not be harvested")
	}

	g.WithWriteLock(func() {
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
		g.EmitEvent(Event{Kind: EventTapCard, CardID: id})
	})
	if queued() != 1 {
		t.Fatalf("a solved Case's trigger must be harvested: queued %d, want 1", queued())
	}
}

// TestGatedTriggersDoNotFireFromADeclaredZone is the FIFTH place an
// object becomes triggered abilities: #925's declared-zone walk
// (trigger_zones.go), which reads a card as it sits in a graveyard or
// in exile rather than on the battlefield.
//
// It is the one that would have been easy to miss, and missing it is
// the structural bug ADR 0046 and ADR 0071 are both written against:
// a second, parallel `CatalogTriggers` read there would fire a Case's
// "Solved — …" ability off a Case in a graveyard, where CR 400.7 has
// already taken the designation away. Routing it through
// TriggersForCard makes the answer fall out instead of needing a rule.
func TestGatedTriggersDoNotFireFromADeclaredZone(t *testing.T) {
	const oracle = "designation-gated-zone-oracle"
	for _, tc := range []struct {
		name   string
		gate   Designation
		solved bool
		want   int
	}{
		{"ungated trigger fires from the graveyard", Designation{}, false, 1},
		{"gated trigger does not fire from the graveyard", CaseSolved(), false, 0},
		// CR 400.7 clears Solved on the way out, so a card in a
		// graveyard can never satisfy the gate. The flag is set by
		// hand here to prove the GATE is what refuses it, not the
		// absence of a way to set the field.
		{"a graveyard card carrying the flag is the gate's call", CaseSolved(), true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner := g.Seats[0]
			var w zoneWatch
			ability := w.ability(ZoneGraveyard, EventDrawCard)
			ability.ActiveWhen = tc.gate
			withZoneTriggerCard(t, oracle, []TriggeredAbility{ability})
			owner.Graveyard.PushTop(Card{
				InstanceID: uuid.New(),
				Name:       "Gated Zone Watcher",
				OracleID:   oracle,
				TypeLine:   "Enchantment — Case",
				Owner:      owner.ID,
				Controller: owner.ID,
				Solved:     tc.solved,
			})

			g.WithWriteLock(func() {
				g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID})
			})

			if w.fired != tc.want {
				t.Fatalf("trigger fired %d times, want %d", w.fired, tc.want)
			}
		})
	}
}

// TestGatedActivatedAbilitiesAreNotEnumerated — the activation path's
// half. Because every consumer (the engine, the legal-move
// enumerator, the lobby lookup, the wire) reads this one accessor,
// this is the #544 invariant: nothing can offer a move the engine
// would refuse.
func TestGatedActivatedAbilitiesAreNotEnumerated(t *testing.T) {
	prev := CatalogActivatedAbilities
	CatalogActivatedAbilities = func(key string) []ActivatedAbilityShape {
		if key != gatedOracle {
			return nil
		}
		return []ActivatedAbilityShape{
			{Label: "always"},
			{Label: "solved only", ActiveWhen: CaseSolved()},
		}
	}
	t.Cleanup(func() { CatalogActivatedAbilities = prev })

	unsolved := Card{OracleID: gatedOracle, Name: "Test Case", TypeLine: "Enchantment — Case"}
	if got := ActivatedAbilitiesForCard(unsolved); len(got) != 1 || got[0].Label != "always" {
		t.Errorf("unsolved: got %d abilities, want only the ungated one", len(got))
	}
	solved := unsolved
	solved.Solved = true
	if got := ActivatedAbilitiesForCard(solved); len(got) != 2 {
		t.Errorf("solved: got %d abilities, want both", len(got))
	}
}

// TestUngatedAbilitiesAreHandedBackUnchanged is the allocation
// contract: a card with no gate anywhere gets the catalog's own
// slice, not a copy of it. Every card in the catalog but a handful
// takes this path on every gather.
func TestUngatedAbilitiesAreHandedBackUnchanged(t *testing.T) {
	statics := []StaticAbility{{Layer: Layer6Ability}, {Layer: Layer7PT}}
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != plainOracle {
			return nil
		}
		return statics
	})
	got := StaticAbilitiesForCard(Card{OracleID: plainOracle})
	if len(got) != 2 || &got[0] != &statics[0] {
		t.Error("an ungated list must be handed back as-is, with no copy")
	}
}

// --- the designations themselves -------------------------------------

// TestSetClassLevelAnnouncesAndInvalidates: the event is what "when
// this Class becomes level N" watches AND what keeps gated statics
// from going stale.
func TestSetClassLevelAnnouncesAndInvalidates(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Class", TypeLine: "Enchantment — Class",
		Owner: seat, Controller: seat,
	})

	var before, after uint64
	g.ReadSnapshot(func() { before = g.layerVersion.Load() })
	g.WithWriteLock(func() {
		if err := g.SetClassLevelForEffect(id, 2); err != nil {
			t.Fatalf("SetClassLevelForEffect: %v", err)
		}
		after = g.layerVersion.Load()
	})
	if after <= before {
		t.Error("a level change must bump the layer version or gated statics go stale")
	}

	var level int
	g.ReadSnapshot(func() { level = g.ClassLevelFor(id) })
	if level != 2 {
		t.Errorf("ClassLevelFor = %d, want 2", level)
	}

	seen := false
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventClassLevel && ev.CardID == id && ev.Amount == 2 && ev.Actor == seat {
				seen = true
			}
		}
	})
	if !seen {
		t.Error("no EventClassLevel(2) in the log")
	}
}

// TestSolveCaseIsIdempotent — CR 719.3b: a solved Case stays solved,
// so a second solve must not re-announce.
func TestSolveCaseIsIdempotent(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Case", TypeLine: "Enchantment — Case",
		Owner: seat, Controller: seat,
	})
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			if err := g.SolveCaseForEffect(id); err != nil {
				t.Fatalf("SolveCaseForEffect: %v", err)
			}
		}
	})
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventCaseSolved && ev.CardID == id {
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("EventCaseSolved emitted %d times, want exactly 1", n)
	}
	solved := false
	g.ReadSnapshot(func() { solved = g.IsSolved(id) })
	if !solved {
		t.Error("IsSolved = false after solving")
	}
}

// TestHarnessForEffectIsIdempotent — CR 701.64a: "harness [this
// permanent]" means "if this permanent isn't harnessed, it becomes
// harnessed", so nothing stops the ability being activated a second
// time and it must not re-announce.
func TestHarnessForEffectIsIdempotent(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Rock", TypeLine: "Artifact",
		Owner: seat, Controller: seat,
	})
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			if err := g.HarnessForEffect(id); err != nil {
				t.Fatalf("HarnessForEffect: %v", err)
			}
		}
	})
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventHarnessed && ev.CardID == id {
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("EventHarnessed emitted %d times, want exactly 1", n)
	}
	harnessed := false
	g.ReadSnapshot(func() { harnessed = g.IsHarnessed(id) })
	if !harnessed {
		t.Error("IsHarnessed = false after harnessing")
	}
}

// TestDesignationsClearOnBattlefieldLeave — CR 400.7. This is also
// what makes both non-copiable without the copy path knowing
// anything: a designation is battlefield state on a permanent, not a
// characteristic of a card.
func TestDesignationsClearOnBattlefieldLeave(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Class", TypeLine: "Enchantment — Class",
		Owner: seat, Controller: seat,
	})
	g.WithWriteLock(func() {
		if err := g.SetClassLevelForEffect(id, 3); err != nil {
			t.Fatalf("SetClassLevelForEffect: %v", err)
		}
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
		// ADR 0071 amendment (#1321): harnessed is the same shape.
		if err := g.HarnessForEffect(id); err != nil {
			t.Fatalf("HarnessForEffect: %v", err)
		}
	})
	var moved Card
	g.WithWriteLock(func() {
		var err error
		moved, err = MoveCard(g.Battlefield, g.Seats[0].Graveyard, id)
		if err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if moved.ClassLevel != 0 || moved.Solved || moved.Harnessed {
		t.Errorf("designations survived the zone change: level %d, solved %v, harnessed %v",
			moved.ClassLevel, moved.Solved, moved.Harnessed)
	}
	if got := ClassLevelOf(moved); got != 1 {
		t.Errorf("ClassLevelOf a card off the battlefield = %d, want 1 (CR 716.2b)", got)
	}
}

// TestDesignationsSurviveCloneAndSnapshot — ADR 0071 decision 6, the
// undo and the deploy. Both zero values are LEGAL states, so a
// restore that dropped them would come back wrong and say nothing.
func TestDesignationsSurviveCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Class", TypeLine: "Enchantment — Class",
		Owner: seat, Controller: seat,
	})
	g.WithWriteLock(func() {
		if err := g.SetClassLevelForEffect(id, 3); err != nil {
			t.Fatalf("SetClassLevelForEffect: %v", err)
		}
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
		if err := g.HarnessForEffect(id); err != nil {
			t.Fatalf("HarnessForEffect: %v", err)
		}
	})

	find := func(cards []Card) Card {
		for _, c := range cards {
			if c.InstanceID == id {
				return c
			}
		}
		return Card{}
	}

	cloned := find(g.Clone().Battlefield.Cards)
	if cloned.ClassLevel != 3 || !cloned.Solved || !cloned.Harnessed {
		t.Errorf("clone lost the designations: level %d, solved %v, harnessed %v",
			cloned.ClassLevel, cloned.Solved, cloned.Harnessed)
	}

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	var back Card
	restored.ReadSnapshot(func() { back = find(restored.Battlefield.Cards) })
	if back.ClassLevel != 3 || !back.Solved || !back.Harnessed {
		t.Errorf("snapshot round-trip lost the designations: level %d, solved %v, harnessed %v",
			back.ClassLevel, back.Solved, back.Harnessed)
	}
}

// TestClassLevelIsNotACopiableValue — CR 716.2c. A copy of a level-3
// Class is level 1, and the reason is structural rather than a rule
// in the copy path: the level is not among the copiable values
// because it is not a printed characteristic.
func TestClassLevelIsNotACopiableValue(t *testing.T) {
	source := Card{
		InstanceID: uuid.New(),
		Name:       "Test Class",
		TypeLine:   "Enchantment — Class",
		OracleID:   gatedOracle,
		ClassLevel: 3,
		Solved:     true,
		Harnessed:  true,
	}
	copied := CopiableValuesOf(source)
	clone := Card{InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter"}
	clone.applyCopy(copied, source)
	if clone.ClassLevel != 0 || clone.Solved || clone.Harnessed {
		t.Errorf("copy took the designations: level %d, solved %v, harnessed %v — CR 716.2c / 719.3b / 701.64 say it must not",
			clone.ClassLevel, clone.Solved, clone.Harnessed)
	}
	if got := ClassLevelOf(clone); got != 1 {
		t.Errorf("a copy of a level-3 Class is level %d, want 1", got)
	}
}

// TestUndoAcrossALevelUpRewindsTheLevel is the undo path: a Room's
// undo is Clone() before the action and RestoreFrom() after, so a
// designation that survives one and not the other would take back
// the mana and keep the level.
func TestUndoAcrossALevelUpRewindsTheLevel(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Test Class", TypeLine: "Enchantment — Class",
		Owner: seat, Controller: seat,
	})

	pre := g.Clone()
	g.WithWriteLock(func() {
		if err := g.SetClassLevelForEffect(id, 2); err != nil {
			t.Fatalf("SetClassLevelForEffect: %v", err)
		}
	})
	if got := levelOfLocked(g, id); got != 2 {
		t.Fatalf("level before the undo = %d, want 2", got)
	}

	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if got := levelOfLocked(g, id); got != 1 {
		t.Errorf("level after the undo = %d, want 1 — the level-up was taken back", got)
	}
}

// levelOfLocked reads a battlefield permanent's level under the read
// lock ClassLevelFor's contract asks for.
func levelOfLocked(g *Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() { n = g.ClassLevelFor(id) })
	return n
}
