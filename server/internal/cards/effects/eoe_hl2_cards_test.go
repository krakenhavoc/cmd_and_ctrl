package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_hl2_cards_test.go — the second batch of Edge of Eternities cards from
// H to L, and three E–G leftovers (#4461). Each test plays the real catalog
// entry and asserts what its text says, and what it refuses.

const (
	eoeHL2StationOracle      = "cda39514-86c9-483e-8d95-fd08138ce33a"
	eoeHL2LoadingZoneOracle  = "16156274-8dc0-439c-94c4-e8c89bbb5687"
	eoeHL2ScoutlanderOracle  = "b3f3d55c-a3f6-441b-a380-60c0734dc514"
	eoeHL2MemorialOracle     = "4fa826ca-d361-4391-ad0d-989ebcfa4a91"
	eoeHL2HarrierOracle      = "1972e0db-e595-4ec2-b97c-902561bb275b"
	eoeHL2ContainmentOracle  = "8054092a-1940-4f4b-991d-81f4d384cdb1"
	eoeHL2LithobrakingOracle = "118c4be1-2204-4133-a80e-990ef96c7e35"
	eoeHL2CruiserOracle      = "90c5ea2b-d004-4219-b60c-d51bc1a5ee6b"
	eoeHL2FaithfulOracle     = "df7500e8-76b4-49b3-a427-9a28ffa14637"
	eoeHL2WorldsireOracle    = "e7f53928-a4cc-41ed-877d-b72d786e77d9"
)

func TestEoeHL2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeHL2StationOracle:      "Infinite Guideline Station",
		eoeHL2LoadingZoneOracle:  "Loading Zone",
		eoeHL2ScoutlanderOracle:  "Larval Scoutlander",
		eoeHL2MemorialOracle:     "Kavaron, Memorial World",
		eoeHL2HarrierOracle:      "Kavaron Harrier",
		eoeHL2ContainmentOracle:  "Hardlight Containment",
		eoeHL2LithobrakingOracle: "Lithobraking",
		eoeHL2CruiserOracle:      "Entropic Battlecruiser",
		eoeHL2FaithfulOracle:     "Faller's Faithful",
		eoeHL2WorldsireOracle:    "Famished Worldsire",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if ac := game.AlternativeCostByKey(eoeHL2LoadingZoneOracle, "warp"); ac == nil || ac.ManaCost != "{G}" {
		t.Errorf("Loading Zone: warp %+v, want {G}", ac)
	}
	if spec, _ := Lookup(eoeHL2LithobrakingOracle); !spec.Purpose.DeclaresSweep() {
		t.Error("Lithobraking declares a sweep")
	}
}

// --- helpers ---------------------------------------------------------------

// eoeHL2Ready clears summoning sickness so the permanent can attack.
func eoeHL2Ready(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).SummonedThisTurn = false })
}

// eoeHL2Count counts the battlefield permanents controller controls whose
// type line names the subtype.
func eoeHL2Count(g *game.Game, controller uuid.UUID, subtype string) (n, tapped int) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.HasSubtype(subtype) {
				n++
				if c.Tapped {
					tapped++
				}
			}
		}
	})
	return n, tapped
}

