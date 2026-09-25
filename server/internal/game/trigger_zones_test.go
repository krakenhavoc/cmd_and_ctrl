package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_zones_test.go — #925: a triggered ability that watches from
// a zone other than the battlefield (CR 113.6).
//
// Every case here uses a synthetic CardDef through the CatalogTriggers
// hook plus the boot-time zone index, exactly as triggers_test.go
// stubs the harvester's catalog. The cards that ship on this
// (Narcomoeba, Bloodghast) are covered in the effects package.

// withTriggerZones swaps the process-wide zone index for the duration
// of the test and indexes `triggers` under `oracle`, the way
// effects.Register does at boot. Restores the previous index via
// t.Cleanup so a case cannot leak a walked zone into the next one.
func withTriggerZones(t *testing.T, oracle string, triggers []TriggeredAbility) {
	t.Helper()
	prev := triggerZones
	triggerZones = newTriggerZoneIndex()
	t.Cleanup(func() { triggerZones = prev })
	IndexTriggerZones(oracle, triggers)
}

// withZoneTriggerCard stubs CatalogTriggers and the index together —
// the pairing every case below wants, and the pairing Register
// guarantees in production.
func withZoneTriggerCard(t *testing.T, oracle string, triggers []TriggeredAbility) {
	t.Helper()
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return triggers
	})
	withTriggerZones(t, oracle, triggers)
}

// graveyardWatcher is the shape under test: an ability that declares
// the graveyard and watches a draw. `fired` counts the dispatches and
// `seen` records what the harvest handed AppliesTo.
type zoneWatch struct {
	fired      int
	seenSource Card
	seenLKI    Characteristic
}

func (w *zoneWatch) ability(zone ZoneKind, kind EventKind) TriggeredAbility {
	return TriggeredAbility{
		Zones:   []ZoneKind{zone},
		Watches: []EventKind{kind},
		AppliesTo: func(_ Event, source *Card, lki Characteristic, _ *Game) bool {
			w.seenSource, w.seenLKI = *source, lki
			return true
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			w.fired++
			return newTriggeredItemForTest(source, "Zone watcher", func(*Game, *StackItem) error { return nil })
		},
	}
}

// TestGraveyardTriggerFiresFromTheGraveyardOnly is the core of #925:
// the same card, the same ability, fires from the graveyard and from
// nowhere else. A declared zone list IS the list — a card sitting on
// the battlefield or in hand with a graveyard-declared trigger is
// silent, which is what keeps "return this card from your graveyard"
// from firing off a permanent that has nothing to return.
func TestGraveyardTriggerFiresFromTheGraveyardOnly(t *testing.T) {
	const oracle = "test-graveyard-zone-oracle"
	for _, tc := range []struct {
		name string
		zone func(g *Game) *Zone
		want int
	}{
		{"graveyard", func(g *Game) *Zone { return g.Seats[0].Graveyard }, 1},
		{"battlefield", func(g *Game) *Zone { return g.Battlefield }, 0},
		{"hand", func(g *Game) *Zone { return g.Seats[0].Hand }, 0},
		{"exile", func(g *Game) *Zone { return g.Exile }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner := g.Seats[0]
			var w zoneWatch
			withZoneTriggerCard(t, oracle, []TriggeredAbility{w.ability(ZoneGraveyard, EventDrawCard)})
			tc.zone(g).PushTop(Card{
				InstanceID: uuid.New(),
				Name:       "Zone Watcher",
				OracleID:   oracle,
				TypeLine:   "Creature — Test",
				Owner:      owner.ID,
				Controller: owner.ID,
			})

			g.WithWriteLock(func() {
				g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID})
			})

			if w.fired != tc.want {
				t.Fatalf("trigger fired %d times from the %s, want %d", w.fired, tc.name, tc.want)
			}
		})
	}
}

