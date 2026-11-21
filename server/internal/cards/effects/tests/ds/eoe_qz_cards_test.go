package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_qz_cards_test.go — Edge of Eternities cards whose names start with Q
// to Z (#4461). Each test plays the real catalog entry and asserts what its
// text says, and what it refuses.

const (
	eoeQZRadiantStrikeOracle = "186efd85-83d4-40cc-b617-76f32dbe99ce"
	eoeQZRedTigerOracle      = "94c9aea7-564d-4fc8-ae14-7b22cba18b95"
	eoeQZRemnantOracle       = "69e8333a-bddf-49e2-9c39-3dc46fec81e0"
	eoeQZRigForWarOracle     = "05bbd557-e2c0-4855-8fe6-00527fa39878"
	eoeQZSamisOracle         = "5f532b7f-cb76-4be7-b361-0106c561dcb2"
	eoeQZShatteredOracle     = "709b9f76-0ea5-456c-b1db-2cf517370fde"
	eoeQZWhaleOracle         = "076865e9-ee10-4a87-9180-4274202d37c1"
	eoeQZPilotOracle         = "fd6afb5e-d99a-44a7-9af4-9f064da95d63"
	eoeQZThawbringerOracle   = "0863c4bc-1e3e-4ff9-8877-88d4021c85e5"
	eoeQZVoteOutOracle       = "7efe9fe5-d93c-46a6-afbd-9655a902c6bc"
	eoeQZRammerOracle        = "03259ab0-caa8-4620-b070-127e2c712252"
	eoeQZSweeperOracle       = "320ed2ec-ef06-41fb-ac8a-72579f11544c"
)

func TestEoeQZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeQZRadiantStrikeOracle: "Radiant Strike",
		eoeQZRedTigerOracle:      "Red Tiger Mechan",
		eoeQZRemnantOracle:       "Remnant Elemental",
		eoeQZRigForWarOracle:     "Rig for War",
		eoeQZSamisOracle:         "Sami's Curiosity",
		eoeQZShatteredOracle:     "Shattered Wings",
		eoeQZWhaleOracle:         "Starbreach Whale",
		eoeQZPilotOracle:         "Starfighter Pilot",
		eoeQZThawbringerOracle:   "Thawbringer",
		eoeQZVoteOutOracle:       "Vote Out",
		eoeQZRammerOracle:        "Wedgelight Rammer",
		eoeQZSweeperOracle:       "Wurmwall Sweeper",
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
		eoeQZRedTigerOracle: "{1}{R}",
		eoeQZWhaleOracle:    "{1}{U}",
	} {
		if ac := game.AlternativeCostByKey(oracle, "warp"); ac == nil || ac.ManaCost != cost {
			t.Errorf("%s: warp %+v, want %s", oracle, ac, cost)
		}
	}
}

// --- Radiant Strike ----------------------------------------------------------

func TestEoeQZRadiantStrikeDestroysATappedCreatureAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Radiant Strike", "Instant", eoeQZRadiantStrikeOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the tapped creature is destroyed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestEoeQZRadiantStrikeDestroysAnUntappedArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Radiant Strike", "Instant", eoeQZRadiantStrikeOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
}

func TestEoeQZRadiantStrikeCannotTargetAnUntappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Radiant Strike", "Instant", eoeQZRadiantStrikeOracle, effectstest.CardRefs(bear)); err == nil {
		t.Error("an untapped creature is not a legal target")
	}
}

func TestEoeQZRadiantStrikeGainsNothingWhenItsTargetUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Radiant Strike", "Instant", eoeQZRadiantStrikeOracle, effectstest.CardRefs(bear))
	if err := g.TapCard(bear, false); err != nil {
		t.Fatalf("untap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature untapped in response is no longer a legal target")
	}
	if me.Life != life {
		t.Errorf("the spell did not resolve, so no life: %d -> %d", life, me.Life)
	}
}

// --- Red Tiger Mechan --------------------------------------------------------

func TestEoeQZRedTigerMechanHasHasteAndWarpsOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tiger := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Red Tiger Mechan", TypeLine: "Artifact Creature — Robot Cat", ManaCost: "{3}{R}",
		OracleID: eoeQZRedTigerOracle, Power: 3, Toughness: 3}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tiger) {
		t.Fatal("the warped Mechan never entered")
	}
	if !effectstest.HasEffectiveKeyword(t, g, tiger, "haste") {
		t.Error("Red Tiger Mechan has haste")
	}
	effectstest.AdvanceThroughEndStep(t, g)
	if !g.Exile.Contains(tiger) {
		t.Error("the warped Mechan is exiled at the end step")
	}
}

