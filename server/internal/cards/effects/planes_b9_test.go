package effects

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b9_test.go — Planechase plane cards, batch 9 (Planechase Anthology
// N–S, ADR 0216 slice 4, #4870). Each test plays the real catalog entry
// face up on a Planechase table (planes_harness_test.go).

var (
	planeNephalia = planeFixture{"72af7529-75bb-488f-801d-7148631b9e9b", "Nephalia", "Plane — Innistrad"}
	planeTazeem   = planeFixture{"97f708de-9271-4137-a9fe-dda25b922900", "Tazeem", "Plane — Zendikar"}
	planeSanctum  = planeFixture{"f0c8ce35-b627-485c-bd64-d7eae47b8efa", "Sanctum of Serra", "Plane — Serra's Realm"}
	planeNorns    = planeFixture{"9d58f647-3f83-4f65-a944-c6236db7c422", "Norn's Dominion", "Plane — New Phyrexia"}
	planeOnakke   = planeFixture{"69f06edb-dade-4e94-b563-e36a05e567f0", "Onakke Catacomb", "Plane — Shandalar"}
	planeRavens   = planeFixture{"b53e06af-be18-4f75-9dd0-7494ec867b13", "Raven's Run", "Plane — Shadowmoor"}
	planeShiv     = planeFixture{"70edae19-e278-401d-9772-508bd66320a6", "Shiv", "Plane — Dominaria"}
	planeVelisVel = planeFixture{"1c4e174e-19fb-4c0a-b762-ef6d927392c9", "Velis Vel", "Plane — Lorwyn"}
	planeZephyr   = planeFixture{"2d84be6d-49a5-4200-8cd3-d8653dfd0899", "The Zephyr Maze", "Plane — Kyneth"}
	planeHippo    = planeFixture{"cde8d8c4-5e2b-452f-b3e9-bb9f2e745ec9", "The Hippodrome", "Plane — Segovia"}
	planeTakenuma = planeFixture{"de53c948-b371-4f79-a965-33c44d0b44de", "Takenuma", "Plane — Kamigawa"}
	planeTember   = planeFixture{"bbb1d13c-a3ce-4151-883a-cce9bbd584d0", "Tember City", "Plane — Kinshala"}
	planeBarony   = planeFixture{"5503c308-560d-4d5f-a86a-4b06d9947ddc", "The Dark Barony", "Plane — Ulgrotha"}
	planeEonFog   = planeFixture{"a04d05ed-1b9b-4407-8780-c75c102a94dd", "The Eon Fog", "Plane — Equilor"}
	planeFourth   = planeFixture{"13e801cd-e537-4ea6-8c3d-21b142bdc026", "The Fourth Sphere", "Plane — Phyrexia"}
	planeOtaria   = planeFixture{"29e3371d-aa41-46a9-921a-96565be47eff", "Otaria", "Plane — Dominaria"}
	planeTrail    = planeFixture{"b2c95c7a-a04c-46ef-a879-6eb08e444093", "Trail of the Mage-Rings", "Plane — Vryn"}
	planeStairs   = planeFixture{"7a425df4-7010-462e-9cb5-20389bac721c", "Stairs to Infinity", "Plane — Xerex"}
	planePrahv    = planeFixture{"34a31865-b701-4ce1-8912-29fc41ecdda3", "Prahv", "Plane — Ravnica"}
	planeOrzhova  = planeFixture{"c7241e4b-d2c7-4801-96a3-91bb998953a1", "Orzhova", "Plane — Ravnica"}
)

// planeLand puts a basic land on the battlefield under `owner`.
func planeLand(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: owner, Controller: owner,
	})
}

// --- Nephalia ----------------------------------------------------------------

func TestNephaliaMillsSevenThenReturnsOneAtRandomAtYourEnd(t *testing.T) {
	requirePlaneRegisteredFull(t, planeNephalia)
	g, plane := newPlaneGame(t, planeNephalia)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lib, hand, yard := me.Library.Size(), me.Hand.Size(), me.Graveyard.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib-7 {
		t.Errorf("library %d, want %d after milling seven", got, lib-7)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: one milled card returned", got, hand+1)
	}
	if got := me.Graveyard.Size(); got != yard+6 {
		t.Errorf("graveyard %d, want %d", got, yard+6)
	}
	_ = plane
}

