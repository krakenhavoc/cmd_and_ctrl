package testsms

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a7_cards_test.go — Marvel Super Heroes, half A, seventh batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA7RestorativeOracle = "f7b2606c-42b8-46d6-bcaa-8d4506b348eb"
	mshA7RickJonesOracle   = "4bccfa67-0c6f-42a1-b8d4-0842a8230d1e"
	mshA7DroneOracle       = "c1669020-8f81-471a-9083-7e9ce3e8f387"
	mshA7HelicarrierOracle = "0f2335bf-6a13-4483-997a-fd19d30c4d48"
	mshA7DinosaurOracle    = "6e21ebdf-b97c-494e-a1d5-19db1b7d3016"
	mshA7ShangChiOracle    = "700ef9c8-698b-45f7-a1fd-dec74e553224"
	mshA7ShuriOracle       = "a8c047fc-20a9-4c62-b8a6-9ead56d60ad0"
	mshA7StatureOracle     = "bd49a05c-ed51-4d69-ab9b-11192442ec17"
	mshA7SuperSpeedOracle  = "88923ca1-a793-42f0-b9f8-ed9ff9c1185d"
	mshA7SuperSuitOracle   = "606092eb-8fe3-4ef7-aaf1-2d3b2687d27a"
)

func TestMshA7EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA7RestorativeOracle: "Restorative Technique",
		mshA7RickJonesOracle:   "Rick Jones, Destined Sidekick",
		mshA7DroneOracle:       "S.H.I.E.L.D. Deployment Drone",
		mshA7HelicarrierOracle: "S.H.I.E.L.D. Helicarrier",
		mshA7DinosaurOracle:    "Savage Land Dinosaur",
		mshA7ShangChiOracle:    "Shang-Chi, Master of Kung Fu",
		mshA7ShuriOracle:       "Shuri, Wakandan Inventor",
		mshA7StatureOracle:     "Stature, Size Shifter",
		mshA7SuperSpeedOracle:  "Super Speed",
		mshA7SuperSuitOracle:   "Super Suit",
	}
	for oracle, name := range want {
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

// --- Restorative Technique -------------------------------------------------------------------

func TestMshA7RestorativeTechniqueGainsLifeFetchesALandForTheTargetAndGrowsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ids := effectstest.SeedSearchLibrary(opp,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Restorative Technique", "Sorcery", mshA7RestorativeOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID},
		{Kind: game.TargetCard, ID: bear},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+2 {
		t.Errorf("the targeted player's life is %d, want %d", opp.Life, life+2)
	}
	c := effectstest.SearchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the targeted player was not asked to search; the search is theirs, not the caster's")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the caster was asked to search")
	}
	if err := g.ResolveSearchLibrary(c.ID, opp.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	forest, ok := effectstest.BattlefieldCard(g, ids[0])
	if !ok || forest.Controller != opp.ID || !forest.Tapped {
		t.Errorf("the land: on battlefield %v, %+v; want it under %s, tapped", ok, forest, opp.ID)
	}
	if n := effectstest.PlusOnes(g, bear); n != 1 {
		t.Errorf("the creature has %d +1/+1 counters, want 1", n)
	}
}

func TestMshA7RestorativeTechniqueTheCreatureIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Restorative Technique", "Sorcery", mshA7RestorativeOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life is %d, want %d with no creature chosen", me.Life, life+2)
	}
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Error("no search was offered with the up-to-one creature left empty")
	}
}

// --- Rick Jones, Destined Sidekick -----------------------------------------------------------

func mshA7RickJonesMill(t *testing.T, g *game.Game) (me *game.Player, hero, enchant, bear, land uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	rick := effectstest.PushCatalogPermanent(g, me.ID, "Rick Jones, Destined Sidekick", "Legendary Creature — Human Advisor", mshA7RickJonesOracle, false)
	effectstest.AdvanceToMain(t, g)
	// Pushed bottom to top: the land is on top, the hero fourth.
	hero = effectstest.SeedLibraryTop(me, "Hero Card", "Creature — Human Hero", "")
	enchant = effectstest.SeedLibraryTop(me, "Plain Enchantment", "Enchantment", "")
	bear = effectstest.SeedLibraryTop(me, "Bear Card", "Creature — Bear", "")
	land = effectstest.SeedLibraryTop(me, "Forest Card", "Basic Land — Forest", "")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, rick, 0, game.ActivateAbilityParams{})
	return me, hero, enchant, bear, land
}

