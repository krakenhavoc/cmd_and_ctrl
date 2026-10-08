package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// speed_cards_test.go — ADR 0138's proof cards (#2122). Each test sets
// the controller's speed directly (SetSpeedForTest, through the one
// write and its event) and checks the card's "Max speed —" ability is
// off below 4 and on at 4. The speed rules themselves are pinned in
// game/speed_test.go.

const (
	gastalRaiderOracle       = "26c09127-8225-4ba2-9f97-f5558b1c41b5"
	perilousSnareOracle      = "cff7499a-44b0-4a2a-b96c-ed2cafb1a90d"
	muragandaRacewayOracle   = "b5fa5651-d714-44d6-867b-be0e3224b7ed"
	amonkhetRacewayOracle    = "fe174586-d36b-40f6-babd-1f98e76eec22"
	vnwxtOracle              = "6454c457-a278-431a-9e8d-bfd1a966bdde"
	avishkarRacewayOracle    = "e2c35551-1ba5-4424-baf9-821b49bbcc8c"
	theSpeedDemonOracle      = "c9fb22b4-dc22-4291-91f2-89c07cee7569"
	racersScoreboardOracle   = "54c2b3b3-e247-4146-8045-d49480b795f2"
	startingColumnOracle     = "d608a3fa-faee-44a4-9ab6-be701d3b7a49"
	aetherSyphonOracle       = "d8956775-fda8-475d-9e8a-b61196a958ea"
	burnoutBashtronautOracle = "1b122fbb-c2e1-42d1-bfc5-fcacacdfa0bc"
	goblinSurveyorOracle     = "237758f3-84b2-4e5b-a2c7-b28c893e2790"
)

// speedGame is a catalog game walked to the active seat's precombat
// main (every max-speed sorcery ability is legal there, and a manual
// draw is not swallowed by the turn's draw step).
func speedGame(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	return g, g.Seats[g.Turn.ActiveSeat]
}

// pushSpeedCard puts a catalog card on the battlefield as a non-sick
// permanent and runs the state-based actions, so its start your
// engines! has given its controller speed.
func pushSpeedCard(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner.ID, Controller: owner.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	g.RunStateChecksForTest()
	return id
}

func TestSpeedCardsDeclareStartYourEngines(t *testing.T) {
	for _, oracle := range []string{
		gastalRaiderOracle, perilousSnareOracle, muragandaRacewayOracle, amonkhetRacewayOracle,
		vnwxtOracle, avishkarRacewayOracle, theSpeedDemonOracle, racersScoreboardOracle,
		startingColumnOracle, aetherSyphonOracle, burnoutBashtronautOracle, goblinSurveyorOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		found := false
		for _, kw := range spec.PrintedKeywords {
			if kw == game.KeywordStartYourEngines {
				found = true
			}
		}
		if !found {
			t.Errorf("%s does not declare start your engines!", spec.Name)
		}
	}
}

// An uncatalogued speed card still starts its controller's speed: the
// deck importer stamps the canonical keyword from Scryfall's keywords
// array, and the state-based action reads it (CR 702.179a).
func TestImportedUncataloguedSpeedCardStartsSpeed(t *testing.T) {
	g, me := speedGame(t)
	row := themeDeckRow("Test Racer", "", "Creature — Test", "{1}")
	row.Power, row.Toughness = "1", "1"
	row.Keywords = []string{"Start your engines!", "Max speed"}
	row.OracleText = "Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)\nMax speed — This creature gets +1/+1."
	c := importThemeCard(row, me.ID)
	found := false
	for _, kw := range c.Keywords {
		if kw == game.KeywordStartYourEngines {
			found = true
		}
	}
	if !found {
		t.Fatalf("the importer did not stamp start your engines!: keywords = %v", c.Keywords)
	}
	pushBattlefieldCardWithTimestamp(g, c)
	g.RunStateChecksForTest()
	if me.Speed != 1 {
		t.Errorf("speed = %d, want 1 from an uncatalogued start-your-engines card", me.Speed)
	}
}

func TestGastalRaiderGivesSpeedAndIsBiggerAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	id := pushSpeedCard(t, g, me, "Gastal Raider", "Creature — Vampire Rogue", gastalRaiderOracle, 2, 1)
	if me.Speed != 1 {
		t.Fatalf("speed = %d, want 1 from Gastal Raider's start your engines!", me.Speed)
	}
	if p, tt := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 2 || tt != 1 {
		t.Errorf("below max speed: %d/%d, want 2/1", p, tt)
	}
	if eotHasAbility(effectiveAbilities(t, g, id), "menace") {
		t.Error("menace below max speed")
	}

	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if p, tt := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 3 || tt != 2 {
		t.Errorf("at max speed: %d/%d, want 3/2", p, tt)
	}
	if !eotHasAbility(effectiveAbilities(t, g, id), "menace") {
		t.Error("no menace at max speed")
	}

	spec, _ := Lookup(gastalRaiderOracle)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil {
		t.Error("Gastal Raider's enters trigger should target an opponent")
	}
}

func TestPerilousSnareCounterNeedsMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	snare := pushSpeedCard(t, g, me, "Perilous Snare", "Artifact", perilousSnareOracle, 0, 0)
	bear := pushSpeedCard(t, g, me, "Bear", "Creature — Bear", "", 2, 2)

	params := game.ActivateAbilityParams{Targets: cardRefs(bear)}
	if err := g.ActivateCatalogAbility(me.ID, snare, 0, params); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("activate below max speed: err = %v, want ErrConditionNotMet", err)
	}
	if row := activatedRowOf(t, g, me.ID, snare, 0); !row.ConditionUnmet {
		t.Error("the max-speed row is not greyed below max speed")
	}

	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if row := activatedRowOf(t, g, me.ID, snare, 0); row.ConditionUnmet {
		t.Error("the max-speed row is still greyed at max speed")
	}
	if err := g.ActivateCatalogAbility(me.ID, snare, 0, params); err != nil {
		t.Fatalf("activate at max speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the target = %d, want 1", got)
	}
}

func TestMuragandaRacewayTwoColorlessAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	land := pushSpeedCard(t, g, me, "Muraganda Raceway", "Land", muragandaRacewayOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("{C}{C} below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	activateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{})
	if got := len(me.ManaPool); got != 2 {
		t.Errorf("pool holds %d mana, want 2", got)
	}
}

func TestAmonkhetRacewayGrantsHasteAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	land := pushSpeedCard(t, g, me, "Amonkhet Raceway", "Land", amonkhetRacewayOracle, 0, 0)
	bear := pushSpeedCard(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	params := game.ActivateAbilityParams{Targets: cardRefs(bear)}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, params); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, params); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !eotHasAbility(effectiveAbilities(t, g, bear), "haste") {
		t.Error("the target did not gain haste")
	}
}

func TestVnwxtDrawsTwoOnlyAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	pushSpeedCard(t, g, me, "Vnwxt, Verbose Host", "Legendary Creature — Homunculus", vnwxtOracle, 0, 4)
	draw := func() int {
		before := me.Hand.Size()
		g.WithWriteLock(func() {
			if err := g.DrawNForEffect(me.ID, 1); err != nil {
				t.Fatalf("DrawNForEffect: %v", err)
			}
		})
		return me.Hand.Size() - before
	}
	if got := draw(); got != 1 {
		t.Errorf("below max speed a draw drew %d, want 1", got)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if got := draw(); got != 2 {
		t.Errorf("at max speed a draw drew %d, want 2", got)
	}
}

func TestAvishkarRacewayRummagesAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	land := pushSpeedCard(t, g, me, "Avishkar Raceway", "Land", avishkarRacewayOracle, 0, 0)
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	pitch := me.Hand.Cards[0].InstanceID
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand = %d, want %d (one discarded, one drawn)", got, before)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestTheSpeedDemonDrawsAndLosesXAtEndStep(t *testing.T) {
	g, me := speedGame(t)
	pushSpeedCard(t, g, me, "The Speed Demon", "Legendary Creature — Demon", theSpeedDemonOracle, 5, 5)
	g.SetSpeedForTest(me.ID, 3)
	hand, life := me.Hand.Size(), me.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3 (speed 3)", got)
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
}

func TestRacersScoreboardDiscountsOnlyAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	pushSpeedCard(t, g, me, "Racers' Scoreboard", "Artifact", racersScoreboardOracle, 0, 0)
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("below max speed: %d, want 2", got)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 1 {
		t.Errorf("at max speed: %d, want 1", got)
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if got := priceInHand(t, g, opp, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("an opponent's spell: %d, want 2", got)
	}
}

func TestStartingColumnSacrificeNeedsMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	col := pushSpeedCard(t, g, me, "Starting Column", "Artifact", startingColumnOracle, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, col, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	if !g.Battlefield.Contains(col) {
		t.Fatal("a refused activation paid its sacrifice")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, col, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	if g.Battlefield.Contains(col) {
		t.Error("Starting Column was not sacrificed")
	}
}

func TestAetherSyphonMillsOnlyAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	pushSpeedCard(t, g, me, "Aether Syphon", "Artifact", aetherSyphonOracle, 0, 0)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	drawAndSettle := func() {
		g.WithWriteLock(func() {
			if err := g.DrawNForEffect(me.ID, 1); err != nil {
				t.Fatalf("DrawNForEffect: %v", err)
			}
		})
		passPriorityAroundTable(t, g)
	}
	before := opp.Graveyard.Size()
	drawAndSettle()
	if got := opp.Graveyard.Size() - before; got != 0 {
		t.Errorf("below max speed an opponent milled %d", got)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	before = opp.Graveyard.Size()
	drawAndSettle()
	if got := opp.Graveyard.Size() - before; got != 2 {
		t.Errorf("at max speed an opponent milled %d, want 2", got)
	}
	if got := me.Graveyard.Size(); got != 0 {
		t.Errorf("the Syphon's controller milled %d", got)
	}
}

func TestBurnoutBashtronautDoubleStrikeAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	id := pushSpeedCard(t, g, me, "Burnout Bashtronaut", "Creature — Goblin Warrior", burnoutBashtronautOracle, 1, 1)
	if eotHasAbility(effectiveAbilities(t, g, id), "double strike") {
		t.Error("double strike below max speed")
	}
	if !eotHasAbility(effectiveAbilities(t, g, id), "menace") {
		t.Error("no printed menace")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if !eotHasAbility(effectiveAbilities(t, g, id), "double strike") {
		t.Error("no double strike at max speed")
	}
}

func TestGoblinSurveyorDrawsFromTheGraveyardAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	id := pushCatalogGraveyardCard(me, "Goblin Surveyor", "Creature — Goblin Scout", goblinSurveyorOracle, 3, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("the Surveyor is still in the graveyard")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}