// --- Remnant Elemental -------------------------------------------------------

func TestEoeQZRemnantElementalGetsPlusTwoZeroOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elemental := effectstest.B12Push(g, me.ID, "Remnant Elemental", "Creature — Elemental", eoeQZRemnantOracle, 0, 4)
	if !effectstest.HasEffectiveKeyword(t, g, elemental, "reach") {
		t.Error("Remnant Elemental has reach")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B31PlayLand(t, g, "Mountain", "Basic Land — Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, elemental); p != 2 || tough != 4 {
		t.Errorf("after a land the Elemental is %d/%d, want 2/4", p, tough)
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	if p, tough := effectstest.EoeCDPT(t, g, elemental); p != 0 || tough != 4 {
		t.Errorf("the pump ends with the turn: %d/%d, want 0/4", p, tough)
	}
}

func TestEoeQZRemnantElementalIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elemental := effectstest.B12Push(g, me.ID, "Remnant Elemental", "Creature — Elemental", eoeQZRemnantOracle, 0, 4)
	effectstest.B13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EoeCDPT(t, g, elemental); p != 0 {
		t.Errorf("an opponent's land pumped it to power %d", p)
	}
}

// --- Rig for War -------------------------------------------------------------

func TestEoeQZRigForWarPumpsAndGrantsReachAndFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rig for War", "Instant", eoeQZRigForWarOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 5/2", p, tough)
	}
	for _, kw := range []string{"reach", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the Bear lacks %s", kw)
		}
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	if p, _ := effectstest.EoeCDPT(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Errorf("the pump and reach end with the turn: power %d", p)
	}
}

// --- Sami's Curiosity --------------------------------------------------------

func TestEoeQZSamisCuriosityGainsTwoAndMakesALander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Sami's Curiosity", "Sorcery", eoeQZSamisOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Fatalf("Landers %d, want 1", got)
	}
	if c := effectstest.LayeredCard(t, g, effectstest.AvB1Named(g, me.ID, "Lander")); !c.IsArtifact() || c.IsCreature() {
		t.Error("the Lander is a noncreature artifact")
	}
}

// --- Shattered Wings ---------------------------------------------------------

func TestEoeQZShatteredWingsDestroysAFlierThenSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Top")
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Shattered Wings", "Sorcery", eoeQZShatteredOracle, effectstest.CardRefs(bird))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) || !opp.Graveyard.Contains(bird) {
		t.Error("the creature with flying is destroyed")
	}
	effectstest.EoeQZSurveil(t, g, me.ID, 1)
}

func TestEoeQZShatteredWingsDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	aura := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Shattered Wings", "Sorcery", eoeQZShatteredOracle, effectstest.CardRefs(aura))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment is destroyed")
	}
}

func TestEoeQZShatteredWingsCannotTargetACreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Shattered Wings", "Sorcery", eoeQZShatteredOracle, effectstest.CardRefs(bear)); err == nil {
		t.Error("a creature without flying is not a legal target")
	}
}

// --- Starbreach Whale --------------------------------------------------------

func TestEoeQZStarbreachWhaleSurveilsTwoAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	whale := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Starbreach Whale", TypeLine: "Creature — Whale", ManaCost: "{4}{U}",
		OracleID: eoeQZWhaleOracle, Power: 3, Toughness: 5}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, whale, "flying") {
		t.Error("Starbreach Whale has flying")
	}
	c := effectstest.EoeQZSurveil(t, g, me.ID, 2)
	effectstest.EoeQZKeepOnTop(t, g, me.ID, c)
	effectstest.AdvanceThroughEndStep(t, g)
	if !g.Exile.Contains(whale) {
		t.Error("the warped Whale is exiled at the end step")
	}
}

// --- Starfighter Pilot -------------------------------------------------------

func TestEoeQZStarfighterPilotSurveilsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	pilot := effectstest.B12Push(g, me.ID, "Starfighter Pilot", "Creature — Human Pilot", eoeQZPilotOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pilot, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoeQZSurveil(t, g, me.ID, 1)
}

