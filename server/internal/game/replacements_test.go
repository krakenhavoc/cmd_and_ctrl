package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// replacements_test.go covers the S17 sub-PR 2 engine skeleton:
// core apply-loop, once-per-event tracking, commander-zone
// built-in (refactored from S13.1), CR 616 order prompt, and
// ResolveReplacementOrder validation. No catalog cards register
// replacements yet — coverage here uses RegisterReplacementForTest
// to inject synthetic replacements directly on the game.

// TestReplacementsZeroPassthrough — with no catalog / test
// replacements registered, every mutation flows byte-for-byte
// identically to pre-S17 behavior. The commander-zone built-in is
// present but only fires for asCommanderMove + commander cards.
func TestReplacementsZeroPassthrough(t *testing.T) {
	g := newActiveGame(t)

	// AddCounter on a plain card.
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := g.Battlefield.Cards[0].Counters["+1/+1"]; got != 1 {
		t.Errorf("counter = %d, want 1", got)
	}

	// ChangePlayerLife.
	newLife, err := g.ChangePlayerLife(owner.ID, -3)
	if err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
	if newLife != 37 {
		t.Errorf("life = %d, want 37", newLife)
	}
}

// TestCommanderZoneBuiltInRoutesToCommandZone verifies the S17
// refactor preserves S13.1 semantics: a commander moved with
// asCommander=true routes to the owner's command zone.
func TestCommanderZoneBuiltInRoutesToCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID:  cmdID,
		Name:        "Atraxa",
		TypeLine:    "Legendary Creature — Angel",
		Power:       4,
		Toughness:   4,
		Owner:       owner.ID,
		Controller:  owner.ID,
		IsCommander: true,
	})

	err := g.MoveCardByIDAsCommander(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID},
		cmdID,
		true,
	)
	if err != nil {
		t.Fatalf("MoveCardByIDAsCommander: %v", err)
	}
	if !owner.Command.Contains(cmdID) {
		t.Errorf("commander did not land in command zone")
	}
	if owner.Graveyard.Contains(cmdID) {
		t.Errorf("commander leaked into graveyard")
	}
}

// TestCommanderZoneBuiltInSkippedWhenAsCommanderFalse verifies the
// built-in's AppliesTo gate — if the caller didn't flag the move,
// the commander goes to graveyard like any other card. Preserves
// the S13.1 semantic.
func TestCommanderZoneBuiltInSkippedWhenAsCommanderFalse(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID:  cmdID,
		Name:        "Atraxa",
		TypeLine:    "Legendary Creature — Angel",
		Power:       4,
		Toughness:   4,
		Owner:       owner.ID,
		Controller:  owner.ID,
		IsCommander: true,
	})

	err := g.MoveCardByIDAsCommander(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID},
		cmdID,
		false,
	)
	if err != nil {
		t.Fatalf("MoveCardByIDAsCommander: %v", err)
	}
	if !owner.Graveyard.Contains(cmdID) {
		t.Errorf("commander did not land in graveyard (asCommander=false)")
	}
	if owner.Command.Contains(cmdID) {
		t.Errorf("commander routed to command zone without asCommander flag")
	}
}

// TestReplacementInjectionMutatesCounterDelta wires a synthetic
// "doubler" replacement into the game and verifies the counter
// placement multiplies through the pipeline.
func TestReplacementInjectionMutatesCounterDelta(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterName == "+1/+1"
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta *= 2
			return nil
		},
		Label: "Test doubler",
	})
	g.mu.Unlock()

	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := g.Battlefield.Cards[0].Counters["+1/+1"]; got != 2 {
		t.Errorf("counter = %d, want 2 (doubled)", got)
	}
}

// TestReplacementCancelsEvent — a cancel-returning replacement
// suppresses the mutation entirely. No event, no mutation.
func TestReplacementCancelsEvent(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		Label: "Solemnity-style cancel",
	})
	g.mu.Unlock()

	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := g.Battlefield.Cards[0].Counters["+1/+1"]; got != 0 {
		t.Errorf("counter = %d, want 0 (canceled)", got)
	}
}

// TestReplacementOncePerEvent — once an effect fires on an event,
// it can't fire a second time for the same event (CR 614.5 /
// 616.1). Proven by an always-applying replacement that would
// otherwise spin.
func TestReplacementOncePerEvent(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	fireCount := 0
	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			fireCount++
			ev.CounterDelta += 1
			return nil
		},
		Label: "Always-applies",
	})
	g.mu.Unlock()

	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if fireCount != 1 {
		t.Errorf("replacement fired %d times, want 1 (once-per-event)", fireCount)
	}
	if got := g.Battlefield.Cards[0].Counters["+1/+1"]; got != 2 {
		t.Errorf("counter = %d, want 2 (1 original + 1 replacement bump)", got)
	}
}

