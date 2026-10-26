package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_doors_counter_life_restore_test.go — #4253 slice 7: the "then" of a
// counter placement and of a life change is a registered tail, so the CR 616
// order prompt that two different counter (or life) replacements ask is a
// restore point. Each test takes a STRICT restore while the prompt is open,
// answers it on the restored table and checks the card still finishes off the
// disk. (strictRestoreWithFrame, answerOrderBySource and the replacement
// seeding helpers are in replacement_resume_restore_test.go,
// counter_tail_cards_test.go and doubling_season_test.go.)

// counterPlusOneProbeOracle is a test card: "If a counter would be put on a
// permanent you control, put one more counter on it instead." Beside Doubling
// Season it makes a placement on any permanent (an enchantment, a land) ask
// for the order, which Hardened Scales only does for a creature.
const counterPlusOneProbeOracle = "test-4253s7-counter-plus-one"

func init() {
	Register(Spec{
		OracleID: counterPlusOneProbeOracle,
		Name:     "Counter Plus One Probe",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Counter Plus One Probe: one more counter",
		}},
	})
}

// counterOrderBoard puts Hardened Scales and Doubling Season under `who`: a
// +1/+1 counter placement on one of their creatures asks for the order.
func counterOrderBoard(g *game.Game, who uuid.UUID) (hs, ds uuid.UUID) {
	hs = seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", who)
	ds = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", who)
	return hs, ds
}

// Now for Wrath, Now for Ruin! puts a counter on each creature in turn, each
// behind the last one's placement (counter/put-on-each-next, the list still
// owed in its Carry): every placement's order prompt is a restore point, and
// the vigilance and the Ring's temptation run once the last has landed
// (now-for-wrath/vigilance-and-temptation, the creatures in its Carry).
func TestNowForWrathIsARestorePointAtEachCountersOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hs, ds := counterOrderBoard(g, me.ID)
	a := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	castCatalogSpell(t, g, "Now for Wrath, Now for Ruin!", "Sorcery", nowForWrathOracle, nil)
	passPriorityAroundTable(t, g)

	for placed := 1; placed <= 2; placed++ {
		if !orderPromptOpen(g) {
			t.Fatalf("placement %d: no order prompt is open: %+v", placed, g.PendingChoices)
		}
		g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
		if placed == 2 && countersOn(g, b, "+1/+1") != 0 {
			t.Fatalf("the second creature has counters before its own order was answered")
		}
		answerOrderBySource(t, g, ds, hs)
	}
	ringSettle(t, g)
	answerRing(t, g, g.Seats[0].ID, a)
	ringSettle(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := countersOn(g, id, "+1/+1"); got != 3 {
			t.Errorf("a creature has %d counters, want one doubled and added to", got)
		}
		if !hasKeywordOnBattlefield(t, g, id, "vigilance") {
			t.Errorf("a creature lacks vigilance: the restored tail did not run on every creature")
		}
	}
	if ringCount(g, g.Seats[0].ID) != 1 {
		t.Error("the Ring did not tempt after the restored counters")
	}
}

// Invigorating Surge's "double the number of +1/+1 counters" waits for the
// first counter (counter/double-plus-one, the creature in its Carry): both
// the first placement and the doubling are restore points.
func TestInvigoratingSurgeIsARestorePointAtTheCountersOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 3) })
	hs, ds := counterOrderBoard(g, me.ID)
	castCatalogSpell(t, g, "Invigorating Surge", "Instant", b34InvigoratingSurgeOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)

	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if got := countersOn(g, bear, "+1/+1"); got != 3 {
		t.Fatalf("the restored table placed counters before the order: %d", got)
	}
	// 1 → 2 (Doubling Season) → 3 (Hardened Scales): 6 counters.
	answerOrderBySource(t, g, ds, hs)
	if got := countersOn(g, bear, "+1/+1"); got != 6 {
		t.Fatalf("%d counters after the first placement, want 6", got)
	}
	if !orderPromptOpen(g) {
		t.Fatalf("the restored tail did not start the doubling: %+v", g.PendingChoices)
	}
	// The doubling is a plain placement, with no tail of its own.
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounter)
	// 6 → 12 → 13: 19 counters.
	answerOrderBySource(t, g, ds, hs)
	if got := countersOn(g, bear, "+1/+1"); got != 19 {
		t.Errorf("%d counters after the doubling, want 19", got)
	}
}

