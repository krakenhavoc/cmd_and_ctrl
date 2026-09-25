package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// source_object_test.go — #1418, CR 400.7 / CR 113.7a / CR 603.7d /
// CR 603.12. Every ability item names the OBJECT its source was, not
// only the card, so a resolving ability can tell "this permanent" from
// a new object the same card has become. The card-level proof is in
// cards/effects/source_object_test.go (Mana Vault, living weapon,
// Hero's Blade, Uthros Research Craft).

// damageWatcher installs a catalog trigger on EventDealDamage — an
// event that is NOT about the source, which is exactly the trigger the
// event's own object snapshot cannot name "this" for.
func damageWatcher(t *testing.T) {
	t.Helper()
	triggerCatalog(t, TriggeredAbility{
		Watches:   []EventKind{EventDealDamage},
		AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "watcher")
		},
	})
}

// fireDamageWatcher emits one damage event and returns the single
// trigger the harvester queued.
func fireDamageWatcher(t *testing.T, g *Game, actor uuid.UUID) *StackItem {
	t.Helper()
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 1, Actor: actor})
	})
	item := queuedTrigger(g)
	if item == nil {
		t.Fatal("the trigger did not fire")
	}
	return item
}

// flickerRaw bounces `id` to its owner's hand and puts it straight back:
// the same card, a new object (CR 400.7).
func flickerRaw(t *testing.T, g *Game, owner, id uuid.UUID) {
	t.Helper()
	hand := ZoneRef{Kind: ZoneHand, Owner: owner}
	bf := ZoneRef{Kind: ZoneBattlefield}
	if err := g.MoveCardByID(bf, hand, id); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	if err := g.MoveCardByID(hand, bf, id); err != nil {
		t.Fatalf("return: %v", err)
	}
}

// sourceRead is SourceObjectForEffect + PermanentForEffect +
// AbilitySourceGoneForEffect under one write lock.
func sourceRead(g *Game, item *StackItem) (ref ObjectRef, refOK bool, info PermanentInfo, infoOK bool, gone bool) {
	g.WithWriteLock(func() {
		ref, refOK = g.SourceObjectForEffect(item)
		if refOK {
			info, infoOK = g.PermanentForEffect(ref)
		}
		gone = g.AbilitySourceGoneForEffect(item)
	})
	return
}

// The control: a trigger whose source is still where it was names the
// live permanent, and every read answers exactly as it did before.
func TestATriggerNamesItsLiveSourceObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	damageWatcher(t)
	src := watcherNamed(g, me.ID, "Watcher")
	want := refOf(t, g, src)

	item := fireDamageWatcher(t, g, me.ID)
	if item.SourceObject != want {
		t.Fatalf("SourceObject = %+v, want the live watcher %+v", item.SourceObject, want)
	}
	ref, ok, info, infoOK, gone := sourceRead(g, item)
	if !ok || ref != want {
		t.Errorf("SourceObjectForEffect = %+v, %v; want %+v", ref, ok, want)
	}
	if !infoOK || info.Left {
		t.Errorf("PermanentForEffect = %+v, %v; want the live permanent", info, infoOK)
	}
	if gone {
		t.Error("a source that never moved reads as gone")
	}
}

// The headline: the source leaves and comes back while its trigger
// waits. The card is on the battlefield under the same ID, and it is
// not the object the trigger came from — every read answers with the
// departed object, never the new one.
func TestATriggerWhoseSourceLeftAndCameBackReadsTheDepartedObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	damageWatcher(t)
	src := watcherNamed(g, me.ID, "Watcher")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == src {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	first := refOf(t, g, src)
	item := fireDamageWatcher(t, g, me.ID)

	flickerRaw(t, g, me.ID, src)
	if now := refOf(t, g, src); now == first {
		t.Fatal("setup: the returned card should be a new object")
	}

	ref, ok, info, infoOK, gone := sourceRead(g, item)
	if !ok || ref != first {
		t.Errorf("SourceObjectForEffect = %+v, %v; want the object it triggered from, %+v", ref, ok, first)
	}
	if !infoOK || !info.Left || !info.Tapped {
		t.Errorf("PermanentForEffect = %+v, %v; want the departed, tapped object's record", info, infoOK)
	}
	if !gone {
		t.Error("AbilitySourceGoneForEffect: a new object with the same ID is not the trigger's source (CR 400.7)")
	}
}