// eoeHL2Hand puts a named card in p's hand.
func eoeHL2Hand(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Infinite Guideline Station --------------------------------------------

func TestEoeHL2InfiniteGuidelineStationCountsYourMulticoloredPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C", "D", "E")
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gold Bear", TypeLine: "Creature — Bear",
		ManaCost: "{G}{W}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gold Charm", TypeLine: "Enchantment",
		ManaCost: "{U}{B}", Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mono Bear", TypeLine: "Creature — Bear",
		ManaCost: "{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Gold", TypeLine: "Enchantment",
		ManaCost: "{R}{W}", Owner: opp.ID, Controller: opp.ID})

	station := avB1Cast(t, g, me, game.Card{Name: "Infinite Guideline Station", TypeLine: "Legendary Artifact — Spacecraft",
		Colors: []string{"W", "U", "B", "R", "G"}, OracleID: eoeHL2StationOracle, Power: 7, Toughness: 15}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n, tapped := eoeHL2Count(g, me.ID, "Robot"); n != 3 || tapped != 3 {
		t.Fatalf("Robots %d (%d tapped), want 3 tapped: the Station and your two gold permanents", n, tapped)
	}
	if layeredCard(t, g, station).IsCreature() {
		t.Fatal("below 12 it is not a creature")
	}
	if err := g.AddCounter(station, game.CounterCharge, 12); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at12 := layeredCard(t, g, station)
	if !at12.IsCreature() || !game.HasKeyword(&at12, "flying") || at12.CurrentPower() != 7 || at12.CurrentToughness() != 15 {
		t.Fatalf("at 12+ it is a 7/15 flyer, got %d/%d", at12.CurrentPower(), at12.CurrentToughness())
	}
	eoeHL2Ready(g, station)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, station)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("attacking drew %d, want 3 (colorless Robots and mono permanents don't count)", got)
	}
}

// --- Loading Zone ------------------------------------------------------------

func TestEoeHL2LoadingZoneDoublesCountersOnCreaturesSpacecraftAndPlanets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Loading Zone", "Enchantment", eoeHL2LoadingZoneOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ship := b12Permanent(g, me.ID, "My Ship", "Artifact — Spacecraft")
	planet := b12Permanent(g, me.ID, "My Planet", "Land — Planet")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, tc := range []struct {
		id        uuid.UUID
		kind      string
		put, want int
		permanent string
	}{
		{bear, game.CounterPlusOne, 1, 2, "your creature"},
		{ship, game.CounterCharge, 3, 6, "your Spacecraft"},
		{planet, game.CounterCharge, 2, 4, "your Planet"},
		{rock, game.CounterCharge, 1, 1, "your other artifact"},
		{theirs, game.CounterPlusOne, 1, 1, "an opponent's creature"},
	} {
		if err := g.AddCounter(tc.id, tc.kind, tc.put); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
		if got := countersOn(g, tc.id, tc.kind); got != tc.want {
			t.Errorf("%d counters put on %s: %d there, want %d", tc.put, tc.permanent, got, tc.want)
		}
	}
	if err := g.AddCounter(bear, game.CounterPlusOne, -1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("removing one counter left %d, want 1: a removal is not doubled", got)
	}
}

// --- Larval Scoutlander ------------------------------------------------------

func TestEoeHL2LarvalScoutlanderTradesALandForTwoBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	swamp := pushLandFor(g, me.ID, "Wastes Land", "Land")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Temple", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Larval Scoutlander", "Artifact — Spacecraft", eoeHL2ScoutlanderOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if hasID(c.ChooseCards, rock) {
		t.Error("an artifact that isn't a Lander is offered")
	}
	answerOwnPermanents(t, g, me.ID, swamp)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(swamp) {
		t.Fatal("the land is sacrificed")
	}
	if sc := searchChoiceFor(g, me.ID); sc != nil && searchOptionNamed(g, sc, "Temple") != uuid.Nil {
		t.Error("a nonbasic land is offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest", "Island")
	passPriorityAroundTable(t, g)
	for _, name := range []string{"Forest", "Island"} {
		id := avB1Named(g, me.ID, name)
		if id == uuid.Nil || !avB1Card(t, g, id).Tapped {
			t.Errorf("%s is not on the battlefield tapped", name)
		}
	}
}