func TestEoeQZStarfighterPilotSurveilsWhenTappedForACost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	pilot := effectstest.B12Push(g, me.ID, "Starfighter Pilot", "Creature — Human Pilot", eoeQZPilotOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.TapForTest(t, g, pilot)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoeQZSurveil(t, g, me.ID, 1)
}

// --- Thawbringer -------------------------------------------------------------

func TestEoeQZThawbringerSurveilsAsItEntersAndAsItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Thawbringer", "Creature — Insect Scout", eoeQZThawbringerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoeQZKeepOnTop(t, g, me.ID, effectstest.EoeQZSurveil(t, g, me.ID, 1))
	thaw := effectstest.AvB1Named(g, me.ID, "Thawbringer")
	if thaw == uuid.Nil {
		t.Fatal("Thawbringer is not on the battlefield")
	}
	effectstest.B15Destroy(t, g, thaw)
	effectstest.EoeQZSurveil(t, g, me.ID, 1)
}

// --- Vote Out ----------------------------------------------------------------

func TestEoeQZVoteOutIsConvokedAndDestroysACreature(t *testing.T) {
	g, me, opp := effectstest.TapTable(t)
	victim := effectstest.Pr7Creature(g, opp.ID, "Their Ogre", 4)
	spell := effectstest.HandCardOf(g, me, "Vote Out", "Sorcery", "{3}{B}", eoeQZVoteOutOracle)
	black := effectstest.Pr7Creature(g, me.ID, "Black One", 1, "B")
	w1 := effectstest.Pr7Creature(g, me.ID, "White One", 1, "W")
	w2 := effectstest.Pr7Creature(g, me.ID, "White Two", 1, "W")
	w3 := effectstest.Pr7Creature(g, me.ID, "White Three", 1, "W")
	cast := func(tap ...uuid.UUID) error {
		return g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, TapIDs: tap, Targets: effectstest.CardRefs(victim)})
	}
	if err := cast(w1, w2, w3); err == nil {
		t.Fatal("three creatures and no mana paid {3}{B}")
	}
	if err := cast(black, w1, w2, w3); err != nil {
		t.Fatalf("four creatures, one black, convoke {3}{B}: %v", err)
	}
	for _, id := range []uuid.UUID{black, w1, w2, w3} {
		if !effectstest.TappedForTest(t, g, id) {
			t.Error("a convoking creature was not tapped")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
		t.Error("the target creature is destroyed")
	}
}

// --- Wedgelight Rammer -------------------------------------------------------

func TestEoeQZWedgelightRammerMakesARobotAndFliesWithFirstStrikeAtNine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ship := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Wedgelight Rammer", TypeLine: "Artifact — Spacecraft", ManaCost: "{3}{W}",
		OracleID: eoeQZRammerOracle, Power: 3, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	robot := effectstest.AvB1Named(g, me.ID, "Robot")
	if robot == uuid.Nil {
		t.Fatal("no Robot token")
	}
	if c := effectstest.LayeredCard(t, g, robot); !c.IsArtifact() || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the token is a 2/2 artifact creature, got %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
	if err := g.AddCounter(ship, game.CounterCharge, 8); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Error("below 9 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at9 := effectstest.LayeredCard(t, g, ship)
	if !at9.IsCreature() || at9.CurrentPower() != 3 || at9.CurrentToughness() != 4 {
		t.Errorf("at 9+ it is a 3/4 creature, got %d/%d", at9.CurrentPower(), at9.CurrentToughness())
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !game.HasKeyword(&at9, kw) {
			t.Errorf("at 9+ it has %s", kw)
		}
	}
}

// --- Wurmwall Sweeper --------------------------------------------------------

func TestEoeQZWurmwallSweeperSurveilsTwoAndFliesAtFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	ship := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Wurmwall Sweeper", TypeLine: "Artifact — Spacecraft", ManaCost: "{2}",
		OracleID: eoeQZSweeperOracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoeQZKeepOnTop(t, g, me.ID, effectstest.EoeQZSurveil(t, g, me.ID, 2))
	if err := g.AddCounter(ship, game.CounterCharge, 3); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Error("below 4 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at4 := effectstest.LayeredCard(t, g, ship)
	if !at4.IsCreature() || !game.HasKeyword(&at4, "flying") || at4.CurrentPower() != 2 || at4.CurrentToughness() != 2 {
		t.Errorf("at 4+ it is a 2/2 flying creature, got %d/%d", at4.CurrentPower(), at4.CurrentToughness())
	}
}
