package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_event_test.go — #1223, the triggering event as data on the
// item (CR 603.2, CR 603.10).
//
// Each test drives the real harvester with a real event, because the
// claim being tested is about WHERE the value comes from — a test
// that stamped the context by hand would prove nothing about the
// dispatch that has to stamp it.

// triggerCatalog installs `abilities` as the catalog's triggers for
// every card and restores the previous hook when the test ends.
func triggerCatalog(t *testing.T, abilities ...TriggeredAbility) {
	t.Helper()
	prev := CatalogTriggers
	CatalogTriggers = func(string) []TriggeredAbility { return abilities }
	t.Cleanup(func() { CatalogTriggers = prev })
}

// watcherNamed is a battlefield permanent with a catalog key, so
// TriggersForCard finds the installed triggers for it.
func watcherNamed(g *Game, controller uuid.UUID, name string) uuid.UUID {
	c := NewCard(name, controller)
	c.TypeLine = "Artifact"
	c.Controller = controller
	c.OracleID = "watcher-" + name
	return pushPermanent(g, c)
}

// queuedTrigger is the single item the harvester queued, or nil.
func queuedTrigger(g *Game) *StackItem {
	if len(g.PendingTriggers) != 1 {
		return nil
	}
	return g.PendingTriggers[0]
}

// onlyAbilityOnStack is the single ability item in StackMeta, or nil.
// A trigger that finished its CR 603.3d target walk has already been
// drained onto the stack by the time the answering call returns.
func onlyAbilityOnStack(g *Game) *StackItem {
	var out *StackItem
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind != StackItemSpell {
				out = it
			}
		}
	})
	return out
}

// pushAbilityItem puts an ability item on the stack by hand and
// returns it. `kind` is StackItemActivated or StackItemTriggered.
func pushAbilityItem(g *Game, kind StackItemKind, controller uuid.UUID, source uuid.UUID, label string, effect func(*Game, *StackItem) error) *StackItem {
	item := &StackItem{
		ID:           uuid.New(),
		Kind:         kind,
		Controller:   controller,
		Owner:        controller,
		SourceCardID: source,
		Label:        label,
		Effect:       effect,
	}
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*StackItem)
		}
		item.Seq = g.nextStackSeqLocked()
		g.StackMeta[item.ID] = item
	})
	return item
}

// bearNamed is a 2/2 creature card for the targeting tests.
func bearNamed(controller uuid.UUID, name string) Card {
	c := NewCard(name, controller)
	c.TypeLine = "Creature — Bear"
	c.Controller = controller
	c.Power, c.Toughness = 2, 2
	return c
}

// TestATriggersItemCarriesTheTriggeringEvent — the damage amount, the
// printed shape this seam was filed for ("whenever ~ deals damage …
// that much").
func TestATriggersItemCarriesTheTriggeringEvent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	triggerCatalog(t, TriggeredAbility{
		Watches:   []EventKind{EventDealDamage},
		AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "watcher", nil)
		},
	})
	watcherNamed(g, me.ID, "Watcher")

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 6, Actor: me.ID, Combat: true})
	})

	item := queuedTrigger(g)
	if item == nil {
		t.Fatal("the trigger did not fire")
	}
	if item.Trigger == nil {
		t.Fatal("the item carries no triggering event")
	}
	if got := item.Trigger.Amount(); got != 6 {
		t.Errorf("Amount() = %d, want 6", got)
	}
	if !item.Trigger.Event.Combat {
		t.Error("the whole event travels, Combat included")
	}
	if !item.Trigger.Fired() {
		t.Error("Fired() is false for a real triggering event")
	}
}

