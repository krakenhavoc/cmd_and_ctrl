package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// trigger_suppression_test.go — #1735, the engine half of "[an event]
// doesn't cause abilities to trigger" (CR 603.2). The suppressors here
// are built in this package, because the effects-package helpers
// (SuppressesEntering, SuppressesDying, OfOpponentsPermanents) can't be
// imported from it. They test the same things those helpers test: the
// event, the subject's type, and for the opponents-only form the
// source's controller. The card half (Torpor Orb, Hushbringer, Hushwing
// Gryff, Tocatli Honor Guard, Elesh Norn) is in
// cards/effects/trigger_suppression_test.go.

const (
	suppOrb      = "supp-orb"       // creatures entering don't cause abilities to trigger
	suppHush     = "supp-hush"      // ... entering or dying
	suppElesh    = "supp-elesh"     // opponents' permanents; any permanent entering; plus the doubler
	suppSelfETB  = "supp-self-etb"  // "When this enters, ..."
	suppWatcher  = "supp-watcher"   // "Whenever another permanent enters, ..."
	suppOptional = "supp-optional"  // "Whenever another permanent enters, you may ..."
	suppTargeted = "supp-targeted"  // "Whenever another permanent enters, target player ..."
	suppDraw     = "supp-draw"      // "Whenever a player draws a card, ..."
	suppDiesSelf = "supp-dies-self" // "When this dies, ..."
	suppDiesAny  = "supp-dies-any"  // "Whenever another creature dies, ..."
	suppLeaves   = "supp-leaves"    // "Whenever another permanent leaves the battlefield, ..."
	suppOneOrMor = "supp-one-more"  // "Whenever one or more other permanents enter, ..." (once per batch)
)

func withCatalogTriggerSuppressors(t *testing.T, fn func(string) []TriggerSuppressor) {
	t.Helper()
	previous := CatalogTriggerSuppressors
	CatalogTriggerSuppressors = fn
	t.Cleanup(func() { CatalogTriggerSuppressors = previous })
}

func suppEnteringEvent(ev Event) bool {
	return ev.Kind == EventETB || ev.Kind == EventTokenCreated || (ev.Kind == EventZoneMove && ev.NewZone == ZoneBattlefield)
}

func suppDyingEvent(ev Event) bool {
	if ev.Kind == EventLTB {
		return ev.NewZone == ZoneGraveyard
	}
	return ev.Kind == EventZoneMove && ev.OldZone == ZoneBattlefield && ev.NewZone == ZoneGraveyard
}

// suppEntering is "[creatures / permanents] entering don't cause
// abilities to trigger".
func suppEntering(creaturesOnly bool) TriggerSuppressor {
	return TriggerSuppressor{Label: "entering", Suppresses: func(_ *Game, q TriggerSuppressionQuery) bool {
		return q.HasSubject && suppEnteringEvent(q.Event) &&
			(!creaturesOnly || slices.Contains(q.SubjectLKI.Types, "Creature"))
	}}
}

// suppDying is Hushbringer's second half.
func suppDying() TriggerSuppressor {
	return TriggerSuppressor{Label: "dying", Suppresses: func(_ *Game, q TriggerSuppressionQuery) bool {
		return q.HasSubject && suppDyingEvent(q.Event) && slices.Contains(q.SubjectLKI.Types, "Creature")
	}}
}

// suppOpponents is "abilities of permanents your opponents control".
func suppOpponents(s TriggerSuppressor) TriggerSuppressor {
	inner := s.Suppresses
	s.Suppresses = func(g *Game, q TriggerSuppressionQuery) bool {
		return q.SourceIsPermanent && q.SourceLKI.Controller != q.SuppressorLKI.Controller && inner(g, q)
	}
	return s
}

