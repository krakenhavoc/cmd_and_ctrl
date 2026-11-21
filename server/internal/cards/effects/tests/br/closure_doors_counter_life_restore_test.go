package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Now for Wrath, Now for Ruin! puts a counter on each creature in turn, each
// behind the last one's placement (counter/put-on-each-next, the list still
// owed in its Carry): every placement's order prompt is a restore point, and
// the vigilance and the Ring's temptation run once the last has landed
// (now-for-wrath/vigilance-and-temptation, the creatures in its Carry).
func TestNowForWrathIsARestorePointAtEachCountersOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hs, ds := effectstest.CounterOrderBoard(g, me.ID)
	a := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Now for Wrath, Now for Ruin!", "Sorcery", effectstest.NowForWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for placed := 1; placed <= 2; placed++ {
		if !effectstest.OrderPromptOpen(g) {
			t.Fatalf("placement %d: no order prompt is open: %+v", placed, g.PendingChoices)
		}
		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
		if placed == 2 && effectstest.CountersOn(g, b, "+1/+1") != 0 {
			t.Fatalf("the second creature has counters before its own order was answered")
		}
		effectstest.AnswerOrderBySource(t, g, ds, hs)
	}
	effectstest.RingSettle(t, g)
	effectstest.AnswerRing(t, g, g.Seats[0].ID, a)
	effectstest.RingSettle(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.CountersOn(g, id, "+1/+1"); got != 3 {
			t.Errorf("a creature has %d counters, want one doubled and added to", got)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "vigilance") {
			t.Errorf("a creature lacks vigilance: the restored tail did not run on every creature")
		}
	}
	if effectstest.RingCount(g, g.Seats[0].ID) != 1 {
		t.Error("the Ring did not tempt after the restored counters")
	}
}

// Invigorating Surge's "double the number of +1/+1 counters" waits for the
// first counter (counter/double-plus-one, the creature in its Carry): both
// the first placement and the doubling are restore points.
func TestInvigoratingSurgeIsARestorePointAtTheCountersOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 3) })
	hs, ds := effectstest.CounterOrderBoard(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Invigorating Surge", "Instant", effectstest.B34InvigoratingSurgeOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if got := effectstest.CountersOn(g, bear, "+1/+1"); got != 3 {
		t.Fatalf("the restored table placed counters before the order: %d", got)
	}
	// 1 → 2 (Doubling Season) → 3 (Hardened Scales): 6 counters.
	effectstest.AnswerOrderBySource(t, g, ds, hs)
	if got := effectstest.CountersOn(g, bear, "+1/+1"); got != 6 {
		t.Fatalf("%d counters after the first placement, want 6", got)
	}
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("the restored tail did not start the doubling: %+v", g.PendingChoices)
	}
	// The doubling is a plain placement, with no tail of its own.
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounter)
	// 6 → 12 → 13: 19 counters.
	effectstest.AnswerOrderBySource(t, g, ds, hs)
	if got := effectstest.CountersOn(g, bear, "+1/+1"); got != 19 {
		t.Errorf("%d counters after the doubling, want 19", got)
	}
}

// Insight Engine's charge counter, then a card for each charge counter
// (counter/draw-per-counter): the order prompt is a restore point and the
// restored draw reads what the counter landed at.
func TestInsightEngineIsARestorePointAtTheChargeOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	engine := effectstest.B21Push(g, me.ID, "Insight Engine", "Artifact", effectstest.B23InsightEngineOracle, 0, 0)
	probe := effectstest.SeedReplacementPermanent(g, effectstest.CounterPlusOneProbeOracle, "Counter Plus One Probe", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if got := g.Seats[0].Hand.Size(); got != hand {
		t.Fatalf("the restored table drew %d cards before the order", got-hand)
	}
	// 1 → 2 (Doubling Season) → 3 (the probe).
	effectstest.AnswerOrderBySource(t, g, ds, probe)
	if got := effectstest.CountersOn(g, engine, "charge"); got != 3 {
		t.Fatalf("%d charge counters, want 3", got)
	}
	if got := g.Seats[0].Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want one for each of the 3 charge counters", got)
	}
}