// Insight Engine's charge counter, then a card for each charge counter
// (counter/draw-per-counter): the order prompt is a restore point and the
// restored draw reads what the counter landed at.
func TestInsightEngineIsARestorePointAtTheChargeOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	engine := b21Push(g, me.ID, "Insight Engine", "Artifact", b23InsightEngineOracle, 0, 0)
	probe := seedReplacementPermanent(g, counterPlusOneProbeOracle, "Counter Plus One Probe", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if got := g.Seats[0].Hand.Size(); got != hand {
		t.Fatalf("the restored table drew %d cards before the order", got-hand)
	}
	// 1 → 2 (Doubling Season) → 3 (the probe).
	answerOrderBySource(t, g, ds, probe)
	if got := countersOn(g, engine, "charge"); got != 3 {
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
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	blood := seedReplacementPermanent(g, bloodletterOracle, "Bloodletter of Aclazotz", me.ID)
	probe := seedReplacementPermanent(g, lifeLossPlusOneProbeOracle, "Life Loss Probe", me.ID)
	myLife, foeLife := me.Life, foe.Life
	if _, err := castWithTapParams(t, g, "Devouring Greed", "Sorcery — Arcane", "{2}{B}{B}", devouringGreedOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
	if g.Seats[0].Life != myLife {
		t.Fatalf("the restored table gained life before the order")
	}
	// The probe first: (2+1)×2 = 6.
	answerOrderBySource(t, g, probe, blood)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs, ds := counterOrderBoard(g, me.ID)
	land := pushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	five := b39Creature(g, me.ID, "Five", "Creature — Beast", 5, 5)
	castCatalogSpell(t, g, "Earthshape", "Instant", earthshapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)

	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if hasKeywordOnBattlefield(t, g, five, "hexproof") {
		t.Fatal("the restored table shielded creatures before the counters landed")
	}
	// 3 → 6 → 7: a 7/7 land shields the 5-power creature.
	answerOrderBySource(t, g, ds, hs)
	if p, _, _, _ := earthbentBody(t, g, land); p != 7 {
		t.Fatalf("that land's power = %d, want 7", p)
	}
	if !hasKeywordOnBattlefield(t, g, five, "hexproof") || !hasKeywordOnBattlefield(t, g, five, "indestructible") {
		t.Error("the restored tail did not shield the 5-power creature")
	}
	if !playerHasHexproof(g, g.Seats[g.Turn.ActiveSeat]) {
		t.Error("you gain hexproof once the restored sentence runs")
	}
}

// Awaken's animation waits for its counters (engine/awaken-animate, the land
// and its entry stamp in the tail): the order prompt is a restore point and
// the restored land is a 0/0 Elemental creature with the counters on it.
func TestAwakenIsARestorePointAtTheCountersOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// The land is not a creature while the counters go on (CR 608.2c), so
	// Hardened Scales does not apply: Doubling Season and the probe do.
	probe := seedReplacementPermanent(g, counterPlusOneProbeOracle, "Counter Plus One Probe", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	land := pushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	g.WithWriteLock(func() {
		if err := g.AwakenForEffect(me.ID, uuid.Nil, land, 2); err != nil {
			t.Fatalf("AwakenForEffect: %v", err)
		}
	})

	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if _, _, _, elemental, _ := awakenedLand(t, g, land); elemental {
		t.Fatal("the restored land was animated before its counters landed")
	}
	// 2 → 4 → 5.
	answerOrderBySource(t, g, ds, probe)
	if counters, p, tough, elemental, _ := awakenedLand(t, g, land); counters != 5 || p != 5 || tough != 5 || !elemental {
		t.Errorf("land: %d counters, %d/%d, elemental %v; want a 5/5 Elemental", counters, p, tough, elemental)
	}
}

// Mystic Remora's cumulative upkeep charges for the age counters that are
// there once the new one has landed (cumulative-upkeep/<label>, the asking
// trigger riding the tail): the order prompt over the age counter is a
// restore point and the restored payment is for all three.
func TestCumulativeUpkeepIsARestorePointAtTheAgeCounterOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	probe := seedReplacementPermanent(g, counterPlusOneProbeOracle, "Counter Plus One Probe", owner.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", owner.ID)
	remora := pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, remora) == nil {
		t.Fatal("the cumulative upkeep trigger was not put on the stack")
	}
	passPriorityAroundTable(t, g)

	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if got := ageCountersOn(t, g, remora); got != 0 {
		t.Fatalf("the restored table placed %d age counters before the order", got)
	}
	// 1 → 2 (Doubling Season) → 3 (the probe).
	answerOrderBySource(t, g, ds, probe)
	if got := ageCountersOn(t, g, remora); got != 3 {
		t.Fatalf("%d age counters, want 3", got)
	}
	choice := remoraUpkeepPrompt(t, g, owner.ID)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	hs, ds := counterOrderBoard(g, me.ID)
	army := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie Army", TypeLine: "Token Creature — Zombie Army",
		Power: 0, Toughness: 0, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 1},
		Owner:    me.ID, Controller: me.ID,
	})
	wall := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 6, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Widespread Brutality", "Sorcery", widespreadBrutalityOracl, nil)
	passPriorityAroundTable(t, g)

	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if w := findBattlefieldCardByID(g, wall); w == nil || w.DamageMarked != 0 {
		t.Fatal("the restored table dealt damage with the amass counters still owed")
	}
	// Amass 2 → 4 → 5, onto the one it had: a 6/6 Army.
	answerOrderBySource(t, g, ds, hs)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, army, game.CounterPlusOne); got != 6 {
		t.Fatalf("Army counters = %d, want 6", got)
	}
	if findBattlefieldCardByID(g, wall) != nil {
		t.Error("the 0/6 wall survived a 6-power Army: the restored tail read the pre-amass power")
	}
}

