package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b158_cards_test.go — GAP batch 158 (#5948): six Jumpstart 2022 red
// creatures that were not in the catalog. Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	gapB158RunnerOracle    = "72b600e3-d671-4279-8ca6-9a3d6e774b53"
	gapB158MolochOracle    = "4056c1dd-9db4-4ff6-9dc5-a69e8da1d2a6"
	gapB158WhelpOracle     = "482f5fc8-ca3b-4280-8cf7-a1c55f8b9f12"
	gapB158SpewerOracle    = "1a668cef-bdc5-410b-b9e3-044e4203a0aa"
	gapB158PranksterOracle = "58cbac5c-d58e-4794-87f3-9563fce1b53f"
	gapB158CourierOracle   = "34e2d6eb-aad4-4776-872e-cebab3938aa0"
	gapB158SarkhanOracle   = "074cbe14-a0c8-4772-955a-b4052333d6ce"
	gapB158ChandraOracle   = "882619f7-69b4-4ce0-be4e-1e34e608f925"
)

func TestGapB158EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB158RunnerOracle:    "Rigging Runner",
		gapB158MolochOracle:    "Rooting Moloch",
		gapB158WhelpOracle:     "Sarkhan's Whelp",
		gapB158SpewerOracle:    "Spear Spewer",
		gapB158PranksterOracle: "Spiteful Prankster",
		gapB158CourierOracle:   "Torch Courier",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

func TestGapB158RiggingRunnerEntersWithACounterOnlyAfterAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.AvB16CastCreature(t, g, "Rigging Runner", "Creature — Goblin Pirate", gapB158RunnerOracle, 1, 1)
	if n := effectstest.CountersOn(g, first, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters without an attack this turn, want 0", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, first, "first strike") {
		t.Error("the Runner has no first strike")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	second := effectstest.AvB16CastCreature(t, g, "Rigging Runner", "Creature — Goblin Pirate", gapB158RunnerOracle, 1, 1)
	if n := effectstest.CountersOn(g, second, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after attacking this turn, want 1", n)
	}
}

func TestGapB158RootingMolochExilesACyclingCardAndLetsYouPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cycler := effectstest.PushToGraveyard(g, me, game.Card{Name: "Scrounging Skyray", TypeLine: "Creature — Bird", OracleID: effectstest.ScroungingSkyrayOracle, Power: 2, Toughness: 2})
	plain := effectstest.PushToGraveyard(g, me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.CastHoldCreature(t, g, "Rooting Moloch", "Creature — Lizard", gapB158MolochOracle, 4, 5)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: plain}}); err == nil {
		t.Fatal("a card with no cycling ability was accepted as a target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: cycler}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, cycler); z != game.ZoneExile {
		t.Fatalf("the cycling card is in zone %v, want exile", z)
	}
	if z := effectstest.ZoneOfCard(g, plain); z != game.ZoneGraveyard {
		t.Errorf("the other card is in zone %v, want the graveyard", z)
	}
	if err := g.CastSpell(me.ID, cycler, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Errorf("could not play the exiled card: %v", err)
	}
}

func TestGapB158RootingMolochCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Rooting Moloch", "Creature — Lizard", gapB158MolochOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled Moloch is not in the graveyard")
	}
}

func TestGapB158SarkhansWhelpPingsWhenYouActivateASarkhanAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	whelp := effectstest.B12Push(g, me.ID, "Sarkhan's Whelp", "Creature — Dragon", gapB158WhelpOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, whelp, "flying") {
		t.Error("the Whelp has no flying")
	}
	sarkhan := effectstest.PushCatalogPermanent(g, me.ID, "Sarkhan, Fireblood", "Legendary Planeswalker — Sarkhan", gapB158SarkhanOracle, false)
	chandra := effectstest.PushCatalogPermanent(g, me.ID, "Chandra, Novice Pyromancer", "Legendary Planeswalker — Chandra", gapB158ChandraOracle, false)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(sarkhan, game.CounterLoyalty, 7)
		_ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 5)
	})
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != life {
		t.Fatal("the Whelp triggered off a Chandra ability")
	}
	effectstest.B16Activate(t, g, me.ID, sarkhan, 2, game.ActivateAbilityParams{})
	effectstest.Vow1Pick(t, g, me, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
}

func TestGapB158SpearSpewerPingsEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spewer := effectstest.B12Push(g, me.ID, "Spear Spewer", "Creature — Goblin Warrior", gapB158SpewerOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, spewer, "defender") {
		t.Error("the Spewer has no defender")
	}
	effectstest.AdvanceToMain(t, g)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.B16Activate(t, g, me.ID, spewer, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		if p.Life != lives[i]-1 {
			t.Errorf("%s: life %d -> %d, want -1", p.Name, lives[i], p.Life)
		}
	}
	if !effectstest.B16Tapped(t, g, spewer) {
		t.Error("the Spewer did not tap")
	}
}

func TestGapB158SpitefulPranksterHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Push(g, me.ID, "Spiteful Prankster", "Creature — Devil", gapB158PranksterOracle, 3, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Spiteful Prankster", "Creature — Devil", gapB158PranksterOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "first strike") {
		t.Error("no first strike during its controller's turn")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("first strike during an opponent's turn")
	}
}

func TestGapB158SpitefulPranksterPingsWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prankster := effectstest.B12Push(g, me.ID, "Spiteful Prankster", "Creature — Devil", gapB158PranksterOracle, 3, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.B25Destroy(g, bear)
	effectstest.Vow1Pick(t, g, me, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1 after another creature died", life, opp.Life)
	}
	effectstest.B25Destroy(g, prankster)
	g.RunStateChecksForTest()
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the Prankster triggered off its own death")
	}
}

func TestGapB158TorchCourierGivesAnotherCreatureHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	courier := effectstest.B12Push(g, me.ID, "Torch Courier", "Creature — Goblin", gapB158CourierOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, courier, "haste") {
		t.Error("the Courier has no haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, courier, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(courier)}); err == nil {
		t.Fatal("the Courier could target itself")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Fatal("the Bear has haste before the ability")
	}
	effectstest.B16Activate(t, g, me.ID, courier, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(courier) {
		t.Error("the sacrifice was not paid as a cost")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear did not gain haste")
	}
}