// Devouring Greed's gain is "that much" (life/gain-equal-to-lost, told the
// life that moved): the loss's order prompt is a restore point and the
// restored gain is what the opponent really lost.
func TestDevouringGreedIsARestorePointAtTheLifeLossOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	blood := effectstest.SeedReplacementPermanent(g, effectstest.BloodletterOracle, "Bloodletter of Aclazotz", me.ID)
	probe := effectstest.SeedReplacementPermanent(g, effectstest.LifeLossPlusOneProbeOracle, "Life Loss Probe", me.ID)
	myLife, foeLife := me.Life, foe.Life
	if _, err := effectstest.CastWithTapParams(t, g, "Devouring Greed", "Sorcery — Arcane", "{2}{B}{B}", effectstest.DevouringGreedOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
	if g.Seats[0].Life != myLife {
		t.Fatalf("the restored table gained life before the order")
	}
	// The probe first: (2+1)×2 = 6.
	effectstest.AnswerOrderBySource(t, g, probe, blood)
	if got := foeLife - g.Seats[1].Life; got != 6 {
		t.Fatalf("the opponent lost %d, want 6", got)
	}
	if got := g.Seats[0].Life - myLife; got != 6 {
		t.Errorf("the caster gained %d, want what the opponent really lost", got)
	}
}

// Earthshape's "Then each creature you control with power less than or equal
// to that land's power…" is the earthbend counters' tail (earthshape/shield,
// the land in its Carry, behind engine/earthbend-counters-placed): the order
// prompt is a restore point and the restored shield reads the land at the
// power the counters gave it.
func TestEarthshapeIsARestorePointAtTheCountersOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs, ds := effectstest.CounterOrderBoard(g, me.ID)
	land := effectstest.PushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	five := effectstest.B39Creature(g, me.ID, "Five", "Creature — Beast", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Earthshape", "Instant", effectstest.EarthshapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if effectstest.HasKeywordOnBattlefield(t, g, five, "hexproof") {
		t.Fatal("the restored table shielded creatures before the counters landed")
	}
	// 3 → 6 → 7: a 7/7 land shields the 5-power creature.
	effectstest.AnswerOrderBySource(t, g, ds, hs)
	if p, _, _, _ := effectstest.EarthbentBody(t, g, land); p != 7 {
		t.Fatalf("that land's power = %d, want 7", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, five, "hexproof") || !effectstest.HasKeywordOnBattlefield(t, g, five, "indestructible") {
		t.Error("the restored tail did not shield the 5-power creature")
	}
	if !effectstest.PlayerHasHexproof(g, g.Seats[g.Turn.ActiveSeat]) {
		t.Error("you gain hexproof once the restored sentence runs")
	}
}

// Awaken's animation waits for its counters (engine/awaken-animate, the land
// and its entry stamp in the tail): the order prompt is a restore point and
// the restored land is a 0/0 Elemental creature with the counters on it.
func TestAwakenIsARestorePointAtTheCountersOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// The land is not a creature while the counters go on (CR 608.2c), so
	// Hardened Scales does not apply: Doubling Season and the probe do.
	probe := effectstest.SeedReplacementPermanent(g, effectstest.CounterPlusOneProbeOracle, "Counter Plus One Probe", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	land := effectstest.PushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	g.WithWriteLock(func() {
		if err := g.AwakenForEffect(me.ID, uuid.Nil, land, 2); err != nil {
			t.Fatalf("AwakenForEffect: %v", err)
		}
	})

	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if _, _, _, elemental, _ := effectstest.AwakenedLand(t, g, land); elemental {
		t.Fatal("the restored land was animated before its counters landed")
	}
	// 2 → 4 → 5.
	effectstest.AnswerOrderBySource(t, g, ds, probe)
	if counters, p, tough, elemental, _ := effectstest.AwakenedLand(t, g, land); counters != 5 || p != 5 || tough != 5 || !elemental {
		t.Errorf("land: %d counters, %d/%d, elemental %v; want a 5/5 Elemental", counters, p, tough, elemental)
	}
}

