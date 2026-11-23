package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_s_cards_test.go — Edge of Eternities cards whose names start with S
// (#4461), the ones the Q–Z batch left. Each test plays the real catalog
// entry and asserts what its text says, and what it refuses.

const (
	eoeSCryologistOracle   = "be3038ff-b39e-4d2e-97ee-85de074ccfd3"
	eoeSSkystingerOracle   = "6bb7419b-e085-4616-8e5c-4ed08cb2b4c1"
	eoeSAgrarianOracle     = "003d9c17-5d5c-469f-ad87-d78faf55e025"
	eoeSImpactOracle       = "c6d89c4a-351f-4d63-ba57-b978a695191e"
	eoeSExpansionistOracle = "944a91ab-c217-4c1d-b920-2e5503c34546"
	eoeSLightsmithOracle   = "132de6c2-4ca2-41da-9c78-caa59838b641"
	eoeSVoidbornOracle     = "4f754f11-3b6d-4a4e-86d7-7495e475ef74"
	eoeSDirgecraftOracle   = "f1d826d1-d2da-46ce-b1db-81393c88d865"
	eoeSCullerOracle       = "a6c28bde-10d9-471b-acd0-32a1c6195805"
	eoeSStarwinderOracle   = "face2dc0-37ba-4efe-b900-b5521567c943"
	eoeSVondamOracle       = "28225bae-2d96-4352-b256-0c2ff4ef65e9"
	eoeSAnomalyOracle      = "241117a0-8305-48bb-9cb9-9b01e86e08f7"
)

func TestEoeSEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeSCryologistOracle:   "Sinister Cryologist",
		eoeSSkystingerOracle:   "Skystinger",
		eoeSAgrarianOracle:     "Seedship Agrarian",
		eoeSImpactOracle:       "Seedship Impact",
		eoeSExpansionistOracle: "Sunstar Expansionist",
		eoeSLightsmithOracle:   "Sunstar Lightsmith",
		eoeSVoidbornOracle:     "Susurian Voidborn",
		eoeSDirgecraftOracle:   "Susurian Dirgecraft",
		eoeSCullerOracle:       "Swarm Culler",
		eoeSStarwinderOracle:   "Starwinder",
		eoeSVondamOracle:       "Syr Vondam, the Lucent",
		eoeSAnomalyOracle:      "Space-Time Anomaly",
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
		eoeSCryologistOracle: "{U}",
		eoeSVoidbornOracle:   "{B}",
		eoeSStarwinderOracle: "{2}{U}{U}",
	} {
		if ac := game.AlternativeCostByKey(oracle, "warp"); ac == nil || ac.ManaCost != cost {
			t.Errorf("%s: warp %+v, want %s", oracle, ac, cost)
		}
	}
}

// --- Sinister Cryologist ----------------------------------------------------

func TestEoeSSinisterCryologistShrinksAnOpponentsCreatureAndWarpsOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	cryo := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Sinister Cryologist", TypeLine: "Creature — Jellyfish Wizard", ManaCost: "{2}{U}",
		OracleID: eoeSCryologistOracle, Power: 2, Toughness: 3}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("the controller's own creature is offered")
	}
	effectstest.PickIfAsked(t, g, me.ID, giant)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, giant); p != 1 || tough != 4 {
		t.Errorf("the target is %d/%d, want 1/4", p, tough)
	}
	if p, _ := effectstest.EoeCDPT(t, g, mine); p != 2 {
		t.Errorf("my creature's power %d, want 2", p)
	}
	effectstest.AdvanceThroughEndStep(t, g)
	if !g.Exile.Contains(cryo) {
		t.Error("the warped Cryologist is exiled at the end step")
	}
}

// --- Skystinger ---------------------------------------------------------------

func TestEoeSSkystingerGetsPlusFiveWhenItBlocksAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sky := effectstest.B12Push(g, opp.ID, "Skystinger", "Creature — Insect Warrior", eoeSSkystingerOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, sky, "reach") {
		t.Error("Skystinger has reach")
	}
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AttackIntoBlocks(t, g, opp.ID, bird)
	if err := g.DeclareBlocker(sky, bird); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, sky); p != 8 || tough != 3 {
		t.Errorf("blocking a flier, Skystinger is %d/%d, want 8/3", p, tough)
	}
}

func TestEoeSSkystingerIgnoresACreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sky := effectstest.B12Push(g, opp.ID, "Skystinger", "Creature — Insect Warrior", eoeSSkystingerOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AttackIntoBlocks(t, g, opp.ID, bear)
	if err := g.DeclareBlocker(sky, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EoeCDPT(t, g, sky); p != 3 {
		t.Errorf("blocking a creature without flying pumped it to power %d", p)
	}
}

// --- Seedship Agrarian --------------------------------------------------------

func TestEoeSSeedshipAgrarianMakesALanderWhenTappedAndGrowsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	agrarian := effectstest.B12Push(g, me.ID, "Seedship Agrarian", "Creature — Insect Scientist", eoeSAgrarianOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.TapForTest(t, g, agrarian)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Fatalf("Landers %d, want 1", got)
	}
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, agrarian, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
}

func TestEoeSSeedshipAgrarianIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	agrarian := effectstest.B12Push(g, me.ID, "Seedship Agrarian", "Creature — Insect Scientist", eoeSAgrarianOracle, 3, 3)
	effectstest.B13PlayAs(t, g, 1, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, agrarian, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's land put %d counters on it", got)
	}
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 0 {
		t.Errorf("an untapped Agrarian made %d Landers", got)
	}
}

// --- Seedship Impact ----------------------------------------------------------

func TestEoeSSeedshipImpactDestroysACheapArtifactAndMakesALander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Seedship Impact", "Instant", eoeSImpactOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("Landers %d, want 1 for a mana value 2 target", got)
	}
}

func TestEoeSSeedshipImpactMakesNoLanderForAnExpensiveEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Enchantment", TypeLine: "Enchantment",
		ManaCost: "{2}{W}", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Seedship Impact", "Instant", eoeSImpactOracle, effectstest.CardRefs(aura))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment is destroyed")
	}
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 0 {
		t.Errorf("Landers %d, want 0 for a mana value 3 target", got)
	}
}

func TestEoeSSeedshipImpactCannotTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Seedship Impact", "Instant", eoeSImpactOracle, effectstest.CardRefs(bear)); err == nil {
		t.Error("a creature is not a legal target")
	}
}

// --- Sunstar Expansionist -----------------------------------------------------

func TestEoeSSunstarExpansionistMakesALanderWhenBehindOnLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	exp := effectstest.CastCatalogSpell(t, g, "Sunstar Expansionist", "Creature — Human Knight", eoeSExpansionistOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(exp) {
		t.Fatal("the Expansionist did not resolve")
	}
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("Landers %d, want 1", got)
	}
}

func TestEoeSSunstarExpansionistMakesNoLanderWhenEven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.CastCatalogSpell(t, g, "Sunstar Expansionist", "Creature — Human Knight", eoeSExpansionistOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EoeCDCount(g, me.ID, "Lander"); got != 0 {
		t.Errorf("Landers %d, want 0", got)
	}
}

func TestEoeSSunstarExpansionistGetsPlusOneZeroOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	exp := effectstest.B12Push(g, me.ID, "Sunstar Expansionist", "Creature — Human Knight", eoeSExpansionistOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.B31PlayLand(t, g, "Plains", "Basic Land — Plains", "")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, exp); p != 3 || tough != 3 {
		t.Errorf("after a land the Expansionist is %d/%d, want 3/3", p, tough)
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	if p, _ := effectstest.EoeCDPT(t, g, exp); p != 2 {
		t.Errorf("the pump ends with the turn: power %d, want 2", p)
	}
}

// --- Sunstar Lightsmith -------------------------------------------------------

func TestEoeSSunstarLightsmithGrowsAndDrawsOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	smith := effectstest.B12Push(g, me.ID, "Sunstar Lightsmith", "Creature — Human Artificer", eoeSLightsmithOracle, 3, 3)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Sami's Curiosity", "Sorcery", eoeQZSamisOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, smith, game.CounterPlusOne); got != 0 || me.Hand.Size() != hand {
		t.Fatalf("the first spell: %d counters, hand %d -> %d", got, hand, me.Hand.Size())
	}
	effectstest.CastCatalogSpell(t, g, "Sami's Curiosity", "Sorcery", eoeQZSamisOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, smith, game.CounterPlusOne); got != 1 {
		t.Errorf("the second spell: %d +1/+1 counters, want 1", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the second spell: hand %d, want %d", got, hand+1)
	}
	effectstest.CastCatalogSpell(t, g, "Sami's Curiosity", "Sorcery", eoeQZSamisOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, smith, game.CounterPlusOne); got != 1 {
		t.Errorf("the third spell triggered it too: %d counters", got)
	}
}

// --- Susurian Voidborn --------------------------------------------------------

func TestEoeSSusurianVoidbornDrainsWhenAnArtifactOrItselfDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	voidborn := effectstest.B12Push(g, me.ID, "Susurian Voidborn", "Creature — Vampire Soldier", eoeSVoidbornOracle, 2, 2)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, dying := range []uuid.UUID{theirs, rock, voidborn} {
		myLife, oppLife := me.Life, opp.Life
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dying) })
		effectstest.PassPriorityAroundTable(t, g)
		if dying == theirs {
			if effectstest.LatestPickTarget(g, me.ID) != nil {
				t.Fatal("an opponent's creature dying triggered the Voidborn")
			}
			continue
		}
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != myLife+1 || opp.Life != oppLife-1 {
			t.Errorf("after a death: me %d -> %d, opponent %d -> %d", myLife, me.Life, oppLife, opp.Life)
		}
	}
}

// --- Susurian Dirgecraft ------------------------------------------------------