func TestNephaliaChaosReturnsTargetCard(t *testing.T) {
	g, plane := newPlaneGame(t, planeNephalia)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatal("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the targeted card did not return to hand")
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	rollChaos(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Error("a face-down plane's chaos ability triggered")
	}
}

// --- Tazeem ------------------------------------------------------------------

func TestTazeemStopsBlocksAndChaosDrawsPerLand(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTazeem)
	g, _ := newPlaneGame(t, planeTazeem)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	for range 3 {
		planeLand(g, me.ID)
	}
	planeLand(g, opp.ID)
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand %d, want %d: a card per land you control", got, before+3)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, raider); err == nil {
		t.Error("a creature blocked on Tazeem")
	}
}

// planeEffective is the battlefield card's characteristics after the layer
// pass.
func planeEffective(t *testing.T, g *game.Game, id uuid.UUID) game.Characteristic {
	t.Helper()
	var out game.Characteristic
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.Effective(), true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return out
}

// onBattlefield reports whether `id` is still on the battlefield.
func planeOnBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			found = found || c.InstanceID == id
		}
	})
	return found
}

// planeHasKeyword reports whether `abilities` lists `kw`.
func planeHasKeyword(abilities []string, kw string) bool {
	return slices.ContainsFunc(abilities, func(a string) bool { return strings.EqualFold(a, kw) })
}

// --- Sanctum of Serra --------------------------------------------------------

func TestSanctumOfSerraDestroysNonlandsWhenYouPlaneswalkAway(t *testing.T) {
	requirePlaneRegisteredFull(t, planeSanctum)
	g, _ := newPlaneGame(t, planeSanctum)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opponentOf(g).ID, "Rock", "Artifact")
	land := planeLand(g, me.ID)
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	if planeOnBattlefield(g, bear) || planeOnBattlefield(g, rock) {
		t.Error("a nonland permanent survived the planeswalk away")
	}
	if !planeOnBattlefield(g, land) {
		t.Error("a land was destroyed")
	}
}

func TestSanctumOfSerraChaosSetsLifeTo20(t *testing.T) {
	g, _ := newPlaneGame(t, planeSanctum)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Life = 14
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if me.Life != 20 {
		t.Errorf("life %d, want 20", me.Life)
	}
	me.Life = 31
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if me.Life != 31 {
		t.Errorf("life %d after declining, want 31", me.Life)
	}
}

// --- Norn's Dominion ---------------------------------------------------------

func TestNornsDominionSparesFatedPermanentsThenClearsTheCounters(t *testing.T) {
	requirePlaneRegisteredFull(t, planeNorns)
	g, _ := newPlaneGame(t, planeNorns)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	doomed := b12Creature(g, me.ID, "Doomed", "Creature — Bear", 2, 2)
	fated := b12Creature(g, opponentOf(g).ID, "Fated", "Creature — Bear", 2, 2)
	land := planeLand(g, me.ID)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(fated, "fate", 1)
		_ = g.AddCounterForEffect(land, "fate", 2)
	})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	pc := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(pc.ID, pc.Chooser, cardRef(doomed)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if got := planeEffective(t, g, doomed); got.Power != 2 {
		t.Fatalf("unexpected power %d", got.Power)
	}
	if c, _ := g.LookupCardForEffect(doomed); c.Counters["fate"] != 1 {
		t.Errorf("chaos left %d fate counters on the target, want 1", c.Counters["fate"])
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	if !planeOnBattlefield(g, doomed) || !planeOnBattlefield(g, fated) {
		t.Error("a permanent with a fate counter was destroyed")
	}
	for _, id := range []uuid.UUID{doomed, fated, land} {
		if c, _ := g.LookupCardForEffect(id); c.Counters["fate"] != 0 {
			t.Errorf("a fate counter survived the planeswalk away (%d)", c.Counters["fate"])
		}
	}
	bare := b12Creature(g, me.ID, "Bare", "Creature — Bear", 1, 1)
	_ = bare
}

// --- Onakke Catacomb ---------------------------------------------------------

func TestOnakkeCatacombMakesCreaturesBlackDeathtouchersAndChaosPumps(t *testing.T) {
	requirePlaneRegisteredFull(t, planeOnakke)
	g, _ := newPlaneGame(t, planeOnakke)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	for _, id := range []uuid.UUID{mine, theirs} {
		c := planeEffective(t, g, id)
		if len(c.Colors) != 1 || !strings.EqualFold(c.Colors[0], "B") && c.Colors[0] != "Black" {
			t.Errorf("colors %v, want only black", c.Colors)
		}
		if !planeHasKeyword(c.Abilities, "deathtouch") {
			t.Errorf("no deathtouch: %v", c.Abilities)
		}
	}
	if c := planeEffective(t, g, rock); len(c.Colors) != 0 {
		t.Errorf("a noncreature permanent became %v", c.Colors)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("your creature has power %d, want 3", got)
	}
	if got := effectivePower(t, g, theirs); got != 3 {
		t.Errorf("their creature has power %d, want 3", got)
	}
	if c := planeEffective(t, g, mine); !planeHasKeyword(c.Abilities, "first strike") {
		t.Error("your creature has no first strike")
	}
	late := b12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, late); got != 2 {
		t.Errorf("a creature gained later has power %d, want 2", got)
	}
}