// TestReplacementOrderPromptQueues — two simultaneously applicable
// replacements trigger a CR 616 prompt queued on
// g.PendingChoices. No mutation runs until the chooser resolves.
func TestReplacementOrderPromptQueues(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	g.mu.Lock()
	// Two replacements that both apply. Apply functions mutate ev
	// rather than cancel, so the apply-loop would iterate through
	// both — but CR 616 queues a prompt first.
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta *= 2
			return nil
		},
		Label: "Doubler",
	})
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta += 1
			return nil
		},
		Label: "Adder",
	})
	g.mu.Unlock()

	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}

	// No mutation yet — the prompt is queued.
	if got := g.Battlefield.Cards[0].Counters["+1/+1"]; got != 0 {
		t.Errorf("counter = %d, want 0 (prompt pending)", got)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(g.PendingChoices))
	}
	pc := g.PendingChoices[0]
	if pc.Kind != PendingChoiceReplacementOrder {
		t.Errorf("pending choice kind = %q, want %q", pc.Kind, PendingChoiceReplacementOrder)
	}
	if len(pc.ReplacementEffectIDs) != 2 {
		t.Errorf("ReplacementEffectIDs len = %d, want 2", len(pc.ReplacementEffectIDs))
	}
	if pc.Chooser != owner.ID {
		t.Errorf("chooser = %s, want %s (counter target controller)", pc.Chooser, owner.ID)
	}
}

// TestResolveReplacementOrderRejectsWrongChooser — the prompt
// names a specific chooser; any other player submission bounces
// with ErrNotTheChooser.
func TestResolveReplacementOrderRejectsWrongChooser(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	other := g.Seats[1]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.mu.Lock()
	for i := 0; i < 2; i++ {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventCounterPlaced},
			AppliesTo: func(_ *ReplacementEvent, _ *Game, _ *Card) bool { return true },
			Replace:   func(_ *ReplacementEvent, _ *Game, _ *Card) error { return nil },
			Label:     "test",
		})
	}
	g.mu.Unlock()
	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	pc := g.PendingChoices[0]

	err := g.ResolveReplacementOrder(pc.ID, other.ID, pc.ReplacementEffectIDs)
	if !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("err = %v, want ErrNotTheChooser", err)
	}
}

// TestResolveReplacementOrderRejectsWrongLength — mismatched order
// length bounces with ErrInvalidParam.
func TestResolveReplacementOrderRejectsWrongLength(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.mu.Lock()
	for i := 0; i < 2; i++ {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventCounterPlaced},
			AppliesTo: func(_ *ReplacementEvent, _ *Game, _ *Card) bool { return true },
			Replace:   func(_ *ReplacementEvent, _ *Game, _ *Card) error { return nil },
			Label:     "test",
		})
	}
	g.mu.Unlock()
	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	pc := g.PendingChoices[0]

	err := g.ResolveReplacementOrder(pc.ID, owner.ID, []ReplacementEffectID{pc.ReplacementEffectIDs[0]})
	if !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam (short order)", err)
	}
}

// TestResolveReplacementOrderRejectsWrongIDs — permutation check
// catches IDs that weren't in the original set.
func TestResolveReplacementOrderRejectsWrongIDs(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.mu.Lock()
	for i := 0; i < 2; i++ {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventCounterPlaced},
			AppliesTo: func(_ *ReplacementEvent, _ *Game, _ *Card) bool { return true },
			Replace:   func(_ *ReplacementEvent, _ *Game, _ *Card) error { return nil },
			Label:     "test",
		})
	}
	g.mu.Unlock()
	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	pc := g.PendingChoices[0]

	// Same count but fabricate an unknown ID.
	err := g.ResolveReplacementOrder(pc.ID, owner.ID, []ReplacementEffectID{pc.ReplacementEffectIDs[0], 99999})
	if !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam (bad ID)", err)
	}
}

// TestReplacementEffectIDRoundTrip — the wire form is stable.
func TestReplacementEffectIDRoundTrip(t *testing.T) {
	cases := []ReplacementEffectID{0, 1, 42, 123456789, 1 << 40, 1 << 62}
	for _, want := range cases {
		s := ReplacementEffectIDToString(want)
		got, err := ReplacementEffectIDFromString(s)
		if err != nil {
			t.Errorf("%d: parse %q: %v", want, s, err)
			continue
		}
		if got != want {
			t.Errorf("%d → %q → %d", want, s, got)
		}
	}
}

// TestReplacementOptionMetaForBuiltin — the view layer can recover
// the built-in's label for the prompt wire projection.
func TestReplacementOptionMetaForBuiltin(t *testing.T) {
	g := NewGame()
	// Built-in 0 = commander-zone.
	label, srcID := g.ReplacementOptionMetaForEffect(builtinReplacementIDBase)
	if label == "" {
		t.Errorf("built-in label empty")
	}
	if srcID != (uuid.UUID{}) {
		t.Errorf("built-in srcID = %v, want zero", srcID)
	}
}
