package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// global_series_b203_cards_test.go — Global Series Jiang Yanggu & Mu
// Yanling, S82 older-sets batch 203 (#5528).

const (
	gs1B203VortexOracle  = "373c6cd0-0bf7-4dab-bc18-7ecf5f015e10"
	gs1B203FoxOracle     = "51f4eb1f-abf2-425a-b78e-cafac177da89"
	gs1B203JourneyOracle = "e87606d5-5a23-4507-aece-8fdd044e7185"
	gs1B203MuOracle      = "34e1477a-c0fb-4364-bb93-04a85c0c2517"
	gs1B203JiangOracle   = "b15c8f43-685d-4901-9cad-47bc6dc6346f"
	gs1B203DeerOracle    = "2d7bef31-3447-4de6-9d0e-bd6749e7ecd5"
	gs1B203ArmorOracle   = "9eb58db8-7934-485c-8606-fb1a6cc60d42"
	gs1B203DrownOracle   = "a3874282-9887-4b53-b80e-564ac99cca23"
	gs1B203YakOracle     = "c2b659ef-68f6-452c-94b5-2872f4e7b069"
)

func TestGs1B203EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gs1B203VortexOracle:  "Rhythmic Water Vortex",
		gs1B203FoxOracle:     "Nine-Tail White Fox",
		gs1B203JourneyOracle: "Journey for the Elixir",
		gs1B203MuOracle:      "Mu Yanling",
		gs1B203JiangOracle:   "Jiang Yanggu",
		gs1B203DeerOracle:    "Sacred White Deer",
		gs1B203ArmorOracle:   "Hardened-Scale Armor",
		gs1B203DrownOracle:   "Drown in Shapelessness",
		gs1B203YakOracle:     "Earth-Origin Yak",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

func TestGs1B203NineTailWhiteFoxDrawsOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fox := effectstest.B12Push(g, me.ID, "Nine-Tail White Fox", "Creature — Fox Spirit", gs1B203FoxOracle, 2, 2)
	hand := me.Hand.Size()
	effectstest.DealCombatDamage(g, me.ID, fox, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestGs1B203EarthOriginYakPumpsYourCreaturesWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastRingCreature(t, g, "Earth-Origin Yak", "Creature — Ox", gs1B203YakOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tough)
	}
}

func TestGs1B203DrownInShapelessnessBouncesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Drown in Shapelessness", "Instant", gs1B203DrownOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the Bear was not returned to its owner's hand")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Drown in Shapelessness", "Instant", gs1B203DrownOracle, nil); err == nil {
		t.Error("cast with no target")
	}
}

func TestGs1B203HardenedScaleArmorGivesPlusThreePlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hardened-Scale Armor", "Enchantment — Aura", gs1B203ArmorOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
}

func TestGs1B203RhythmicWaterVortexBouncesTwoThenFindsMuYanling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	a := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Spared", "Creature — Bear", 2, 2)
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Mu Yanling", "Legendary Planeswalker — Yanling"),
		effectstest.SearchTestLand("Grizzly Bears", "Creature — Bear"),
	)
	graveyardMu := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: graveyardMu, Name: "Mu Yanling", TypeLine: "Legendary Planeswalker — Yanling", Owner: me.ID, Controller: me.ID})
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "U", 2)
	effectstest.CastFromHandForTest(t, g, me, "Rhythmic Water Vortex", "Sorcery", "{3}{U}{U}", gs1B203VortexOracle,
		game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(c) {
		t.Error("exactly the two targets should have been bounced")
	}
	if effectstest.SearchPromptFor(g, me.ID) == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Mu Yanling")
	effectstest.PassPriorityAroundTable(t, g)
	got := false
	for _, h := range me.Hand.Cards {
		got = got || h.Name == "Mu Yanling"
	}
	if !got {
		t.Error("Mu Yanling is not in hand")
	}
}