// --- Raven's Run -------------------------------------------------------------

func TestRavensRunChaosPutsOneTwoAndThreeCounters(t *testing.T) {
	requirePlaneRegisteredFull(t, planeRavens)
	g, _ := newPlaneGame(t, planeRavens)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 9, 9)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 9, 9)
	c := b12Creature(g, opponentOf(g).ID, "C", "Creature — Bear", 9, 9)
	if cc := planeEffective(t, g, a); !planeHasKeyword(cc.Abilities, "wither") {
		t.Error("a creature lacks wither")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for slot, id := range []uuid.UUID{a, b, c} {
		ch := pickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot}); err != nil {
			t.Fatalf("ResolvePickTarget %d: %v", slot, err)
		}
	}
	passPriorityAroundTable(t, g)
	for i, id := range []uuid.UUID{a, b, c} {
		if got, _ := g.LookupCardForEffect(id); got.Counters[game.CounterMinusOne] != i+1 {
			t.Errorf("creature %d has %d -1/-1 counters, want %d", i, got.Counters[game.CounterMinusOne], i+1)
		}
	}
}

// --- Shiv --------------------------------------------------------------------

func TestShivGivesEveryCreatureARedPumpAndChaosMakesADragon(t *testing.T) {
	requirePlaneRegisteredFull(t, planeShiv)
	g, plane := newPlaneGame(t, planeShiv)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.BumpLayerVersionForTest()
	floatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the granted pump cannot be activated: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := mkmMPPT(t, g, bear); p != 3 {
		t.Errorf("power %d after the pump, want 3", p)
	}
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatal("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Dragon"); n != 1 {
		t.Errorf("%d Dragons, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	g.BumpLayerVersionForTest()
	floatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the pump survived planeswalking away from Shiv")
	}
}

// --- Velis Vel ---------------------------------------------------------------

func TestVelisVelCountsSharedTypesAndChaosGivesAllTypes(t *testing.T) {
	requirePlaneRegisteredFull(t, planeVelisVel)
	g, _ := newPlaneGame(t, planeVelisVel)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	s1 := b12Creature(g, me.ID, "Shaman One", "Creature — Elemental Shaman", 2, 2)
	s2 := b12Creature(g, opp.ID, "Shaman Two", "Creature — Elemental Shaman", 2, 2)
	sp := b12Creature(g, me.ID, "Spirit", "Creature — Elemental Spirit", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for id, want := range map[uuid.UUID]int{s1: 4, s2: 4, sp: 4, bear: 2} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	// The Bear now shares a type with every other creature: +3/+3.
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("the Bear with all creature types has power %d, want 5", got)
	}
}

// --- The Zephyr Maze ---------------------------------------------------------

func TestZephyrMazeFlyersGrowGroundedShrinkAndChaosGrantsFlying(t *testing.T) {
	requirePlaneRegisteredFull(t, planeZephyr)
	g, _ := newPlaneGame(t, planeZephyr)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	flyer := pushVanillaCreature(g, me.ID, "Flyer", 2, 2)
	walker := pushVanillaCreature(g, me.ID, "Walker", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flyer {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	g.BumpLayerVersionForTest()
	if got := effectivePower(t, g, flyer); got != 4 {
		t.Errorf("flyer power %d, want 4", got)
	}
	if got := effectivePower(t, g, walker); got != 1 {
		t.Errorf("grounded creature power %d, want 1", got)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, walker)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, walker); got != 5 {
		t.Errorf("the creature given flying has power %d, want 5", got)
	}
}

// --- The Hippodrome ----------------------------------------------------------

func TestHippodromeShrinksPowerAndChaosMayDestroyAPowerlessCreature(t *testing.T) {
	requirePlaneRegisteredFull(t, planeHippo)
	g, _ := newPlaneGame(t, planeHippo)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	small := pushVanillaCreature(g, me.ID, "Small", 3, 3)
	big := pushVanillaCreature(g, me.ID, "Big", 7, 7)
	g.BumpLayerVersionForTest()
	if got := effectivePower(t, g, small); got != -2 {
		t.Errorf("power %d, want -2", got)
	}
	// A creature whose power is above 0 is a legal target but is not destroyed.
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, big)
	passPriorityAroundTable(t, g)
	if !planeOnBattlefield(g, big) {
		t.Fatal("a creature with positive power was destroyed")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, small)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if !planeOnBattlefield(g, small) {
		t.Fatal("declining still destroyed the creature")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, small)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if planeOnBattlefield(g, small) {
		t.Error("the powerless creature was not destroyed")
	}
}

// --- Takenuma ----------------------------------------------------------------

func TestTakenumaDrawsForTheCreaturesControllerAndChaosBounces(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTakenuma)
	g, _ := newPlaneGame(t, planeTakenuma)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	oppHand, myHand := opp.Hand.Size(), me.Hand.Size()
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{theirs, rock}) })
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != oppHand+1 {
		t.Errorf("the creature's controller drew %d, want 1", got-oppHand)
	}
	if got := me.Hand.Size(); got != myHand {
		t.Errorf("the planar controller drew %d for another player's creature", got-myHand)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("the creature did not return to its owner's hand")
	}
	if got := me.Hand.Size(); got != myHand+2 {
		t.Errorf("hand %d, want %d: the bounced creature and the card its leaving drew", got, myHand+2)
	}
}

