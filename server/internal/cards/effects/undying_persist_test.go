package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undying_persist_test.go — #2075, CR 702.93 and CR 702.79 (ADR 0113
// §4). Undying and persist are canonical keyword tokens whose dies
// triggers the engine derives from the departed permanent's last-known
// ability list (game/undying_persist.go). One test per interaction the
// ADR lists, then one per card that lands with the seam.

const (
	oracleGleefulArsonist      = "9025b3ce-f6ae-431d-9eb5-02d713b8c9b8"
	oraclePersistentConstrictr = "bf0ee9da-6071-440c-9149-d441f730fcfb"
	oracleMurderousRedcap      = "a498bc70-36e7-4454-bc44-906893df38b8"

	// undyingLordOracle is a test-only lord: "Other creatures you
	// control have undying." Mikaeus, the Unhallowed's grant, without
	// the rest of Mikaeus.
	undyingLordOracle = "test-2075-undying-lord"
)

func init() {
	Register(Spec{
		OracleID: undyingLordOracle,
		Name:     "Undying Lord",
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, game.KeywordUndying),
		},
	})
}

// pushDiesKeywordCreature is a creature with no catalog entry whose
// printed keywords the deck importer stamped — Young Wolf's and Safehold
// Elite's case, the proof that they need no card file.
func pushDiesKeywordCreature(g *game.Game, controller uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Wolf",
		Power: power, Toughness: toughness, Keywords: keywords,
		Owner: controller, Controller: controller,
	})
}

// diesReturnItems counts the undying and persist triggers waiting on
// the stack or headed for it.
func diesReturnItems(g *game.Game) int {
	n := 0
	is := func(item *game.StackItem) bool {
		return item != nil && (item.Body == "undying/return" || item.Body == "persist/return")
	}
	for _, item := range g.StackMeta {
		if is(item) {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if is(item) {
			n++
		}
	}
	return n
}

func addCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, kind, n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
}

func effectErrors(g *game.Game) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventEffectError {
				out = append(out, ev.ErrorMsg)
			}
		}
	})
	return out
}

// CR 702.93a / 702.79a: undying returns the creature with a +1/+1
// counter, persist with a -1/-1 counter, each under its owner's control.
func TestUndyingAndPersistReturnTheCreatureWithTheirCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	wolf := pushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)
	elite := pushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)

	b25Destroy(g, wolf)
	if n := diesReturnItems(g); n != 1 {
		t.Fatalf("undying put %d triggers on the stack, want 1", n)
	}
	if onBattlefield(g, wolf) {
		t.Fatal("the creature came back before its trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, wolf) || countersOn(g, wolf, game.CounterPlusOne) != 1 {
		t.Fatalf("undying: on battlefield %v with %d +1/+1 counters, want true and 1",
			onBattlefield(g, wolf), countersOn(g, wolf, game.CounterPlusOne))
	}

	b25Destroy(g, elite)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, elite) || countersOn(g, elite, game.CounterMinusOne) != 1 {
		t.Fatalf("persist: on battlefield %v with %d -1/-1 counters, want true and 1",
			onBattlefield(g, elite), countersOn(g, elite, game.CounterMinusOne))
	}
}

// The intervening "if" (CR 603.4): a creature that died with the
// keyword's own counter on it does not trigger, and a creature exiled
// instead of dying was never put into a graveyard.
func TestUndyingAndPersistDoNotTriggerWithTheirCounterOrWhenExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	wolf := pushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)
	elite := pushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)
	exiled := pushDiesKeywordCreature(g, me, "Kitchen Wolf", 2, 2, game.KeywordPersist)
	addCounters(t, g, wolf, game.CounterPlusOne, 1)
	addCounters(t, g, elite, game.CounterMinusOne, 1)

	b25Destroy(g, wolf)
	b25Destroy(g, elite)
	exileCreature(t, g, exiled)
	if n := diesReturnItems(g); n != 0 {
		t.Fatalf("%d dies-return triggers, want none", n)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{wolf, elite, exiled} {
		if onBattlefield(g, id) {
			t.Errorf("%s came back", id)
		}
	}
}

// CR 704.5q makes the loops: an undying creature that came back with
// its +1/+1 counter and then got a -1/-1 counter has neither, so it
// dies with no +1/+1 counter and returns again.
func TestUndyingReturnsAgainAfterTheCountersAnnihilate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	wolf := pushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)

	b25Destroy(g, wolf)
	passPriorityAroundTable(t, g)
	if countersOn(g, wolf, game.CounterPlusOne) != 1 {
		t.Fatal("setup: the first return has no +1/+1 counter")
	}
	addCounters(t, g, wolf, game.CounterMinusOne, 1)
	g.RunStateChecksForTest()
	if p, m := countersOn(g, wolf, game.CounterPlusOne), countersOn(g, wolf, game.CounterMinusOne); p != 0 || m != 0 {
		t.Fatalf("after CR 704.5q: %d +1/+1 and %d -1/-1 counters, want none", p, m)
	}
	b25Destroy(g, wolf)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, wolf) || countersOn(g, wolf, game.CounterPlusOne) != 1 {
		t.Fatal("the annihilated undying creature did not return a second time")
	}
}

