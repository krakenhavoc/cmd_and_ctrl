package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CR 702.93a / 702.79a: undying returns the creature with a +1/+1
// counter, persist with a -1/-1 counter, each under its owner's control.
func TestUndyingAndPersistReturnTheCreatureWithTheirCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	wolf := effectstest.PushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)
	elite := effectstest.PushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)

	effectstest.B25Destroy(g, wolf)
	if n := effectstest.DiesReturnItems(g); n != 1 {
		t.Fatalf("undying put %d triggers on the stack, want 1", n)
	}
	if effects.OnBattlefield(g, wolf) {
		t.Fatal("the creature came back before its trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, wolf) || effectstest.CountersOn(g, wolf, game.CounterPlusOne) != 1 {
		t.Fatalf("undying: on battlefield %v with %d +1/+1 counters, want true and 1",
			effects.OnBattlefield(g, wolf), effectstest.CountersOn(g, wolf, game.CounterPlusOne))
	}

	effectstest.B25Destroy(g, elite)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, elite) || effectstest.CountersOn(g, elite, game.CounterMinusOne) != 1 {
		t.Fatalf("persist: on battlefield %v with %d -1/-1 counters, want true and 1",
			effects.OnBattlefield(g, elite), effectstest.CountersOn(g, elite, game.CounterMinusOne))
	}
}

// The intervening "if" (CR 603.4): a creature that died with the
// keyword's own counter on it does not trigger, and a creature exiled
// instead of dying was never put into a graveyard.
func TestUndyingAndPersistDoNotTriggerWithTheirCounterOrWhenExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	wolf := effectstest.PushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)
	elite := effectstest.PushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)
	exiled := effectstest.PushDiesKeywordCreature(g, me, "Kitchen Wolf", 2, 2, game.KeywordPersist)
	effectstest.AddCounters(t, g, wolf, game.CounterPlusOne, 1)
	effectstest.AddCounters(t, g, elite, game.CounterMinusOne, 1)

	effectstest.B25Destroy(g, wolf)
	effectstest.B25Destroy(g, elite)
	effectstest.ExileCreature(t, g, exiled)
	if n := effectstest.DiesReturnItems(g); n != 0 {
		t.Fatalf("%d dies-return triggers, want none", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{wolf, elite, exiled} {
		if effects.OnBattlefield(g, id) {
			t.Errorf("%s came back", id)
		}
	}
}

// CR 704.5q makes the loops: an undying creature that came back with
// its +1/+1 counter and then got a -1/-1 counter has neither, so it
// dies with no +1/+1 counter and returns again.
func TestUndyingReturnsAgainAfterTheCountersAnnihilate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	wolf := effectstest.PushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)

	effectstest.B25Destroy(g, wolf)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, wolf, game.CounterPlusOne) != 1 {
		t.Fatal("setup: the first return has no +1/+1 counter")
	}
	effectstest.AddCounters(t, g, wolf, game.CounterMinusOne, 1)
	g.RunStateChecksForTest()
	if p, m := effectstest.CountersOn(g, wolf, game.CounterPlusOne), effectstest.CountersOn(g, wolf, game.CounterMinusOne); p != 0 || m != 0 {
		t.Fatalf("after CR 704.5q: %d +1/+1 and %d -1/-1 counters, want none", p, m)
	}
	effectstest.B25Destroy(g, wolf)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, wolf) || effectstest.CountersOn(g, wolf, game.CounterPlusOne) != 1 {
		t.Fatal("the annihilated undying creature did not return a second time")
	}
}