// suppEleshDoubler is Elesh Norn's first paragraph: a permanent
// entering, an ability of a permanent the doubler's controller controls.
func suppEleshDoubler() TriggerDoubler {
	return TriggerDoubler{Label: "Test Elesh", Applies: func(_ *Game, q TriggerDoublingQuery) bool {
		return !q.FromSpell && q.HasSubject && suppEnteringEvent(q.Event) && q.SourceLKI.Controller == q.DoublerLKI.Controller
	}}
}

func suppTrigger(label string, watches EventKind, applies func(ev Event, source *Card) bool) TriggeredAbility {
	return TriggeredAbility{
		Key:     label,
		Watches: []EventKind{watches},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return applies(ev, source)
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, label)
		},
	}
}

func anotherObject(ev Event, source *Card) bool { return ev.CardID != source.InstanceID }
func thisObject(ev Event, source *Card) bool    { return ev.CardID == source.InstanceID }

// installSuppressionCatalog stubs the three catalog hooks the harvest
// reads with this file's fixtures.
func installSuppressionCatalog(t *testing.T) {
	t.Helper()
	triggers := map[string][]TriggeredAbility{
		suppSelfETB: {suppTrigger("self etb", EventETB, thisObject)},
		suppWatcher: {suppTrigger("watch", EventETB, anotherObject)},
		suppDraw:    {suppTrigger("draw watch", EventDrawCard, func(Event, *Card) bool { return true })},
		suppDiesSelf: {suppTrigger("dies self", EventLTB, func(ev Event, source *Card) bool {
			return thisObject(ev, source) && ev.NewZone == ZoneGraveyard
		})},
		suppDiesAny: {suppTrigger("dies watch", EventLTB, func(ev Event, source *Card) bool {
			return anotherObject(ev, source) && ev.NewZone == ZoneGraveyard
		})},
		suppLeaves: {suppTrigger("leaves watch", EventLTB, anotherObject)},
	}
	optional := suppTrigger("optional watch", EventETB, anotherObject)
	optional.OptionalPrompt = &TriggerOptionalPrompt{Question: "Gain 1 life?"}
	triggers[suppOptional] = []TriggeredAbility{optional}
	targeted := suppTrigger("targeted watch", EventETB, anotherObject)
	targeted.Targets = &TargetSpec{Mode: "player", Players: true, Min: 1, Max: 1}
	triggers[suppTargeted] = []TriggeredAbility{targeted}
	oneOrMore := suppTrigger("one or more", EventETB, anotherObject)
	oneOrMore.OncePerBatch = true
	triggers[suppOneOrMor] = []TriggeredAbility{oneOrMore}

	suppressors := map[string][]TriggerSuppressor{
		suppOrb:   {suppEntering(true)},
		suppHush:  {suppEntering(true), suppDying()},
		suppElesh: {suppOpponents(suppEntering(false))},
	}
	withCatalogTriggers(t, func(key string) []TriggeredAbility { return triggers[key] })
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key == suppElesh {
			return []TriggerDoubler{suppEleshDoubler()}
		}
		return nil
	})
	withCatalogTriggerSuppressors(t, func(key string) []TriggerSuppressor { return suppressors[key] })
}

// suppInHand parks a card in its owner's hand and returns its ID.
func suppInHand(p *Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// suppPutFromHand is a permanent entering: one ETB, harvested at once.
func suppPutFromHand(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(id, HandEntryOptions{}); err != nil {
			t.Fatalf("put from hand: %v", err)
		}
	})
}

// suppQueued counts the instances of `label` that triggered: queued,
// on the stack, or waiting on a prompt.
func suppQueued(g *Game, label string) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.Label == label {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Label == label {
			n++
		}
	}
	return n
}

// --- entering ----------------------------------------------------------