// Mystic Remora's cumulative upkeep charges for the age counters that are
// there once the new one has landed (cumulative-upkeep/<label>, the asking
// trigger riding the tail): the order prompt over the age counter is a
// restore point and the restored payment is for all three.
func TestCumulativeUpkeepIsARestorePointAtTheAgeCounterOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	probe := effectstest.SeedReplacementPermanent(g, effectstest.CounterPlusOneProbeOracle, "Counter Plus One Probe", owner.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", owner.ID)
	remora := effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, remora) == nil {
		t.Fatal("the cumulative upkeep trigger was not put on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if got := effectstest.AgeCountersOn(t, g, remora); got != 0 {
		t.Fatalf("the restored table placed %d age counters before the order", got)
	}
	// 1 → 2 (Doubling Season) → 3 (the probe).
	effectstest.AnswerOrderBySource(t, g, ds, probe)
	if got := effectstest.AgeCountersOn(t, g, remora); got != 3 {
		t.Fatalf("%d age counters, want 3", got)
	}
	choice := effectstest.RemoraUpkeepPrompt(t, g, owner.ID)
	parsed, err := game.ParseCost(choice.PayCost)
	if err != nil {
		t.Fatalf("cost %q does not parse: %v", choice.PayCost, err)
	}
	if parsed.Generic != 3 {
		t.Errorf("the restored upkeep costs %q, want {1} once per age counter (3)", choice.PayCost)
	}
}

// Widespread Brutality's amass onto an Army it already has: the counters'
// order prompt is a restore point, and the restored tail deals the damage the
// grown Army's power measures (engine/amass-counters-placed).
func TestAmassIsARestorePointAtTheCountersOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	hs, ds := effectstest.CounterOrderBoard(g, me.ID)
	army := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie Army", TypeLine: "Token Creature — Zombie Army",
		Power: 0, Toughness: 0, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 1},
		Owner:    me.ID, Controller: me.ID,
	})
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 6, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Widespread Brutality", "Sorcery", effectstest.WidespreadBrutalityOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if w := effectstest.FindBattlefieldCardByID(g, wall); w == nil || w.DamageMarked != 0 {
		t.Fatal("the restored table dealt damage with the amass counters still owed")
	}
	// Amass 2 → 4 → 5, onto the one it had: a 6/6 Army.
	effectstest.AnswerOrderBySource(t, g, ds, hs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, army, game.CounterPlusOne); got != 6 {
		t.Fatalf("Army counters = %d, want 6", got)
	}
	if effectstest.FindBattlefieldCardByID(g, wall) != nil {
		t.Error("the 0/6 wall survived a 6-power Army: the restored tail read the pre-amass power")
	}
}

// Explore's +1/+1 counter, then the graveyard question (engine/explore-ask,
// the explorer and the revealed card in its Carry): the counter's order prompt
// is a restore point and the restored answer still opens the question.
func TestExploreIsARestorePointAtTheCounterOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hs, ds := effectstest.CounterOrderBoard(g, me.ID)
	explorer := effectstest.CtrlPushCreature(g, me.ID, "Explorer")
	top := effectstest.SeedLibraryTop(me, "Some Sorcery", "Sorcery", "{1}")
	epoch := 0
	if c, ok := g.LookupCardForEffect(explorer); ok {
		epoch = c.ObjectEpoch
	}
	g.WithWriteLock(func() {
		if err := g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: explorer, Epoch: epoch}, me.ID); err != nil {
			t.Fatalf("ExploreForEffect: %v", err)
		}
	})

	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if n := len(effectstest.RandomEvents(g, game.EventExplored)); n != 0 {
		t.Fatalf("explored %d times before the counter landed", n)
	}
	effectstest.AnswerOrderBySource(t, g, ds, hs)
	if got := effectstest.CountersOn(g, explorer, game.CounterPlusOne); got != 3 {
		t.Errorf("the explorer has %d counters, want 3", got)
	}
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || len(ask.OptionCarry) != 2 || ask.OptionCarry[0] != top || ask.OptionCarry[1] != explorer {
		t.Fatalf("the restored tail did not ask the graveyard question: %+v", ask)
	}
}