// TestExileTriggerFiresFromExile is suspend's shape (CR 702.62b, the
// #659 consumer): "at the beginning of your upkeep, remove a time
// counter from this card" — an ability on a card in exile, which is
// the only place a suspended card ever is.
func TestExileTriggerFiresFromExile(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-exile-zone-oracle"

	var w zoneWatch
	withZoneTriggerCard(t, oracle, []TriggeredAbility{w.ability(ZoneExile, EventBeginUpkeep)})
	g.Exile.PushTop(Card{
		InstanceID: uuid.New(),
		Name:       "Suspended Card",
		OracleID:   oracle,
		TypeLine:   "Sorcery",
		Owner:      owner.ID,
	})

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventBeginUpkeep, Actor: owner.ID})
	})

	if w.fired != 1 {
		t.Fatalf("exile trigger fired %d times, want 1", w.fired)
	}
	if w.seenSource.Controller != owner.ID {
		t.Errorf("source controller = %s, want the owner %s (CR 108.4)", w.seenSource.Controller, owner.ID)
	}
}

// TestZoneTriggerSourceIsTheCardInThatZone pins what the harvest
// hands the ability: the card AS IT SITS IN THE ZONE, with its own
// characteristics — not a battlefield LKI snapshot, which is
// harvestLTB's business and belongs to a permanent that just left.
//
// And CR 108.4: a card outside the battlefield has no controller, so
// "you" is its OWNER. The stack item the trigger queues goes to the
// owner, which is what makes Bloodghast answer its owner's land drop
// after an opponent's Bribery gave the battlefield copy away.
func TestZoneTriggerSourceIsTheCardInThatZone(t *testing.T) {
	g := newActiveGame(t)
	owner, opponent := g.Seats[0], g.Seats[1]
	const oracle = "test-graveyard-lki-oracle"
	cardID := uuid.New()

	var w zoneWatch
	withZoneTriggerCard(t, oracle, []TriggeredAbility{w.ability(ZoneGraveyard, EventDrawCard)})
	owner.Graveyard.PushTop(Card{
		InstanceID: cardID,
		Name:       "Graveyard Watcher",
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  3,
		Owner:      owner.ID,
		// Stale from a battlefield the card has already left — the
		// harvest must not read "you" off this.
		Controller: opponent.ID,
	})

	pendingBefore := len(g.PendingTriggers)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID})
	})

	if w.fired != 1 {
		t.Fatalf("graveyard trigger fired %d times, want 1", w.fired)
	}
	if w.seenSource.InstanceID != cardID {
		t.Errorf("source = %s, want the graveyard card %s", w.seenSource.InstanceID, cardID)
	}
	if w.seenSource.Controller != owner.ID {
		t.Errorf("source controller = %s, want the OWNER %s (CR 108.4)", w.seenSource.Controller, owner.ID)
	}
	if w.seenLKI.Power != 2 || w.seenLKI.Toughness != 3 {
		t.Errorf("source LKI = %d/%d, want the graveyard card's own 2/3",
			w.seenLKI.Power, w.seenLKI.Toughness)
	}
	if got := len(g.PendingTriggers) - pendingBefore; got != 1 {
		t.Fatalf("PendingTriggers delta = %d, want 1", got)
	}
	if queued := g.PendingTriggers[len(g.PendingTriggers)-1]; queued.Controller != owner.ID {
		t.Errorf("queued item controller = %s, want the owner %s", queued.Controller, owner.ID)
	}
}

// TestBattlefieldWalkIsUnchangedByAZoneTrigger is the cost assertion.
// The index answers "which non-battlefield zones does THIS event kind
// have to be walked for", and the answer for an event kind nothing
// declares is none — so the graveyards are not walked at all and the
// battlefield's per-event cost is exactly what it was before #925.
func TestBattlefieldWalkIsUnchangedByAZoneTrigger(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-zone-index-cost-oracle"

	var w zoneWatch
	ability := w.ability(ZoneGraveyard, EventDrawCard)
	withZoneTriggerCard(t, oracle, []TriggeredAbility{ability})
	owner.Graveyard.PushTop(Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
	})

	if got := triggerZones.zonesFor(EventDrawCard); len(got) != 1 || got[0] != ZoneGraveyard {
		t.Fatalf("zonesFor(EventDrawCard) = %v, want [graveyard]", got)
	}
	for _, kind := range []EventKind{EventETB, EventLTB, EventCast, EventDealDamage, EventAttack} {
		if got := triggerZones.zonesFor(kind); len(got) != 0 {
			t.Fatalf("zonesFor(%s) = %v, want none — only a declared kind pays for a walk", kind, got)
		}
	}

	// An ETB is an event no card watches from another zone, so the
	// graveyard walk never reaches the watcher's AppliesTo.
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: uuid.New(), Actor: owner.ID})
	})
	if w.fired != 0 || w.seenSource.OracleID != "" {
		t.Fatalf("an unindexed event kind walked the graveyard anyway (fired %d, saw %q)",
			w.fired, w.seenSource.OracleID)
	}
}

