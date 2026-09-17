package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func withCatalogTriggerDoublers(t *testing.T, fn func(string) []TriggerDoubler) {
	t.Helper()
	previous := CatalogTriggerDoublers
	CatalogTriggerDoublers = fn
	t.Cleanup(func() { CatalogTriggerDoublers = previous })
}

func TestTriggerDoublerSnapshotJSONRestoreKeepsAttribution(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const sourceOracle, doublerOracle = "restore-source", "restore-doubler"
	source := seedTriggerSource(g, owner, sourceOracle)
	doubler := pushBattlefieldForTest(g, owner.ID, "Restore Doubler", "Artifact", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != sourceOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventETB}, AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "restorable", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != doublerOracle {
			return nil
		}
		return []TriggerDoubler{{Label: "Restore Doubler", Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
	})
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: source}) })
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.PendingTriggers; len(got) != 2 || got[1].DoubledBy != doubler || got[1].DoubledByName != "Restore Doubler" {
		t.Fatalf("restored attribution = %#v", got)
	}
}

func TestTriggerDoublersStackAndTargetPromptsAreIndependent(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const sourceOracle = "targeted-source"
	_ = seedTriggerSource(g, owner, sourceOracle)
	ids := []uuid.UUID{
		pushBattlefieldForTest(g, owner.ID, "One", "Artifact", "doubler-one"),
		pushBattlefieldForTest(g, owner.ID, "Two", "Artifact", "doubler-two"),
		pushBattlefieldForTest(g, owner.ID, "Three", "Artifact", "doubler-three"),
	}
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != sourceOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventDrawCard}, AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true }, Targets: &TargetSpec{Mode: "player", Players: true, Min: 1, Max: 1}, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "targeted", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == "doubler-one" || key == "doubler-two" || key == "doubler-three" {
			return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
		}
		return nil
	})
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID}) })
	if got := len(g.PendingChoices); got != 4 {
		t.Fatalf("target prompts = %d, want 4", got)
	}
	seen := map[uuid.UUID]bool{}
	for _, choice := range g.PendingChoices[1:] {
		id, _ := choice.TriggerDoubler()
		seen[id] = true
	}
	for _, id := range ids {
		if !seen[id] {
			t.Fatalf("missing doubled target prompt for %s", id)
		}
	}
	first, second := g.PendingChoices[0], g.PendingChoices[1]
	if err := g.ResolvePickTarget(first.ID, owner.ID, TargetRef{Kind: TargetPlayer, ID: owner.ID}); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 3 {
		t.Fatalf("resolving one target choice consumed %d prompts", 4-len(g.PendingChoices))
	}
	if err := g.ResolvePickTarget(second.ID, owner.ID, TargetRef{Kind: TargetPlayer, ID: owner.ID}); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingTriggers)+len(g.StackMeta) != 2 {
		t.Fatalf("independent target choices built %d items", len(g.PendingTriggers)+len(g.StackMeta))
	}
}

func TestTriggerDoublerEmptyTargetsDropEveryInstance(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const sourceOracle, doublerOracle = "empty-target-source", "empty-target-doubler"
	seedTriggerSource(g, owner, sourceOracle)
	pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != sourceOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventDrawCard}, AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true }, Targets: &TargetSpec{Zones: []ZoneKind{ZoneBattlefield}, CardOK: func(*Game, uuid.UUID, Card, ZoneKind) bool { return false }, Min: 1, Max: 1}, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "never", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == doublerOracle {
			return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
		}
		return nil
	})
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID}) })
	if len(g.PendingChoices) != 0 || len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatalf("empty target trigger was queued: choices=%d pending=%d stack=%d", len(g.PendingChoices), len(g.PendingTriggers), len(g.StackMeta))
	}
}