func TestEoeHL2LarvalScoutlanderSacrificesALanderAndMayDecline(t *testing.T) {
	t.Run("lander", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		lander := b12Permanent(g, me.ID, "Lander", "Token Artifact — Lander")
		castCatalogSpell(t, g, "Larval Scoutlander", "Artifact — Spacecraft", eoeHL2ScoutlanderOracle, nil)
		passPriorityAroundTable(t, g)
		answerOwnPermanents(t, g, me.ID, lander)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(lander) || searchChoiceFor(g, me.ID) == nil {
			t.Error("sacrificing a Lander searches")
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
		castCatalogSpell(t, g, "Larval Scoutlander", "Artifact — Spacecraft", eoeHL2ScoutlanderOracle, nil)
		passPriorityAroundTable(t, g)
		answerOwnPermanents(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(land) || searchChoiceFor(g, me.ID) != nil {
			t.Error("declining sacrificed the land or searched")
		}
	})
}

// --- Kavaron, Memorial World -------------------------------------------------

func TestEoeHL2KavaronMemorialWorldMakesARobotAndPumpsTheTeamAtTwelve(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	world := b12Push(g, me.ID, "Kavaron, Memorial World", "Land — Planet", eoeHL2MemorialOracle, 0, 0)
	fodder := pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	activate := func() error {
		return g.ActivateCatalogAbility(me.ID, world, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}})
	}
	fillPoolColored(me, "R", 2)
	if err := activate(); err == nil {
		t.Fatal("the 12+ ability was activated with no charge counters")
	}
	if err := g.AddCounter(world, game.CounterCharge, 12); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if err := activate(); err != nil {
		t.Fatalf("activate at 12: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the land is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	robots := battlefieldIDsNamed(g, "Robot")
	if len(robots) != 1 {
		t.Fatalf("%d Robots, want 1", len(robots))
	}
	for _, id := range []uuid.UUID{bear, robots[0]} {
		if !hasKeywordOnBattlefield(t, g, id, "haste") {
			t.Errorf("%s has no haste", cardByID(g, id).Name)
		}
	}
	if p, tough := eoeCDPT(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 3/2", p, tough)
	}
	if p, _ := eoeCDPT(t, g, robots[0]); p != 3 {
		t.Errorf("the Robot made first is pumped too: power %d, want 3", p)
	}
	if p, _ := eoeCDPT(t, g, theirs); p != 2 || hasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("an opponent's creature is pumped")
	}
}

// --- Kavaron Harrier ---------------------------------------------------------

func TestEoeHL2KavaronHarrierPaysTwoForAnAttackingRobotForOneCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harrier := b12Push(g, me.ID, "Kavaron Harrier", "Artifact Creature — Robot Soldier", eoeHL2HarrierOracle, 2, 1)
	life := opp.Life
	declareAttack(t, g, opp.ID, harrier)
	passPriorityAroundTable(t, g)
	fillPoolColored(me, "C", 2)
	answerPayUnless(t, g, me.ID, true)
	passAndSettleAttackSplit(t, g, me.ID)
	var robot uuid.UUID
	for _, id := range battlefieldIDsNamed(g, "Robot") {
		if id != harrier {
			robot = id
		}
	}
	if robot == uuid.Nil {
		t.Fatal("no Robot token")
	}
	if tok := cardByID(g, robot); !tok.Tapped || tok.AttackingTarget != opp.ID {
		t.Errorf("the Robot is not tapped and attacking: %+v", tok)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("defender at %d, want %d (the Harrier and the Robot)", opp.Life, life-4)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(robot) {
		t.Error("the Robot is sacrificed at end of combat")
	}
	if !g.Battlefield.Contains(harrier) {
		t.Error("the Harrier itself is not sacrificed")
	}
}

func TestEoeHL2KavaronHarrierDeclinedMakesNoRobot(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harrier := b12Push(g, me.ID, "Kavaron Harrier", "Artifact Creature — Robot Soldier", eoeHL2HarrierOracle, 2, 1)
	declareAttack(t, g, opp.ID, harrier)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Robot")); n != 0 {
		t.Errorf("%d Robot tokens after declining, want 0", n)
	}
}

// --- Hardlight Containment ---------------------------------------------------

