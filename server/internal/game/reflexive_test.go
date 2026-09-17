package game

import (
	"testing"

	"github.com/google/uuid"
)

// reflexive_test.go covers the CR 603.12 reflexive trigger seam
// (#636): a trigger created by a resolving effect goes on the stack
// above whatever the parent was sitting on, chooses its targets when
// it gets there (CR 603.3d) rather than when the parent was
// announced, asks its "you may" question, is dropped without a prompt
// when nothing is legal, survives Clone / RestoreFrom, and works when
// its source card has already left.

// queueAbility puts an ability item on the stack the way the
// harvester does — onto PendingTriggers, then drained at the next
// state-check boundary. One item at a time, so the CR 603.3b ordering
// prompt never opens and the stack order is the call order.
func queueAbility(g *Game, controller, source uuid.UUID, label string, effect func(*Game, *StackItem) error) *StackItem {
	item := &StackItem{
		Kind:         StackItemTriggered,
		Controller:   controller,
		Owner:        controller,
		SourceCardID: source,
		Label:        label,
		Effect:       effect,
	}
	g.WithWriteLock(func() {
		g.queueHarvestedTriggerLocked(item)
		g.runStateChecksLocked()
	})
	return item
}

// creatureSpec is "target creature" — the clause a reflexive trigger
// in these tests picks with.
func creatureSpec() *TargetSpec {
	return &TargetSpec{
		Mode:   "creature",
		Label:  "target creature",
		Zones:  []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsCreature() },
		Min:    1, Max: 1,
	}
}

// pickTargetPrompt returns the outstanding pick_target choice, or nil.
func pickTargetPrompt(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePickTarget {
			return c
		}
	}
	return nil
}

// triggerPrompt returns the outstanding "you may" choice, or nil.
func triggerPrompt(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceTriggerPrompt {
			return c
		}
	}
	return nil
}

// TestReflexiveTriggerLandsAboveTheParentAndWaitsForPriority is the
// whole point of the seam: the follow-up is a separate stack item
// that resolves BEFORE whatever the parent was sitting on, and only
// after everyone has passed — not inside the parent's resolution.
func TestReflexiveTriggerLandsAboveTheParentAndWaitsForPriority(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	source := pushCreatureToBattlefield(t, g, g.Seats[0])

	var order []string
	// Beneath: an unrelated ability already on the stack.
	queueAbility(g, me, source, "beneath", func(*Game, *StackItem) error {
		order = append(order, "beneath")
		return nil
	})
	// Above it: the parent, which creates the reflexive trigger.
	queueAbility(g, me, source, "parent", func(g *Game, item *StackItem) error {
		order = append(order, "parent")
		g.QueueReflexiveTriggerForEffect(item, ReflexiveTrigger{
			Label: "reflexive — when you do",
			Effect: func(*Game, *StackItem) error {
				order = append(order, "reflexive")
				return nil
			},
		})
		return nil
	})
	if len(g.StackMeta) != 2 {
		t.Fatalf("setup: %d items on the stack, want 2", len(g.StackMeta))
	}

	// One pass around the table resolves the parent.
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(order) != 1 || order[0] != "parent" {
		t.Fatalf("after one round of passes, order=%v, want [parent]", order)
	}
	// The response window: the reflexive trigger is ON the stack and
	// has not resolved. Folding it into the parent is what this test
	// exists to catch.
	if len(g.StackMeta) != 2 {
		t.Fatalf("stack holds %d items, want the reflexive trigger above the one beneath", len(g.StackMeta))
	}
	var reflexive *StackItem
	for _, it := range g.StackMeta {
		if it.Label == "reflexive — when you do" {
			reflexive = it
		}
	}
	if reflexive == nil {
		t.Fatal("no reflexive trigger on the stack")
	}
	if reflexive.Kind != StackItemTriggered {
		t.Errorf("reflexive item Kind=%q, want %q", reflexive.Kind, StackItemTriggered)
	}
	if reflexive.Controller != me || reflexive.SourceCardID != source {
		t.Errorf("reflexive item controller/source = %v/%v, want %v/%v",
			reflexive.Controller, reflexive.SourceCardID, me, source)
	}

	settleStack(t, g)
	want := []string{"parent", "reflexive", "beneath"}
	if len(order) != len(want) {
		t.Fatalf("order=%v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order=%v, want %v", order, want)
		}
	}
}