func TestTriggerDoublerUsesBatchLKIAndDoesNotRediscoverRemovedDoubler(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const watcherOracle, doublerOracle = "wipe-watcher", "wipe-doubler"
	watcher := seedTriggerSource(g, owner, watcherOracle)
	doubler := pushBattlefieldForTest(g, owner.ID, "Silenced Doubler", "Artifact", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != watcherOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventLTB}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool { return ev.CardID == doubler }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "watch", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != doublerOracle {
			return nil
		}
		return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == doubler {
				effective := g.Battlefield.Cards[i].Effective()
				effective.AbilitiesRemoved = true
				g.Battlefield.Cards[i].effective = &effective
			}
		}
		// The watcher leaves first. Its matching trigger is therefore found
		// through the simultaneous-exit copy, after the doubler has reached
		// its destination zone. The batch LKI must still suppress it.
		g.DestroyPermanentsForEffect([]uuid.UUID{watcher, doubler})
	})
	if got := len(g.PendingTriggers) + len(g.StackMeta); got != 1 {
		t.Fatalf("removed doubler changed a simultaneous watcher into %d instances, want 1", got)
	}
}

func TestTriggerDoublerReadsLivingCandidateDuringAnotherCardsLTB(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const watcherOracle, doublerOracle = "living-watcher", "living-doubler"
	watcher := seedTriggerSource(g, owner, watcherOracle)
	doubler := pushBattlefieldForTest(g, owner.ID, "Living Doubler", "Artifact", doublerOracle)
	victim := pushBattlefieldForTest(g, owner.ID, "Victim", "Creature", "")
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != watcherOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventLTB}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool { return ev.CardID == victim }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "living watch", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != doublerOracle {
			return nil
		}
		return []TriggerDoubler{{Applies: func(_ *Game, q TriggerDoublingQuery) bool {
			return q.Doubler.InstanceID == doubler && q.Subject == victim && q.SubjectLKI.Controller == owner.ID
		}}}
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	if got := len(g.PendingTriggers); got != 2 {
		t.Fatalf("living doubler LTB instances = %d, want 2", got)
	}
	if g.PendingTriggers[1].DoubledBy != doubler {
		t.Fatalf("living doubler attribution = %#v", g.PendingTriggers[1])
	}
	_ = watcher
}

func TestTriggerDoublerLTBPassPrefersSimultaneousSubjectSnapshot(t *testing.T) {
	g := newActiveGame(t)
	owner, changed := g.Seats[0], g.Seats[1]
	victim := pushBattlefieldForTest(g, owner.ID, "Animated Subject", "Creature", "")
	g.WithWriteLock(func() {
		old := g.Battlefield.Cards[0].Effective()
		old.Controller = changed.ID
		g.simultaneousExit = []Card{{InstanceID: victim, Controller: changed.ID, effective: &old}}
		g.lastKnownBattlefield = map[uuid.UUID]Characteristic{victim: {Controller: owner.ID}}
		pass := g.newHarvestPassLocked(Event{Kind: EventLTB, CardID: victim})
		if !pass.hasSubject || pass.subjectLKI.Controller != changed.ID {
			t.Fatalf("LTB subject LKI = %#v, want simultaneous controller %s", pass.subjectLKI, changed.ID)
		}
	})
}

func TestTriggerDoublerCountsItselfWhenItDies(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const watcherOracle, doublerOracle = "dying-watcher", "dying-doubler"
	_ = seedTriggerSource(g, owner, watcherOracle)
	doubler := pushBattlefieldForTest(g, owner.ID, "Dying Doubler", "Artifact", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != watcherOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventLTB}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool { return ev.CardID == doubler }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "death", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == doublerOracle {
			return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
		}
		return nil
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(doubler) })
	if got := len(g.PendingTriggers); got != 2 {
		t.Fatalf("dying doubler instances = %d, want 2", got)
	}
	if got := g.PendingTriggers[1].DoubledBy; got != doubler {
		t.Fatalf("dying doubler attribution = %s, want %s", got, doubler)
	}
}

