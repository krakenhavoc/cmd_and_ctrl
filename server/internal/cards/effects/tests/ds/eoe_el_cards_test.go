package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_el_cards_test.go — Edge of Eternities cards whose names start with E
// to L (#4461). Each test plays the real catalog entry and asserts what its
// text says, and what it refuses.

const (
	eoeELEjectOracle         = "da363da1-aecf-4d49-8e77-8dc56981a741"
	eoeELTerrabotanistOracle = "92226fd2-ad93-4722-89a0-ca88ea03e1b4"
	eoeELWayfarerOracle      = "86746a84-3e8a-44b4-8494-60cbee9816b5"
	eoeELWurmOracle          = "8735c62d-1508-4bdc-a5d0-f7c68d892a5e"
	eoeELGravkillOracle      = "15c6837c-775b-4b4a-a2f7-9fd79215fa05"
	eoeELMonoistOracle       = "9c009b3f-6580-4ab8-84e8-d4cf7a456a02"
	eoeELHonorOracle         = "48c0cd06-465d-4901-bf23-948a003afc0b"
	eoeELHullcarverOracle    = "931c2153-e5e5-4642-bff0-f9dadcc24c9a"
	eoeELCrasherOracle       = "b9650a0e-5795-42ff-84c6-6040e5cf4c90"
	eoeELGalebladeOracle     = "d6273621-c36c-491b-85ca-37445ed30097"
	eoeELTenderfootOracle    = "94023b23-339a-426e-85cc-cc361ef68754"
	eoeELManeuversOracle     = "4c376349-1a50-4215-8a40-f702da81eb7b"
	eoeELLuminaryOracle      = "11809d9b-30f3-49d5-9119-e45b4de044c8"
)

func TestEoeELEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeELEjectOracle:         "Emergency Eject",
		eoeELTerrabotanistOracle: "Eumidian Terrabotanist",
		eoeELWayfarerOracle:      "Galactic Wayfarer",
		eoeELWurmOracle:          "Germinating Wurm",
		eoeELGravkillOracle:      "Gravkill",
		eoeELMonoistOracle:       "Gravpack Monoist",
		eoeELHonorOracle:         "Honor",
		eoeELHullcarverOracle:    "Hullcarver",
		eoeELCrasherOracle:       "Icecave Crasher",
		eoeELGalebladeOracle:     "Illvoi Galeblade",
		eoeELTenderfootOracle:    "Intrepid Tenderfoot",
		eoeELManeuversOracle:     "Invasive Maneuvers",
		eoeELLuminaryOracle:      "Knight Luminary",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for oracle, cost := range map[string]string{
		eoeELWurmOracle:     "{1}{G}",
		eoeELLuminaryOracle: "{1}{W}",
	} {
		if ac := game.AlternativeCostByKey(oracle, "warp"); ac == nil || ac.ManaCost != cost {
			t.Errorf("%s: warp %+v, want %s", oracle, ac, cost)
		}
	}
}

// --- helpers ---------------------------------------------------------------

// eoeELDies reports whether a permanent is in its owner's graveyard after
// the state-based actions.
func eoeELDies(g *game.Game, owner *game.Player, id uuid.UUID) bool {
	g.RunStateChecksForTest()
	return !g.Battlefield.Contains(id) && owner.Graveyard.Contains(id)
}

// --- Emergency Eject ---------------------------------------------------------

func TestEoeELEmergencyEjectDestroysAndGivesItsControllerALander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Emergency Eject", "Instant", eoeELEjectOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
	if got := effectstest.EoeCDCount(g, opp.ID, "Lander"); got != 1 {
		t.Errorf("its controller has %d Landers, want 1", got)
	}
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 0 {
		t.Errorf("the caster has %d Landers, want 0", got)
	}
}

func TestEoeELEmergencyEjectGivesTheLanderEvenIfNothingIsDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 4, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Emergency Eject", "Instant", eoeELEjectOracle, effectstest.CardRefs(wall))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Error("an indestructible creature survives")
	}
	if got := effectstest.EoeCDCount(g, opp.ID, "Lander"); got != 1 {
		t.Errorf("its controller still gets a Lander: %d, want 1", got)
	}
}

func TestEoeELEmergencyEjectCannotTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	if err := effectstest.CastCatalogSpellErr(t, g, "Emergency Eject", "Instant", eoeELEjectOracle, effectstest.CardRefs(land)); err == nil {
		t.Error("a land is not a nonland permanent")
	}
}

// --- Eumidian Terrabotanist --------------------------------------------------

func TestEoeELEumidianTerrabotanistGainsOneForEachOfYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Eumidian Terrabotanist", "Creature — Insect Druid", eoeELTerrabotanistOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestEoeELEumidianTerrabotanistIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Eumidian Terrabotanist", "Creature — Insect Druid", eoeELTerrabotanistOracle, 2, 3)
	life := me.Life
	effectstest.B13PlayAs(t, g, 1, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an opponent's land changed your life %d -> %d", life, me.Life)
	}
}

// --- Galactic Wayfarer -------------------------------------------------------

func TestEoeELGalacticWayfarerMakesALanderAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Galactic Wayfarer", "Creature — Human Scout", eoeELWayfarerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	lander := effectstest.AvB1Named(g, me.ID, "Lander")
	if lander == uuid.Nil {
		t.Fatal("no Lander token")
	}
	if c := effectstest.LayeredCard(t, g, lander); !c.IsArtifact() || c.IsCreature() {
		t.Error("the Lander is a noncreature artifact")
	}
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("Landers %d, want 1", got)
	}
}

// --- Germinating Wurm --------------------------------------------------------

func TestEoeELGerminatingWurmGainsTwoLifeAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Germinating Wurm", "Creature — Plant Wurm", eoeELWurmOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestEoeELGerminatingWurmWarpedStillGainsTheLifeThenLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	wurm := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Germinating Wurm", TypeLine: "Creature — Plant Wurm", ManaCost: "{4}{G}",
		OracleID: eoeELWurmOracle, Power: 5, Toughness: 5}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) || me.Life != life+2 {
		t.Fatalf("the warped Wurm enters and gains 2: on battlefield %v, life %d -> %d", g.Battlefield.Contains(wurm), life, me.Life)
	}
	effectstest.AdvanceThroughEndStep(t, g)
	if !g.Exile.Contains(wurm) {
		t.Error("the warped Wurm is exiled at the end step")
	}
}

// --- Gravkill ----------------------------------------------------------------

func TestEoeELGravkillExilesACreatureOrASpacecraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Gravkill", "Instant", eoeELGravkillOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the creature is exiled")
	}
	ship := effectstest.B12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	effectstest.CastCatalogSpell(t, g, "Gravkill", "Instant", eoeELGravkillOracle, effectstest.CardRefs(ship))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ship) {
		t.Error("a Spacecraft that isn't a creature is exiled too")
	}
}

func TestEoeELGravkillCannotTargetAnotherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Gravkill", "Instant", eoeELGravkillOracle, effectstest.CardRefs(rock)); err == nil {
		t.Error("an artifact that is neither a creature nor a Spacecraft is not a legal target")
	}
}

// --- Gravpack Monoist --------------------------------------------------------

func TestEoeELGravpackMonoistLeavesATappedRobotWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	monoist := effectstest.B12Push(g, me.ID, "Gravpack Monoist", "Creature — Human Scout", eoeELMonoistOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, monoist, "flying") {
		t.Error("Gravpack Monoist has flying")
	}
	effectstest.B15Destroy(t, g, monoist)
	robot := effectstest.AvB1Named(g, me.ID, "Robot")
	if robot == uuid.Nil {
		t.Fatal("no Robot token")
	}
	if c := effectstest.LayeredCard(t, g, robot); !c.Tapped || !c.IsArtifact() || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("a tapped 2/2 Robot artifact creature: tapped %v, %d/%d", c.Tapped, c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Honor -------------------------------------------------------------------

func TestEoeELHonorPutsACounterAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Honor", "Sorcery", eoeELHonorOracle, effectstest.CardRefs(bear))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3 with a +1/+1 counter", p, tough)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Hullcarver --------------------------------------------------------------

func TestEoeELHullcarverHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	carver := effectstest.B12Push(g, me.ID, "Hullcarver", "Artifact Creature — Robot Assassin", eoeELHullcarverOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, carver, "deathtouch") {
		t.Error("Hullcarver has deathtouch")
	}
}

// --- Icecave Crasher ---------------------------------------------------------

