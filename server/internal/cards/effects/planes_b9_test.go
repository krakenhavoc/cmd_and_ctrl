package effects

import (
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