func TestTriggerDoublerLegendChoiceCountsSurvivorAndLeavingCopy(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const watcherOracle, doublerOracle = "legend-watcher", "legend-doubler"
	_ = seedTriggerSource(g, owner, watcherOracle)
	keep := pushBattlefieldForTest(g, owner.ID, "Legendary Doubler", "Legendary Creature", doublerOracle)
	drop := pushBattlefieldForTest(g, owner.ID, "Legendary Doubler", "Legendary Creature", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != watcherOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventLTB}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool { return ev.CardID == drop }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "legend death", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == doublerOracle {
			return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
		}
		return nil
	})
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	choice := legendPromptFor(g, owner.ID)
	if choice == nil {
		t.Fatal("legend choice was not queued")
	}
	if err := g.ResolveLegendRule(choice.ID, owner.ID, keep); err != nil {
		t.Fatal(err)
	}
	items := append([]*StackItem(nil), g.PendingTriggers...)
	for _, item := range g.StackMeta {
		items = append(items, item)
	}
	if got := len(items); got != 3 {
		t.Fatalf("legend-rule death instances = %d, want 3", got)
	}
	seen := map[uuid.UUID]bool{}
	for _, item := range items {
		if item.DoubledBy == uuid.Nil {
			continue
		}
		seen[item.DoubledBy] = true
	}
	if !seen[keep] || !seen[drop] {
		t.Fatalf("legend-rule doubler attribution = %#v", seen)
	}
}

func TestTriggerDoublerSeesBattlefieldToGraveyardZoneMoveSubjectAndLeaver(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const watcherOracle, doublerOracle = "zone-watcher", "zone-doubler"
	_ = seedTriggerSource(g, owner, watcherOracle)
	doubler := pushBattlefieldForTest(g, owner.ID, "Zone Doubler", "Artifact", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != watcherOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventZoneMove}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool {
			return ev.OldZone == ZoneBattlefield && ev.NewZone == ZoneGraveyard && ev.CardID == doubler
		}, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "move death", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == doublerOracle {
			return []TriggerDoubler{{Applies: func(_ *Game, q TriggerDoublingQuery) bool {
				return q.Subject == doubler && q.SubjectLKI.Controller == owner.ID
			}}}
		}
		return nil
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(doubler) })
	if got := len(g.PendingTriggers); got != 2 {
		t.Fatalf("battlefield move instances = %d, want 2", got)
	}

	// An identically shaped non-battlefield move must not recover a doubler
	// from stale destination data.
	g = newActiveGame(t)
	owner = g.Seats[0]
	_ = seedTriggerSource(g, owner, watcherOracle)
	graveDoubler := NewCard("Former Doubler", owner.ID)
	graveDoubler.OracleID = doublerOracle
	owner.Graveyard.PushTop(graveDoubler)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != watcherOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventZoneMove}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool { return ev.CardID == graveDoubler.InstanceID }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "not a death", nil)
		}}}
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: graveDoubler.InstanceID, OldZone: ZoneHand, NewZone: ZoneGraveyard})
	})
	if got := len(g.PendingTriggers); got != 1 {
		t.Fatalf("non-battlefield move instances = %d, want 1", got)
	}
}