func TestEoeHL2HardlightContainmentExilesUntilItLeavesAndWardsItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	construct := b12Creature(g, me.ID, "My Construct", "Artifact Creature — Construct", 2, 2)
	theirRock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	if err := castCatalogSpellErr(t, g, "Hardlight Containment", "Enchantment — Aura", eoeHL2ContainmentOracle, cardRefs(theirRock)); err == nil {
		t.Fatal("an artifact you don't control was accepted")
	}
	aura := castCatalogSpell(t, g, "Hardlight Containment", "Enchantment — Aura", eoeHL2ContainmentOracle, cardRefs(construct))
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if pick := latestPickTarget(g, me.ID); pick != nil && hasID(pick.PickTargetCards, mine) {
		t.Error("a creature you control is offered as a target")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the opposing creature was not exiled")
	}

	castAtWardedCreature(t, g, opp, construct)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's spell at the enchanted artifact")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(construct) {
		t.Fatal("a declined ward {1} counters the spell")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if got := eoeCDCount(g, opp.ID, "Their Threat"); got != 1 {
		t.Errorf("after the Aura left, %d copies of the creature are back under its owner, want 1", got)
	}
}

// --- Lithobraking ------------------------------------------------------------

func TestEoeHL2LithobrakingSacrificesTheLanderToSweepForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	castCatalogSpell(t, g, "Lithobraking", "Instant", eoeHL2LithobrakingOracle, nil)
	passPriorityAroundTable(t, g)
	landers := battlefieldIDsNamed(g, "Lander")
	if len(landers) != 1 {
		t.Fatalf("%d Landers, want 1", len(landers))
	}
	answerOwnPermanents(t, g, me.ID, landers[0])
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(landers[0]) {
		t.Error("the Lander is sacrificed")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("2 damage kills each 2/2")
	}
	if !g.Battlefield.Contains(big) || cardByID(g, big).DamageMarked != 2 {
		t.Error("a 3/3 survives with 2 damage")
	}
}

func TestEoeHL2LithobrakingDeclinedDealsNoDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Lithobraking", "Instant", eoeHL2LithobrakingOracle, nil)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Lander")) != 1 {
		t.Error("the Lander stays when you decline")
	}
	if !g.Battlefield.Contains(bear) || cardByID(g, bear).DamageMarked != 0 {
		t.Error("declining dealt damage")
	}
}

// --- Entropic Battlecruiser --------------------------------------------------

func TestEoeHL2EntropicBattlecruiserPunishesDiscardsFromItsFirstCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := b12Push(g, me.ID, "Entropic Battlecruiser", "Artifact — Spacecraft", eoeHL2CruiserOracle, 3, 10)
	pitch := func(p *game.Player) {
		id := uuid.New()
		p.Hand.Cards = p.Hand.Cards[:0]
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Pitched", TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
		g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(p.ID, 1) })
		passPriorityAroundTable(t, g)
	}
	life := opp.Life
	pitch(opp)
	if opp.Life != life {
		t.Fatalf("with no charge counter a discard cost %d life", life-opp.Life)
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	pitch(opp)
	if opp.Life != life-3 {
		t.Errorf("at 1+ an opponent's discard: life %d, want %d", opp.Life, life-3)
	}
	mine := me.Life
	pitch(me)
	if me.Life != mine {
		t.Errorf("your own discard cost you %d life", mine-me.Life)
	}
}

func TestEoeHL2EntropicBattlecruiserAttackMakesEachOpponentDiscardOrLoseThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	ship := b12Push(g, me.ID, "Entropic Battlecruiser", "Artifact — Spacecraft", eoeHL2CruiserOracle, 3, 10)
	if err := g.AddCounter(ship, game.CounterCharge, 8); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at8 := layeredCard(t, g, ship)
	if !at8.IsCreature() || !game.HasKeyword(&at8, "flying") || !game.HasKeyword(&at8, "deathtouch") {
		t.Fatal("at 8+ it is a creature with flying and deathtouch")
	}
	for _, p := range g.Seats[1:] {
		p.Hand.Cards = p.Hand.Cards[:0]
	}
	card := eoeHL2Hand(opp, "Their Card", "Creature — Bear")
	myLife, oppLife, thirdLife := me.Life, opp.Life, third.Life
	declareAttack(t, g, opp.ID, ship)
	passPriorityAroundTable(t, g)
	if third.Life != thirdLife-3 {
		t.Errorf("an opponent with an empty hand: life %d, want %d", third.Life, thirdLife-3)
	}
	answerDiscard(t, g, opp.ID, card)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(card) {
		t.Fatal("the opponent with a card discarded it")
	}
	if opp.Life != oppLife-3 {
		t.Errorf("a discard it forces triggers the 1+ line: life %d, want %d", opp.Life, oppLife-3)
	}
	if me.Life != myLife {
		t.Errorf("you lost %d life", myLife-me.Life)
	}
}