// --- Tember City -------------------------------------------------------------

func TestTemberCityPingsLandTappersAndChaosIsAnEdict(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTember)
	g, _ := newPlaneGame(t, planeTember)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine1 := b12Permanent(g, me.ID, "Mine", "Artifact")
	_ = mine1
	theirs := b12Permanent(g, opp.ID, "Theirs", "Artifact")
	theirLand := planeLand(g, opp.ID)
	myLand := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	lives := [2]int{me.Life, opp.Life}
	if err := g.ActivateManaAbility(me.ID, myLand, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != lives[0]-1 || opp.Life != lives[1] {
		t.Errorf("lives %d/%d after tapping a land, want %d/%d", me.Life, opp.Life, lives[0]-1, lives[1])
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.ID != me.ID {
			answerSacrifice(t, g, p.ID, theirs)
		}
	}
	passPriorityAroundTable(t, g)
	if planeOnBattlefield(g, theirs) {
		t.Error("the opponent kept their nonland permanent")
	}
	if !planeOnBattlefield(g, theirLand) || !planeOnBattlefield(g, mine1) {
		t.Error("a land or the planar controller's permanent was sacrificed")
	}
}

// --- The Dark Barony ---------------------------------------------------------

func TestDarkBaronyCostsLifeForNonblackCardsInGraveyardsAndChaosDiscards(t *testing.T) {
	requirePlaneRegisteredFull(t, planeBarony)
	g, _ := newPlaneGame(t, planeBarony)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	g.WithWriteLock(func() {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dark One", TypeLine: "Creature — Bat", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	})
	before := opp.Life
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("a black card cost %d life", before-opp.Life)
	}
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d after two nonblack cards, want %d", opp.Life, before-2)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if discardOwed(g, p.ID) != 1 {
			t.Fatalf("%s was not asked to discard a card", p.Name)
		}
		answerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
	}
}

// --- The Eon Fog -------------------------------------------------------------

