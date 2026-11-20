package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// speed_slice_b_test.go — the 13 speed cards of #2711 slice speed-b, on
// ADR 0138's seam. Each test sets the controller's speed directly
// (SetSpeedForTest) and checks the "Max speed —" ability is off below 4.

const (
	embalmedAscendantOracle  = "2f4d7988-3f5b-45d7-9838-02897d7180be"
	slickImitatorOracle      = "5482ddb1-7f7d-492a-aa3d-67c037062cc3"
	prideOfTheRoadOracle     = "2af236b0-6b30-4ddc-8735-864d2ffad68e"
	outpaceOblivionOracle    = "4d92f48d-c851-4b71-b7fa-916459928d59"
	risenNecroregentOracle   = "527cc22f-b471-4a84-93a8-be56193a9723"
	walkingSarcophagusOracle = "791b8ac8-8db8-44b4-8c85-fcc9017c17ef"
	endriderSpikespitterOrac = "817f3b8b-4399-4c28-9453-7f57df4a7360"
	leoninSurveyorOracle     = "97c29db0-b3e1-4704-afe8-5218250c044c"
	streakingOilgorgerOracle = "c163a95e-9414-4b8c-b5a2-1e7674530404"
	mutantSurveyorOracle     = "f35aace7-57b5-44b8-91d2-4710eb7d5548"
	swiftwingAssailantOracle = "46b36d25-25bd-48c5-9a53-75230bb27dee"
	loxodonSurveyorOracle    = "970f26d4-c916-444a-a2d2-4c1f5bb5556b"
	glitchGhostSurveyorOrac  = "f0f83ab8-6185-4139-b41e-a11500036915"
)

func TestSpeedSliceBCardsDeclareStartYourEngines(t *testing.T) {
	for _, oracle := range []string{
		embalmedAscendantOracle, slickImitatorOracle, prideOfTheRoadOracle, outpaceOblivionOracle,
		risenNecroregentOracle, walkingSarcophagusOracle, endriderSpikespitterOrac, leoninSurveyorOracle,
		streakingOilgorgerOracle, mutantSurveyorOracle, swiftwingAssailantOracle, loxodonSurveyorOracle,
		glitchGhostSurveyorOrac,
	} {
		spec, ok := effects.Lookup(oracle)
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

func TestWalkingSarcophagusGrowsAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	id := pushSpeedCard(t, g, me, "Walking Sarcophagus", "Artifact Creature — Zombie Cat", walkingSarcophagusOracle, 2, 1)
	if effectstest.EffectivePower(t, g, id) != 2 || effectstest.EffectiveToughness(t, g, id) != 1 {
		t.Errorf("below max speed = %d/%d, want 2/1", effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id))
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if effectstest.EffectivePower(t, g, id) != 3 || effectstest.EffectiveToughness(t, g, id) != 3 {
		t.Errorf("at max speed = %d/%d, want 3/3", effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id))
	}
}

func TestSwiftwingAssailantGetsToughnessAndVigilanceAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	id := pushSpeedCard(t, g, me, "Swiftwing Assailant", "Creature — Bird Warrior", swiftwingAssailantOracle, 3, 3)
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, id), "vigilance") || effectstest.EffectiveToughness(t, g, id) != 3 {
		t.Error("vigilance or toughness bonus below max speed")
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, id), "flying") {
		t.Error("printed flying missing")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, id), "vigilance") {
		t.Error("no vigilance at max speed")
	}
	if effectstest.EffectivePower(t, g, id) != 3 || effectstest.EffectiveToughness(t, g, id) != 4 {
		t.Errorf("at max speed = %d/%d, want 3/4", effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id))
	}
}

func TestStreakingOilgorgerHasLifelinkAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	id := pushSpeedCard(t, g, me, "Streaking Oilgorger", "Creature — Vampire", streakingOilgorgerOracle, 3, 3)
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, id), "lifelink") {
		t.Error("lifelink below max speed")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, id), "lifelink") {
		t.Error("no lifelink at max speed")
	}
}

func TestLeoninSurveyorHasFirstStrikeOnlyDuringItsControllersTurn(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	opp := speedSliceAOpponent(g)
	mine := pushSpeedCard(t, g, me, "Leonin Surveyor", "Creature — Cat Scout", leoninSurveyorOracle, 2, 2)
	theirs := pushSpeedCard(t, g, opp, "Leonin Surveyor", "Creature — Cat Scout", leoninSurveyorOracle, 2, 2)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, mine), "first strike") {
		t.Error("no first strike on its controller's turn")
	}
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, theirs), "first strike") {
		t.Error("first strike on an opponent's turn")
	}
}

// The four Surveyors with the graveyard draw.
func TestSurveyorsDrawFromTheGraveyardAtMaxSpeed(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Leonin Surveyor", "Creature — Cat Scout", leoninSurveyorOracle},
		{"Mutant Surveyor", "Creature — Mutant Scout", mutantSurveyorOracle},
		{"Loxodon Surveyor", "Creature — Elephant Scout", loxodonSurveyorOracle},
		{"Glitch Ghost Surveyor", "Creature — Spirit Scout", glitchGhostSurveyorOrac},
	} {
		t.Run(c.name, func(t *testing.T) {
			g, me := effectstest.SpeedGame(t)
			id := effectstest.PushCatalogGraveyardCard(me, c.name, c.typeLine, c.oracle, 2, 2)
			idx := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: c.oracle})) - 1
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
				t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
			}
			g.SetSpeedForTest(me.ID, game.MaxSpeed)
			hand := me.Hand.Size()
			if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("at max speed: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if me.Graveyard.Contains(id) {
				t.Error("the Surveyor is still in the graveyard")
			}
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d, want 1", got)
			}
		})
	}
}