func TestMshA7RickJonesMillsFourAndOffersOnlyHeroesAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, hero, enchant, bear, land := mshA7RickJonesMill(t, g)
	for _, id := range []uuid.UUID{hero, enchant, bear, land} {
		if !me.Graveyard.Contains(id) {
			t.Fatalf("card %s was not milled", id)
		}
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt after the mill")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[hero] || !offered[enchant] {
		t.Errorf("offered %v, want exactly the Hero and the enchantment", offered)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, enchant)
	if !me.Hand.Contains(enchant) || me.Graveyard.Contains(enchant) {
		t.Error("the chosen enchantment did not reach the hand")
	}
	if !me.Graveyard.Contains(hero) {
		t.Error("the Hero that was not chosen left the graveyard")
	}
}

func TestMshA7RickJonesMayDeclineAndOnlyTheMilledCardsCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rick := effectstest.PushCatalogPermanent(g, me.ID, "Rick Jones, Destined Sidekick", "Legendary Creature — Human Advisor", mshA7RickJonesOracle, false)
	effectstest.AdvanceToMain(t, g)
	// A Hero already in the graveyard is not "among those cards".
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Hero", TypeLine: "Creature — Human Hero", Owner: me.ID, Controller: me.ID})
	for i := 0; i < 4; i++ {
		effectstest.SeedLibraryTop(me, "Bear Card", "Creature — Bear", "")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, rick, 0, game.ActivateAbilityParams{})
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatalf("a prompt was offered %v with no Hero or enchantment milled", pick.ChooseCards)
	}
	if !effectstest.B16Tapped(t, g, rick) {
		t.Error("the {T} cost was not paid")
	}
}

func TestMshA7RickJonesMayPickNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, hero, enchant, _, _ := mshA7RickJonesMill(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if me.Hand.Contains(hero) || me.Hand.Contains(enchant) {
		t.Error("a card reached the hand though none was chosen")
	}
}

// --- S.H.I.E.L.D. Deployment Drone ------------------------------------------------------------

func TestMshA7DeploymentDroneFliesAndMakesASoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	drone := effectstest.B12PlayFromHand(t, g, "S.H.I.E.L.D. Deployment Drone", "Artifact Creature — Robot", mshA7DroneOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !mshA3HasKeyword(g, drone, "flying") {
		t.Error("the Drone has no flying")
	}
	if n := effectstest.MshB1Tokens(g, me.ID, "Soldier"); n != 1 {
		t.Fatalf("%d Soldier tokens, want 1", n)
	}
	soldier := effectstest.MshB1Token(t, g, me.ID, "Soldier")
	if soldier.Power != 1 || soldier.Toughness != 1 || len(soldier.Colors) != 1 || soldier.Colors[0] != "W" {
		t.Errorf("the token is %+v, want a white 1/1", soldier)
	}
}

// --- S.H.I.E.L.D. Helicarrier -----------------------------------------------------------------

func TestMshA7HelicarrierMakesTwoSoldiersAtOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	car := effectstest.B12PlayFromHand(t, g, "S.H.I.E.L.D. Helicarrier", "Artifact — Vehicle", mshA7HelicarrierOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.MshB1Tokens(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("%d Soldier tokens, want 2", n)
	}
	if !mshA3HasKeyword(g, car, "flying") {
		t.Error("the Helicarrier has no flying")
	}
}

func TestMshA7HelicarrierCrewSixNeedsSixPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	car := effectstest.PushCatalogPermanent(g, me.ID, "S.H.I.E.L.D. Helicarrier", "Artifact — Vehicle", mshA7HelicarrierOracle, false)
	effectstest.AdvanceToMain(t, g)
	weak := effectstest.PushCrewerForTest(g, me.ID, "Weak Crew", 5)
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{weak}}); err == nil {
		t.Fatal("crewed with total power 5")
	}
	strong := effectstest.PushCrewerForTest(g, me.ID, "Strong Crew", 3)
	effectstest.CrewBy(t, g, me.ID, car, weak, strong)
	effectstest.ResolveTopOfStack(t, g)
	c, ok := effectstest.BattlefieldCard(g, car)
	if !ok || !c.IsCreature() {
		t.Error("the Helicarrier is not a creature after crewing with power 8")
	}
}

// --- Savage Land Dinosaur ---------------------------------------------------------------------

func TestMshA7SavageLandDinosaurTramplesAndBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dino := effectstest.PushCatalogPermanent(g, me.ID, "Savage Land Dinosaur", "Creature — Dinosaur", mshA7DinosaurOracle, false)
	if !mshA3HasKeyword(g, dino, "trample") {
		t.Error("the Dinosaur has no trample")
	}
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := effectstest.CycleFromHand(t, g, "Savage Land Dinosaur", "Creature — Dinosaur", mshA7DinosaurOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("no basic land reached the hand")
	}
}