func TestEonFogSkipsUntapAndChaosUntapsYourPermanents(t *testing.T) {
	requirePlaneRegisteredFull(t, planeEonFog)
	g, _ := newPlaneGame(t, planeEonFog)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	next := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, next.ID, "Rock", "Artifact")
	mine := b12Permanent(g, me.ID, "Mine", "Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].Tapped = true
		}
	})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(mine); c.Tapped {
		t.Error("chaos left the planar controller's permanent tapped")
	}
	if c, _ := g.LookupCardForEffect(rock); !c.Tapped {
		t.Error("chaos untapped another player's permanent")
	}
	if active := toNextUpkeep(t, g); active.ID != next.ID {
		t.Fatalf("active player %s, want %s", active.Name, next.Name)
	}
	if c, _ := g.LookupCardForEffect(rock); !c.Tapped {
		t.Error("a permanent untapped in an untap step on The Eon Fog")
	}
}

// --- The Fourth Sphere -------------------------------------------------------

func TestFourthSphereSacrificesANonblackCreatureAtUpkeepAndChaosMakesAZombie(t *testing.T) {
	requirePlaneRegisteredFull(t, planeFourth)
	g, _ := newPlaneGame(t, planeFourth)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	next := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	black := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black", TypeLine: "Creature — Bat",
		Colors: []string{"B"}, Power: 1, Toughness: 1, Owner: next.ID, Controller: next.ID})
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: next.ID, Controller: next.ID})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Zombie"); n != 1 {
		t.Fatalf("%d Zombies, want 1", n)
	}
	toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, next.ID, green)
	passPriorityAroundTable(t, g)
	if planeOnBattlefield(g, green) {
		t.Error("the nonblack creature was not sacrificed")
	}
	if !planeOnBattlefield(g, black) {
		t.Error("the black creature was sacrificed")
	}
}

// --- Orzhova -----------------------------------------------------------------

func TestOrzhovaChaosExilesOpponentCreatureCardsAndPlaneswalkingReturnsAll(t *testing.T) {
	requirePlaneRegisteredFull(t, planeOrzhova)
	g, _ := newPlaneGame(t, planeOrzhova)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mineDead := pushGraveyardCardWithTypeLineID(me, "My Bear", "Creature — Bear")
	theirDead := pushGraveyardCardWithTypeLineID(opp, "Their Bear", "Creature — Bear")
	theirSpell := pushGraveyardCardWithTypeLineID(opp, "Their Ritual", "Sorcery")
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(theirSpell)); err == nil {
		t.Fatal("a noncreature card was accepted")
	}
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(mineDead)); err == nil {
		t.Fatal("the planar controller's own graveyard was accepted")
	}
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(theirDead)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	for pickTargetOpen(g) {
		ch := pickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{}); err != nil {
			break
		}
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirDead) {
		t.Error("the opponent's creature card was not exiled")
	}
	other := pushGraveyardCardWithTypeLineID(opp, "Other Bear", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	if !planeOnBattlefield(g, mineDead) || !planeOnBattlefield(g, other) {
		t.Error("creature cards did not return from every graveyard")
	}
	if !opp.Graveyard.Contains(theirSpell) {
		t.Error("a noncreature card left the graveyard")
	}
}

func pickTargetOpen(g *game.Game) bool {
	return latestChoiceOfKind(g, game.PendingChoicePickTarget) != nil
}

// --- Otaria ------------------------------------------------------------------

func TestOtariaGivesEveryGraveyardsInstantsAndSorceriesFlashbackAndChaosTakesExtraTurns(t *testing.T) {
	requirePlaneRegisteredFull(t, planeOtaria)
	g, _ := newPlaneGame(t, planeOtaria)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	spell := pushGraveyardCardWithTypeLineID(me, "Dead Sorcery", "Sorcery")
	theirs := pushGraveyardCardWithTypeLineID(opp, "Dead Instant", "Instant")
	bear := pushGraveyardCardWithTypeLineID(me, "Dead Bear", "Creature — Bear")
	for _, c := range []struct {
		owner *game.Player
		id    uuid.UUID
		want  bool
	}{{me, spell, true}, {opp, theirs, true}, {me, bear, false}} {
		perm := grantedPermissionOn(g, c.owner.ID, c.id, game.ZoneGraveyard)
		if perm.Granted() != c.want {
			t.Errorf("flashback granted = %v on %v, want %v", perm.Granted(), c.id, c.want)
		}
		if c.want && (perm.AltCostKey != "flashback" || perm.Cost != "" || !perm.ExileOnResolution) {
			t.Errorf("the grant is %+v, want flashback for the mana cost, exiled after", perm)
		}
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback the sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Error("a sorcery cast with flashback is exiled as it leaves the stack")
	}
	spell = pushGraveyardCardWithTypeLineID(me, "Another Sorcery", "Sorcery")
	rollChaos(t, g, me.ID)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := len(g.ExtraTurnsQueued()); n != 2 {
		t.Errorf("%d extra turns queued by two chaos rolls, want 2", n)
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if perm := grantedPermissionOn(g, me.ID, spell, game.ZoneGraveyard); perm.Granted() {
		t.Error("flashback outlived the plane")
	}
}

// --- Trail of the Mage-Rings -------------------------------------------------

func TestTrailOfTheMageRingsGivesEveryPlayersInstantsAndSorceriesRebound(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTrail)
	g, _ := newPlaneGame(t, planeTrail)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	advanceToMain(t, g)
	instant := castPlainSpell(t, g, me, "Instant")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(instant) || reboundTriggersFor(g, instant) != 1 {
		t.Fatal("an instant cast from hand on Trail of the Mage-Rings did not rebound")
	}
	creature := castPlainSpell(t, g, me, "Creature — Bear")
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(creature) {
		t.Error("a creature spell rebounded")
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = (seat + 1) % len(g.Seats) })
	theirs := castPlainSpell(t, g, opp, "Instant")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || reboundTriggersFor(g, theirs) != 1 {
		t.Error("another player's instant did not rebound")
	}
}

