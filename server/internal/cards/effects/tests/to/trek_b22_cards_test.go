package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b22_cards_test.go — Star Trek, batch 22 (#4973). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	trekB22SaruOracle        = "b6a21e53-6f62-4874-b9cc-84ae177a0da6"
	trekB22SaurianOracle     = "00f4cab8-b135-4118-8e3a-92097178abb6"
	trekB22SpeciesOracle     = "645a5619-d0b2-4153-aae0-ca2adc6f5e3d"
	trekB22WormholeOracle    = "dba55309-cb21-42db-8968-8da07227a1be"
	trekB22MissionOracle     = "59a4c4cf-07b4-4009-bbea-7da7b31f09a1"
	trekB22TillyOracle       = "d4c1dbc8-23d8-49e0-bc2d-d005e9215172"
	trekB22LiquidatorsOracle = "2f32de5c-4d85-42bc-b1bf-f762a46e5efc"
	trekB22ConquerorOracle   = "306002e6-d68d-47ad-879b-e6060efebeff"
	trekB22CourageOracle     = "101b2387-58a1-4ef9-8395-0024713e217b"
	trekB22DiscoveryOracle   = "13dbfe3f-1e04-4c4f-953c-7c949dba9d33"
)

func TestTrekB22EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB22SaruOracle:        "Saru, Unthreatened",
		trekB22SaurianOracle:     "Saurian Explorer",
		trekB22SpeciesOracle:     "Species 8472",
		trekB22WormholeOracle:    "Stable Wormhole",
		trekB22MissionOracle:     "Support Mission",
		trekB22TillyOracle:       "Sylvia Tilly, Joy of Discovery",
		trekB22LiquidatorsOracle: "Syndicate Liquidators",
		trekB22ConquerorOracle:   "Temporal Conqueror",
		trekB22CourageOracle:     "Test of Courage",
		trekB22DiscoveryOracle:   "U.S.S. Discovery, at Black Alert",
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
}

// trekB22Enter puts a plain creature onto the battlefield under `owner`
// from their hand, so it ENTERS (and fires enters triggers).
func trekB22Enter(t *testing.T, g *game.Game, owner *game.Player, name string) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.Power, c.Toughness = "Creature — Bear", 2, 2
	owner.Hand.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	return id
}