// --- Shang-Chi, Master of Kung Fu -------------------------------------------------------------

func mshA7ShangChiTable(t *testing.T) (g *game.Game, me *game.Player, shang uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	shang = effectstest.PushCatalogPermanent(g, me.ID, "Shang-Chi, Master of Kung Fu", "Legendary Creature — Human Warrior Hero", mshA7ShangChiOracle, false)
	return g, me, shang
}

func TestMshA7ShangChiLetsSickCreaturesPayTapCosts(t *testing.T) {
	g, me, _ := mshA7ShangChiTable(t)
	elves := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, true)
	if err := g.ActivateManaAbility(me.ID, elves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Llanowar Elves' {T} ability the turn it arrived, under Shang-Chi: %v", err)
	}
	other := effectstest.NewCatalogGame(t)
	otherMe := other.Seats[other.Turn.ActiveSeat]
	lone := effectstest.PushCatalogPermanent(other, otherMe.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, true)
	if err := other.ActivateManaAbility(otherMe.ID, lone, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrSummoningSick) {
		t.Errorf("without Shang-Chi: err = %v, want ErrSummoningSick", err)
	}
}

func TestMshA7ShangChiMakesTwoManaOfOneColor(t *testing.T) {
	g, me, shang := mshA7ShangChiTable(t)
	effectstest.B28TapForMana(t, g, me.ID, shang, "R")
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool %v, want two {R}", got)
	}
	for _, tok := range me.ManaPool.Units() {
		if len(tok.Restrictions) != 1 {
			t.Errorf("token %+v, want one spend restriction", tok)
		}
	}
}

func TestMshA7ShangChiManaPaysCreatureAbilitiesButNotSpells(t *testing.T) {
	g, me, shang := mshA7ShangChiTable(t)
	rick := effectstest.PushCatalogPermanent(g, me.ID, "Rick Jones, Destined Sidekick", "Legendary Creature — Human Advisor", mshA7RickJonesOracle, false)
	effectstest.AdvanceToMain(t, g)
	// A bear spell cannot be cast with the restricted mana alone.
	effectstest.B28TapForMana(t, g, me.ID, shang, "G")
	bear := effectstest.HandCardFull(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", nil)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Shang-Chi's mana paid for a creature spell")
	}
	// It does pay an ability of a creature source: {3} with the two plus one more.
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, rick, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Rick Jones' {3}, {T} ability paid with Shang-Chi's mana: %v", err)
	}
}

// --- Shuri, Wakandan Inventor -----------------------------------------------------------------

func TestMshA7ShuriMakesArtifactSpellsCostOneLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, "", "Artifact — Equipment", "{3}"); got != 3 {
		t.Fatalf("an artifact spell costs %d before Shuri, want 3", got)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Shuri, Wakandan Inventor", "Legendary Creature — Human Artificer Hero", mshA7ShuriOracle, false)
	if got := effectstest.SelfPricedMV(t, g, me, "", "Artifact — Equipment", "{3}"); got != 2 {
		t.Errorf("an artifact spell costs %d with Shuri, want 2", got)
	}
	if got := effectstest.SelfPricedMV(t, g, me, "", "Creature — Bear", "{3}"); got != 3 {
		t.Errorf("a creature spell costs %d with Shuri, want 3", got)
	}
}

func TestMshA7ShuriCopiesASecondArtifactMinusLegendaryUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shuri := effectstest.PushCatalogPermanent(g, me.ID, "Shuri, Wakandan Inventor", "Legendary Creature — Human Artificer Hero", mshA7ShuriOracle, false)
	plain := effectstest.DcSeed(g, me.ID, "Plain Gadget", "Artifact", "", 0, 0)
	grand := effectstest.DcSeed(g, me.ID, "Grand Relic", "Legendary Artifact", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.DcMana(t, g, me.ID, "{C}")
	effectstest.B16Activate(t, g, me.ID, shuri, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(plain, grand)})
	c := effectstest.DcCard(t, g, plain)
	if c.Name != "Grand Relic" {
		t.Errorf("the first artifact is %q, want a copy of Grand Relic", c.Name)
	}
	if c.IsLegendary() {
		t.Error("the copy is legendary, but the text says it isn't")
	}
	if !effectstest.DcCard(t, g, grand).IsLegendary() {
		t.Error("the copied artifact lost its legendary supertype")
	}
	if !effectstest.B16Tapped(t, g, shuri) {
		t.Error("Shuri was not tapped for the cost")
	}
	effectstest.DcCleanup(g)
	if got := effectstest.DcCard(t, g, plain).Name; got != "Plain Gadget" {
		t.Errorf("after the end of turn the artifact is %q, want Plain Gadget back", got)
	}
}