// The entering creature's OWN "when this enters" is stopped: the
// creature is on the battlefield with the suppressor after the event,
// and that is the board an enters trigger is judged on (CR 603.10).
// A noncreature entering under a creatures-only suppressor still
// triggers, which pins that the filter reads the SUBJECT.
func TestTriggerSuppressorStopsTheEnteringCreaturesOwnETB(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "Orb", "Artifact", suppOrb)

	creature := suppInHand(me, "Mulldrifter", "Creature — Elemental", suppSelfETB)
	suppPutFromHand(t, g, creature)
	if got := suppQueued(g, "self etb"); got != 0 {
		t.Fatalf("a creature's own enters trigger triggered %d times under the suppressor, want 0", got)
	}

	artifact := suppInHand(me, "Trinket", "Artifact", suppSelfETB)
	suppPutFromHand(t, g, artifact)
	if got := suppQueued(g, "self etb"); got != 1 {
		t.Fatalf("a noncreature's enters trigger triggered %d times under a creatures-only suppressor, want 1", got)
	}
}

// Another permanent's "whenever a creature enters" is stopped too, and
// it is the suppressor that stops it: once the suppressor is gone the
// same watcher triggers on the next creature.
func TestTriggerSuppressorStopsAnotherPermanentsEntersWatcher(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	installSuppressionCatalog(t)
	orb := pushBattlefieldForTest(g, me.ID, "Orb", "Artifact", suppOrb)
	seedTriggerSource(g, opp, suppWatcher)

	suppPutFromHand(t, g, suppInHand(me, "Bear", "Creature — Bear", ""))
	if got := suppQueued(g, "watch"); got != 0 {
		t.Fatalf("the watcher triggered %d times under the suppressor, want 0", got)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(orb); err != nil {
			t.Fatalf("destroy the suppressor: %v", err)
		}
	})
	suppPutFromHand(t, g, suppInHand(me, "Second Bear", "Creature — Bear", ""))
	if got := suppQueued(g, "watch"); got != 1 {
		t.Fatalf("the watcher triggered %d times once the suppressor was gone, want 1", got)
	}
}

// Elesh Norn's form reads the SOURCE's controller: an opponent's
// permanents' abilities are stopped, the suppressor's controller's are
// not, and an emblem's ability is not an ability of a permanent at all.
func TestOpponentsOnlySuppressorScopesByTheSourcesController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	installSuppressionCatalog(t)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		switch key {
		case suppSelfETB:
			return []TriggeredAbility{suppTrigger("self etb", EventETB, thisObject)}
		case suppWatcher:
			return []TriggeredAbility{suppTrigger("watch", EventETB, anotherObject)}
		case EmblemKey(suppWatcher):
			return []TriggeredAbility{suppTrigger("emblem watch", EventETB, anotherObject)}
		}
		return nil
	})
	withCatalogTriggerDoublers(t, func(string) []TriggerDoubler { return nil })
	pushBattlefieldForTest(g, me.ID, "Elesh", "Legendary Creature — Phyrexian Praetor", suppElesh)
	mine := seedTriggerSource(g, me, suppWatcher)
	theirs := seedTriggerSource(g, opp, suppWatcher)
	giveTimingEmblem(g, opp, suppWatcher)

	// The opponent's creature enters: its own trigger and the opponent's
	// watcher are stopped; mine and the opponent's emblem are not.
	suppPutFromHand(t, g, suppInHand(opp, "Their Mulldrifter", "Creature — Elemental", suppSelfETB))
	if got := suppQueued(g, "self etb"); got != 0 {
		t.Errorf("the opponent's own enters trigger triggered %d times, want 0", got)
	}
	bySource := map[uuid.UUID]int{}
	for _, it := range g.PendingTriggers {
		if it.Label == "watch" {
			bySource[it.SourceCardID]++
		}
	}
	if bySource[mine] != 1 || bySource[theirs] != 0 {
		t.Errorf("watchers triggered mine=%d theirs=%d, want 1 and 0", bySource[mine], bySource[theirs])
	}
	if got := suppQueued(g, "emblem watch"); got != 1 {
		t.Errorf("the opponent's emblem triggered %d times, want 1: an emblem is not a permanent", got)
	}
}

