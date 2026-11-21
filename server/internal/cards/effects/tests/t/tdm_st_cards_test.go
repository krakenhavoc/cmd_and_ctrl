package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_st_cards_test.go — Tarkir: Dragonstorm, cards from S to T (#4463).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tdmSaguPummelerOracle       = "00f1d167-c143-471d-bd52-da2a65d087c4"
	tdmSaltRoadPackbeastOracle  = "50261b9d-48f2-4bad-a35b-14f13b3cdd32"
	tdmSandskitterOutriderOracl = "eae480fe-d9fe-4dbe-8b78-b65f398b0af3"
	tdmSarkhansResolveOracle    = "30f058f8-5777-4c25-b881-8f1d0ce17d3a"
	tdmShockBrigadeOracle       = "8676e917-17dd-4047-99c4-402f2e909e16"
	tdmShockingSharpshooterOrcl = "fbb7735f-ec91-4ed4-b735-d94a76e68a80"
	tdmSibsigAppraiserOracle    = "5b51958d-533a-45e4-aac1-3c157670978b"
	tdmSinkholeSurveyorOracle   = "83a47724-b8d1-4884-9442-b62aa1510dc3"
	tdmSkirmishRhinoOracle      = "40eb7f43-7cee-41a1-a4d1-f34ee125ccd1"
	tdmStormplainDetainmentOrcl = "4ad9c550-5605-4ffe-8e27-e9f2d9ada000"
	tdmSummitIntimidatorOracle  = "b772d35a-d365-47ce-b1dd-a2ba574c74ea"
	tdmTemurTawnybackOracle     = "889d26d6-9e6c-4cc0-8e8e-dd932f705a20"
)

func TestTdmSTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmSaguPummelerOracle:       "Sagu Pummeler",
		tdmSaltRoadPackbeastOracle:  "Salt Road Packbeast",
		tdmSandskitterOutriderOracl: "Sandskitter Outrider",
		tdmSarkhansResolveOracle:    "Sarkhan's Resolve",
		tdmShockBrigadeOracle:       "Shock Brigade",
		tdmShockingSharpshooterOrcl: "Shocking Sharpshooter",
		tdmSibsigAppraiserOracle:    "Sibsig Appraiser",
		tdmSinkholeSurveyorOracle:   "Sinkhole Surveyor",
		tdmSkirmishRhinoOracle:      "Skirmish Rhino",
		tdmStormplainDetainmentOrcl: "Stormplain Detainment",
		tdmSummitIntimidatorOracle:  "Summit Intimidator",
		tdmTemurTawnybackOracle:     "Temur Tawnyback",
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

// --- fixtures ------------------------------------------------------------------

// tdmSTCast puts a creature card with its printed body in the active
// seat's hand and casts it from a main phase, paying nothing.
func tdmSTCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
		})
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// --- Sagu Pummeler -------------------------------------------------------------

func TestTdmSaguPummelerHasReachAndRenewsCountersAndReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beast := effectstest.B12Push(g, me.ID, "Sagu Pummeler", "Creature — Beast", tdmSaguPummelerOracle, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, beast, "reach") {
		t.Error("Sagu Pummeler has no reach")
	}
	bear := effectstest.SeedBear(g, me.ID)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Sagu Pummeler", "Creature — Beast",
		tdmSaguPummelerOracle, 4, 4, "{G}{G}{G}{G}{G}", game.ActivateAbilityParams{
			Targets: effectstest.CardRefs(bear),
		})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile the Pummeler")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterReach); got != 1 {
		t.Errorf("%d reach counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the reach counter did not give reach")
	}
}

func TestTdmSaguPummelerRenewCostsFourGFromTheGraveyardOnly(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tdmSaguPummelerOracle})
	if len(abilities) != 1 || abilities[0].Cost.Mana != "{4}{G}" ||
		!game.AbilityFunctionsFromZone(abilities[0], game.ZoneGraveyard) ||
		game.AbilityFunctionsFromZone(abilities[0], game.ZoneBattlefield) {
		t.Fatalf("abilities %+v, want one {4}{G} ability that works only from the graveyard", abilities)
	}
}

// --- Salt Road Packbeast -------------------------------------------------------

var tdmPackbeastCard = game.Card{Name: "Salt Road Packbeast", TypeLine: "Creature — Beast",
	ManaCost: "{5}{W}", OracleID: tdmSaltRoadPackbeastOracle, Power: 4, Toughness: 3}

func TestTdmSaltRoadPackbeastCostsOneLessPerCreatureAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.SeedBear(g, me.ID)
	}
	effectstest.SeedBear(g, opp.ID)
	effectstest.TopLibraryCard(me, "Drawn", "Sorcery")
	// Three creatures: {5}{W} is {2}{W}.
	tmcB6Pool(t, g, me, "{W}{C}")
	if _, err := tmcB6Cast(g, me, tdmPackbeastCard, game.CastSpellParams{}); err == nil {
		t.Fatal("the Packbeast was cast for {1}{W} with three creatures (an opponent's counted)")
	}
	tmcB6Pool(t, g, me, "{C}")
	if _, err := tmcB6Cast(g, me, tdmPackbeastCard, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for {2}{W}: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn on entry", hand, me.Hand.Size())
	}
}

