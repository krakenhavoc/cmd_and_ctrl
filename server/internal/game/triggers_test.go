package game

import (
	"testing"

	"github.com/google/uuid"
)

// triggers_test.go covers the S19 sub-PR 1 dispatcher framework:
// the triggerHarvester listener, the CatalogTriggers hook, and the
// LKI snapshot mechanism. No catalog cards declare TriggeredAbility
// in sub-PR 1 — coverage here uses a stub CatalogTriggers function
// that returns synthetic abilities for a fixed oracle ID.
//
// Per-card trigger coverage lands in sub-PRs 3-7 (ETB, dies, upkeep,
// cast, combat); the sub-PR 8 test bundle layers APNAP-ordering and
// fallback-to-manual coverage on top.

// withCatalogTriggers swaps the package-level CatalogTriggers hook
// for the duration of the test. Restores the previous value via
// t.Cleanup so test order doesn't leak state between cases. Pattern
// mirrors how RegisterReplacementForTest scopes test replacements,
// adapted for the function-variable hook shape.
func withCatalogTriggers(t *testing.T, fn func(oracleID string) []TriggeredAbility) {
	t.Helper()
	prev := CatalogTriggers
	CatalogTriggers = fn
	t.Cleanup(func() { CatalogTriggers = prev })
}

// TestTriggerHarvesterFiresOnETB pins the basic dispatcher contract:
// a card on the battlefield with a TriggeredAbility watching EventETB
// gets its Build callback invoked when the corresponding event fires,
// and the returned StackItem lands on g.PendingTriggers.
func TestTriggerHarvesterFiresOnETB(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-etb-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Test ETB Creature",
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	var buildCalls int
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(ev Event, source *Card, _ Characteristic, g *Game) *StackItem {
				buildCalls++
				return &StackItem{
					Kind:         StackItemTriggered,
					Controller:   source.Controller,
					Owner:        source.Owner,
					SourceCardID: source.InstanceID,
					Label:        "Test ETB trigger",
				}
			},
		}}
	})

	pendingBefore := len(g.PendingTriggers)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID, Actor: owner.ID})
	})

	if buildCalls != 1 {
		t.Fatalf("Build calls: got %d, want 1", buildCalls)
	}
	if got := len(g.PendingTriggers) - pendingBefore; got != 1 {
		t.Fatalf("PendingTriggers delta: got %d, want 1", got)
	}
	queued := g.PendingTriggers[len(g.PendingTriggers)-1]
	if queued.SourceCardID != cardID {
		t.Errorf("queued source = %s, want %s", queued.SourceCardID, cardID)
	}
	if queued.Kind != StackItemTriggered {
		t.Errorf("queued kind = %q, want %q", queued.Kind, StackItemTriggered)
	}
	if queued.Label != "Test ETB trigger" {
		t.Errorf("queued label = %q, want %q", queued.Label, "Test ETB trigger")
	}
	if queued.ID == uuid.Nil {
		t.Errorf("queued ID is nil — harvester should mint one when Build leaves it zero")
	}
}

// TestTriggerHarvesterSkipsUnwatchedKinds keeps the harvester cheap:
// a TriggeredAbility that watches EventETB only does not fire on
// EventDealDamage even though the source is on the battlefield.
// Watches is the pre-filter; without it the per-event walk would
// dispatch to every card on every emit.
func TestTriggerHarvesterSkipsUnwatchedKinds(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-etb-only-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	var buildCalls int
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				buildCalls++
				return &StackItem{SourceCardID: source.InstanceID}
			},
		}}
	})

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Actor: owner.ID, Amount: 3})
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID})
	})

	if buildCalls != 0 {
		t.Errorf("Build calls on unwatched events: got %d, want 0", buildCalls)
	}
}

// TestTriggerHarvesterAppliesToFalseSuppresses confirms the per-event
// predicate gates Build — an ability whose AppliesTo returns false
// for a given event leaves PendingTriggers untouched.
func TestTriggerHarvesterAppliesToFalseSuppresses(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-applies-false-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches:   []EventKind{EventETB},
			AppliesTo: func(_ Event, _ *Card, _ Characteristic, _ *Game) bool { return false },
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return &StackItem{SourceCardID: source.InstanceID}
			},
		}}
	})

	pendingBefore := len(g.PendingTriggers)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID, Actor: owner.ID})
	})
	if got := len(g.PendingTriggers) - pendingBefore; got != 0 {
		t.Fatalf("PendingTriggers delta: got %d, want 0 (AppliesTo returned false)", got)
	}
}

// TestTriggerHarvesterBuildNilSuppresses confirms an optional-trigger
// pattern: Build returning nil (controller declined) leaves the queue
// untouched even when AppliesTo returned true.
func TestTriggerHarvesterBuildNilSuppresses(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-build-nil-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			Build:   func(_ Event, _ *Card, _ Characteristic, _ *Game) *StackItem { return nil },
		}}
	})

	pendingBefore := len(g.PendingTriggers)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID, Actor: owner.ID})
	})
	if got := len(g.PendingTriggers) - pendingBefore; got != 0 {
		t.Fatalf("PendingTriggers delta: got %d, want 0 (Build returned nil)", got)
	}
}