// TestReflexiveTriggerPicksItsTargetWhenItGoesOnTheStack: CR 603.3d.
// The legal set is computed when the trigger is created — after the
// parent resolved — and the controller answers a pick_target prompt,
// not the parent's announce.
func TestReflexiveTriggerPicksItsTargetWhenItGoesOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	source := pushCreatureToBattlefield(t, g, g.Seats[0])

	var chosen uuid.UUID
	var latecomer uuid.UUID
	queueAbility(g, me, source, "parent", func(g *Game, item *StackItem) error {
		// The parent's resolution creates the creature the trigger
		// will target — which is precisely what a clause on the
		// parent could not have seen.
		c := NewCard("Latecomer", g.Seats[1].ID)
		c.TypeLine = "Creature — Test"
		c.Power, c.Toughness = 1, 1
		g.Battlefield.PushTop(c)
		latecomer = c.InstanceID
		g.QueueReflexiveTriggerForEffect(item, ReflexiveTrigger{
			Label:   "reflexive — target creature",
			Targets: creatureSpec(),
			Effect: func(_ *Game, it *StackItem) error {
				if len(it.Targets) > 0 {
					chosen = it.Targets[0].ID
				}
				return nil
			},
		})
		return nil
	})
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}

	prompt := pickTargetPrompt(g)
	if prompt == nil {
		t.Fatal("no pick_target prompt after the parent resolved")
	}
	if prompt.Chooser != me {
		t.Errorf("prompt chooser %v, want the parent's controller %v", prompt.Chooser, me)
	}
	if prompt.Reason != "target creature" {
		t.Errorf("prompt label %q, want the clause's label", prompt.Reason)
	}
	// The set was computed NOW: the creature the parent just made is
	// in it.
	found := false
	for _, id := range prompt.PickTargetCards {
		if id == latecomer {
			found = true
		}
	}
	if !found {
		t.Errorf("legal set %v does not hold the creature the parent created (%v)", prompt.PickTargetCards, latecomer)
	}
	// Nothing is on the stack yet — the trigger is being put there.
	if len(g.StackMeta) != 0 {
		t.Errorf("stack holds %d items while the target is being chosen", len(g.StackMeta))
	}

	if err := g.ResolvePickTarget(prompt.ID, me, TargetRef{Kind: TargetCard, ID: latecomer}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	settleStack(t, g)
	if chosen != latecomer {
		t.Errorf("trigger resolved against %v, want %v", chosen, latecomer)
	}
}

// TestReflexiveTriggerWithNoLegalTargetIsDropped: CR 603.3d — a
// targeted trigger with nothing legal to target is never put on the
// stack, and the controller is not asked.
func TestReflexiveTriggerWithNoLegalTargetIsDropped(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	// An empty battlefield, so "target creature" has an empty legal
	// set at the moment the trigger would be created.
	g.WithWriteLock(func() { g.Battlefield.Cards = nil })

	fired := 0
	queueAbility(g, me, uuid.Nil, "parent", func(g *Game, item *StackItem) error {
		g.QueueReflexiveTriggerForEffect(item, ReflexiveTrigger{
			Label:   "reflexive — target creature",
			Targets: creatureSpec(),
			Effect: func(*Game, *StackItem) error {
				fired++
				return nil
			},
		})
		return nil
	})
	settleStack(t, g)

	if p := pickTargetPrompt(g); p != nil {
		t.Errorf("a target prompt was opened with no legal targets: %+v", p)
	}
	if fired != 0 {
		t.Errorf("the trigger resolved %d times with no legal target, want 0", fired)
	}
	if len(g.StackMeta)+len(g.PendingTriggers) != 0 {
		t.Errorf("the dropped trigger left something on the stack")
	}
}

