package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Star Trek batch 30 (#4973): Temporal Anomaly, Unexplored Territory,
// The Unicomplex and Vulcan.

const (
	trekB30TemporalAnomalyOracle     = "cf0d1859-699e-47d4-bc91-7c5ff61fabe9"
	trekB30UnexploredTerritoryOracle = "6e5d4bbd-4999-4acd-8775-25e02d901035"
	trekB30UnicomplexOracle          = "b38fcf9a-464a-432a-b13c-9a206d52dcd9"
	trekB30VulcanOracle              = "c7b6bc1f-9161-415e-a097-d9c8cd81d4ae"
)

func trekB30Land(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	id := b12Permanent(g, owner, name, "Land")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].OracleID = oracle
		}
	}
	return id
}

func TestTrekB30AllRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB30TemporalAnomalyOracle:     "Temporal Anomaly",
		trekB30UnexploredTerritoryOracle: "Unexplored Territory",
		trekB30UnicomplexOracle:          "The Unicomplex",
		trekB30VulcanOracle:              "Vulcan",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestTrekB30VulcanEntersTappedGainsOneAndTapsForWOrU(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	id := playLandFromHand(t, g, "Vulcan", trekB30VulcanOracle)
	top100AssertEnteredTapped(t, g, id, "Vulcan")
	if triggerOnStack(g, id) == nil {
		t.Error("the life is a trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
	spec, _ := Lookup(trekB30VulcanOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U}" {
		t.Errorf("mana abilities %+v, want one producing {W|U}", spec.ManaAbilities)
	}
}

func TestTrekB30UnexploredTerritoryFetchesABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := trekB30Land(g, me.ID, "Unexplored Territory", trekB30UnexploredTerritoryOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	b02bSetTapped(g, land, false)
	me.ManaPool = game.ManaPool{}
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Command Tower", "Land"),
	)
	advanceToMain(t, g)
	// No {1} to pay: a strict activation is refused and costs nothing.
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the fetch must cost {1}")
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("a refused activation sacrificed the land")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Fatal("the land is sacrificed as a cost")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two basics: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("only basic land cards are offered")
	}
	answerSearchNamed(t, g, me.ID, "Island")
	island := findBattlefieldByName(g, "Island")
	if island == uuid.Nil || !b16Tapped(t, g, island) {
		t.Error("the basic enters tapped")
	}
}

func trekB30ActivateAnomaly(t *testing.T, g *game.Game, me *game.Player, lib ...game.Card) []uuid.UUID {
	t.Helper()
	land := trekB30Land(g, me.ID, "Temporal Anomaly", trekB30TemporalAnomalyOracle)
	ids := mm2Library(me, lib...)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the activation must cost {3}{G}")
	}
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{3}{G} needs green")
	}
	b06AddMana(me, "G")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(land) {
		t.Fatal("the land is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	return ids
}

func trekB30NewPlanet(t *testing.T, g *game.Game) {
	t.Helper()
	planet := findBattlefieldByName(g, "New Planet")
	if planet == uuid.Nil {
		t.Fatal("no New Planet token")
	}
	c, _ := battlefieldCard(g, planet)
	if !c.IsToken() || !c.HasSubtype("planet") || !c.IsLand() || !c.Tapped {
		t.Errorf("New Planet %+v: want a tapped Planet land token", c)
	}
}

func TestTrekB30TemporalAnomalyMillsFourTakesAPermanentAndMakesAPlanet(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := trekB30ActivateAnomaly(t, g, me,
		mm2Spell("Spell A"), game.Card{Name: "Prize", TypeLine: "Artifact"}, mm2Spell("Spell B"), mm2Spell("Spell C"), mm2Spell("Deep"))
	if findBattlefieldByName(g, "New Planet") != uuid.Nil {
		t.Error("the token is made after the pick, in printed order")
	}
	if c := chooseCardsChoiceFor(g, me.ID); c == nil || len(c.ChooseCards) != 1 {
		t.Fatalf("the pick offers the one milled permanent card, got %+v", c)
	}
	answerChooseCards(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the permanent card should be in hand")
	}
	if me.Library.Size() != 1 || !me.Graveyard.Contains(ids[0]) || me.Graveyard.Contains(ids[4]) {
		t.Errorf("exactly the top four should be milled (library %d)", me.Library.Size())
	}
	trekB30NewPlanet(t, g)
}

func TestTrekB30TemporalAnomalyWithNoPermanentStillMakesAPlanet(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB30ActivateAnomaly(t, g, me, mm2Spell("A"), mm2Spell("B"), mm2Spell("C"), mm2Spell("D"))
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no pick should be offered when nothing milled is a permanent")
	}
	trekB30NewPlanet(t, g)
}

func TestTrekB30UnicomplexEntersTappedWithFewerThanTwoOtherLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	id := playLandFromHand(t, g, "The Unicomplex", trekB30UnicomplexOracle)
	top100AssertEnteredTapped(t, g, id, "The Unicomplex")
	if modePickChoiceFor(g, me.ID) != nil || triggerOnStack(g, id) != nil {
		t.Error("entering tapped, the trigger does not fire")
	}
	spec, _ := Lookup(trekB30UnicomplexOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B}" {
		t.Errorf("mana abilities %+v, want one producing {B}", spec.ManaAbilities)
	}
}

func TestTrekB30UnicomplexUntappedTucksACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	bear := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	batch01GraveyardCard(me, "Old Spell", "Instant")
	id := playLandFromHand(t, g, "The Unicomplex", trekB30UnicomplexOracle)
	if c, ok := battlefieldCard(g, id); !ok || c.Tapped {
		t.Fatal("two other lands: it enters untapped")
	}
	chooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bear) {
		t.Error("the creature card is still in the graveyard")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != bear {
		t.Error("the creature card did not go on top of the library")
	}
}

func TestTrekB30UnicomplexUntappedSurveilsWithNoCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	batch01GraveyardCard(me, "Old Spell", "Instant")
	mm2Library(me, mm2Spell("Card A"), mm2Spell("Card B"), mm2Spell("Card C"))
	id := playLandFromHand(t, g, "The Unicomplex", trekB30UnicomplexOracle)
	if c, ok := battlefieldCard(g, id); !ok || c.Tapped {
		t.Fatal("two other lands: it enters untapped")
	}
	// No creature card to target: only the surveil can be chosen.
	chooseModeNow(t, g, me.ID, []int{1}, 1)
	passPriorityAroundTable(t, g)
	sv := latestChoiceOfKind(g, game.PendingChoiceSurveil)
	if sv == nil {
		t.Fatal("no surveil prompt")
	}
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	if len(sv.ScryCards) != 1 || sv.ScryCards[0] != top {
		t.Errorf("surveil 1 looks at the top card, got %v", sv.ScryCards)
	}
}