// A trigger that is not caused by anything entering is untouched by an
// entering suppressor, even with the suppressor on the battlefield: a
// draw, and a creature dying.
func TestEnteringSuppressorLeavesOtherTriggersAlone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "Orb", "Artifact", suppOrb)
	seedTriggerSource(g, me, suppDraw)
	seedTriggerSource(g, me, suppDiesAny)
	victim := pushBattlefieldForTest(g, me.ID, "Victim", "Creature — Bear", "")

	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventDrawCard, Actor: me.ID}) })
	if got := suppQueued(g, "draw watch"); got != 1 {
		t.Errorf("a draw trigger triggered %d times under an entering suppressor, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	if got := suppQueued(g, "dies watch"); got != 1 {
		t.Errorf("a dies trigger triggered %d times under an entering suppressor, want 1", got)
	}
}

// A suppressed "you may" is never asked: the ability did not trigger,
// so there is no instance to put a question to.
func TestSuppressedOptionalTriggerIsNotPrompted(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "Orb", "Artifact", suppOrb)
	seedTriggerSource(g, opp, suppOptional)

	suppPutFromHand(t, g, suppInHand(me, "Bear", "Creature — Bear", ""))
	if len(g.PendingChoices) != 0 || suppQueued(g, "optional watch") != 0 {
		t.Fatalf("a suppressed optional trigger left choices=%d queued=%d, want none",
			len(g.PendingChoices), suppQueued(g, "optional watch"))
	}
}

// A suppressed targeted trigger never asks for a target either.
func TestSuppressedTargetedTriggerDoesNotAskForATarget(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "Orb", "Artifact", suppOrb)
	seedTriggerSource(g, opp, suppTargeted)

	suppPutFromHand(t, g, suppInHand(me, "Bear", "Creature — Bear", ""))
	if len(g.PendingChoices) != 0 || suppQueued(g, "targeted watch") != 0 {
		t.Fatalf("a suppressed targeted trigger left choices=%d queued=%d, want none",
			len(g.PendingChoices), suppQueued(g, "targeted watch"))
	}
}

// --- composition with the doubler -------------------------------------

// Elesh Norn across two players: my watcher triggers twice (my doubler),
// the opponent's does not trigger at all (my suppressor). Nothing the
// opponent controls is doubled into existence.
func TestEleshDoublesMineAndSuppressesTheOpponents(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "Elesh", "Legendary Creature — Phyrexian Praetor", suppElesh)
	mine := seedTriggerSource(g, me, suppWatcher)
	theirs := seedTriggerSource(g, opp, suppWatcher)

	suppPutFromHand(t, g, suppInHand(opp, "Their Bear", "Creature — Bear", ""))
	bySource := map[uuid.UUID]int{}
	for _, it := range g.PendingTriggers {
		bySource[it.SourceCardID]++
	}
	if bySource[mine] != 2 || bySource[theirs] != 0 {
		t.Fatalf("watchers triggered mine=%d theirs=%d, want 2 (doubled) and 0 (suppressed)",
			bySource[mine], bySource[theirs])
	}
}

// Two Elesh Norns under different players: each player's ability is
// stopped by the OTHER player's Elesh Norn, so nothing triggers, and a
// player's own Elesh Norn has nothing to double. The suppressor is
// asked first; a doubler never adds an instance to an ability that did
// not trigger.
func TestTwoEleshNornsStopEachOthersTriggers(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "My Elesh", "Legendary Creature — Phyrexian Praetor", suppElesh)
	pushBattlefieldForTest(g, opp.ID, "Their Elesh", "Legendary Creature — Phyrexian Praetor", suppElesh)
	seedTriggerSource(g, me, suppWatcher)
	seedTriggerSource(g, opp, suppWatcher)

	suppPutFromHand(t, g, suppInHand(me, "Bear", "Creature — Bear", suppSelfETB))
	if got := len(g.PendingTriggers) + len(g.StackMeta) + len(g.PendingChoices); got != 0 {
		t.Fatalf("with an Elesh Norn on each side, %d instances triggered, want 0", got)
	}
}