// TestTriggerZoneIndexStoresOnlyDeclaredZones pins the index's two
// questions against a card that declares nothing: a catalog of
// ordinary battlefield triggers leaves the index empty, so the added
// harvest is a single map lookup per event for every table that plays
// no graveyard trigger at all.
func TestTriggerZoneIndexStoresOnlyDeclaredZones(t *testing.T) {
	withTriggerZones(t, "test-plain-oracle", []TriggeredAbility{{
		Watches: []EventKind{EventETB},
		Build:   func(Event, *Card, Characteristic, *Game) *StackItem { return nil },
	}})
	if got := triggerZones.zonesFor(EventETB); len(got) != 0 {
		t.Fatalf("an undeclared trigger indexed %v", got)
	}
	if triggerZones.declares("test-plain-oracle") {
		t.Fatal("an undeclared trigger marked its oracle key as a zone watcher")
	}
}

// TestTriggerZoneUnsupportedNamesTheAlternative keeps the boot guard
// honest: the battlefield and the stack each have a different way of
// being said, and Register quotes this reason at the card author.
func TestTriggerZoneUnsupportedNamesTheAlternative(t *testing.T) {
	for _, zone := range supportedTriggerZones {
		if why := TriggerZoneUnsupported(zone); why != "" {
			t.Errorf("TriggerZoneUnsupported(%s) = %q, want it supported", zone, why)
		}
	}
	for _, zone := range []ZoneKind{ZoneBattlefield, ZoneStack, ZoneHand, ZoneLibrary, ZoneCommand} {
		if TriggerZoneUnsupported(zone) == "" {
			t.Errorf("TriggerZoneUnsupported(%s) = \"\", want a reason", zone)
		}
	}
}

// TestGraveyardTriggerPromptSurvivesUndo: a graveyard trigger takes
// the ordinary dispatch, so its CR 603.5 "you may" is an ordinary
// PendingChoice — and a game cloned while that prompt is open rewinds
// to the prompt and replays it, exactly as a battlefield trigger's
// does. This is the undo property the whole harvest reuse buys.
func TestGraveyardTriggerPromptSurvivesUndo(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-graveyard-prompt-oracle"
	cardID := uuid.New()

	var built int
	withZoneTriggerCard(t, oracle, []TriggeredAbility{{
		Zones:          []ZoneKind{ZoneGraveyard},
		Watches:        []EventKind{EventDrawCard},
		OptionalPrompt: &TriggerOptionalPrompt{Question: "Return it?"},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			built++
			return newTriggeredItemForTest(source, "Graveyard prompt", func(*Game, *StackItem) error { return nil })
		},
	}})
	owner.Graveyard.PushTop(Card{
		InstanceID: cardID,
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
	})

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID})
	})
	choice := pendingChoiceOfKind(g, PendingChoiceTriggerPrompt)
	if choice == nil {
		t.Fatal("a graveyard trigger with an OptionalPrompt queued no prompt")
	}
	if choice.Chooser != owner.ID {
		t.Errorf("prompt chooser = %s, want the card's owner %s", choice.Chooser, owner.ID)
	}
	promptOpen := g.Clone()

	if err := g.ResolveTriggerPrompt(choice.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	if built != 1 {
		t.Fatalf("Build ran %d times on the first answer, want 1", built)
	}

	built = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	replay := pendingChoiceOfKind(g, PendingChoiceTriggerPrompt)
	if replay == nil {
		t.Fatal("the rewind lost the graveyard trigger's prompt")
	}
	if err := g.ResolveTriggerPrompt(replay.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt (replay, declined): %v", err)
	}
	if built != 0 {
		t.Errorf("Build ran %d times on a declined replay, want 0", built)
	}
}