// TestADiesTriggerCarriesTheObjectAsItWas — CR 603.10. The creature
// is in a graveyard by the time anything reads this, and the numbers
// that matter are the ones it had on the battlefield.
func TestADiesTriggerCarriesTheObjectAsItWas(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	triggerCatalog(t, TriggeredAbility{
		Watches:   []EventKind{EventLTB},
		AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "watcher", nil)
		},
	})
	watcherNamed(g, me.ID, "Watcher")

	dying := NewCard("Grizzly Bears", me.ID)
	dying.TypeLine = "Creature — Bear"
	dying.Controller = me.ID
	dying.Power, dying.Toughness = 2, 2
	dying.ManaCost = "{1}{G}"
	dyingID := pushPermanent(g, dying)

	g.WithWriteLock(func() {
		g.recomputeLayersLocked()
		// The real exit, so the LKI snapshot is taken the way every
		// battlefield exit in the engine takes it.
		g.battlefieldExitLocked(dyingID)
		card, err := g.Battlefield.Remove(dyingID)
		if err != nil {
			t.Fatalf("Battlefield.Remove: %v", err)
		}
		me.Graveyard.PushTop(card)
		g.EmitEvent(Event{
			Kind: EventLTB, CardID: dyingID, Actor: me.ID,
			OldZone: ZoneBattlefield, NewZone: ZoneGraveyard,
		})
	})

	item := queuedTrigger(g)
	if item == nil || item.Trigger == nil {
		t.Fatal("the dies-trigger did not fire, or carried no event")
	}
	obj := item.Trigger.Object
	if obj == nil {
		t.Fatal("the dies-trigger carried no object snapshot")
	}
	if obj.ID != dyingID {
		t.Errorf("snapshot names %v, want the card that left (%v)", obj.ID, dyingID)
	}
	if obj.ManaValue != 2 {
		t.Errorf("mana value = %d, want 2 — the number Scrap Trawler's clause reads", obj.ManaValue)
	}
	if !obj.HasType("creature") {
		t.Errorf("types = %v, want the creature it was", obj.Types)
	}
	if obj.Name != "Grizzly Bears" {
		t.Errorf("name = %q, want the name it had", obj.Name)
	}
	if obj.Controller != me.ID {
		t.Errorf("controller = %v, want the seat that controlled it", obj.Controller)
	}
}

// TestATriggersEventSurvivesACloneAndARestore — the field is data, so
// an undo and a deploy restore both bring it back. Before #1223 the
// event lived in a closure, which has no wire form at all.
func TestATriggersEventSurvivesACloneAndARestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Source", me.ID))
	item := pushAbilityItem(g, StackItemTriggered, me.ID, src, "trigger", nil)
	g.WithWriteLock(func() {
		item.Trigger = &TriggerContext{
			Event:  Event{Kind: EventDealDamage, Amount: 9},
			Object: &ObjectSnapshot{Name: "Bear", Types: []string{"creature"}, ManaValue: 2},
		}
	})

	clone := g.Clone()
	var cloned *StackItem
	clone.WithWriteLock(func() { cloned = clone.StackMeta[item.ID] })
	if cloned == nil || cloned.Trigger == nil || cloned.Trigger.Amount() != 9 {
		t.Fatalf("the clone lost the triggering event: %+v", cloned)
	}
	if cloned.Trigger == item.Trigger {
		t.Error("the clone shares the live item's context pointer")
	}

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	var back *StackItem
	restored.WithWriteLock(func() { back = restored.StackMeta[item.ID] })
	if back == nil || back.Trigger == nil {
		t.Fatal("a restored game lost the triggering event")
	}
	if back.Trigger.Amount() != 9 {
		t.Errorf("restored amount = %d, want 9", back.Trigger.Amount())
	}
	if back.Trigger.Object == nil || back.Trigger.Object.ManaValue != 2 || back.Trigger.Object.Name != "Bear" {
		t.Errorf("restored object snapshot = %+v, want the one that was captured", back.Trigger.Object)
	}
}