// A suppressed match does not spend a "whenever one or more" slot: the
// creature member of the batch is stopped, and the artifact member
// after it still triggers the ability once. Asking the suppressor
// AFTER the once-per-batch guard would have spent the slot on an event
// that triggered nothing.
func TestSuppressedMatchDoesNotSpendTheOncePerBatchSlot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "Orb", "Artifact", suppOrb)
	seedTriggerSource(g, me, suppOneOrMor)
	creature := suppInHand(me, "Bear", "Creature — Bear", "")
	artifact := suppInHand(me, "Trinket", "Artifact", "")

	g.WithWriteLock(func() {
		err := g.PutOntoBattlefieldTogetherThenForEffect([]BatchEntry{
			{CardID: creature, From: ZoneHand},
			{CardID: artifact, From: ZoneHand},
		}, ZoneEntryOptions{}, func(*Game, []uuid.UUID) error { return nil })
		if err != nil {
			t.Fatalf("put together: %v", err)
		}
	})
	if got := suppQueued(g, "one or more"); got != 1 {
		t.Fatalf("the one-or-more watcher triggered %d times, want 1 (for the artifact)", got)
	}
}

// --- which board is asked ------------------------------------------------

// A suppressor entering WITH the creature applies: after the event both
// are on the battlefield, and that is the board an enters trigger is
// judged on (CR 603.10). This is Torpor Orb and a creature returned
// together, and it is Elesh Norn's own entry.
func TestSuppressorEnteringWithTheCreatureApplies(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	orb := suppInHand(me, "Orb", "Artifact", suppOrb)
	creature := suppInHand(me, "Mulldrifter", "Creature — Elemental", suppSelfETB)

	g.WithWriteLock(func() {
		err := g.PutOntoBattlefieldTogetherThenForEffect([]BatchEntry{
			{CardID: creature, From: ZoneHand},
			{CardID: orb, From: ZoneHand},
		}, ZoneEntryOptions{}, func(*Game, []uuid.UUID) error { return nil })
		if err != nil {
			t.Fatalf("put together: %v", err)
		}
	})
	if got := suppQueued(g, "self etb"); got != 0 {
		t.Fatalf("a creature entering together with the suppressor triggered %d times, want 0", got)
	}
}

// The rule this PR chose for a suppressor that leaves the battlefield
// at the same time as a permanent enters: an enters trigger is judged on
// the board AFTER the event (CR 603.10), and a suppressor that has
// already left is not on it. So the creature's own trigger happens.
// The suppressor here is a member of an open simultaneous exit, already
// moved to its owner's graveyard, when the entry is announced. That is
// the one way the harvest can still see a permanent that has left. No
// production path announces an entry while an exit batch is open today,
// so the test builds the board the rule is about.
//
// A suppressor still on the battlefield when the entry is announced
// applies, even if it is on its way out (the second half).
func TestSuppressorThatLeftInTheSameExitDoesNotStopAnEntry(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	orb := pushBattlefieldForTest(g, me.ID, "Orb", "Artifact", suppOrb)
	gone := suppInHand(me, "Mulldrifter", "Creature — Elemental", suppSelfETB)

	g.WithWriteLock(func() {
		closeBatch := g.beginSimultaneousExitLocked([]uuid.UUID{orb})
		defer closeBatch()
		left, err := g.Battlefield.Remove(orb)
		if err != nil {
			t.Fatalf("remove the suppressor: %v", err)
		}
		me.Graveyard.PushTop(left)
		if _, err := g.PutFromHandOntoBattlefieldForEffect(gone, HandEntryOptions{}); err != nil {
			t.Fatalf("put from hand: %v", err)
		}
	})
	if got := suppQueued(g, "self etb"); got != 1 {
		t.Fatalf("a suppressor that had left still stopped an entry: %d instances, want 1", got)
	}

	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	orb2 := pushBattlefieldForTest(g2, me2.ID, "Orb", "Artifact", suppOrb)
	staying := suppInHand(me2, "Mulldrifter", "Creature — Elemental", suppSelfETB)
	g2.WithWriteLock(func() {
		closeBatch := g2.beginSimultaneousExitLocked([]uuid.UUID{orb2})
		defer closeBatch()
		if _, err := g2.PutFromHandOntoBattlefieldForEffect(staying, HandEntryOptions{}); err != nil {
			t.Fatalf("put from hand: %v", err)
		}
	})
	if got := suppQueued(g2, "self etb"); got != 0 {
		t.Fatalf("a suppressor still on the battlefield did not stop the entry: %d instances, want 0", got)
	}
}