// A "when this dies" trigger names the permanent that DIED — its
// battlefield epoch off the record — not the graveyard card the harvest
// finds, which is a new object too.
func TestADiesTriggerNamesThePermanentThatLeft(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	triggerCatalog(t, TriggeredAbility{
		Watches: []EventKind{EventLTB},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.CardID == source.InstanceID
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "dies")
		},
	})
	bear := bearNamed(me.ID, "Dying Bear")
	bear.OracleID = "watcher-dying-bear"
	id := pushPermanent(g, bear)
	onBattlefield := refOf(t, g, id)

	destroy(t, g, id)
	item := queuedTrigger(g)
	if item == nil {
		t.Fatal("the dies-trigger did not fire")
	}
	if item.SourceObject != onBattlefield {
		t.Errorf("SourceObject = %+v, want the permanent that died %+v", item.SourceObject, onBattlefield)
	}
	if _, _, info, ok, _ := sourceRead(g, item); !ok || !info.Left || info.Power != 2 {
		t.Errorf("SourcePermanent read = %+v, %v; want the dead bear's record", info, ok)
	}
}

// The object is taken when the ability TRIGGERS, not when the item is
// finally built. An optional trigger waits on its "you may" prompt; a
// card that moved in that window is a new object, and the trigger still
// names the one it came from — the dispatch's value copy of the source,
// not a board read at Build.
func TestAnOptionalTriggerNamesTheObjectThatTriggeredNotTheOneAtBuild(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	triggerCatalog(t, TriggeredAbility{
		Watches:        []EventKind{EventDealDamage},
		AppliesTo:      func(Event, *Card, Characteristic, *Game) bool { return true },
		OptionalPrompt: &TriggerOptionalPrompt{Question: "watch?"},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "watcher")
		},
	})
	src := watcherNamed(g, me.ID, "Watcher")
	first := refOf(t, g, src)
	var prompt *PendingChoice
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 1, Actor: me.ID})
		for _, c := range g.PendingChoices {
			if c.Kind == PendingChoiceTriggerPrompt {
				prompt = c
			}
		}
		// The move a prompt window cannot normally host, done under
		// the lock so the gate does not refuse it.
		if err := g.BounceToHandForEffect(src); err != nil {
			t.Fatalf("bounce: %v", err)
		}
		if _, err := MoveCard(me.Hand, g.Battlefield, src); err != nil {
			t.Fatalf("return: %v", err)
		}
	})
	if prompt == nil {
		t.Fatal("the optional trigger queued no prompt")
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("answer yes: %v", err)
	}
	var item *StackItem
	g.WithWriteLock(func() {
		for _, it := range append(append([]*StackItem(nil), g.PendingTriggers...), mapValues(g.StackMeta)...) {
			if it != nil && it.SourceCardID == src {
				item = it
			}
		}
	})
	if item == nil {
		t.Fatal("no item after the yes")
	}
	if item.SourceObject != first {
		t.Errorf("SourceObject = %+v, want the object that triggered %+v", item.SourceObject, first)
	}
}