func TestTrailOfTheMageRingsChaosMayTutorAnInstantOrSorcery(t *testing.T) {
	g, _ := newPlaneGame(t, planeTrail)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: bolt, Name: "Test Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not reach the hand")
	}
}

// --- Stairs to Infinity ------------------------------------------------------

func TestStairsToInfinityDrawsOnEveryPlanarRollAndHasNoHandLimit(t *testing.T) {
	requirePlaneRegisteredFull(t, planeStairs)
	g, _ := newPlaneGame(t, planeStairs)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	before := me.Hand.Size()
	rollBlank(t, g, me.ID)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a blank roll, want %d", got, before+1)
	}
	rollBlank(t, g, opp.ID)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("another player's roll drew a card for the planar controller")
	}
	for range 10 {
		pushHandCard(g, me)
	}
	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepCleanup)
	if me.Hand.Size() < 8 || discardOwed(g, me.ID) != 0 {
		t.Errorf("hand %d, discard owed %d: a hand limit applied", me.Hand.Size(), discardOwed(g, me.ID))
	}
}

func TestStairsToInfinityChaosRevealsTheTopPlaneAndMayBottomIt(t *testing.T) {
	g, _ := newPlaneGame(t, planeStairs)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	top, _ := g.PlanarDeck.Top()
	rollChaos(t, g, me.ID)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if now, _ := g.PlanarDeck.Top(); now.InstanceID != top.InstanceID {
		t.Error("declining moved the planar deck's top card")
	}
	rollChaos(t, g, me.ID)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if now, _ := g.PlanarDeck.Top(); now.InstanceID == top.InstanceID {
		t.Error("the revealed plane is still on top of the planar deck")
	}
	if bottom, _ := g.PlanarDeck.Bottom(); bottom.InstanceID != top.InstanceID {
		t.Error("the revealed plane is not on the bottom of the planar deck")
	}
}

// --- Prahv -------------------------------------------------------------------

func TestPrahvCastingStopsAttackingAndAttackingStopsCasting(t *testing.T) {
	requirePlaneRegisteredFull(t, planePrahv)
	g, _ := newPlaneGame(t, planePrahv)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	if err := jmpB4CastSandboxSpell(t, g, me); err != nil {
		t.Fatalf("casting before attacking: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, opp.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("a player who cast a spell attacked: %v", err)
	}

	// The next turn: the new planar controller attacks first, then can't cast.
	next := toNextUpkeep(t, g)
	other := b12Creature(g, next.ID, "Other Raider", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, me.ID); err != nil {
		t.Fatalf("a player who cast nothing may attack: %v", err)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if err := jmpB4CastSandboxSpell(t, g, next); err == nil {
		t.Fatal("a player who attacked with creatures cast a spell")
	}
}

func TestPrahvChaosGainsLifeForTheHandSize(t *testing.T) {
	g, _ := newPlaneGame(t, planePrahv)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	life, hand := me.Life, me.Hand.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Life != life+hand {
		t.Errorf("life %d, want %d", me.Life, life+hand)
	}
}