// TestTriggerHarvesterLKIPopulatedOnLTB exercises the CR 603.10
// snapshot: a creature with a "when ~ dies" trigger reads the
// captured LKI inside Build, not the post-move characteristics.
// The dispatcher pre-stamps lastKnownBattlefield in the LTB-emitting
// path (effect_api.go::ExileCardForEffect via the test's
// DestroyPermanentForEffect call), and the harvester reads it when
// EventLTB fires.
func TestTriggerHarvesterLKIPopulatedOnLTB(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-ltb-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		Name:       "Solemn Standin",
		TypeLine:   "Creature — Test",
		Power:      3,
		Toughness:  3,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	var seenPower int
	var seenName string
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventLTB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, lki Characteristic, _ *Game) *StackItem {
				seenPower = lki.Power
				seenName = lki.Name
				return &StackItem{
					Kind:         StackItemTriggered,
					SourceCardID: source.InstanceID,
					Controller:   source.Controller,
					Owner:        source.Owner,
					Label:        "dies-trigger",
				}
			},
		}}
	})

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(cardID); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	if seenPower != 3 {
		t.Errorf("LKI Power: got %d, want 3 (snapshot should preserve battlefield characteristics)", seenPower)
	}
	if seenName != "Solemn Standin" {
		t.Errorf("LKI Name: got %q, want %q", seenName, "Solemn Standin")
	}
	// LKI map should be cleared after dispatch — keeping it would
	// leak across subsequent LTB events for the same instance ID.
	if _, ok := g.lastKnownBattlefield[cardID]; ok {
		t.Errorf("lastKnownBattlefield[%s] still populated after dispatch — should be cleared", cardID)
	}
	// And the trigger should have made it onto the queue.
	found := false
	for _, item := range g.PendingTriggers {
		if item.SourceCardID == cardID && item.Label == "dies-trigger" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("dies-trigger missing from PendingTriggers")
	}
}

// TestTriggerHarvesterOptionalPromptQueuesChoice exercises the S19
// sub-PR 2 OptionalPrompt branch: a TriggeredAbility with a non-nil
// OptionalPrompt queues a PendingChoiceTriggerPrompt instead of
// firing Build immediately. PendingTriggers stays empty until the
// chooser answers; on `apply: true` the deferred Build runs and the
// StackItem appears on the queue.
func TestTriggerHarvesterOptionalPromptQueuesChoice(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-optional-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		Name:       "Optional Pal",
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	var buildCalls int
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				buildCalls++
				return &StackItem{
					Kind:         StackItemTriggered,
					Controller:   source.Controller,
					Owner:        source.Owner,
					SourceCardID: source.InstanceID,
					Label:        "Optional Pal trigger",
				}
			},
			OptionalPrompt: &TriggerOptionalPrompt{Question: "Fire Optional Pal trigger?"},
		}}
	})

	pendingBefore := len(g.PendingTriggers)
	choicesBefore := len(g.PendingChoices)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID, Actor: owner.ID})
	})

	if buildCalls != 0 {
		t.Errorf("Build calls before resolve: got %d, want 0 (deferred until apply=true)", buildCalls)
	}
	if got := len(g.PendingTriggers) - pendingBefore; got != 0 {
		t.Errorf("PendingTriggers delta before resolve: got %d, want 0", got)
	}
	if got := len(g.PendingChoices) - choicesBefore; got != 1 {
		t.Fatalf("PendingChoices delta: got %d, want 1", got)
	}
	choice := g.PendingChoices[len(g.PendingChoices)-1]
	if choice.Kind != PendingChoiceTriggerPrompt {
		t.Errorf("queued choice kind: got %q, want %q", choice.Kind, PendingChoiceTriggerPrompt)
	}
	if choice.Chooser != owner.ID {
		t.Errorf("queued choice chooser: got %s, want %s (source.Controller default)", choice.Chooser, owner.ID)
	}
	if choice.Reason != "Fire Optional Pal trigger?" {
		t.Errorf("queued choice reason: got %q, want %q", choice.Reason, "Fire Optional Pal trigger?")
	}
	if choice.Source != cardID {
		t.Errorf("queued choice source: got %s, want %s", choice.Source, cardID)
	}

	if err := g.ResolveTriggerPrompt(choice.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt(apply=true): %v", err)
	}
	if buildCalls != 1 {
		t.Errorf("Build calls after apply=true: got %d, want 1", buildCalls)
	}
	if got := len(g.PendingTriggers) - pendingBefore; got != 1 {
		t.Errorf("PendingTriggers delta after apply=true: got %d, want 1", got)
	}
	if got := len(g.PendingChoices) - choicesBefore; got != 0 {
		t.Errorf("PendingChoices delta after resolve: got %d, want 0 (entry should be dequeued)", got)
	}
}