// The persist ruling: a creature with a +1/+1 counter that gets enough
// -1/-1 counters to die still had -1/-1 counters as it last existed,
// so persist does not trigger.
func TestPersistCreatureKilledByMinusCountersOverAPlusCounterStaysDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	elite := effectstest.PushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)
	effectstest.AddCounters(t, g, elite, game.CounterPlusOne, 1)
	effectstest.AddCounters(t, g, elite, game.CounterMinusOne, 3)
	g.RunStateChecksForTest()
	if effects.OnBattlefield(g, elite) {
		t.Fatal("setup: a 3/3 with three -1/-1 counters survived")
	}
	if n := effectstest.DiesReturnItems(g); n != 0 {
		t.Fatalf("persist triggered %d times on a creature that died with -1/-1 counters", n)
	}
}

// CR 704.3: every state-based action in one check happens at once, so
// a creature that dies in the same check that would annihilate its
// counters dies with BOTH kinds. A 1/1 undying creature with a +1/+1
// counter and two -1/-1 counters had a +1/+1 counter as it last existed,
// and undying does not trigger.
func TestACreatureThatDiesWithBothKindsOfCounterTriggersNeither(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	wolf := effectstest.PushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying, game.KeywordPersist)
	effectstest.AddCounters(t, g, wolf, game.CounterPlusOne, 1)
	effectstest.AddCounters(t, g, wolf, game.CounterMinusOne, 2)
	g.RunStateChecksForTest()
	if effects.OnBattlefield(g, wolf) {
		t.Fatal("setup: a 1/1 with +1/+1 and two -1/-1 counters survived")
	}
	if n := effectstest.DiesReturnItems(g); n != 0 {
		t.Fatalf("%d dies-return triggers from a creature that died with both kinds of counter, want none", n)
	}
}

// CR 111.7: a token with persist triggers, and there is nothing to
// return — no error either.
func TestATokenWithPersistTriggersAndReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
	errs := len(effectstest.EffectErrors(g))
	effectstest.B25Destroy(g, token)
	if n := effectstest.DiesReturnItems(g); n != 1 {
		t.Fatalf("a dying persist token put %d triggers on the stack, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, token) {
		t.Error("a token that ceased to exist came back")
	}
	if got := effectstest.EffectErrors(g); len(got) != errs {
		t.Errorf("effect errors: %v", got[errs:])
	}
}

// CR 400.7: a card that left the graveyard in response is a new object
// and stays where it is — exiled, or exiled and put back.
func TestPersistReturnsNothingOnceTheCardLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	gone := effectstest.PushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)
	back := effectstest.PushDiesKeywordCreature(g, me, "Kitchen Wolf", 2, 2, game.KeywordPersist)
	effectstest.B25Destroy(g, gone)
	effectstest.B25Destroy(g, back)
	if n := effectstest.DiesReturnItems(g); n != 2 {
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
	effectstest.SettleProwess(t, g)
	if effects.OnBattlefield(g, gone) || effects.OnBattlefield(g, back) {
		t.Errorf("a card that left the graveyard came back (exiled %v, put back %v)",
			effects.OnBattlefield(g, gone), effects.OnBattlefield(g, back))
	}
}

// CR 603.3a: a stolen Murderous Redcap's persist trigger is the
// thief's, and the Redcap returns under its owner's control.
func TestAStolenRedcapPersistsUnderItsOwnersControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	redcap := effectstest.B11Push(g, me.ID, "Murderous Redcap", "Creature — Goblin Assassin", effectstest.OracleMurderousRedcap, 2, 2)
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), redcap, thief.ID, game.IndefiniteDuration(), "test — steal") {
			t.Fatal("setup: the steal did not register")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	effectstest.B25Destroy(g, redcap)
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
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(redcap)
	if !ok || !effects.OnBattlefield(g, redcap) {
		t.Fatal("the Redcap did not return")
	}
	if c.Controller != me.ID {
		t.Errorf("the Redcap returned under %s, want its owner", c.Controller)
	}
	if effectstest.CountersOn(g, redcap, game.CounterMinusOne) != 1 {
		t.Error("the Redcap returned without its -1/-1 counter")
	}
}