// The persist ruling: a creature with a +1/+1 counter that gets enough
// -1/-1 counters to die still had -1/-1 counters as it last existed,
// so persist does not trigger.
func TestPersistCreatureKilledByMinusCountersOverAPlusCounterStaysDead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	elite := pushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)
	addCounters(t, g, elite, game.CounterPlusOne, 1)
	addCounters(t, g, elite, game.CounterMinusOne, 3)
	g.RunStateChecksForTest()
	if onBattlefield(g, elite) {
		t.Fatal("setup: a 3/3 with three -1/-1 counters survived")
	}
	if n := diesReturnItems(g); n != 0 {
		t.Fatalf("persist triggered %d times on a creature that died with -1/-1 counters", n)
	}
}

// CR 704.3: every state-based action in one check happens at once, so
// a creature that dies in the same check that would annihilate its
// counters dies with BOTH kinds. A 1/1 undying creature with a +1/+1
// counter and two -1/-1 counters had a +1/+1 counter as it last existed,
// and undying does not trigger.
func TestACreatureThatDiesWithBothKindsOfCounterTriggersNeither(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	wolf := pushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying, game.KeywordPersist)
	addCounters(t, g, wolf, game.CounterPlusOne, 1)
	addCounters(t, g, wolf, game.CounterMinusOne, 2)
	g.RunStateChecksForTest()
	if onBattlefield(g, wolf) {
		t.Fatal("setup: a 1/1 with +1/+1 and two -1/-1 counters survived")
	}
	if n := diesReturnItems(g); n != 0 {
		t.Fatalf("%d dies-return triggers from a creature that died with both kinds of counter, want none", n)
	}
}

// CR 111.7: a token with persist triggers, and there is nothing to
// return — no error either.
func TestATokenWithPersistTriggersAndReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me, game.Card{
			Name: "Persist Spirit", TypeLine: "Token Creature — Spirit",
			Power: 1, Toughness: 1, Keywords: []string{game.KeywordPersist},
		}, 1)
	})
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Persist Spirit" {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("setup: no token")
	}
	errs := len(effectErrors(g))
	b25Destroy(g, token)
	if n := diesReturnItems(g); n != 1 {
		t.Fatalf("a dying persist token put %d triggers on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, token) {
		t.Error("a token that ceased to exist came back")
	}
	if got := effectErrors(g); len(got) != errs {
		t.Errorf("effect errors: %v", got[errs:])
	}
}

// CR 400.7: a card that left the graveyard in response is a new object
// and stays where it is — exiled, or exiled and put back.
func TestPersistReturnsNothingOnceTheCardLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	gone := pushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)
	back := pushDiesKeywordCreature(g, me, "Kitchen Wolf", 2, 2, game.KeywordPersist)
	b25Destroy(g, gone)
	b25Destroy(g, back)
	if n := diesReturnItems(g); n != 2 {
		t.Fatalf("setup: %d triggers, want 2", n)
	}
	yard := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me}
	exile := game.ZoneRef{Kind: game.ZoneExile}
	for _, id := range []uuid.UUID{gone, back} {
		if err := g.MoveCardByID(yard, exile, id); err != nil {
			t.Fatalf("exile from graveyard: %v", err)
		}
	}
	if err := g.MoveCardByID(exile, yard, back); err != nil {
		t.Fatalf("put back: %v", err)
	}
	settleProwess(t, g)
	if onBattlefield(g, gone) || onBattlefield(g, back) {
		t.Errorf("a card that left the graveyard came back (exiled %v, put back %v)",
			onBattlefield(g, gone), onBattlefield(g, back))
	}
}