// TestATargetClauseCanReadTheTriggeringEvent — the row's own sentence:
// "a triggered ability's target clause is static and cannot read the
// event that fired it". TargetsFrom is the answer, and the spec it
// builds is the one the CR 608.2b re-check reads back.
func TestATargetClauseCanReadTheTriggeringEvent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	small := pushPermanent(g, bearNamed(me.ID, "Small"))
	big := pushPermanent(g, bearNamed(me.ID, "Big"))
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case small:
				g.Battlefield.Cards[i].Power = 1
			case big:
				g.Battlefield.Cards[i].Power = 9
			}
		}
	})

	triggerCatalog(t, TriggeredAbility{
		Watches:   []EventKind{EventDealDamage},
		AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
		// "target creature with power less than the damage dealt" —
		// a clause that cannot exist without the event.
		TargetsFrom: func(tc TriggerContext, _ *Card, _ *Game) *TargetSpec {
			limit := tc.Event.Amount
			return &TargetSpec{
				Mode: "creature", Label: "target creature with small power",
				Zones: []ZoneKind{ZoneBattlefield},
				CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
					return c.IsCreature() && c.Power < limit
				},
				Min: 1, Max: 1,
			}
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "watcher", nil)
		},
	})
	watcherNamed(g, me.ID, "Watcher")

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 4, Actor: me.ID})
	})

	choice := pendingOfKind(t, g, PendingChoicePickTarget)
	if len(choice.PickTargetCards) != 1 || choice.PickTargetCards[0] != small {
		t.Fatalf("the clause offered %v, want only the power-1 creature", choice.PickTargetCards)
	}
	if err := g.ResolvePickTargets(choice.ID, me.ID, []TargetRef{{Kind: TargetCard, ID: small}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	// Answering is the moment the ability is put on the stack (CR
	// 603.3), and finishPickTargetLocked drains the queue there — so
	// the item is in StackMeta by the time this reads it.
	item := onlyAbilityOnStack(g)
	if item == nil {
		t.Fatal("answering the pick put no ability on the stack")
	}
	if item.Trigger == nil || item.Trigger.Amount() != 4 {
		t.Errorf("the event did not survive the CR 603.3d prompt: %+v", item.Trigger)
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != small {
		t.Errorf("targets = %+v, want the picked creature", item.Targets)
	}
}

// TestAnUnfillableEventBuiltClauseRemovesTheTrigger — CR 603.3d, over
// a clause the event cut. Nothing is prompted and nothing reaches the
// stack, which is the difference between Scrap Trawler doing nothing
// and Scrap Trawler wedging the table on an unanswerable prompt.
func TestAnUnfillableEventBuiltClauseRemovesTheTrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushPermanent(g, bearNamed(me.ID, "Big"))

	triggerCatalog(t, TriggeredAbility{
		Watches:   []EventKind{EventDealDamage},
		AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
		TargetsFrom: func(tc TriggerContext, _ *Card, _ *Game) *TargetSpec {
			limit := tc.Event.Amount
			return &TargetSpec{
				Mode: "creature", Label: "target creature with small power",
				Zones: []ZoneKind{ZoneBattlefield},
				CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
					return c.IsCreature() && c.Power < limit
				},
				Min: 1, Max: 1,
			}
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "watcher", nil)
		},
	})
	watcherNamed(g, me.ID, "Watcher")

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 1, Actor: me.ID})
	})

	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("%d prompts opened for a clause with no legal target (CR 603.3d)", n)
	}
	if n := len(g.PendingTriggers); n != 0 {
		t.Errorf("%d triggers queued for a clause with no legal target", n)
	}
}

// TestAReflexiveTriggerCarriesNoTriggeringEvent — CR 603.12. It was
// created by a resolution, not by an event, and the synthetic Event
// the dispatch carries for it must not be handed to a card as though
// something had happened on the board.
func TestAReflexiveTriggerCarriesNoTriggeringEvent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Source", me.ID))
	parent := pushAbilityItem(g, StackItemActivated, me.ID, src, "parent", nil)

	g.WithWriteLock(func() {
		ok := g.QueueReflexiveTriggerForEffect(parent, ReflexiveTrigger{
			Label: "when you do",
			Body:  testBody(func(*Game, *StackItem) error { return nil }),
		})
		if !ok {
			t.Fatal("QueueReflexiveTriggerForEffect refused a well-formed declaration")
		}
	})
	item := queuedTrigger(g)
	if item == nil {
		t.Fatal("the reflexive trigger was not queued")
	}
	if item.Trigger != nil {
		t.Errorf("a reflexive trigger carries an event it never had: %+v", item.Trigger)
	}
}

// TestObjectSnapshotSharesAPermanentType — CR 110.4a, the read
// Cloudstone Curio's clause is built on. An instant card shares
// nothing with anything: only the six permanent types count.
func TestObjectSnapshotSharesAPermanentType(t *testing.T) {
	artifactCreature := &ObjectSnapshot{Types: []string{"artifact", "creature"}}

	plainArtifact := NewCard("Sol Ring", uuid.New())
	plainArtifact.TypeLine = "Artifact"
	if !artifactCreature.SharesPermanentTypeWith(plainArtifact) {
		t.Error("an artifact creature shares 'artifact' with an artifact")
	}

	bear := NewCard("Bear", uuid.New())
	bear.TypeLine = "Creature — Bear"
	if !artifactCreature.SharesPermanentTypeWith(bear) {
		t.Error("an artifact creature shares 'creature' with a creature")
	}

	land := NewCard("Forest", uuid.New())
	land.TypeLine = "Basic Land — Forest"
	if artifactCreature.SharesPermanentTypeWith(land) {
		t.Error("an artifact creature shares no type with a land")
	}

	instantOnly := &ObjectSnapshot{Types: []string{"instant"}}
	if instantOnly.SharesPermanentTypeWith(bear) {
		t.Error("an instant has no permanent type to share")
	}
	if got := instantOnly.PermanentTypes(); len(got) != 0 {
		t.Errorf("PermanentTypes() = %v, want none", got)
	}
}