// trekB22ToGraveyard puts a plain card in p's graveyard.
func trekB22ToGraveyard(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// trekB22NewPlanets is the New Planet tokens `controller` controls,
// through the layered view.
func trekB22NewPlanets(t *testing.T, g *game.Game, controller uuid.UUID) []game.Card {
	t.Helper()
	return trekB10Tokens(t, g, controller, "New Planet")
}

// trekB22CheckNewPlanet asserts the printed New Planet: a tapped,
// colourless Planet land with a mana ability.
func trekB22CheckNewPlanet(t *testing.T, g *game.Game, p game.Card) {
	t.Helper()
	if !p.Tapped {
		t.Error("the New Planet entered untapped")
	}
	if !p.IsLand() || !p.HasSubtype("Planet") || p.IsCreature() || len(p.Colors) != 0 {
		t.Errorf("the New Planet is %q colours %v", p.TypeLine, p.Colors)
	}
	if len(game.ManaAbilitiesForCard(p)) != 1 {
		t.Errorf("the New Planet has %d mana abilities, want 1", len(game.ManaAbilitiesForCard(p)))
	}
}

// --- Saru, Unthreatened --------------------------------------------------------

func TestTrekB22SaruGivesAnotherCreatureYouControlHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	saru := trekB10Cast(t, g, "Saru, Unthreatened", "Legendary Creature — Kelpien Officer", "{2}{G}", trekB22SaruOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for Saru's trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, saru) || effectstest.HasID(pick.PickTargetCards, theirs) || !effectstest.HasID(pick.PickTargetCards, mine) {
		t.Fatalf("want only another creature I control: %v", pick.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, mine); !game.HasKeyword(&c, "hexproof") {
		t.Error("the target did not gain hexproof")
	}
	if c := effectstest.LayeredCard(t, g, saru); game.HasKeyword(&c, "hexproof") {
		t.Error("Saru gained hexproof himself")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.PowerUpPassTo(t, g, 1)
	if c := effectstest.LayeredCard(t, g, mine); game.HasKeyword(&c, "hexproof") {
		t.Error("hexproof outlasted the turn")
	}
}

func TestTrekB22SaruHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spec, _ := effects.Lookup(trekB22SaruOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Fatalf("Saru's printed keywords %v, want flash", spec.PrintedKeywords)
	}
	saru := effectstest.B27Push(g, g.Seats[0].ID, "Saru, Unthreatened", "Legendary Creature — Kelpien Officer", trekB22SaruOracle, "{2}{G}", 3, 2, "G")
	if c := effectstest.LayeredCard(t, g, saru); !game.HasKeyword(&c, "flash") {
		t.Error("Saru on the battlefield reads no flash")
	}
}

// --- Saurian Explorer ----------------------------------------------------------

func TestTrekB22SaurianExplorerGrowsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dino := effectstest.B27Push(g, me.ID, "Saurian Explorer", "Creature — Dinosaur Scout", trekB22SaurianOracle, "{3}{G}", 4, 3, "G")
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, dino); c.CurrentPower() != 6 || c.CurrentToughness() != 5 {
		t.Errorf("after a land entered: %d/%d, want 6/5", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.B12Permanent(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest")
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.PowerUpPassTo(t, g, 1)
	if c := effectstest.LayeredCard(t, g, dino); c.CurrentPower() != 4 || c.CurrentToughness() != 3 {
		t.Errorf("next turn: %d/%d, want 4/3", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestTrekB22SaurianExplorerMakesANewPlanetAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dino := effectstest.B27Push(g, me.ID, "Saurian Explorer", "Creature — Dinosaur Scout", trekB22SaurianOracle, "{3}{G}", 4, 3, "G")
	effectstest.AdvanceToMain(t, g)

	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, dino, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for five mana")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, dino, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	planets := trekB22NewPlanets(t, g, me.ID)
	if len(planets) != 1 {
		t.Fatalf("%d New Planets, want 1", len(planets))
	}
	trekB22CheckNewPlanet(t, g, planets[0])

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, dino, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Species 8472 --------------------------------------------------------------

func TestTrekB22SpeciesGainsLifePerDifferentlyNamedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Mystic Gate", "Land")
	effectstest.B12Permanent(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp")
	life := me.Life
	trekB10Cast(t, g, "Species 8472", "Creature — Insect Beast", "{4}{G}{G}", trekB22SpeciesOracle, 7, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life; got != life+3 {
		t.Errorf("life %d, want %d (three land names)", got, life+3)
	}
}

func TestTrekB22SpeciesCantBeBlockedByArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	species := effectstest.B27Push(g, me.ID, "Species 8472", "Creature — Insect Beast", trekB22SpeciesOracle, "{4}{G}{G}", 7, 5, "G")
	golem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(species, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(golem, species); err == nil {
		t.Error("an artifact creature blocked it")
	}
	if err := g.DeclareBlocker(bear, species); err != nil {
		t.Errorf("a non-artifact creature could not block it: %v", err)
	}
}

func TestTrekB22SpeciesRefusesOpponentsArtifactAbilitiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	species := effectstest.B27Push(g, me.ID, "Species 8472", "Creature — Insect Beast", trekB22SpeciesOracle, "{4}{G}{G}", 7, 5, "G")
	theirBallista := effectstest.B27Push(g, opp.ID, "Walking Ballista", "Artifact Creature — Construct", effectstest.WalkingBallistaOracle, "{X}{X}", 0, 0)
	myBallista := effectstest.B27Push(g, me.ID, "Walking Ballista", "Artifact Creature — Construct", effectstest.WalkingBallistaOracle, "{X}{X}", 0, 0)
	for _, id := range []uuid.UUID{theirBallista, myBallista} {
		effectstest.SetCountersOn(g, id, map[string]int{game.CounterPlusOne: 2})
	}
	effectstest.AdvanceToMain(t, g)

	// An opponent's artifact's ability can't target it…
	if err := g.ActivateCatalogAbility(opp.ID, theirBallista, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(species), Strict: true}); err == nil {
		t.Error("an opponent's artifact ability targeted Species 8472")
	}
	// …but can target another creature, and its controller.
	other := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(opp.ID, theirBallista, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other), Strict: true}); err != nil {
		t.Errorf("the Ballista could not target another creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// My own artifact's ability can.
	if err := g.ActivateCatalogAbility(me.ID, myBallista, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(species), Strict: true}); err != nil {
		t.Errorf("my own artifact ability could not target it: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, species); c.DamageMarked != 1 {
		t.Errorf("damage on Species 8472 %d, want 1", c.DamageMarked)
	}
}

func TestTrekB22SpeciesAdmitsArtifactSpellsAndNonArtifactAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	species := effectstest.B27Push(g, me.ID, "Species 8472", "Creature — Insect Beast", trekB22SpeciesOracle, "{4}{G}{G}", 7, 5, "G")
	spec, _ := effects.Lookup(trekB22SpeciesOracle)
	if len(spec.TargetingRestrictions) != 1 {
		t.Fatalf("%d targeting restrictions, want 1", len(spec.TargetingRestrictions))
	}
	r := spec.TargetingRestrictions[0]
	self := effectstest.LayeredCard(t, g, species)
	artifact := game.Card{InstanceID: uuid.New(), Name: "Their Relic", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID}
	bear := game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID}
	q := func(by game.TargetSource) game.TargetingQuery {
		return game.TargetingQuery{Game: g, Card: self, Zone: game.ZoneBattlefield, Controller: by.Controller, Source: self, By: by}
	}
	if !r.Forbids(q(game.SourceObject(opp.ID, &artifact))) {
		t.Error("an opponent's artifact ability is not refused")
	}
	if r.Forbids(q(game.SourceCast(opp.ID, &artifact))) {
		t.Error("an opponent's artifact SPELL is refused")
	}
	if r.Forbids(q(game.SourceObject(opp.ID, &bear))) {
		t.Error("an opponent's non-artifact ability is refused")
	}
	if r.Forbids(q(game.SourceObject(me.ID, &artifact))) {
		t.Error("my own artifact ability is refused")
	}
	other := effectstest.LayeredCard(t, g, effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2))
	if r.Forbids(game.TargetingQuery{Game: g, Card: other, Zone: game.ZoneBattlefield, Controller: opp.ID, Source: self, By: game.SourceObject(opp.ID, &artifact)}) {
		t.Error("the restriction covers another creature")
	}
}

// --- Stable Wormhole -----------------------------------------------------------

func TestTrekB22StableWormholeMakesOneNewPlanetPerAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Stable Wormhole", "Enchantment", trekB22WormholeOracle, "{2}{G}", 0, 0, "G")
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	planets := trekB22NewPlanets(t, g, me.ID)
	if len(planets) != 1 {
		t.Fatalf("%d New Planets after one attack with two creatures, want 1", len(planets))
	}
	trekB22CheckNewPlanet(t, g, planets[0])
}