func TestMutantSurveyorPumpsItselfForTwo(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	id := pushSpeedCard(t, g, me, "Mutant Surveyor", "Creature — Mutant Scout", mutantSurveyorOracle, 2, 3)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, id) != 3 || effectstest.EffectiveToughness(t, g, id) != 4 {
		t.Errorf("after the pump = %d/%d, want 3/4", effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id))
	}
}

func TestRisenNecroregentMakesAZombieAtMaxSpeedOnly(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	pushSpeedCard(t, g, me, "Risen Necroregent", "Creature — Zombie Cat Knight", risenNecroregentOracle, 5, 4)
	before := effectstest.CountOnBattlefield(g, "Zombie", me.ID)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Zombie", me.ID) - before; got != 0 {
		t.Errorf("below max speed made %d Zombies", got)
	}
	g2, me2 := effectstest.SpeedGame(t)
	pushSpeedCard(t, g2, me2, "Risen Necroregent", "Creature — Zombie Cat Knight", risenNecroregentOracle, 5, 4)
	g2.SetSpeedForTest(me2.ID, game.MaxSpeed)
	effectstest.AdvanceToEndStep(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CountOnBattlefield(g2, "Zombie", me2.ID); got != 1 {
		t.Errorf("at max speed made %d Zombies, want 1", got)
	}
}

func TestEndriderSpikespitterImpulseExilesAtMaxSpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	pushSpeedCard(t, g, me, "Endrider Spikespitter", "Creature — Human Mercenary", endriderSpikespitterOrac, 3, 4)
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib {
		t.Errorf("below max speed the library went %d to %d", lib, me.Library.Size())
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	// Next of seat 1's upkeeps.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	lib = me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	// The draw step happens after the upkeep, so the library lost the
	// exiled card and nothing else yet.
	if got := lib - me.Library.Size(); got != 1 {
		t.Errorf("at max speed the library lost %d, want 1", got)
	}
}

func TestPrideOfTheRoadGivesDoubleStrikeAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	pushSpeedCard(t, g, me, "Pride of the Road", "Creature — Zombie Cat Warrior", prideOfTheRoadOracle, 2, 5)
	ally := pushSpeedCard(t, g, me, "Ally", "Creature — Bear", "", 2, 2)
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTarget(t, g, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, ally), "double strike") {
		t.Error("the chosen creature has no double strike at max speed")
	}
}

func TestPrideOfTheRoadDoesNothingBelowMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	pushSpeedCard(t, g, me, "Pride of the Road", "Creature — Zombie Cat Warrior", prideOfTheRoadOracle, 2, 5)
	ally := pushSpeedCard(t, g, me, "Ally", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, ally), "double strike") {
		t.Error("double strike below max speed")
	}
}

func TestEmbalmedAscendantMakesAZombieAndDrainsOnlyAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	opp := speedSliceAOpponent(g)
	effectstest.CastCatalogSpell(t, g, "Embalmed Ascendant", "Creature — Zombie", embalmedAscendantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Zombie", me.ID); got != 1 {
		t.Fatalf("ETB Zombie tokens = %d, want 1", got)
	}
	zombie := effectstest.BattlefieldIDsNamed(g, "Zombie")[0]
	lifeMe, lifeOpp := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lifeMe || opp.Life != lifeOpp {
		t.Errorf("drained below max speed: me %d->%d opp %d->%d", lifeMe, me.Life, lifeOpp, opp.Life)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := lifeOpp - opp.Life; got != 1 {
		t.Errorf("opponent lost %d, want 1", got)
	}
	if got := me.Life - lifeMe; got != 1 {
		t.Errorf("you gained %d, want 1", got)
	}
	// An opponent's creature dying is not "you control".
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	lifeOpp = opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lifeOpp {
		t.Error("an opponent's creature dying drained")
	}
}

func TestSlickImitatorCopiesYourSpellAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	imitator := pushSpeedCard(t, g, me, "Slick Imitator", "Creature — Ooze", slickImitatorOracle, 1, 3)
	spell := effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.FloatForTest(g, me, "C")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: spell}}
	if err := g.ActivateCatalogAbility(me.ID, imitator, 0, game.ActivateAbilityParams{Targets: target}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, imitator, 0, game.ActivateAbilityParams{Targets: target}); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, imitator) {
		t.Error("the Imitator was not sacrificed")
	}
	if got := effectstest.CountOnBattlefield(g, "Grizzly Bears", me.ID); got != 2 {
		t.Errorf("Grizzly Bears on the battlefield = %d, want 2 (the spell and its token copy)", got)
	}
}

func TestOutpaceOblivionBlastsACreatureOnEntry(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	opp := speedSliceAOpponent(g)
	victim := effectstest.PushCatalogPermanent(g, opp.ID, "Big Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 5, 5
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Outpace Oblivion", "Enchantment", outpaceOblivionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, victim) {
		t.Error("the targeted creature survived 5 damage")
	}
	_ = me
}

func TestOutpaceOblivionHitsOnlyPlayersWithoutMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	opp := speedSliceAOpponent(g)
	id := pushSpeedCard(t, g, me, "Outpace Oblivion", "Enchantment", outpaceOblivionOracle, 0, 0)
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	lifeMe, lifeOpp := me.Life, opp.Life
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lifeMe {
		t.Errorf("a player with max speed lost %d", lifeMe-me.Life)
	}
	if got := lifeOpp - opp.Life; got != 2 {
		t.Errorf("a player without max speed lost %d, want 2", got)
	}
	if effects.OnBattlefield(g, id) {
		t.Error("the enchantment was not sacrificed")
	}
}