// --- Sandskitter Outrider ------------------------------------------------------

func TestTdmSandskitterOutriderEnduresTwoWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rider := tdmSTCast(t, g, "Sandskitter Outrider", "Creature — Goblin Soldier", tdmSandskitterOutriderOracl, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no endure choice on entry")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if got := effectstest.CountersOn(g, rider, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if n := len(effectstest.SocB8Named(g, me.ID, "Spirit")); n != 0 {
		t.Errorf("%d Spirits after choosing the counters", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, rider, "menace") {
		t.Error("Sandskitter Outrider has no menace")
	}
}

func TestTdmSandskitterOutriderEnduresTwoWithASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rider := tdmSTCast(t, g, "Sandskitter Outrider", "Creature — Goblin Soldier", tdmSandskitterOutriderOracl, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	spirits := effectstest.SocB8Named(g, me.ID, "Spirit")
	if len(spirits) != 1 || spirits[0].Power != 2 || spirits[0].Toughness != 2 ||
		len(spirits[0].Colors) != 1 || spirits[0].Colors[0] != "W" {
		t.Fatalf("Spirits %+v, want one 2/2 white Spirit", spirits)
	}
	if got := effectstest.CountersOn(g, rider, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters after choosing the Spirit", got)
	}
}

// --- Sarkhan's Resolve ---------------------------------------------------------

func TestTdmSarkhansResolvePumpsThreeThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.CastModal(t, g, "Sarkhan's Resolve", "Instant", tdmSarkhansResolveOracle, []int{0}, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the pump mode destroyed its target")
	}
}

func TestTdmSarkhansResolveDestroysOnlyAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ground := effectstest.SeedBear(g, opp.ID)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	if _, err := effectstest.CastModalErr(t, g, "Sarkhan's Resolve", "Instant", tdmSarkhansResolveOracle, []int{1}, effectstest.CardRefs(ground)); err == nil {
		t.Fatal("the destroy mode targeted a creature without flying")
	}
	effectstest.CastModal(t, g, "Sarkhan's Resolve", "Instant", tdmSarkhansResolveOracle, []int{1}, effectstest.CardRefs(flier))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the flier was not destroyed")
	}
	if !g.Battlefield.Contains(ground) {
		t.Error("the creature without flying was touched")
	}
}

// --- Shock Brigade -------------------------------------------------------------

func TestTdmShockBrigadeMobilizesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brigade := effectstest.B12Push(g, me.ID, "Shock Brigade", "Creature — Goblin Soldier", tdmShockBrigadeOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, brigade, "menace") {
		t.Error("Shock Brigade has no menace")
	}
	effectstest.DeclareAttack(t, g, opp.ID, brigade)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TdmPlaceWarriors(t, g, me.ID)
	warriors := effectstest.SocB8Named(g, me.ID, "Warrior")
	if len(warriors) != 1 {
		t.Fatalf("mobilize 1 made %d Warriors, want 1", len(warriors))
	}
	w := warriors[0]
	if !w.Tapped || w.AttackingTarget == uuid.Nil || w.Power != 1 || w.Toughness != 1 {
		t.Errorf("Warrior %d/%d tapped=%v attacking=%v, want a tapped and attacking 1/1",
			w.Power, w.Toughness, w.Tapped, w.AttackingTarget != uuid.Nil)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB8Named(g, me.ID, "Warrior")); n != 0 {
		t.Errorf("%d Warriors survived the end step", n)
	}
}

// --- Shocking Sharpshooter -----------------------------------------------------

func TestTdmShockingSharpshooterPingsAnOpponentWhenAnotherCreatureOfYoursEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archer := effectstest.B12Push(g, me.ID, "Shocking Sharpshooter", "Creature — Human Archer", tdmShockingSharpshooterOrcl, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, archer, "reach") {
		t.Error("Shocking Sharpshooter has no reach")
	}
	oppLife, myLife := opp.Life, me.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 black Zombie Druid"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != oppLife {
		t.Fatal("an opponent's creature entering triggered the Sharpshooter")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 black Zombie Druid"), 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Fatal("the Sharpshooter targeted its own controller; the target is an opponent")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.PlayerTarget(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 1 {
		t.Errorf("the opponent lost %d life, want 1", got)
	}
	if me.Life != myLife {
		t.Errorf("the Sharpshooter's controller lost %d life", myLife-me.Life)
	}
}

// --- Sibsig Appraiser ----------------------------------------------------------