func TestTrekB22StableWormholeIgnoresAnOpponentsAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Stable Wormhole", "Enchantment", trekB22WormholeOracle, "{2}{G}", 0, 0, "G")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.PowerUpPassTo(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, theirs)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(trekB22NewPlanets(t, g, me.ID)) + len(trekB22NewPlanets(t, g, opp.ID)); n != 0 {
		t.Errorf("%d New Planets after an opponent attacked", n)
	}
}

// --- Support Mission -----------------------------------------------------------

func TestTrekB22SupportMissionCountsQuestsAndGrowsEachArrival(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mission := effectstest.B27Push(g, me.ID, "Support Mission", "Enchantment", trekB22MissionOracle, "{2}{G}", 0, 0, "G")
	effectstest.AdvanceToMain(t, g)

	first := trekB22Enter(t, g, me, "First Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mission, "quest"); got != 1 {
		t.Errorf("quest counters %d, want 1", got)
	}
	if got := effectstest.CounterOn(g, first, game.CounterPlusOne); got != 1 {
		t.Errorf("the first creature has %d +1/+1 counters, want 1", got)
	}

	trekB22Enter(t, g, opp, "Their Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mission, "quest"); got != 1 {
		t.Errorf("an opponent's creature added a quest counter: %d", got)
	}

	effectstest.SetCountersOn(g, mission, map[string]int{"quest": 4})
	fifth := trekB22Enter(t, g, me, "Fifth Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mission, "quest"); got != 5 {
		t.Errorf("quest counters %d, want 5", got)
	}
	if got := effectstest.CounterOn(g, fifth, game.CounterPlusOne); got != 2 {
		t.Errorf("the creature entering with the fifth quest counter has %d +1/+1 counters, want 2", got)
	}
}