// TestReflexiveTriggerOptionalAsksFirst: a reflexive trigger whose
// own text says "you may" queues the CR 603.5 prompt; "no" drops it,
// "yes" puts it on the stack.
func TestReflexiveTriggerOptionalAsksFirst(t *testing.T) {
	for _, tc := range []struct {
		name  string
		apply bool
		want  int
	}{
		{"declined", false, 0},
		{"accepted", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0].ID
			fired := 0
			queueAbility(g, me, uuid.Nil, "parent", func(g *Game, item *StackItem) error {
				g.QueueReflexiveTriggerForEffect(item, ReflexiveTrigger{
					Label:    "reflexive — you may",
					Optional: &TriggerOptionalPrompt{Question: "do the thing?"},
					Effect: func(*Game, *StackItem) error {
						fired++
						return nil
					},
				})
				return nil
			})
			for i := 0; i < len(g.Seats); i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			prompt := triggerPrompt(g)
			if prompt == nil {
				t.Fatal("no trigger prompt for an optional reflexive trigger")
			}
			if prompt.Chooser != me || prompt.Reason != "do the thing?" {
				t.Errorf("prompt chooser/question = %v/%q, want %v/%q", prompt.Chooser, prompt.Reason, me, "do the thing?")
			}
			if fired != 0 {
				t.Fatal("the effect ran before the question was answered")
			}
			if err := g.ResolveTriggerPrompt(prompt.ID, me, tc.apply); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			settleStack(t, g)
			if fired != tc.want {
				t.Errorf("effect ran %d times, want %d", fired, tc.want)
			}
		})
	}
}

// TestReflexiveTriggerCarriesItsPayload: what the parent had to tell
// the trigger rides item.Payload, and a target pick does NOT clobber
// it — the two slots are separate for exactly this reason.
func TestReflexiveTriggerCarriesItsPayload(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	victim := pushCreatureToBattlefield(t, g, g.Seats[1])
	carried := uuid.New()

	var sawPayload []TargetRef
	var sawTarget uuid.UUID
	queueAbility(g, me, uuid.Nil, "parent", func(g *Game, item *StackItem) error {
		g.QueueReflexiveTriggerForEffect(item, ReflexiveTrigger{
			Label:   "reflexive — payload plus target",
			Targets: creatureSpec(),
			Payload: []TargetRef{{Kind: TargetCard, ID: carried}, {Kind: TargetPlayer, ID: me}},
			Effect: func(_ *Game, it *StackItem) error {
				sawPayload = it.Payload
				if len(it.Targets) > 0 {
					sawTarget = it.Targets[0].ID
				}
				return nil
			},
		})
		return nil
	})
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := pickTargetPrompt(g)
	if prompt == nil {
		t.Fatal("no pick_target prompt")
	}
	if err := g.ResolvePickTarget(prompt.ID, me, TargetRef{Kind: TargetCard, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	settleStack(t, g)

	if len(sawPayload) != 2 || sawPayload[0].ID != carried || sawPayload[1].ID != me {
		t.Errorf("payload seen at resolution = %v, want the card then the player", sawPayload)
	}
	if sawTarget != victim {
		t.Errorf("target seen at resolution = %v, want %v", sawTarget, victim)
	}
}

// TestReflexiveTriggerSurvivesCloneAndRestore is the undo contract:
// an item queued by a resolving effect is an ordinary StackItem, so
// the snapshot carries it — payload and all — and it resolves against
// the RESTORED game rather than the one it was created in.
func TestReflexiveTriggerSurvivesCloneAndRestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	carried := uuid.New()

	// Package-level-shaped effect: it takes the live game handed in
	// and reads its input off the item, capturing neither.
	resolvedIn := map[*Game]int{}
	var payloadSeen uuid.UUID
	effect := func(live *Game, it *StackItem) error {
		resolvedIn[live]++
		if len(it.Payload) > 0 {
			payloadSeen = it.Payload[0].ID
		}
		return nil
	}
	queueAbility(g, me, uuid.Nil, "parent", func(g *Game, item *StackItem) error {
		g.QueueReflexiveTriggerForEffect(item, ReflexiveTrigger{
			Label:   "reflexive — survives undo",
			Payload: []TargetRef{{Kind: TargetCard, ID: carried}},
			Effect:  effect,
		})
		return nil
	})
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("stack holds %d items, want the reflexive trigger", len(g.StackMeta))
	}

	snap := g.Clone()
	var snapItem *StackItem
	for _, it := range snap.StackMeta {
		snapItem = it
	}
	if snapItem == nil {
		t.Fatal("the clone lost the reflexive trigger")
	}
	if len(snapItem.Payload) != 1 || snapItem.Payload[0].ID != carried {
		t.Fatalf("clone payload = %v, want [%v]", snapItem.Payload, carried)
	}
	for _, it := range g.StackMeta {
		if &it.Payload[0] == &snapItem.Payload[0] {
			t.Error("clone aliased the original item's payload array")
		}
	}

	// Resolve it in the live game, then undo back to the snapshot and
	// resolve it again — against the restored game.
	settleStack(t, g)
	if resolvedIn[g] != 1 {
		t.Fatalf("live game resolved the trigger %d times, want 1", resolvedIn[g])
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if len(g.StackMeta) != 1 {
		t.Fatalf("RestoreFrom did not put the reflexive trigger back: %d items", len(g.StackMeta))
	}
	payloadSeen = uuid.Nil
	settleStack(t, g)
	if resolvedIn[g] != 2 {
		t.Errorf("restored trigger resolved against the live game %d times, want 2", resolvedIn[g])
	}
	if payloadSeen != carried {
		t.Errorf("restored trigger read payload %v, want %v", payloadSeen, carried)
	}
}

// TestReflexiveTriggerPayloadSurvivesASnapshot is the other half of
// the payload contract: it is game state, so it has to cross the
// JSON snapshot a deploy restores from, not just Clone. The Effect
// does not — that is what the continuation census is for — but a
// payload that vanished would leave a restored trigger resolving
// against nothing.
func TestReflexiveTriggerPayloadSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	carried := uuid.New()
	queueAbility(g, me, uuid.Nil, "parent", func(g *Game, item *StackItem) error {
		g.QueueReflexiveTriggerForEffect(item, ReflexiveTrigger{
			Label:   "reflexive — survives a restart",
			Payload: []TargetRef{{Kind: TargetCard, ID: carried}},
			Effect:  func(*Game, *StackItem) error { return nil },
		})
		return nil
	})
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}

	_, restored := roundTrip(t, g)
	var item *StackItem
	for _, it := range restored.StackMeta {
		if it.Label == "reflexive — survives a restart" {
			item = it
		}
	}
	if item == nil {
		t.Fatal("the restored game lost the reflexive trigger")
	}
	if len(item.Payload) != 1 || item.Payload[0].ID != carried || item.Payload[0].Kind != TargetCard {
		t.Errorf("restored payload = %v, want one card ref for %v", item.Payload, carried)
	}
}