// mapValues lists a stack-meta map's items.
func mapValues(m map[uuid.UUID]*StackItem) []*StackItem {
	out := make([]*StackItem, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// A catalog activation reads the object BEFORE its costs are paid: a
// source sacrificed to its own ability is still "this permanent", as it
// last existed (CR 608.2h), while SourceEpoch keeps the post-cost reading
// its readers want.
func TestAnActivationNamesTheSourceItsCostSacrificed(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	const oracle = "source-object-bomb"
	var seen PermanentInfo
	var seenOK bool
	stubActivatedFor(t, oracle, ActivatedAbilityShape{
		Label: "Sacrifice this: remember it.",
		Cost:  AbilityCost{SacrificeSelf: true},
		Effect: func(g *Game, item *StackItem) error {
			if ref, ok := g.SourceObjectForEffect(item); ok {
				seen, seenOK = g.PermanentForEffect(ref)
			}
			return nil
		},
	})
	bomb := pushGateCard(g, "Bomb", "Artifact", oracle, me.ID)
	want := refOf(t, g, bomb)

	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	item := onlyAbilityOnStack(g)
	if item == nil || item.SourceObject != want {
		t.Fatalf("activated item = %+v, want SourceObject %+v (pre-cost)", item, want)
	}
	if item.SourceEpoch == want.Epoch {
		t.Errorf("SourceEpoch = %d: it is the post-cost reading and should have moved with the sacrifice", item.SourceEpoch)
	}
	for i := 0; i < len(g.Seats) && onlyAbilityOnStack(g) != nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if !seenOK || !seen.Left {
		t.Errorf("at resolution the source reads %+v, %v; want the sacrificed permanent's record", seen, seenOK)
	}
}

// #1404's "Exile this artifact:" cost moves the source off the
// battlefield during the payment, exactly as a sacrifice cost does. The
// stamp is read before the payment, so it names the permanent that paid
// — whose record the read returns — not the card now in exile.
func TestAnExileThisPermanentCostStillNamesThePermanent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushExileThisPermanent(g, me, exileThisPermanentAbility("{1}", false))
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	want := refOf(t, g, src)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(src) {
		t.Fatal("setup: the source is not in exile after paying its cost")
	}
	item := onlyAbilityOnStack(g)
	if item == nil || item.SourceObject != want {
		t.Fatalf("activated item = %+v, want SourceObject %+v (the permanent before the cost)", item, want)
	}
	if _, _, info, ok, gone := sourceRead(g, item); !ok || !info.Left || !gone {
		t.Errorf("the exiled source reads %+v, %v, gone=%v; want the departed permanent's record", info, ok, gone)
	}
}

// The manual announce and the sandbox trigger both stamp the object.
func TestManualActivationAndAnnouncedTriggerNameTheSourceObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Relic", me.ID))
	want := refOf(t, g, src)

	if err := g.ActivateAbility(me.ID, src, AbilityParams{Label: "relic ability"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if got := onlyAbilityOnStack(g); got == nil || got.SourceObject != want {
		t.Errorf("activated item = %+v, want SourceObject %+v", got, want)
	}
	if err := g.AnnounceTrigger(me.ID, src, AbilityParams{Label: "relic trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	found := false
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == StackItemTriggered {
				found = true
				if it.SourceObject != want {
					t.Errorf("announced trigger SourceObject = %+v, want %+v", it.SourceObject, want)
				}
			}
		}
		for _, it := range g.PendingTriggers {
			found = true
			if it.SourceObject != want {
				t.Errorf("announced trigger SourceObject = %+v, want %+v", it.SourceObject, want)
			}
		}
	})
	if !found {
		t.Fatal("the announced trigger is neither pending nor on the stack")
	}

	// A game-built trigger that reaches the queue unstamped gets the
	// object its source is now; one that names its object keeps it.
	g.WithWriteLock(func() {
		g.PendingTriggers = nil
		bare := &StackItem{Controller: me.ID, Owner: me.ID, SourceCardID: src, Label: "bare"}
		g.queueHarvestedTriggerLocked(bare)
		if bare.SourceObject != want {
			t.Errorf("unstamped queued trigger SourceObject = %+v, want %+v", bare.SourceObject, want)
		}
		kept := ObjectRef{ID: src, Epoch: want.Epoch + 7}
		named := &StackItem{Controller: me.ID, Owner: me.ID, SourceCardID: src, SourceObject: kept, Label: "named"}
		g.queueHarvestedTriggerLocked(named)
		if named.SourceObject != kept {
			t.Errorf("a stamped trigger was overwritten: %+v, want %+v", named.SourceObject, kept)
		}
	})
}

// CR 603.7d and CR 603.12: a delayed trigger and a reflexive trigger
// take the source OBJECT of the ability that created them — here an
// ability whose source has since become a new object — not the object
// the card is now.
func TestDelayedAndReflexiveTriggersInheritTheCreatorsSourceObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Source", me.ID))
	first := refOf(t, g, src)
	parent := pushAbilityItem(g, StackItemTriggered, me.ID, src, "parent", nil)
	g.WithWriteLock(func() { parent.SourceObject = first })
	flickerRaw(t, g, me.ID, src)

	var dt *DelayedTrigger
	g.WithWriteLock(func() {
		g.beginResolvingLocked(parent)
		defer g.beginResolvingLocked(nil)
		id := g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, SourceCardID: src, Label: "later", At: StepEnd,
			Body: testBody(func(*Game, *StackItem) error { return nil }),
		})
		for _, d := range g.DelayedTriggers {
			if d.ID == id {
				dt = d
			}
		}
		if !g.QueueReflexiveTriggerForEffect(parent, ReflexiveTrigger{
			Label: "when you do", Body: testBody(func(*Game, *StackItem) error { return nil }),
		}) {
			t.Fatal("QueueReflexiveTriggerForEffect refused a well-formed declaration")
		}
	})
	if dt == nil || dt.SourceObject != first {
		t.Fatalf("delayed trigger = %+v, want SourceObject %+v", dt, first)
	}
	reflexive := queuedTrigger(g)
	if reflexive == nil || reflexive.SourceObject != first {
		t.Errorf("reflexive trigger = %+v, want SourceObject %+v", reflexive, first)
	}

	g.WithWriteLock(func() {
		g.PendingTriggers = nil
		g.fireDelayedTriggersLocked(StepEnd)
	})
	fired := queuedTrigger(g)
	if fired == nil || fired.SourceObject != first {
		t.Errorf("fired delayed trigger = %+v, want SourceObject %+v", fired, first)
	}

	// Scheduled with nothing resolving: the object the card is now.
	var now *DelayedTrigger
	g.WithWriteLock(func() {
		id := g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, SourceCardID: src, Label: "later", At: StepEnd,
			Body: testBody(func(*Game, *StackItem) error { return nil }),
		})
		for _, d := range g.DelayedTriggers {
			if d.ID == id {
				now = d
			}
		}
	})
	if want := refOf(t, g, src); now == nil || now.SourceObject != want {
		t.Errorf("unparented delayed trigger = %+v, want the current object %+v", now, want)
	}
}