func TestTriggerDoublerUsesCopiedLeavingIdentityAndCarriesItThroughSnapshot(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const watcherOracle, cloneOracle, doublerOracle = "copy-watcher", "printed-clone", "copied-doubler"
	_ = seedTriggerSource(g, owner, watcherOracle)
	target := Card{OracleID: doublerOracle, Name: "Teysa", TypeLine: "Creature", Owner: owner.ID, Controller: owner.ID}
	clone := Card{InstanceID: uuid.New(), OracleID: cloneOracle, Name: "Clone", TypeLine: "Creature", Owner: owner.ID, Controller: owner.ID}
	clone.applyCopy(CopiableValuesOf(target), target)
	g.Battlefield.PushTop(clone)
	var builtFrom Card
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		switch key {
		case watcherOracle:
			return []TriggeredAbility{{Watches: []EventKind{EventLTB}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool { return ev.CardID == clone.InstanceID }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "copied death", nil)
			}}}
		case doublerOracle:
			// This source has become a printed Clone by the time EventLTB is
			// harvested. The retained identity must still recover the copied
			// trigger and supply its copied name/controller to Build.
			return []TriggeredAbility{{Watches: []EventKind{EventLTB}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool { return ev.CardID == clone.InstanceID }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				builtFrom = *source
				return nil
			}}}
		}
		return nil
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == doublerOracle {
			return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
		}
		return nil
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(clone.InstanceID) })
	if got := len(g.PendingTriggers); got != 2 {
		t.Fatalf("copied dying doubler instances = %d, want 2", got)
	}
	if builtFrom.Name != "Teysa" || builtFrom.Controller != owner.ID || builtFrom.OracleID != doublerOracle {
		t.Fatalf("copied leaving trigger source = %#v", builtFrom)
	}

	// The narrow LKI metadata is carried exactly like the characteristic map.
	g.WithWriteLock(func() {
		g.lastKnownTriggerIdentity = map[uuid.UUID]triggerIdentityLKI{clone.InstanceID: {OracleID: doublerOracle, AttachedTo: TargetRef{Kind: TargetCard, ID: uuid.New()}}}
	})
	snap := g.CaptureSnapshot()
	if got := snap.LastKnownTriggerIdentity[clone.InstanceID]; got.OracleID != doublerOracle || got.AttachedTo.Kind != TargetCard {
		t.Fatalf("snapshot trigger identity = %#v", got)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.lastKnownTriggerIdentity[clone.InstanceID]; got.OracleID != doublerOracle || got.AttachedTo.Kind != TargetCard {
		t.Fatalf("restored trigger identity = %#v", got)
	}
	if got := g.Clone().lastKnownTriggerIdentity[clone.InstanceID]; got.OracleID != doublerOracle || got.AttachedTo.Kind != TargetCard {
		t.Fatalf("cloned trigger identity = %#v", got)
	}
}

func TestTriggerDoublerHandlesSimultaneousLegendaryEntriesBeforeLegendChoice(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const sourceOracle, doublerOracle = "legend-entry-source", "legend-entry-doubler"
	first := pushBattlefieldForTest(g, owner.ID, "Shared Legend", "Legendary Creature", sourceOracle)
	second := pushBattlefieldForTest(g, owner.ID, "Shared Legend", "Legendary Creature", sourceOracle)
	pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != sourceOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventETB}, AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool { return ev.CardID == source.InstanceID }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "entry", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == doublerOracle {
			return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
		}
		return nil
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: first, Actor: owner.ID})
		g.EmitEvent(Event{Kind: EventETB, CardID: second, Actor: owner.ID})
		g.runStateChecksLocked()
	})
	if got := len(g.PendingTriggers); got != 4 {
		t.Fatalf("two simultaneous entries produced %d instances, want 4", got)
	}
	if legendPromptFor(g, owner.ID) == nil {
		t.Fatal("legend choice was not queued after doubled entry triggers")
	}
}

func TestTriggerDoublerDoesNotApplyToManualOrReflexiveTriggers(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	source := pushBattlefieldForTest(g, owner.ID, "Source", "Artifact", "manual-source")
	pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", "manual-doubler")
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != "manual-doubler" {
			return nil
		}
		return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
	})
	if err := g.AnnounceTrigger(owner.ID, source, AbilityParams{Label: "manual"}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		g.QueueReflexiveTriggerForEffect(&StackItem{SourceCardID: source, Controller: owner.ID, Label: "parent"}, ReflexiveTrigger{Label: "reflexive", Effect: func(*Game, *StackItem) error { return nil }})
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{Controller: owner.ID, SourceCardID: source, Label: "delayed", At: StepEnd, Effect: func(*Game, *StackItem) error { return nil }})
		g.fireDelayedTriggersLocked(StepEnd)
	})
	if got := len(g.PendingTriggers); got != 3 {
		t.Fatalf("manual/reflexive/delayed instances = %d, want 3", got)
	}
	for _, item := range g.PendingTriggers {
		if item.DoubledBy != uuid.Nil {
			t.Fatalf("excluded trigger has doubler attribution: %#v", item)
		}
	}
}