func TestTdmSibsigAppraiserTakesOneOfTheTopTwoAndBinsTheOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lower := effectstest.TopLibraryCard(me, "Lower", "Sorcery")
	upper := effectstest.TopLibraryCard(me, "Upper", "Sorcery")
	tdmSTCast(t, g, "Sibsig Appraiser", "Creature — Zombie Advisor", tdmSibsigAppraiserOracle, 2, 1)
	lib := me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick between the top two cards")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top two", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lower}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(lower) {
		t.Error("the chosen card is not in hand")
	}
	if !me.Graveyard.Contains(upper) {
		t.Error("the other of the top two should be put into the graveyard")
	}
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d -> %d, want exactly the top two gone", lib, me.Library.Size())
	}
}

// --- Sinkhole Surveyor ---------------------------------------------------------

func TestTdmSinkholeSurveyorLosesOneAndEnduresOneWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.B12Push(g, me.ID, "Sinkhole Surveyor", "Creature — Bird Scout", tdmSinkholeSurveyorOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("Sinkhole Surveyor has no flying")
	}
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, bird)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no endure choice when the Surveyor attacked")
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 lost before the endure", life, me.Life)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if got := effectstest.CountersOn(g, bird, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestTdmSinkholeSurveyorCanMakeASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.B12Push(g, me.ID, "Sinkhole Surveyor", "Creature — Bird Scout", tdmSinkholeSurveyorOracle, 1, 3)
	effectstest.DeclareAttack(t, g, opp.ID, bird)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	spirits := effectstest.SocB8Named(g, me.ID, "Spirit")
	if len(spirits) != 1 || spirits[0].Power != 1 || spirits[0].Toughness != 1 {
		t.Fatalf("Spirits %+v, want one 1/1 Spirit", spirits)
	}
	if got := effectstest.CountersOn(g, bird, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters after choosing the Spirit", got)
	}
}

// --- Skirmish Rhino ------------------------------------------------------------

func TestTdmSkirmishRhinoDrainsEachOpponentForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	rhino := tdmSTCast(t, g, "Skirmish Rhino", "Creature — Rhino", tdmSkirmishRhinoOracle, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID] - 2
		if p.ID == me.ID {
			want = before[p.ID] + 2
		}
		if p.Life != want {
			t.Errorf("seat %s life %d -> %d, want %d", p.ID, before[p.ID], p.Life, want)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, rhino, "trample") {
		t.Error("Skirmish Rhino has no trample")
	}
}

// --- Stormplain Detainment -----------------------------------------------------

func TestTdmStormplainDetainmentExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	mine := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	detainment := effectstest.CastCatalogSpell(t, g, "Stormplain Detainment", "Enchantment", tdmStormplainDetainmentOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	for name, id := range map[string]uuid.UUID{"an opponent's land": land, "your own permanent": mine} {
		if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(id)); err == nil {
			t.Fatalf("the Detainment targeted %s", name)
		}
	}
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(relic) || g.Battlefield.Contains(relic) {
		t.Fatal("the opponent's artifact was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(detainment) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Relic", opp.ID); n != 1 {
		t.Errorf("after the Detainment left, %d Relics under their owner's control, want 1", n)
	}
}

// --- Summit Intimidator --------------------------------------------------------

func TestTdmSummitIntimidatorStopsATargetFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := effectstest.B12Creature(g, opp.ID, "Other Wall", "Creature — Wall", 0, 4)
	yeti := tdmSTCast(t, g, "Summit Intimidator", "Creature — Yeti", tdmSummitIntimidatorOracle, 4, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, yeti, "reach")
	effectstest.AssertRestrictions(t, g, wall, game.CantBlock)
	effectstest.AssertRestrictions(t, g, other, 0)
}

// --- Temur Tawnyback -----------------------------------------------------------

var tdmTawnybackCard = game.Card{Name: "Temur Tawnyback", TypeLine: "Creature — Beast",
	ManaCost: "{2/G}{2/U}{2/R}", OracleID: tdmTemurTawnybackOracle, Power: 4, Toughness: 3}

func TestTdmTemurTawnybackPaysHybridAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.TopLibraryCard(me, "Drawn", "Sorcery")
	// {G}, {U}, then two of any type for {2/R}.
	tmcB6Pool(t, g, me, "{G}{U}{C}{C}")
	if _, err := tmcB6Cast(g, me, tdmTawnybackCard, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for {G}{U}{C}{C}: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn before the discard", hand, me.Hand.Size())
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Error("no discard prompt after the draw")
	}
}

func TestTdmTemurTawnybackRefusesTooLittleMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tmcB6Pool(t, g, me, "{G}{U}{C}")
	if _, err := tmcB6Cast(g, me, tdmTawnybackCard, game.CastSpellParams{}); err == nil {
		t.Error("the Tawnyback was cast for three mana with one hybrid paid generic")
	}
}