// --- dying (Hushbringer) -------------------------------------------------

// A creature dying under the dying suppressor triggers nothing: neither
// its own "when this dies" nor another permanent's "whenever a creature
// dies". A creature EXILED is not dying, and a leave trigger still
// happens.
func TestDyingSuppressorStopsDiesTriggers(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "Hush", "Creature — Faerie", suppHush)
	seedTriggerSource(g, me, suppDiesAny)
	seedTriggerSource(g, me, suppLeaves)
	dier := pushBattlefieldForTest(g, me.ID, "Dier", "Creature — Bear", suppDiesSelf)
	exiled := pushBattlefieldForTest(g, me.ID, "Exiled", "Creature — Bear", "")

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dier) })
	for _, label := range []string{"dies self", "dies watch", "leaves watch"} {
		if got := suppQueued(g, label); got != 0 {
			t.Errorf("%q triggered %d times when a creature died under the suppressor, want 0", label, got)
		}
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(exiled) })
	if got := suppQueued(g, "leaves watch"); got != 1 {
		t.Errorf("a creature exiled (not dying) triggered the leave watcher %d times, want 1", got)
	}
}

// A dies trigger looks back in time (CR 603.10a), so the suppressor is
// read as it was immediately BEFORE the event: a Hushbringer that dies
// in the same wipe still stops every death in it, and its own death
// triggers nothing either.
func TestDyingSuppressorLooksBackThroughAWipeAndItsOwnDeath(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	installSuppressionCatalog(t)
	hush := pushBattlefieldForTest(g, me.ID, "Hush", "Creature — Faerie", suppHush)
	artist := seedTriggerSource(g, opp, suppDiesAny)
	dier := pushBattlefieldForTest(g, me.ID, "Dier", "Creature — Bear", suppDiesSelf)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{dier, hush, artist}) })
	if got := len(g.PendingTriggers) + len(g.StackMeta); got != 0 {
		t.Fatalf("a wipe with the suppressor in it triggered %d instances, want 0", got)
	}

	g2 := newActiveGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	hush2 := pushBattlefieldForTest(g2, me2.ID, "Hush", "Creature — Faerie", suppHush)
	seedTriggerSource(g2, opp2, suppDiesAny)
	g2.WithWriteLock(func() { _ = g2.DestroyPermanentForEffect(hush2) })
	if got := suppQueued(g2, "dies watch"); got != 0 {
		t.Fatalf("the suppressor's own death triggered a dies watcher %d times, want 0", got)
	}
}

// An event-conditioned delayed trigger is an ability too: an earthbent
// land that dies under the dying suppressor does not come back, and the
// trigger stays queued rather than firing.
func TestDyingSuppressorStopsAnEarthbendReturn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	pushBattlefieldForTest(g, me.ID, "Hush", "Creature — Faerie", suppHush)
	land := earthbendTestLand(t, g, me, "Forest")
	earthbend(t, g, me, land, 2)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		_ = g.DestroyPermanentForEffect(land)
	})
	settleStack(t, g)
	if findCard(g, land) != nil {
		t.Fatal("the earthbent land came back although its dying caused nothing to trigger")
	}
	if !me.Graveyard.Contains(land) {
		t.Fatal("the earthbent land is not in the graveyard")
	}
	if len(g.DelayedTriggers) == 0 {
		t.Fatal("the delayed return was consumed although it did not trigger")
	}
}