func TestTriggerDoublerOncePerBatchKeepsFirstMatchingInstance(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const sourceOracle, doublerOracle = "later-match-source", "later-match-doubler"
	_ = seedTriggerSource(g, owner, sourceOracle)
	pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
	first := pushBattlefieldForTest(g, owner.ID, "First attacker", "Creature", "")
	later := pushBattlefieldForTest(g, owner.ID, "Later attacker", "Creature", "")
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != sourceOracle {
			return nil
		}
		return []TriggeredAbility{{Key: "one-or-more", OncePerBatch: true, Watches: []EventKind{EventAttack}, AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "one-or-more", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == doublerOracle {
			return []TriggerDoubler{{Applies: func(_ *Game, q TriggerDoublingQuery) bool { return q.Subject == later }}}
		}
		return nil
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventAttack, CardID: first})
		g.EmitEvent(Event{Kind: EventAttack, CardID: later})
	})
	if got := len(g.PendingTriggers); got != 1 {
		t.Fatalf("later qualifying event added %d instances, want first match only", got)
	}
}

func TestTriggerDoublerCoversStackSpellSagaAndEvokeHarvests(t *testing.T) {
	t.Run("stack spell", func(t *testing.T) {
		g := newActiveGame(t)
		owner := g.Seats[0]
		const spellOracle, doublerOracle = "stack-source", "stack-doubler"
		spell := NewCard("Stack Source", owner.ID)
		spell.OracleID, spell.TypeLine = spellOracle, "Creature"
		g.Stack.PushTop(spell)
		pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
		withCatalogTriggers(t, func(key string) []TriggeredAbility {
			if key != spellOracle {
				return nil
			}
			return []TriggeredAbility{{FromStack: true, Watches: []EventKind{EventCast}, AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "cast", nil)
			}}}
		})
		withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
			if key == doublerOracle {
				return []TriggerDoubler{{Applies: func(_ *Game, q TriggerDoublingQuery) bool { return q.FromSpell }}}
			}
			return nil
		})
		g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventCast, CardID: spell.InstanceID, Actor: owner.ID}) })
		if got := len(g.PendingTriggers); got != 2 {
			t.Fatalf("stack spell instances = %d, want 2", got)
		}
	})
	t.Run("saga source", func(t *testing.T) {
		g := newActiveGame(t)
		owner := g.Seats[0]
		const sagaOracle, doublerOracle = "saga-source", "saga-doubler"
		saga := seedTriggerSource(g, owner, sagaOracle)
		pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
		withCatalogTriggers(t, func(key string) []TriggeredAbility {
			if key != sagaOracle {
				return nil
			}
			return []TriggeredAbility{{Watches: []EventKind{EventSagaChapter}, AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool { return ev.Source == source.InstanceID }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "chapter", nil)
			}}}
		})
		withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
			if key == doublerOracle {
				return []TriggerDoubler{{Applies: func(_ *Game, q TriggerDoublingQuery) bool { return q.Event.Source == saga && !q.FromSpell }}}
			}
			return nil
		})
		g.WithWriteLock(func() {
			g.EmitEvent(Event{Kind: EventSagaChapter, Source: saga, CardID: saga, Actor: owner.ID, Amount: 1})
		})
		if got := len(g.PendingTriggers); got != 2 {
			t.Fatalf("saga instances = %d, want 2", got)
		}
	})
	t.Run("evoke", func(t *testing.T) {
		g := newActiveGame(t)
		owner := g.Seats[0]
		const evokeOracle, doublerOracle = "evoke-source", "evoke-doubler"
		evoke := seedTriggerSource(g, owner, evokeOracle)
		pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
		withCatalogAlternativeCosts(t, func(key string) []AlternativeCost {
			if key == evokeOracle {
				return []AlternativeCost{{Key: "evoke", Label: "Evoke", SacrificeOnEntry: true}}
			}
			return nil
		})
		withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
			if key == doublerOracle {
				return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
			}
			return nil
		})
		g.WithWriteLock(func() {
			g.queueAltCostEntryTriggerLocked(*g.findCardByIDLocked(evoke), &StackItem{AltCost: "evoke", Controller: owner.ID})
		})
		if got := len(g.PendingTriggers); got != 2 {
			t.Fatalf("evoke instances = %d, want 2", got)
		}
	})
}

