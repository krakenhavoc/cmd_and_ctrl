package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b34_b_cards_test.go — Ferocification, Hellspur Brute, Dust Animus,
// Highway Robbery, Loan Shark, Take for a Ride, Irascible Wolverine,
// Colossal Rattlewurm and Intrepid Stablemaster (Outlaws of Thunder Junction).

const (
	ferocificationOracle       = "814e06c0-9b61-4a64-8c67-73a8eb00a768"
	hellspurBruteOracle        = "81e5c62c-5aed-41e2-8bd1-610022e880bd"
	dustAnimusOracle           = "494f8c29-7626-43ed-b0b8-fb9afd4cd5d3"
	highwayRobberyOracle       = "df020a47-1784-479b-9882-be77b7f30950"
	loanSharkOracle            = "15030c1d-7c0a-4273-a022-eb98d6c7a1ce"
	takeForARideOracle         = "4f815ca4-0197-48c7-bfa2-5eb937c615e6"
	irascibleWolverineOracle   = "445cd4bd-ba7c-42a2-a794-42aed823d0d3"
	colossalRattlewurmOracle   = "510659d6-69b4-4b04-8836-08b76213954e"
	intrepidStablemasterOracle = "a825bdea-3ea2-46b2-93b9-21586643aab0"
)

func TestStdB34PlotCostsAreThePrintedOnes(t *testing.T) {
	for name, tc := range map[string]struct{ oracle, cost string }{
		"Dust Animus":         {dustAnimusOracle, "{1}{W}"},
		"Highway Robbery":     {highwayRobberyOracle, "{1}{R}"},
		"Loan Shark":          {loanSharkOracle, "{3}{U}"},
		"Irascible Wolverine": {irascibleWolverineOracle, "{2}{R}"},
	} {
		sa := effectstest.PlotOf(tc.oracle)
		if sa == nil || sa.Cost != tc.cost {
			t.Errorf("%s: plot %+v, want cost %s", name, sa, tc.cost)
		}
	}
}

func TestStdB34FerocificationPumpsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Ferocification", "Enchantment", ferocificationOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the trigger did not ask for a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
}

func TestStdB34FerocificationGrantsMenaceAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Ferocification", "Enchantment", ferocificationOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the trigger did not ask for a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"menace", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the creature should have %s", kw)
		}
	}
}

func TestStdB34FerocificationDoesNothingOnAnOpponentsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Ferocification", "Enchantment", ferocificationOracle, 0, 0)
	effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if c := effectstest.ModePickChoiceFor(g, opp.ID); c != nil {
		t.Fatal("Ferocification triggered on another player's turn")
	}
}

func TestStdB34HellspurBruteHasAffinityForOutlaws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	if got := effectstest.SelfPricedMV(t, g, me, hellspurBruteOracle, "Creature — Ogre Warrior", "{4}{R}"); got != 5 {
		t.Errorf("no outlaws: %d, want 5", got)
	}
	effectstest.BlbB1Push(g, me.ID, "Rogue", "Creature — Human Rogue", "", 1, 1)
	effectstest.BlbB1Push(g, me.ID, "Pirate", "Creature — Human Pirate", "", 1, 1)
	effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 1, 1)
	if got := effectstest.SelfPricedMV(t, g, me, hellspurBruteOracle, "Creature — Ogre Warrior", "{4}{R}"); got != 3 {
		t.Errorf("two outlaws: %d, want 3", got)
	}
}

func dustAnimusEnters(t *testing.T, g *game.Game, untappedLands int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	for i := 0; i < untappedLands; i++ {
		effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	}
	return effectstest.CastHoldCreature(t, g, "Dust Animus", "Creature — Spirit", dustAnimusOracle, 3, 3)
}

func TestStdB34DustAnimusEntersBigWithFiveUntappedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := dustAnimusEnters(t, g, 5)
	if got := effectstest.PlusOnes(g, id); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
	if got := effectstest.CounterCount(g, id, game.CounterLifelink); got != 1 {
		t.Errorf("lifelink counters %d, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
		t.Error("the lifelink counter should grant lifelink")
	}
}

func TestStdB34DustAnimusEntersPlainWithFourUntappedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := dustAnimusEnters(t, g, 4)
	if got := effectstest.PlusOnes(g, id) + effectstest.CounterCount(g, id, game.CounterLifelink); got != 0 {
		t.Errorf("%d counters, want none", got)
	}
}

func TestStdB34DustAnimusDoesNotCountTappedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	}
	tapped := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, tapped).Tapped = true })
	id := effectstest.CastHoldCreature(t, g, "Dust Animus", "Creature — Spirit", dustAnimusOracle, 3, 3)
	if got := effectstest.PlusOnes(g, id); got != 0 {
		t.Errorf("%d counters with only four untapped lands, want 0", got)
	}
}

func highwayRobberyCast(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Highway Robbery", "Sorcery", highwayRobberyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestStdB34HighwayRobberyDiscardsToDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := me.Library.Size()
	highwayRobberyCast(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // Discard a card
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d, want %d (two cards drawn)", me.Library.Size(), lib-2)
	}
}

func TestStdB34HighwayRobberySacrificesALandToDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	lib := me.Library.Size()
	highwayRobberyCast(t, g)
	// Hand holds cards, so the options are: neither, discard, sacrifice.
	effectstest.AnswerOptionPick(t, g, me.ID, 2)
	if c := effectstest.LatestChooseCardsFor(g, me.ID); c != nil {
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land}); err != nil {
			t.Fatalf("choose the land: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land should have been sacrificed")
	}
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d, want %d (two cards drawn)", me.Library.Size(), lib-2)
	}
}

func TestStdB34HighwayRobberyDecliningDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	lib := me.Library.Size()
	highwayRobberyCast(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib {
		t.Errorf("library %d, want %d (declined)", me.Library.Size(), lib)
	}
}

func TestStdB34LoanSharkDrawsOnlyAfterTwoSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Loan Shark", "Creature — Shark Pirate", loanSharkOracle, 3, 2)
	if me.Library.Size() != lib {
		t.Errorf("one spell: library %d, want %d (no draw)", me.Library.Size(), lib)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	effectstest.CastWithCost(t, g2, "Cantrip", "Instant", "", "")
	effectstest.PassPriorityAroundTable(t, g2)
	lib2 := me2.Library.Size()
	effectstest.CastHoldCreature(t, g2, "Loan Shark", "Creature — Shark Pirate", loanSharkOracle, 3, 2)
	if me2.Library.Size() != lib2-1 {
		t.Errorf("second spell: library %d, want %d (one draw)", me2.Library.Size(), lib2-1)
	}
}

func TestStdB34TakeForARideStealsUntapsAndGivesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, victim).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Take for a Ride", "Sorcery", takeForARideOracle, effectstest.TwTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, victim)
	if c == nil || c.Controller != me.ID {
		t.Fatalf("control did not change: %+v", c)
	}
	if c.Tapped {
		t.Error("the creature should be untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, victim, "haste") {
		t.Error("the creature should have haste")
	}
}

func TestStdB34TakeForARideNeedsACrimeToBeCastAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	// Seat 1's turn: seat 0 has committed no crime, so it may not cast it.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Take for a Ride", TypeLine: "Sorcery", OracleID: takeForARideOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.TwTarget(victim)}); err == nil {
		t.Fatal("cast at instant speed with no crime committed this turn")
	}
}

func TestStdB34IrascibleWolverineExilesTheTopCardAndMayPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top Card")
	lib := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Irascible Wolverine", "Creature — Wolverine", irascibleWolverineOracle, 3, 2)
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d", me.Library.Size(), lib-1)
	}
	if g.Exile.Size() != 1 {
		t.Errorf("exile holds %d cards, want 1", g.Exile.Size())
	}
}

func TestStdB34ColossalRattlewurmFlashNeedsADesert(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Colossal Rattlewurm", TypeLine: "Creature — Wurm", OracleID: colossalRattlewurmOracle, Owner: me.ID, Controller: me.ID, Power: 6, Toughness: 5})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast at instant speed without a Desert")
	}
	effectstest.B12Permanent(g, me.ID, "Sunscorched Desert", "Land — Desert")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("flash cast with a Desert: %v", err)
	}
}

func TestStdB34ColossalRattlewurmFetchesADesertFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Colossal Rattlewurm", TypeLine: "Creature — Wurm", OracleID: colossalRattlewurmOracle, Owner: me.ID, Controller: me.ID, Power: 6, Toughness: 5})
	desert := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: desert, Name: "Desert", TypeLine: "Land — Desert", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, desert)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the Rattlewurm should be exiled as the cost")
	}
	c := effectstest.FindBattlefieldCardByID(g, desert)
	if c == nil || !c.Tapped {
		t.Errorf("the Desert should be on the battlefield tapped: %+v", c)
	}
}

func TestStdB34IntrepidStablemasterRestrictsItsBigManaToMountsAndVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sm := effectstest.B12Push(g, me.ID, "Intrepid Stablemaster", "Creature — Human Scout", intrepidStablemasterOracle, 1, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, sm, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap for two of one colour: %v", err)
	}
	cost, _ := game.ParseCost("{1}")
	mount := game.Card{Name: "Horse", TypeLine: "Creature — Mount", ManaCost: "{1}"}
	vehicle := game.Card{Name: "Cart", TypeLine: "Artifact — Vehicle", ManaCost: "{1}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(mount)) || !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(vehicle)) {
		t.Error("the mana refused a Mount or Vehicle spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) || me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(vehicle)) {
		t.Error("the mana paid for something that is not a Mount or Vehicle spell")
	}
}

func TestStdB34IntrepidStablemasterTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sm := effectstest.B12Push(g, me.ID, "Intrepid Stablemaster", "Creature — Human Scout", intrepidStablemasterOracle, 1, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, sm, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for G: %v", err)
	}
	cost, _ := game.ParseCost("{G}")
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("the plain {G} should pay for anything")
	}
}