func TestEoeELIcecaveCrasherGetsPlusOneZeroOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crasher := effectstest.B12Push(g, me.ID, "Icecave Crasher", "Creature — Beast", eoeELCrasherOracle, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, crasher, "trample") {
		t.Error("Icecave Crasher has trample")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, crasher); p != 5 || tough != 4 {
		t.Errorf("after a land the Crasher is %d/%d, want 5/4", p, tough)
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	if p, tough := effectstest.EoeCDPT(t, g, crasher); p != 4 || tough != 4 {
		t.Errorf("the pump ends with the turn: %d/%d, want 4/4", p, tough)
	}
}

// --- Illvoi Galeblade --------------------------------------------------------

func TestEoeELIllvoiGalebladeHasFlashAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := effectstest.AvB1Hand(me, game.Card{Name: "Illvoi Galeblade", TypeLine: "Creature — Jellyfish Warrior", ManaCost: "{U}",
		OracleID: eoeELGalebladeOracle, Power: 1, Toughness: 1})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash: cast in combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Galeblade resolved onto the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Illvoi Galeblade has flying")
	}
}

func TestEoeELIllvoiGalebladeSacrificesToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	blade := effectstest.B12Push(g, me.ID, "Illvoi Galeblade", "Creature — Jellyfish Warrior", eoeELGalebladeOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.FillPoolColored(me, "C", 2)
	if err := effectstest.EoeCDActivate(g, me.ID, blade, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(blade) {
		t.Error("the Galeblade is sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Intrepid Tenderfoot -----------------------------------------------------

func TestEoeELIntrepidTenderfootGrowsAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	foot := effectstest.B12Push(g, me.ID, "Intrepid Tenderfoot", "Creature — Insect Citizen", eoeELTenderfootOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		effectstest.FillPoolColored(me, "C", 3)
		if err := effectstest.EoeCDActivate(g, me.ID, foot, 0); err != nil {
			t.Fatalf("activation %d in your main phase: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.EoeCDPT(t, g, foot); p != 4 || tough != 4 {
		t.Errorf("two activations make it %d/%d, want 4/4", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "C", 3)
	if err := effectstest.EoeCDActivate(g, me.ID, foot, 0); err == nil {
		t.Error("activated in combat: 'activate only as a sorcery'")
	}
}

// --- Invasive Maneuvers ------------------------------------------------------

func TestEoeELInvasiveManeuversDealsThreeWithoutASpacecraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.B12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 4)
	effectstest.CastCatalogSpell(t, g, "Invasive Maneuvers", "Instant", eoeELManeuversOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if eoeELDies(g, opp, ogre) {
		t.Fatal("3 damage doesn't kill a 4-toughness creature: an opponent's Spacecraft doesn't count")
	}
	if c := effectstest.AvB1Card(t, g, ogre); c.DamageMarked != 3 {
		t.Errorf("damage marked %d, want 3", c.DamageMarked)
	}
}

func TestEoeELInvasiveManeuversDealsFiveWithYourSpacecraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "My Ship", "Artifact — Spacecraft")
	giant := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Invasive Maneuvers", "Instant", eoeELManeuversOracle, effectstest.CardRefs(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if !eoeELDies(g, opp, giant) {
		t.Error("with a Spacecraft, 5 damage kills a 5-toughness creature")
	}
}

func TestEoeELInvasiveManeuversTargetsOnlyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	if err := effectstest.CastCatalogSpellErr(t, g, "Invasive Maneuvers", "Instant", eoeELManeuversOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("a player is not a target creature")
	}
}

// --- Knight Luminary ---------------------------------------------------------

func TestEoeELKnightLuminaryMakesASoldierThatOutlastsItsWarp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	knight := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Knight Luminary", TypeLine: "Creature — Human Knight", ManaCost: "{3}{W}",
		OracleID: eoeELLuminaryOracle, Power: 3, Toughness: 2}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	soldier := effectstest.AvB1Named(g, me.ID, "Human Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Human Soldier token")
	}
	if c := effectstest.LayeredCard(t, g, soldier); !c.IsCreature() || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("a 1/1 creature token, got %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AdvanceThroughEndStep(t, g)
	if !g.Exile.Contains(knight) {
		t.Error("the warped Knight is exiled at the end step")
	}
	if !g.Battlefield.Contains(soldier) {
		t.Error("the Soldier stays")
	}
}