// CR 122.6: the counter is put on the creature as it enters, so
// Hardened Scales adds one more.
func TestUndyingUnderHardenedScalesReturnsWithTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me)
	wolf := effectstest.PushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)
	effectstest.B25Destroy(g, wolf)
	effectstest.SettleProwess(t, g)
	if n := effectstest.CountersOn(g, wolf, game.CounterPlusOne); n != 2 {
		t.Errorf("undying under Hardened Scales returned with %d +1/+1 counters, want 2", n)
	}
}

// CR 113.2c: a printed undying and a granted one are two triggers; the
// first returns the creature and the second finds a new object and does
// nothing. A creature that died wearing only the grant has it too
// (CR 603.10a).
func TestTwoUndyingInstancesReturnTheCreatureOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Undying Lord", TypeLine: "Creature — Zombie",
		OracleID: effectstest.UndyingLordOracle, Power: 2, Toughness: 2, Owner: me, Controller: me,
	})
	wolf := effectstest.PushDiesKeywordCreature(g, me, "Young Wolf", 1, 1, game.KeywordUndying)
	bear := effectstest.PushDiesKeywordCreature(g, me, "Grizzly Bears", 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	effectstest.B25Destroy(g, wolf)
	if n := effectstest.DiesReturnItems(g); n != 2 {
		t.Fatalf("printed + granted undying put %d triggers on the stack, want 2", n)
	}
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, wolf) || effectstest.CountersOn(g, wolf, game.CounterPlusOne) != 1 {
		t.Fatalf("on battlefield %v with %d counters, want once with one", effects.OnBattlefield(g, wolf), effectstest.CountersOn(g, wolf, game.CounterPlusOne))
	}

	effectstest.B25Destroy(g, bear)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, bear) || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("a creature that died wearing a granted undying did not return")
	}
}

// CR 603.3b: a wipe kills persist creatures on two sides; the active
// player's trigger goes on the stack first, so the other player's
// resolves first.
func TestABoardWipeOrdersPersistTriggersInAPNAPOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ap := g.Seats[g.Turn.ActiveSeat].ID
	nap := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	mine := effectstest.PushDiesKeywordCreature(g, ap, "Safehold Elite", 2, 2, game.KeywordPersist)
	theirs := effectstest.PushDiesKeywordCreature(g, nap, "Kitchen Wolf", 2, 2, game.KeywordPersist)
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
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, mine) || !effects.OnBattlefield(g, theirs) {
		t.Error("both persist creatures should have returned")
	}
}

// A Clone that copied a persist creature has persist (layer 1), returns
// as a Clone, and chooses what to copy again.
func TestACloneOfAPersistCreatureReturnsAndCopiesAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	finks := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kitchen Finks", TypeLine: "Creature — Ouphe",
		OracleID: "oracle-kitchen-finks-2075", ManaCost: "{1}{G/W}{G/W}", Power: 3, Toughness: 2,
		Keywords: []string{game.KeywordPersist}, Owner: me, Controller: me,
	})
	bear := effectstest.SeedCopyableCreature(g, me, "Grizzly Bears", "Creature — Bear", 2, 2)

	clone := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, finks)
	if c := effectstest.CopyBattlefieldCard(t, g, clone); c.Effective().Name != "Kitchen Finks" {
		t.Fatalf("setup: the Clone is %q, want a Kitchen Finks", c.Effective().Name)
	}

	effectstest.B25Destroy(g, clone)
	if n := effectstest.DiesReturnItems(g); n != 1 {
		t.Fatalf("the Clone of a persist creature put %d persist triggers on the stack, want 1", n)
	}
	effectstest.ResolveWithCopyChoice(t, g, bear)
	c := effectstest.CopyBattlefieldCard(t, g, clone)
	if c.Effective().Name != "Grizzly Bears" {
		t.Errorf("the returned Clone is %q, want the Grizzly Bears it chose this time", c.Effective().Name)
	}
	if effectstest.CountersOn(g, clone, game.CounterMinusOne) != 1 {
		t.Error("the returned Clone has no -1/-1 counter")
	}
}