// --- Sylvia Tilly, Joy of Discovery ----------------------------------------------

func TestTrekB22SylviaTillySurveilsTwoOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top", "Second", "Third")
	trekB10Cast(t, g, "Sylvia Tilly, Joy of Discovery", "Legendary Creature — Human Scientist Artificer", "{2}{G}", trekB22TillyOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt: %+v", c)
	}
	gy := me.Graveyard.Size()
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := me.Graveyard.Size(); got != gy+2 {
		t.Errorf("graveyard %d, want %d", got, gy+2)
	}
}

func TestTrekB22SylviaTillyPaysAndExilesForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tilly := effectstest.B27Push(g, me.ID, "Sylvia Tilly, Joy of Discovery", "Legendary Creature — Human Scientist Artificer", trekB22TillyOracle, "{2}{G}", 2, 2, "G")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	dead := trekB22ToGraveyard(g, me, "Dead Card", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if p := effectstest.AnswerPayUnless(t, g, me.ID, true); p == nil {
		t.Fatal("no pay {1} offer at the beginning of combat")
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, dead) {
		t.Fatalf("want a choice of a graveyard card: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dead)
	if !effectstest.InExile(g, dead) {
		t.Fatal("the card was not exiled")
	}
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CounterOn(g, tilly, game.CounterPlusOne); got != 0 {
		t.Errorf("Tilly has %d counters", got)
	}
}

func TestTrekB22SylviaTillyDecliningOrAnEmptyGraveyardDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Sylvia Tilly, Joy of Discovery", "Legendary Creature — Human Scientist Artificer", trekB22TillyOracle, "{2}{G}", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID) != nil {
		t.Fatal("offered the payment with an empty graveyard")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.B27Push(g2, me2.ID, "Sylvia Tilly, Joy of Discovery", "Legendary Creature — Human Scientist Artificer", trekB22TillyOracle, "{2}{G}", 2, 2, "G")
	dead := trekB22ToGraveyard(g2, me2, "Dead Card", "Sorcery")
	effectstest.AdvanceToMain(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	if p := effectstest.AnswerPayUnless(t, g2, me2.ID, false); p == nil {
		t.Fatal("no offer")
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !me2.Graveyard.Contains(dead) || effectstest.LatestPickTarget(g2, me2.ID) != nil {
		t.Error("declining still exiled or asked for a target")
	}
}

// --- Syndicate Liquidators -------------------------------------------------------

func TestTrekB22LiquidatorsPutTwoCountersOnACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liq := effectstest.B27Push(g, me.ID, "Syndicate Liquidators", "Creature — Orion Pirate", trekB22LiquidatorsOracle, "{4}{G}", 4, 4, "G")
	if c := effectstest.LayeredCard(t, g, liq); !game.HasKeyword(&c, "trample") {
		t.Error("no trample")
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no landfall target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, theirs) || !effectstest.HasID(pick.PickTargetCards, mine) || !effectstest.HasID(pick.PickTargetCards, liq) {
		t.Fatalf("want creatures I control: %v", pick.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

// --- Temporal Conqueror ----------------------------------------------------------

func TestTrekB22TemporalConquerorReturnsAMilledLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "Spell A", "Spell B", "Spell C", "Deep Card")
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID == ids[1] {
				me.Library.Cards[i].TypeLine = "Land"
			}
		}
	})
	treasures := len(trekB10Tokens(t, g, me.ID, "Treasure"))
	trekB10Cast(t, g, "Temporal Conqueror", "Creature — Na'kuhl Warrior", "{2}{G}", trekB22ConquerorOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[1] || pick.ChooseMin != 1 {
		t.Fatalf("want a forced choice of the milled land: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) {
		t.Error("the land did not return to hand")
	}
	if !me.Graveyard.Contains(ids[0]) || !me.Graveyard.Contains(ids[2]) || !me.Library.Contains(ids[3]) {
		t.Error("the mill was not three cards")
	}
	if got := len(trekB10Tokens(t, g, me.ID, "Treasure")); got != treasures {
		t.Error("made a Treasure though a land returned")
	}
}

func TestTrekB22TemporalConquerorMakesATreasureWithNoLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Spell A", "Spell B", "Spell C")
	trekB10Cast(t, g, "Temporal Conqueror", "Creature — Na'kuhl Warrior", "{2}{G}", trekB22ConquerorOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("asked for a land with none in the graveyard")
	}
	if got := len(trekB10Tokens(t, g, me.ID, "Treasure")); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Test of Courage -------------------------------------------------------------

func trekB22CastCourage(t *testing.T, g *game.Game, mode int, target uuid.UUID) error {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test of Courage", TypeLine: "Instant", ManaCost: "{G}",
			OracleID: trekB22CourageOracle, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatMana(t, g, me, "{G}")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: effectstest.CardRefs(target), Strict: true})
}