func TestEoeSSusurianDirgecraftEdictsNontokenCreaturesAndFliesAtSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	ship := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Susurian Dirgecraft", TypeLine: "Artifact — Spacecraft", ManaCost: "{4}{B}",
		OracleID: eoeSDirgecraftOracle, Power: 4, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the controller is not asked to sacrifice")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent's only nontoken creature is sacrificed")
	}
	if !g.Battlefield.Contains(goblin) {
		t.Error("a token is not a nontoken creature")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the controller's own creature stays")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 6); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Error("below 7 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at7 := effectstest.LayeredCard(t, g, ship)
	if !at7.IsCreature() || !game.HasKeyword(&at7, "flying") || at7.CurrentPower() != 4 || at7.CurrentToughness() != 3 {
		t.Errorf("at 7+ it is a 4/3 flying creature, got %d/%d", at7.CurrentPower(), at7.CurrentToughness())
	}
}

// --- Swarm Culler -------------------------------------------------------------

func TestEoeSSwarmCullerSacrificesToDrawWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	culler := effectstest.B12Push(g, me.ID, "Swarm Culler", "Creature — Insect Warrior", eoeSCullerOracle, 2, 4)
	if !effectstest.HasEffectiveKeyword(t, g, culler, "flying") {
		t.Error("Swarm Culler has flying")
	}
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.AdvanceToMain(t, g)
	effectstest.TapForTest(t, g, culler)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, relic) || !effectstest.HasID(prompt.ChooseCards, bear) ||
		effectstest.HasID(prompt.ChooseCards, culler) || effectstest.HasID(prompt.ChooseCards, land) {
		t.Fatalf("want a may-sacrifice over another creature or artifact: %+v", prompt)
	}
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestEoeSSwarmCullerDrawsNothingWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	culler := effectstest.B12Push(g, me.ID, "Swarm Culler", "Creature — Insect Warrior", eoeSCullerOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.TapForTest(t, g, culler)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("declining sacrificed the Bear")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("declining drew: hand %d, want %d", got, hand)
	}
}

// --- Starwinder ---------------------------------------------------------------

func TestEoeSStarwinderDrawsThatManyWhenACreatureConnects(t *testing.T) {
	for _, tc := range []struct {
		name   string
		accept bool
		drawn  int
	}{{"accept", true, 3}, {"decline", false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			effectstest.SeedLibrary(me, "A", "B", "C", "D")
			effectstest.B12Push(g, me.ID, "Starwinder", "Creature — Leviathan", eoeSStarwinderOracle, 7, 7)
			ogre := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
			effectstest.AttackIntoBlocks(t, g, them.ID, ogre)
			hand := me.Hand.Size()
			effectstest.AdvanceToStepInTurn(t, g, game.StepCombatDamage)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, tc.accept)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.drawn {
				t.Errorf("drew %d cards, want %d", got, tc.drawn)
			}
		})
	}
}

// --- Syr Vondam, the Lucent ---------------------------------------------------

func TestEoeSSyrVondamPumpsTheOthersAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	vondam := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Syr Vondam, the Lucent", TypeLine: "Legendary Creature — Human Knight", ManaCost: "{2}{W}{B}{B}",
		OracleID: eoeSVondamOracle, Power: 4, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 3 || tough != 2 || !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Errorf("my Bear is %d/%d, want 3/2 with deathtouch", p, tough)
	}
	if p, _ := effectstest.EoeCDPT(t, g, vondam); p != 4 {
		t.Errorf("Syr Vondam pumped itself to power %d", p)
	}
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, vondam, kw) {
			t.Errorf("Syr Vondam has %s", kw)
		}
	}
	if p, _ := effectstest.EoeCDPT(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Errorf("an opponent's creature was pumped: power %d", p)
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	if p, _ := effectstest.EoeCDPT(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Errorf("the pump ends with the turn: power %d", p)
	}
}

func TestEoeSSyrVondamPumpsTheOthersWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vondam := effectstest.B12Push(g, me.ID, "Syr Vondam, the Lucent", "Legendary Creature — Human Knight", eoeSVondamOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, vondam)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EoeCDPT(t, g, bear); p != 3 || !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Errorf("my Bear has power %d, want 3 with deathtouch", p)
	}
}

// --- Space-Time Anomaly -------------------------------------------------------

func TestEoeSSpaceTimeAnomalyMillsYourLifeTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E", "F")
	me.Life = 4
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Space-Time Anomaly", "Sorcery", eoeSAnomalyOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 4 {
		t.Errorf("milled %d cards, want 4", got)
	}
	if got := opp.Graveyard.Size() - graveyard; got != 4 {
		t.Errorf("graveyard grew by %d, want 4", got)
	}
}

func TestEoeSSpaceTimeAnomalyReadsTheLifeTotalAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E", "F")
	me.Life = 5
	library := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Space-Time Anomaly", "Sorcery", eoeSAnomalyOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	me.Life = 2
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d cards, want 2 (the life total as it resolves)", got)
	}
}