// A persist trigger waiting on the stack survives a capture → JSON →
// restore, and the restored game returns the creature.
func TestAPersistTriggerSurvivesASnapshotRoundTrip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	elite := effectstest.PushDiesKeywordCreature(g, me, "Safehold Elite", 2, 2, game.KeywordPersist)
	effectstest.B25Destroy(g, elite)
	for i := 0; i < 4 && len(g.PendingTriggers) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if it := effectstest.TriggerOnStack(g, elite); it == nil || it.Body != "persist/return" {
		t.Fatalf("setup: no persist trigger on the stack: %+v", it)
	}
	restored := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.SettleProwess(t, restored)
	if !effects.OnBattlefield(restored, elite) || effectstest.CountersOn(restored, elite, game.CounterMinusOne) != 1 {
		t.Error("the restored persist trigger did not return the creature with a -1/-1 counter")
	}
}

// --- the cards -------------------------------------------------------

// Gleeful Arsonist: an opponent's noncreature spell, and only that,
// deals damage equal to its power to the caster; undying makes it
// a 2/3 that hits for 2.
func TestGleefulArsonist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arsonist := effectstest.B11Push(g, me.ID, "Gleeful Arsonist", "Creature — Human Wizard", effectstest.OracleGleefulArsonist, 1, 2)
	effectstest.AssertKeywords(t, g, arsonist, game.KeywordUndying)

	life := opp.Life
	effectstest.CastFromSeat(t, g, opp, "Opt", "Instant", "", nil)
	effectstest.SettleProwess(t, g)
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
	effectstest.SettleProwess(t, g)
	if opp.Life != life-1 {
		t.Errorf("an opponent's creature spell dealt damage: life %d", opp.Life)
	}

	effectstest.B25Destroy(g, arsonist)
	effectstest.SettleProwess(t, g)
	if effectstest.CountersOn(g, arsonist, game.CounterPlusOne) != 1 {
		t.Fatal("undying did not return the Arsonist with a +1/+1 counter")
	}
	effectstest.CastFromSeat(t, g, opp, "Opt", "Instant", "", nil)
	effectstest.SettleProwess(t, g)
	if opp.Life != life-3 {
		t.Errorf("after undying: life %d, want %d (2 damage from a 2/3)", opp.Life, life-3)
	}
}

// Persistent Constrictor: the upkeep's player loses 1 life and their
// chosen creature gets a -1/-1 counter; with no creature to choose the
// life is still lost; an illegal target means no life is lost either.
func TestPersistentConstrictor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	constrictor := effectstest.B11Push(g, me.ID, "Persistent Constrictor", "Creature — Zombie Snake", effectstest.OraclePersistentConstrictr, 5, 3)
	effectstest.AssertKeywords(t, g, constrictor, game.KeywordPersist)

	// No creature to choose: the life is lost anyway.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	life := opp.Life
	effectstest.SettleProwess(t, g)
	if opp.Life != life-1 {
		t.Fatalf("no target: life %d → %d, want -1", life, opp.Life)
	}

	// A creature of theirs: the counter goes on it.
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	nextUpkeepOf(t, g, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	for _, id := range pick.PickTargetCards {
		if id == mine {
			t.Error("a creature the upkeep's player does not control was offered")
		}
	}
	life = opp.Life
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.SettleProwess(t, g)
	if opp.Life != life-1 || effectstest.CountersOn(g, bear, game.CounterMinusOne) != 1 {
		t.Fatalf("with a target: life %d → %d, counters %d", life, opp.Life, effectstest.CountersOn(g, bear, game.CounterMinusOne))
	}

	// The target leaves in response: the ability does not resolve.
	nextUpkeepOf(t, g, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	life = opp.Life
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.B25Destroy(g, bear)
	effectstest.SettleProwess(t, g)
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
	effectstest.AdvanceToUpkeepOf(t, g, seat)
}