// Undo and a deploy restore both bring the stamp back — on a stack
// item, a pending trigger and a delayed trigger. An item that never had
// one (a snapshot from before #1418) restores unstamped and falls back
// to the pre-#1418 reads.
func TestSourceObjectRidesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Source", me.ID))
	first := refOf(t, g, src)
	onStack := pushAbilityItem(g, StackItemTriggered, me.ID, src, "on stack", nil)
	legacy := pushAbilityItem(g, StackItemTriggered, me.ID, src, "legacy", nil)
	var dtID uuid.UUID
	g.WithWriteLock(func() {
		onStack.SourceObject = first
		dtID = g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, SourceCardID: src, Label: "later", At: StepEnd,
			Body: testBody(func(*Game, *StackItem) error { return nil }),
		})
	})
	flickerRaw(t, g, me.ID, src)
	// Queued after the flicker: a zone move is a priority boundary that
	// would drain it onto the stack.
	g.WithWriteLock(func() {
		g.PendingTriggers = append(g.PendingTriggers, &StackItem{
			ID: uuid.New(), Kind: StackItemTriggered, Controller: me.ID, Owner: me.ID,
			SourceCardID: src, SourceObject: first, Label: "pending",
		})
	})

	check := func(name string, h *Game) {
		t.Helper()
		h.WithWriteLock(func() {
			if it := h.StackMeta[onStack.ID]; it == nil || it.SourceObject != first {
				t.Errorf("%s: stack item = %+v, want SourceObject %+v", name, it, first)
			} else if !h.AbilitySourceGoneForEffect(it) {
				t.Errorf("%s: the restored stamp does not see the new object", name)
			}
			if len(h.PendingTriggers) != 1 || h.PendingTriggers[0].SourceObject != first {
				t.Errorf("%s: pending trigger lost its SourceObject", name)
			}
			var dt *DelayedTrigger
			for _, d := range h.DelayedTriggers {
				if d.ID == dtID {
					dt = d
				}
			}
			if dt == nil || dt.SourceObject != first {
				t.Errorf("%s: delayed trigger = %+v, want SourceObject %+v", name, dt, first)
			}
			old := h.StackMeta[legacy.ID]
			if old == nil || old.SourceObject != (ObjectRef{}) {
				t.Errorf("%s: an unstamped item came back stamped: %+v", name, old)
			} else if ref, ok := h.SourceObjectForEffect(old); !ok || ref == first {
				t.Errorf("%s: unstamped fallback = %+v, %v; want the card's latest object", name, ref, ok)
			}
		})
	}

	check("clone", g.Clone())
	undo := g.Clone()
	g.WithWriteLock(func() {
		g.StackMeta = nil
		g.PendingTriggers = nil
		g.DelayedTriggers = nil
		g.RestoreFrom(undo)
	})
	check("undo", g)

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	check("snapshot", restored)
}