func TestMshA7ShuriNeedsTwoDifferentArtifactsAndSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shuri := effectstest.PushCatalogPermanent(g, me.ID, "Shuri, Wakandan Inventor", "Legendary Creature — Human Artificer Hero", mshA7ShuriOracle, false)
	plain := effectstest.DcSeed(g, me.ID, "Plain Gadget", "Artifact", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, shuri, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(plain, plain)}); err == nil {
		t.Error("the same artifact was accepted as both targets")
	}
	other := effectstest.DcSeed(g, g.Seats[1].ID, "Their Gadget", "Artifact", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, shuri, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(plain, other)}); err == nil {
		t.Error("an artifact an opponent controls was accepted as the second target")
	}
}

// --- Stature, Size Shifter --------------------------------------------------------------------

func TestMshA7StatureCantBeBlockedWhilePowerIsOneOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stature := effectstest.B12Push(g, me.ID, "Stature, Size Shifter", "Legendary Creature — Human Hero", mshA7StatureOracle, 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	effectstest.BrAttack(t, g, stature)
	refused := effectstest.BrRefusal(t, g.DeclareBlocker(blocker, stature), game.BlockReasonCantBeBlocked)
	if refused.Source != stature {
		t.Errorf("the refusal names %s, want Stature herself", refused.Source)
	}
}

func TestMshA7StatureCanBeBlockedOnceHerPowerIsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stature := effectstest.B12Push(g, me.ID, "Stature, Size Shifter", "Legendary Creature — Human Hero", mshA7StatureOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(uuid.Nil, stature, game.CounterPlusOne, 1) })
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	effectstest.BrAttack(t, g, stature)
	if err := g.DeclareBlocker(blocker, stature); err != nil {
		t.Fatalf("a 2/2 Stature should be blockable: %v", err)
	}
}

func TestMshA7StaturePowerUpPutsXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	stature := effectstest.PushPowerUpPermanent(g, me.ID, "Stature, Size Shifter", "Legendary Creature — Human Hero", "{U}", mshA7StatureOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, stature, 0, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if n := effectstest.PlusOnes(g, stature); n != 3 {
		t.Errorf("%d +1/+1 counters, want 3", n)
	}
	effectstest.FloatMana(t, g, me, "{C}{U}{U}")
	err := g.ActivateCatalogAbility(me.ID, stature, 0, game.ActivateAbilityParams{XValue: 1})
	if !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMshA7StaturePowerUpIsCheaperByHerCostTheTurnSheEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	stature := effectstest.CastPowerUpCreature(t, g, me, "Stature, Size Shifter", "Legendary Creature — Human Hero", "{U}", "{U}", mshA7StatureOracle, 1)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, stature); got != "{X}{U}" {
		t.Errorf("the turn she entered with X=0: %q, want {X}{U} ({X}{U}{U} less her {U})", got)
	}
}

// --- Super Speed ------------------------------------------------------------------------------

func TestMshA7SuperSpeedGivesFirstStrikeThisTurnAndPlusOneHasteForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Super Speed", "Enchantment — Aura", mshA7SuperSpeedOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "first strike") {
		t.Error("the enchanted creature has no first strike the turn the Aura entered")
	}
	if !effectstest.HasKeywordOf(g, bear, "haste") {
		t.Error("the enchanted creature has no haste")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if effectstest.HasKeywordOf(g, bear, "first strike") {
		t.Error("first strike outlasted the turn")
	}
	if !effectstest.HasKeywordOf(g, bear, "haste") {
		t.Error("haste ended with the turn; it lasts as long as the Aura")
	}
}

func TestMshA7SuperSpeedHasFlash(t *testing.T) {
	spec, ok := effects.Lookup(mshA7SuperSpeedOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Super Speed keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Super Suit -------------------------------------------------------------------------------

func TestMshA7SuperSuitAttachesAndUntapsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AssistCardTapped(t, g, bear)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Fatal("setup: the creature should start tapped")
	}
	suit := effectstest.CastCatalogSpell(t, g, "Super Suit", "Artifact — Equipment", mshA7SuperSuitOracle, nil)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature is still tapped")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("equipped Bears are %d/%d, want 3/4", p, tough)
	}
	c, ok := effectstest.BattlefieldCard(g, suit)
	if !ok || !c.IsAttachedTo(bear) {
		t.Errorf("Super Suit attached: %v, %+v", ok, c.AttachedTo)
	}
}