// Explore's +1/+1 counter, then the graveyard question (engine/explore-ask,
// the explorer and the revealed card in its Carry): the counter's order prompt
// is a restore point and the restored answer still opens the question.
func TestExploreIsARestorePointAtTheCounterOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hs, ds := counterOrderBoard(g, me.ID)
	explorer := ctrlPushCreature(g, me.ID, "Explorer")
	top := seedLibraryTop(me, "Some Sorcery", "Sorcery", "{1}")
	epoch := 0
	if c, ok := g.LookupCardForEffect(explorer); ok {
		epoch = c.ObjectEpoch
	}
	g.WithWriteLock(func() {
		if err := g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: explorer, Epoch: epoch}, me.ID); err != nil {
			t.Fatalf("ExploreForEffect: %v", err)
		}
	})

	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	if n := len(randomEvents(g, game.EventExplored)); n != 0 {
		t.Fatalf("explored %d times before the counter landed", n)
	}
	answerOrderBySource(t, g, ds, hs)
	if got := countersOn(g, explorer, game.CounterPlusOne); got != 3 {
		t.Errorf("the explorer has %d counters, want 3", got)
	}
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || len(ask.OptionCarry) != 2 || ask.OptionCarry[0] != top || ask.OptionCarry[1] != explorer {
		t.Fatalf("the restored tail did not ask the graveyard question: %+v", ask)
	}
}