func TestTriggerDoublerQueuesIndependentInstances(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const sourceOracle, doublerOracle = "trigger-source", "test-doubler"
	source := seedTriggerSource(g, owner, sourceOracle)
	doubler := pushBattlefieldForTest(g, owner.ID, "Test Doubler", "Artifact", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != sourceOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventETB}, AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool { return ev.CardID == source.InstanceID }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "draw", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != doublerOracle {
			return nil
		}
		return []TriggerDoubler{{Label: "Test Doubler", Applies: func(_ *Game, q TriggerDoublingQuery) bool {
			return q.SourceLKI.Controller == owner.ID && q.Doubler.InstanceID == doubler
		}}}
	})
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: source, Actor: owner.ID}) })
	if got := len(g.PendingTriggers); got != 2 {
		t.Fatalf("instances = %d, want 2", got)
	}
	if first, second := g.PendingTriggers[0], g.PendingTriggers[1]; first.DoubledBy != uuid.Nil || second.DoubledBy != doubler || second.DoubledByName != "Test Doubler" || first.Label != second.Label {
		t.Fatalf("attribution: first=%#v second=%#v", first, second)
	}
	var triggers int
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventTrigger {
				triggers++
			}
		}
	})
	if triggers != 2 {
		t.Fatalf("EventTrigger count = %d, want 2", triggers)
	}
	if got := g.CaptureSnapshot().PendingTriggers[1]; got.DoubledBy != doubler || got.DoubledByName != "Test Doubler" {
		t.Fatalf("snapshot attribution = %#v", got)
	}
	clone := g.Clone()
	if got := clone.PendingTriggers[1]; got.DoubledBy != doubler || got.DoubledByName != "Test Doubler" {
		t.Fatalf("clone attribution = %#v", got)
	}
}

func TestTriggerDoublerOncePerBatchChecksMatchNotInstance(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const sourceOracle, doublerOracle = "attack-source", "attack-doubler"
	source := seedTriggerSource(g, owner, sourceOracle)
	pushBattlefieldForTest(g, owner.ID, "Doubler", "Creature", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != sourceOracle {
			return nil
		}
		return []TriggeredAbility{{Key: "attack", OncePerBatch: true, Watches: []EventKind{EventAttack}, AppliesTo: func(_ Event, _ *Card, _ Characteristic, _ *Game) bool { return true }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "attack", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != doublerOracle {
			return nil
		}
		return []TriggerDoubler{{Label: "Doubler", Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventAttack, CardID: source})
		g.EmitEvent(Event{Kind: EventAttack, CardID: uuid.New()})
	})
	if got := len(g.PendingTriggers); got != 2 {
		t.Fatalf("once-per-batch instances = %d, want 2", got)
	}
}

func TestTriggerDoublerOptionalPromptsAreIndependent(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const sourceOracle, doublerOracle = "optional-source", "optional-doubler"
	_ = seedTriggerSource(g, owner, sourceOracle)
	pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != sourceOracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventDrawCard}, OptionalPrompt: &TriggerOptionalPrompt{Question: "draw?"}, AppliesTo: func(_ Event, _ *Card, _ Characteristic, _ *Game) bool { return true }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "optional", nil)
		}}}
	})
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != doublerOracle {
			return nil
		}
		return []TriggerDoubler{{Label: "Doubler", Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
	})
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID}) })
	if len(g.PendingChoices) != 2 {
		t.Fatalf("optional prompts = %d, want 2", len(g.PendingChoices))
	}
	first, second := g.PendingChoices[0], g.PendingChoices[1]
	if id, name := second.TriggerDoubler(); id == uuid.Nil || name != "Doubler" {
		t.Fatalf("prompt attribution = %s %q", id, name)
	}
	if err := g.ResolveTriggerPrompt(first.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("declining one removed %d prompts", 2-len(g.PendingChoices))
	}
	if err := g.ResolveTriggerPrompt(second.ID, owner.ID, true); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingTriggers)+len(g.StackMeta) != 1 {
		t.Fatalf("accepted second did not independently build")
	}
}