func TestTrekB22TestOfCourageExilesAGraveyardCardAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := trekB22ToGraveyard(g, opp, "Their Dead", "Creature — Bear")
	life := me.Life
	if err := trekB22CastCourage(t, g, 0, dead); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, dead) {
		t.Error("the card was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestTrekB22TestOfCourageDestroysACheapArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	cheap := effectstest.B27Push(g, opp.ID, "Their Relic", "Artifact", "", "{3}", 0, 0)
	pricey := effectstest.B27Push(g, opp.ID, "Their Monolith", "Artifact", "", "{4}", 0, 0)
	aura := effectstest.B27Push(g, opp.ID, "Their Shrine", "Enchantment", "", "{1}{W}", 0, 0)
	bear := effectstest.B27Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}", 2, 2)
	if err := trekB22CastCourage(t, g, 1, pricey); err == nil {
		t.Error("targeted a mana value 4 artifact")
	}
	if err := trekB22CastCourage(t, g, 1, bear); err == nil {
		t.Error("targeted a creature")
	}
	if err := trekB22CastCourage(t, g, 1, cheap); err != nil {
		t.Fatalf("cast at the artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the artifact survived")
	}
	if err := trekB22CastCourage(t, g, 1, aura); err != nil {
		t.Fatalf("cast at the enchantment: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment survived")
	}
}

// --- U.S.S. Discovery, at Black Alert --------------------------------------------

func TestTrekB22DiscoveryDiscoversFiveOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hit := effectstest.DiscoverLibrary(me, 5)
	trekB10Cast(t, g, "U.S.S. Discovery, at Black Alert", "Legendary Artifact — Spacecraft", "{3}{G}{G}", trekB22DiscoveryOracle, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestTrekB22DiscoveryIsAFlyingCreatureAtSixAndDiscoversOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := effectstest.B27Push(g, me.ID, "U.S.S. Discovery, at Black Alert", "Legendary Artifact — Spacecraft", trekB22DiscoveryOracle, "{3}{G}{G}", 5, 5, "G")
	effectstest.AddCountersForTest(g, ship, game.CounterCharge, 5)
	if c := effectstest.LayeredCard(t, g, ship); c.IsCreature() || game.HasKeyword(&c, "flying") {
		t.Error("a creature or flying at five charge counters")
	}
	effectstest.AddCountersForTest(g, ship, game.CounterCharge, 1)
	c := effectstest.LayeredCard(t, g, ship)
	if !c.IsCreature() || !game.HasKeyword(&c, "flying") || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Fatalf("at six: creature %v flying %v %d/%d", c.IsCreature(), game.HasKeyword(&c, "flying"), c.CurrentPower(), c.CurrentToughness())
	}
	hit := effectstest.DiscoverLibrary(me, 5)
	effectstest.DeclareAttack(t, g, opp.ID, ship)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}