// TestTriggerHarvesterOptionalPromptDeclineDrops confirms that
// `apply: false` discards the trigger without invoking Build —
// PendingTriggers stays empty even though the prompt fired.
func TestTriggerHarvesterOptionalPromptDeclineDrops(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-optional-decline-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	var buildCalls int
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				buildCalls++
				return &StackItem{SourceCardID: source.InstanceID}
			},
			OptionalPrompt: &TriggerOptionalPrompt{Question: "Decline me?"},
		}}
	})

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID, Actor: owner.ID})
	})
	if len(g.PendingChoices) == 0 {
		t.Fatalf("expected a queued PendingChoice")
	}
	choice := g.PendingChoices[len(g.PendingChoices)-1]

	pendingBefore := len(g.PendingTriggers)
	if err := g.ResolveTriggerPrompt(choice.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt(apply=false): %v", err)
	}
	if buildCalls != 0 {
		t.Errorf("Build calls after apply=false: got %d, want 0", buildCalls)
	}
	if got := len(g.PendingTriggers) - pendingBefore; got != 0 {
		t.Errorf("PendingTriggers delta after decline: got %d, want 0", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("PendingChoices after decline: got %d, want 0", len(g.PendingChoices))
	}
}

// TestTriggerHarvesterOptionalPromptChooserOverride confirms the
// Chooser override field routes the prompt to a player other than
// source.Controller (the smothering-tithe-style "opponent decides"
// shape — the first card to use it ships in a later sub-PR).
func TestTriggerHarvesterOptionalPromptChooserOverride(t *testing.T) {
	g := newActiveGame(t)
	source := g.Seats[0]
	opponent := g.Seats[1]
	const oracle = "test-optional-override-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Owner:      source.ID,
		Controller: source.ID,
	})

	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			Build: func(_ Event, src *Card, _ Characteristic, _ *Game) *StackItem {
				return &StackItem{SourceCardID: src.InstanceID}
			},
			OptionalPrompt: &TriggerOptionalPrompt{
				Question: "Opponent decides",
				Chooser: func(_ Event, _ *Card, g *Game) uuid.UUID {
					if len(g.Seats) >= 2 {
						return g.Seats[1].ID
					}
					return uuid.Nil
				},
			},
		}}
	})

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID, Actor: source.ID})
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("PendingChoices: got %d, want 1", len(g.PendingChoices))
	}
	if g.PendingChoices[0].Chooser != opponent.ID {
		t.Errorf("Chooser override: got %s, want %s", g.PendingChoices[0].Chooser, opponent.ID)
	}
}

// TestPendingChoiceKindForTriggerPrompt covers the lookup helper the
// resolve_choice action dispatcher uses to disambiguate between
// PendingChoiceOptionalReplacement (S17) and PendingChoiceTriggerPrompt
// (S19) — both consume the same `{apply: bool}` payload, so the
// dispatcher routes by querying the choice's kind.
func TestPendingChoiceKindForTriggerPrompt(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-kind-lookup-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			Build: func(_ Event, src *Card, _ Characteristic, _ *Game) *StackItem {
				return &StackItem{SourceCardID: src.InstanceID}
			},
			OptionalPrompt: &TriggerOptionalPrompt{Question: "Pick me"},
		}}
	})

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID, Actor: owner.ID})
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("PendingChoices: got %d, want 1", len(g.PendingChoices))
	}
	id := g.PendingChoices[0].ID

	kind, ok := g.PendingChoiceKindFor(id)
	if !ok {
		t.Fatalf("PendingChoiceKindFor: ok=false for live id")
	}
	if kind != PendingChoiceTriggerPrompt {
		t.Errorf("kind: got %q, want %q", kind, PendingChoiceTriggerPrompt)
	}

	if _, ok := g.PendingChoiceKindFor(uuid.New()); ok {
		t.Errorf("PendingChoiceKindFor: ok=true for unknown id")
	}
}

// TestTriggerHarvesterNoCatalogIsNoop confirms graceful degradation:
// without CatalogTriggers wired (nil hook), the harvester runs but
// short-circuits before touching the battlefield. Production runs
// always wire the hook via cards/effects/wire.go init, but tests
// that don't import the catalog leave it nil.
func TestTriggerHarvesterNoCatalogIsNoop(t *testing.T) {
	g := newActiveGame(t)
	withCatalogTriggers(t, nil)

	pendingBefore := len(g.PendingTriggers)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: uuid.New()})
		g.EmitEvent(Event{Kind: EventLTB, CardID: uuid.New()})
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: g.Seats[0].ID})
	})
	if got := len(g.PendingTriggers) - pendingBefore; got != 0 {
		t.Errorf("PendingTriggers delta with nil CatalogTriggers: got %d, want 0", got)
	}
}