func TestGs1B203JourneyForTheElixirFindsALandAndJiangYanggu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Grizzly Bears", "Creature — Bear"),
	)
	jiang := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: jiang, Name: "Jiang Yanggu", TypeLine: "Legendary Planeswalker — Yanggu", Owner: me.ID, Controller: me.ID})
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.CastFromHandForTest(t, g, me, "Journey for the Elixir", "Sorcery", "{2}{G}", gs1B203JourneyOracle,
		game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.SearchPromptFor(g, me.ID)
	if pc == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if pc.SearchZonesPending() {
		t.Fatal("the search asked for zones: it names both")
	}
	var forest, island, bears uuid.UUID
	for _, id := range pc.SearchCards {
		switch effectstest.FindCardAnywhere(t, g, id).Name {
		case "Forest":
			forest = id
		case "Island":
			island = id
		case "Grizzly Bears":
			bears = id
		}
	}
	if forest == uuid.Nil || island == uuid.Nil || bears != uuid.Nil {
		t.Fatalf("offered cards wrong: forest %v island %v bears %v", forest, island, bears)
	}
	if err := g.ResolveSearchLibrary(pc.ID, me.ID, []uuid.UUID{forest, island}); err == nil {
		t.Fatal("two basic land cards were found")
	}
	if err := g.ResolveSearchLibrary(pc.ID, me.ID, []uuid.UUID{forest, jiang}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) || !me.Hand.Contains(jiang) {
		t.Error("the basic land and Jiang Yanggu are not in hand")
	}
}

func TestGs1B203SacredWhiteDeerGainsLifeOnlyWithYourYanggu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	deer := effectstest.B12Push(g, me.ID, "Sacred White Deer", "Creature — Elk", gs1B203DeerOracle, 2, 2)
	effectstest.ApaMana(me, "G", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, deer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without a Yanggu planeswalker")
	}
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Walker", "Legendary Planeswalker — Yanggu")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, game.CounterLoyalty, 3) })
	if err := g.ActivateCatalogAbility(me.ID, deer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with only an opponent's Yanggu planeswalker")
	}
	mine := effectstest.B12Permanent(g, me.ID, "My Walker", "Legendary Planeswalker — Yanggu")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterLoyalty, 3) })
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, deer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func gs1B203Mu(t *testing.T, g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	t.Helper()
	return effectstest.PushCatalogWalkerWithLine(g, owner, "Mu Yanling", "Legendary Planeswalker — Yanling", gs1B203MuOracle, loyalty)
}

func TestGs1B203MuYanlingMinusThreeDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mu := gs1B203Mu(t, g, me.ID, 5)
	effectstest.ToMain(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, mu, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if got := effectstest.LoyaltyCount(g, mu); got != 2 {
		t.Errorf("loyalty %d after −3, want 2", got)
	}
}

func TestGs1B203MuYanlingPlusTwoMakesACreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mu := gs1B203Mu(t, g, me.ID, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, mu, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyCount(g, mu); got != 7 {
		t.Errorf("loyalty %d after +2, want 7", got)
	}
	effectstest.BrAttack(t, g, mine)
	if err := g.DeclareBlocker(blocker, mine); err == nil {
		t.Error("the targeted creature was blocked")
	}
}

func TestGs1B203MuYanlingMinusTenTapsTheirCreaturesAndTakesAnExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mu := gs1B203Mu(t, g, me.ID, 10)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, mu, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FindCardAnywhere(t, g, theirs).Tapped {
		t.Error("the opponent's creature was not tapped")
	}
	if effectstest.FindCardAnywhere(t, g, mine).Tapped {
		t.Error("your own creature was tapped")
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me.Seat, true, "after Mu Yanling's −10")
}

func TestGs1B203JiangYanggu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jiang := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Jiang Yanggu", "Legendary Planeswalker — Yanggu", gs1B203JiangOracle, 9)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, jiang, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d after +1, want 4/4", p, tough)
	}
}

func TestGs1B203JiangYanggusMinusOneCreatesMowu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jiang := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Jiang Yanggu", "Legendary Planeswalker — Yanggu", gs1B203JiangOracle, 9)
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, jiang, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Mowu" && c.IsToken() {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Mowu tokens, want 1", n)
	}
}

func TestGs1B203JiangYanggusMinusOneMakesNoSecondMowu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jiang := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Jiang Yanggu", "Legendary Planeswalker — Yanggu", gs1B203JiangOracle, 9)
	effectstest.B12Creature(g, me.ID, "Mowu", "Legendary Creature — Dog", 3, 3)
	effectstest.ToMain(t, g)
	before := len(g.Battlefield.Cards)
	if err := g.ActivateCatalogAbility(me.ID, jiang, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Battlefield.Cards); got != before {
		t.Errorf("%d permanents, want %d: a second Mowu was made", got, before)
	}
}

func TestGs1B203JiangYanggusMinusFiveGivesTrampleAndLandCountPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jiang := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Jiang Yanggu", "Legendary Planeswalker — Yanggu", gs1B203JiangOracle, 9)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 4; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, jiang, 2, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("−5: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6 (four lands)", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}