// TestReflexiveTriggerFromADepartedSource: a sorcery's reflexive
// trigger has its source in a graveyard, and a land that sacrificed
// itself has one there too. The trigger is the resolving ability's,
// not the permanent's, so it still happens.
func TestReflexiveTriggerFromADepartedSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	source := pushCreatureToBattlefield(t, g, g.Seats[0])

	fired := 0
	var sawSource, sawController uuid.UUID
	queueAbility(g, me, source, "parent", func(g *Game, item *StackItem) error {
		// The source leaves during its own resolution.
		c, err := g.Battlefield.Remove(item.SourceCardID)
		if err != nil {
			return err
		}
		g.Seats[0].Graveyard.PushTop(c)
		g.QueueReflexiveTriggerForEffect(item, ReflexiveTrigger{
			Label: "reflexive — from a graveyard",
			Effect: func(_ *Game, it *StackItem) error {
				fired++
				sawSource, sawController = it.SourceCardID, it.Controller
				return nil
			},
		})
		return nil
	})
	settleStack(t, g)

	if fired != 1 {
		t.Fatalf("the trigger fired %d times from a departed source, want 1", fired)
	}
	if sawSource != source || sawController != me {
		t.Errorf("item source/controller = %v/%v, want %v/%v", sawSource, sawController, source, me)
	}
}

// TestQueueReflexiveTriggerRejectsMalformed: a declaration with no
// parent, no effect or no label is dropped rather than queued, the
// same posture ScheduleDelayedTriggerForEffect takes.
func TestQueueReflexiveTriggerRejectsMalformed(t *testing.T) {
	g := newActiveGame(t)
	parent := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, Controller: g.Seats[0].ID}
	noop := func(*Game, *StackItem) error { return nil }
	g.WithWriteLock(func() {
		if g.QueueReflexiveTriggerForEffect(nil, ReflexiveTrigger{Label: "x", Effect: noop}) {
			t.Error("queued a reflexive trigger with no parent item")
		}
		if g.QueueReflexiveTriggerForEffect(parent, ReflexiveTrigger{Label: "x"}) {
			t.Error("queued a reflexive trigger with no Effect")
		}
		if g.QueueReflexiveTriggerForEffect(parent, ReflexiveTrigger{Effect: noop}) {
			t.Error("queued a reflexive trigger with no Label")
		}
	})
	if len(g.PendingTriggers)+len(g.PendingChoices) != 0 {
		t.Errorf("a malformed declaration left %d triggers and %d choices behind",
			len(g.PendingTriggers), len(g.PendingChoices))
	}
}