// --- Faller's Faithful -------------------------------------------------------

func TestEoeHL2FallersFaithfulDrawsOnlyForAnUndamagedCreature(t *testing.T) {
	for _, damaged := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		seedLibrary(opp, "A", "B", "C")
		victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
		if damaged {
			src := b12Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)
			g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, victim, 1) })
		}
		hand := opp.Hand.Size()
		faithful := castCatalogSpell(t, g, "Faller's Faithful", "Creature — Human Wizard", eoeHL2FaithfulOracle, nil)
		passPriorityAroundTable(t, g)
		b04WaitForPick(t, g, me.ID)
		eoeHL2RefuseSelf(t, g, me.ID, faithful)
		pickCard(t, g, me.ID, victim)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) {
			t.Errorf("damaged=%v: the creature is destroyed", damaged)
		}
		want := 2
		if damaged {
			want = 0
		}
		if got := opp.Hand.Size() - hand; got != want {
			t.Errorf("damaged=%v: its controller drew %d, want %d", damaged, got, want)
		}
	}
}

func TestEoeHL2FallersFaithfulDrawsEvenIfTheCreatureSurvives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "A", "B", "C")
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"indestructible"}})
	hand := opp.Hand.Size()
	castCatalogSpell(t, g, "Faller's Faithful", "Creature — Human Wizard", eoeHL2FaithfulOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("an indestructible creature survives")
	}
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("its controller drew %d, want 2: the draw doesn't depend on the destruction", got)
	}
}

// eoeHL2RefuseSelf asserts that the chooser's open target prompt does not
// offer id.
func eoeHL2RefuseSelf(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	if pick := latestPickTarget(g, chooser); pick != nil && hasID(pick.PickTargetCards, id) {
		t.Error("the Faithful offers itself as the other target")
	}
}

// --- Famished Worldsire ------------------------------------------------------

func TestEoeHL2FamishedWorldsireDevoursLandsAndPutsLandsFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	l1 := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	l2 := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		plTop(me, "Deep Card", "Instant", "{U}")
	}
	spell := plTop(me, "Top Spell", "Sorcery", "{G}")
	landA := plTop(me, "Top Mountain", "Basic Land — Mountain", "")
	landB := plTop(me, "Top Plains", "Basic Land — Plains", "")
	for i := 0; i < 3; i++ {
		plTop(me, "Top Filler", "Creature — Bear", "{1}{G}")
	}

	worldsire := castCatalogSpell(t, g, "Famished Worldsire", "Creature — Leviathan", eoeHL2WorldsireOracle, nil)
	passPriorityAroundTable(t, g)
	c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("no devour prompt")
	}
	if hasID(c.ChooseCards, bear) {
		t.Error("a creature is offered to Devour land")
	}
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, l1, l2)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, worldsire, game.CounterPlusOne); got != 6 {
		t.Fatalf("devouring two lands: %d counters, want 6", got)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land pick from the top six")
	}
	if hasID(pick.ChooseCards, spell) || !hasID(pick.ChooseCards, landA) || !hasID(pick.ChooseCards, landB) || pick.ChooseMin != 0 {
		t.Errorf("offered %v (min %d), want only the two lands with a floor of zero", pick.ChooseCards, pick.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, landA, landB)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{landA, landB} {
		if !g.Battlefield.Contains(id) || !avB1Card(t, g, id).Tapped {
			t.Errorf("%s is not on the battlefield tapped", cardByID(g, id).Name)
		}
	}
	if !me.Library.Contains(spell) {
		t.Error("the nonland card stays in the library")
	}
}