// CR 603.3a: a stolen Murderous Redcap's persist trigger is the
// thief's, and the Redcap returns under its owner's control.
func TestAStolenRedcapPersistsUnderItsOwnersControl(t *testing.T) {
	g := newCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	redcap := b11Push(g, me.ID, "Murderous Redcap", "Creature — Goblin Assassin", oracleMurderousRedcap, 2, 2)
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), redcap, thief.ID, game.IndefiniteDuration(), "test — steal") {
			t.Fatal("setup: the steal did not register")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	b25Destroy(g, redcap)
	var item *game.StackItem
	for _, it := range g.StackMeta {
		if it != nil && it.Body == "persist/return" {
			item = it
		}
	}
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == "persist/return" {
			item = it
		}
	}
	if item == nil {
		t.Fatal("no persist trigger")
	}
	if item.Controller != thief.ID {
		t.Errorf("the persist trigger is controlled by %s, want the thief", item.Controller)
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(redcap)
	if !ok || !onBattlefield(g, redcap) {
		t.Fatal("the Redcap did not return")
	}
	if c.Controller != me.ID {
		t.Errorf("the Redcap returned under %s, want its owner", c.Controller)
	}
	if countersOn(g, redcap, game.CounterMinusOne) != 1 {
		t.Error("the Redcap returned without its -1/-1 counter")
	}
}

// CR 122.6: the counter is put on the creature as it enters, so
// Hardened Scales adds one more.
func TestUndyingUnderHardenedScalesReturnsWithTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me)
	wolf := pushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)
	b25Destroy(g, wolf)
	settleProwess(t, g)
	if n := countersOn(g, wolf, game.CounterPlusOne); n != 2 {
		t.Errorf("undying under Hardened Scales returned with %d +1/+1 counters, want 2", n)
	}
}

// CR 113.2c: a printed undying and a granted one are two triggers; the
// first returns the creature and the second finds a new object and does
// nothing. A creature that died wearing only the grant has it too
// (CR 603.10a).
func TestTwoUndyingInstancesReturnTheCreatureOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Undying Lord", TypeLine: "Creature — Zombie",
		OracleID: undyingLordOracle, Power: 2, Toughness: 2, Owner: me, Controller: me,
	})
	wolf := pushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)
	bear := pushDiesKeywordCreature(g, me, "Grizzly Bears", 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	b25Destroy(g, wolf)
	if n := diesReturnItems(g); n != 2 {
		t.Fatalf("printed + granted undying put %d triggers on the stack, want 2", n)
	}
	settleProwess(t, g)
	if !onBattlefield(g, wolf) || countersOn(g, wolf, game.CounterPlusOne) != 1 {
		t.Fatalf("on battlefield %v with %d counters, want once with one", onBattlefield(g, wolf), countersOn(g, wolf, game.CounterPlusOne))
	}

	b25Destroy(g, bear)
	settleProwess(t, g)
	if !onBattlefield(g, bear) || countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("a creature that died wearing a granted undying did not return")
	}
}

// CR 603.3b: a wipe kills persist creatures on two sides; the active
// player's trigger goes on the stack first, so the other player's
// resolves first.
func TestABoardWipeOrdersPersistTriggersInAPNAPOrder(t *testing.T) {
	g := newCatalogGame(t)
	ap := g.Seats[g.Turn.ActiveSeat].ID
	nap := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	mine := pushDiesKeywordCreature(g, ap, "Safehold Elite", 2, 2, game.KeywordPersist)
	theirs := pushDiesKeywordCreature(g, nap, "Kitchen Wolf", 2, 2, game.KeywordPersist)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{mine, theirs}) })
	for i := 0; i < 4 && len(g.PendingTriggers) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	seq := map[uuid.UUID]int64{}
	for _, it := range g.StackMeta {
		if it != nil && it.Body == "persist/return" {
			seq[it.SourceCardID] = int64(it.Seq)
		}
	}
	sm, okm := seq[mine]
	st, okt := seq[theirs]
	if !okm || !okt {
		t.Fatalf("both persist triggers should be on the stack: %v", seq)
	}
	if sm > st {
		t.Errorf("the active player's trigger (seq %d) is above the other player's (seq %d)", sm, st)
	}
	settleProwess(t, g)
	if !onBattlefield(g, mine) || !onBattlefield(g, theirs) {
		t.Error("both persist creatures should have returned")
	}
}

// A Clone that copied a persist creature has persist (layer 1), returns
// as a Clone, and chooses what to copy again.
func TestACloneOfAPersistCreatureReturnsAndCopiesAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	finks := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kitchen Finks", TypeLine: "Creature — Ouphe",
		OracleID: "oracle-kitchen-finks-2075", ManaCost: "{1}{G/W}{G/W}", Power: 3, Toughness: 2,
		Keywords: []string{game.KeywordPersist}, Owner: me, Controller: me,
	})
	bear := seedCopyableCreature(g, me, "Grizzly Bears", "Creature — Bear", 2, 2)

	clone := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, finks)
	if c := copyBattlefieldCard(t, g, clone); c.Effective().Name != "Kitchen Finks" {
		t.Fatalf("setup: the Clone is %q, want a Kitchen Finks", c.Effective().Name)
	}

	b25Destroy(g, clone)
	if n := diesReturnItems(g); n != 1 {
		t.Fatalf("the Clone of a persist creature put %d persist triggers on the stack, want 1", n)
	}
	resolveWithCopyChoice(t, g, bear)
	c := copyBattlefieldCard(t, g, clone)
	if c.Effective().Name != "Grizzly Bears" {
		t.Errorf("the returned Clone is %q, want the Grizzly Bears it chose this time", c.Effective().Name)
	}
	if countersOn(g, clone, game.CounterMinusOne) != 1 {
		t.Error("the returned Clone has no -1/-1 counter")
	}
}

// A persist trigger waiting on the stack survives a capture → JSON →
// restore, and the restored game returns the creature.
func TestAPersistTriggerSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	elite := pushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)
	b25Destroy(g, elite)
	for i := 0; i < 4 && len(g.PendingTriggers) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if it := triggerOnStack(g, elite); it == nil || it.Body != "persist/return" {
		t.Fatalf("setup: no persist trigger on the stack: %+v", it)
	}
	restored := restoreRoundTrip(t, g, true)
	settleProwess(t, restored)
	if !onBattlefield(restored, elite) || countersOn(restored, elite, game.CounterMinusOne) != 1 {
		t.Error("the restored persist trigger did not return the creature with a -1/-1 counter")
	}
}

// --- the cards -------------------------------------------------------

// Gleeful Arsonist: an opponent's noncreature spell, and only that,
// deals damage equal to its power to the caster; undying makes it
// a 2/3 that hits for 2.
func TestGleefulArsonist(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arsonist := b11Push(g, me.ID, "Gleeful Arsonist", "Creature — Human Wizard", oracleGleefulArsonist, 1, 2)
	assertKeywords(t, g, arsonist, game.KeywordUndying)

	life := opp.Life
	castFromSeat(t, g, opp, "Opt", "Instant", "", nil)
	settleProwess(t, g)
	if opp.Life != life-1 {
		t.Fatalf("an opponent's instant: life %d → %d, want -1", life, opp.Life)
	}
	flashBear := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: flashBear, Name: "Flash Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"flash"}, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, flashBear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a creature spell: %v", err)
	}
	settleProwess(t, g)
	if opp.Life != life-1 {
		t.Errorf("an opponent's creature spell dealt damage: life %d", opp.Life)
	}

	b25Destroy(g, arsonist)
	settleProwess(t, g)
	if countersOn(g, arsonist, game.CounterPlusOne) != 1 {
		t.Fatal("undying did not return the Arsonist with a +1/+1 counter")
	}
	castFromSeat(t, g, opp, "Opt", "Instant", "", nil)
	settleProwess(t, g)
	if opp.Life != life-3 {
		t.Errorf("after undying: life %d, want %d (2 damage from a 2/3)", opp.Life, life-3)
	}
}

// Persistent Constrictor: the upkeep's player loses 1 life and their
// chosen creature gets a -1/-1 counter; with no creature to choose the
// life is still lost; an illegal target means no life is lost either.
func TestPersistentConstrictor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	constrictor := b11Push(g, me.ID, "Persistent Constrictor", "Creature — Zombie Snake", oraclePersistentConstrictr, 5, 3)
	assertKeywords(t, g, constrictor, game.KeywordPersist)

	// No creature to choose: the life is lost anyway.
	advanceToUpkeepOf(t, g, 1)
	life := opp.Life
	settleProwess(t, g)
	if opp.Life != life-1 {
		t.Fatalf("no target: life %d → %d, want -1", life, opp.Life)
	}

	// A creature of theirs: the counter goes on it.
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	nextUpkeepOf(t, g, 1)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	for _, id := range pick.PickTargetCards {
		if id == mine {
			t.Error("a creature the upkeep's player does not control was offered")
		}
	}
	life = opp.Life
	pickCard(t, g, me.ID, bear)
	settleProwess(t, g)
	if opp.Life != life-1 || countersOn(g, bear, game.CounterMinusOne) != 1 {
		t.Fatalf("with a target: life %d → %d, counters %d", life, opp.Life, countersOn(g, bear, game.CounterMinusOne))
	}

	// The target leaves in response: the ability does not resolve.
	nextUpkeepOf(t, g, 1)
	b04WaitForPick(t, g, me.ID)
	life = opp.Life
	pickCard(t, g, me.ID, bear)
	b25Destroy(g, bear)
	settleProwess(t, g)
	if opp.Life != life {
		t.Errorf("an illegal target: life %d → %d, want no loss", life, opp.Life)
	}
}

// nextUpkeepOf leaves the current step first, so a seat already at its
// upkeep is walked round to its next one.
func nextUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToUpkeepOf(t, g, seat)
}
