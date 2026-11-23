package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap_b176_cards_test.go — Jumpstart 2022 cards, S82 GAP batch 176
// (#5948). Every test plays the real catalog entry through the engine.

const (
	gapRuthlessDisposalOracle = "dcd12cb6-6920-4bff-af73-104ad32f7d59"
	gapDeclareDominanceOracle = "db640c5b-4548-44b1-a9cd-cc6838707dbb"
	gapRoarOfChallengeOracle  = "9d83a1d3-ef0f-4134-a38b-c992d396592b"
	gapRootsOfWisdomOracle    = "8f6af612-a4a1-4165-ae4d-ff6353cbaf8d"
	gapRuinInTheirWakeOracle  = "68adbc52-34dd-47c4-bb1c-0c51033fe7cc"
	gapSavagePunchOracle      = "3ca58e01-21c6-4314-ab3d-280ba6f9f509"
	gapScaleTheHeightsOracle  = "13d5b664-db13-406e-9956-3d9fe4177d33"
	gapSoulsMightOracle       = "a3538a5e-3d08-4c61-bcb4-ef1f5e5b659d"
	gapWolfsQuarryOracle      = "c6fead73-0473-4b57-899a-9f8c5b4f896d"
)

func TestGapB176CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapRuthlessDisposalOracle: "Ruthless Disposal", gapDeclareDominanceOracle: "Declare Dominance",
		gapRoarOfChallengeOracle: "Roar of Challenge", gapRootsOfWisdomOracle: "Roots of Wisdom",
		gapRuinInTheirWakeOracle: "Ruin in Their Wake", gapSavagePunchOracle: "Savage Punch",
		gapScaleTheHeightsOracle: "Scale the Heights", gapSoulsMightOracle: "Soul's Might", gapWolfsQuarryOracle: "Wolf's Quarry",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registration %+v", name, spec)
		}
	}
}

func TestGapRuthlessDisposalDiscardsSacrificesAndShrinksTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	a := effectstest.B12Creature(g, opp.ID, "Big One", "Creature — Giant", 8, 8)
	b := effectstest.B12Creature(g, opp.ID, "Big Two", "Creature — Giant", 9, 9)
	c := effectstest.B12Creature(g, opp.ID, "Big Three", "Creature — Giant", 9, 9)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	pitch := effectstest.PushCatalogHandCard(me, "Pitched", "Creature — Bear", "")
	spell := effectstest.PushCatalogHandCard(me, "Ruthless Disposal", "Sorcery", gapRuthlessDisposalOracle)
	two := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	for name, p := range map[string]game.CastSpellParams{
		"no discard":   {Strict: true, Targets: two, SacrificeIDs: []uuid.UUID{fodder}},
		"no sacrifice": {Strict: true, Targets: two, DiscardIDs: []uuid.UUID{pitch}},
		"one target":   {Strict: true, Targets: two[:1], DiscardIDs: []uuid.UUID{pitch}, SacrificeIDs: []uuid.UUID{fodder}},
	} {
		if err := g.CastSpell(me.ID, spell, p); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
	params := game.CastSpellParams{Strict: true, Targets: two, DiscardIDs: []uuid.UUID{pitch}, SacrificeIDs: []uuid.UUID{fodder}}
	if err := g.CastSpell(me.ID, spell, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) || !me.Graveyard.Contains(pitch) {
		t.Error("the costs were not paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a target survived -13/-13")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("an untargeted creature died")
	}
}

func TestGapDeclareDominancePumpsAndLures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Declare Dominance", "Sorcery", gapDeclareDominanceOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, bear); p != 5 || tou != 5 {
		t.Errorf("Bear = %d/%d, want 5/5", p, tou)
	}
	effectstest.ReqToBlockers(t, g, bear)
	if owed := effectstest.MustBlockNow(g); owed[wall] != bear {
		t.Errorf("must_block = %v, want the Wall on the Bear", owed)
	}
}

func TestGapRoarOfChallengeLuresAndGivesIndestructibleOnlyWithFerocious(t *testing.T) {
	for _, tc := range []struct {
		name      string
		ferocious bool
	}{{"without ferocious", false}, {"with ferocious", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.Seats1684(g)
			bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
			wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
			if tc.ferocious {
				effectstest.ReqCreature(g, me.ID, "Big", "", 4, 4)
			}
			effectstest.CastCatalogSpell(t, g, "Roar of Challenge", "Sorcery", gapRoarOfChallengeOracle, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.P1g2HasKeyword(g, bear, "indestructible"); got != tc.ferocious {
				t.Errorf("indestructible = %v, want %v", got, tc.ferocious)
			}
			effectstest.ReqToBlockers(t, g, bear)
			if owed := effectstest.MustBlockNow(g); owed[wall] != bear {
				t.Errorf("must_block = %v, want the Wall on the Bear", owed)
			}
		})
	}
}

func TestGapRootsOfWisdomReturnsALandOrElfElseDraws(t *testing.T) {
	// A land in the graveyard (a Bear is not a legal pick).
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Old Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Roots of Wisdom", "Sorcery", gapRootsOfWisdomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no return pick offered")
	}
	if pick.ChooseMin != 1 {
		t.Errorf("the pick is optional (min %d), it is mandatory", pick.ChooseMin)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(land) || len(me.Hand.Cards) != hand+1 {
		t.Error("the land did not return to hand (and nothing else was drawn)")
	}

	// An Elf counts; with neither a land nor an Elf the spell draws.
	g = effectstest.NewCatalogGame(t)
	me, _, _ = effectstest.BlbB1Seats(g)
	me.Library.Cards = nil
	for i := 0; i < 3; i++ {
		effectstest.StapleLibraryCard(me, "Milled Bear", "Creature — Bear")
	}
	effectstest.StapleLibraryCard(me, "Drawn", "Creature — Bear")
	elf := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: elf, Name: "Old Elf", TypeLine: "Creature — Elf Warrior", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Roots of Wisdom", "Sorcery", gapRootsOfWisdomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestChooseCardsFor(g, me.ID); pick == nil {
		t.Fatal("an Elf card was not offered")
	} else if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{elf}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	g = effectstest.NewCatalogGame(t)
	me, _, _ = effectstest.BlbB1Seats(g)
	hand = len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Roots of Wisdom", "Sorcery", gapRootsOfWisdomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand = %d, want %d: it draws when it can't return anything", len(me.Hand.Cards), hand+1)
	}
}

func TestGapRuinInTheirWakeFetchesToHandOrTappedWithWastes(t *testing.T) {
	// No Wastes: the basic land goes to hand.
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Ruin in Their Wake", "Sorcery", gapRuinInTheirWakeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("without Wastes the basic land should be in hand")
	}
	if g.Battlefield.Contains(forest) {
		t.Error("the land entered the battlefield without Wastes")
	}

	// With Wastes: the basic land enters tapped.
	g = effectstest.NewCatalogGame(t)
	me, _, _ = effectstest.BlbB1Seats(g)
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Wastes", TypeLine: "Basic Land"})
	forest = effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Ruin in Their Wake", "Sorcery", gapRuinInTheirWakeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	got, ok := effectstest.AangCardOnBF(g, forest)
	if !ok || !got.Tapped {
		t.Errorf("with Wastes the Forest should be on the battlefield tapped: %+v %v", got, ok)
	}
	if me.Hand.Contains(forest) {
		t.Error("the land went to hand as well")
	}
}

func TestGapSavagePunchFightsAndPumpsOnlyWithFerocious(t *testing.T) {
	for _, tc := range []struct {
		name      string
		ferocious bool
		survives  bool
	}{{"without ferocious", false, false}, {"with ferocious", true, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp, _ := effectstest.BlbB1Seats(g)
			mine := effectstest.B12Creature(g, me.ID, "My Fighter", "Creature — Bear", 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
			if tc.ferocious {
				effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 4, 4)
			}
			ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 1, 1)
			if err := effectstest.CastCatalogSpellErr(t, g, "Savage Punch", "Sorcery", gapSavagePunchOracle,
				effectstest.CardTargets(mine, ally)); err == nil {
				t.Fatal("Savage Punch fought two creatures I control")
			}
			effectstest.CastCatalogSpell(t, g, "Savage Punch", "Sorcery", gapSavagePunchOracle, effectstest.CardTargets(mine, theirs))
			effectstest.PassPriorityAroundTable(t, g)
			if got := g.Battlefield.Contains(mine); got != tc.survives {
				t.Errorf("my fighter survives = %v, want %v", got, tc.survives)
			}
			if tc.ferocious && g.Battlefield.Contains(theirs) {
				t.Error("their 3/3 survived a 4-power fighter")
			}
		})
	}
}

func TestGapScaleTheHeightsDoesAllFourThings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life, hand := me.Life, len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Scale the Heights", "Sorcery", gapScaleTheHeightsOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want 1", n)
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand = %d, want %d", len(me.Hand.Cards), hand+1)
	}
	land1 := effectstest.PushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	land2 := effectstest.PushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	land3 := effectstest.PushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	for i, id := range []uuid.UUID{land1, land2} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("land %d: %v", i+1, err)
		}
	}
	if err := g.CastSpell(me.ID, land3, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a third land was played: only one additional land is granted")
	}
}

func TestGapScaleTheHeightsWithNoTargetStillGainsDrawsAndGrantsTheLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	life, hand := me.Life, len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Scale the Heights", "Sorcery", gapScaleTheHeightsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || len(me.Hand.Cards) != hand+1 {
		t.Errorf("life %d hand %d, want %d and %d", me.Life, len(me.Hand.Cards), life+2, hand+1)
	}
}

func TestGapSoulsMightPutsPowerManyCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 2)
	effectstest.CastCatalogSpell(t, g, "Soul's Might", "Sorcery", gapSoulsMightOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 3 {
		t.Errorf("counters = %d, want 3 (its power)", n)
	}
	if p, tou := effectstest.PtOf(t, g, bear); p != 6 || tou != 5 {
		t.Errorf("Bear = %d/%d, want 6/5", p, tou)
	}
}

func TestGapWolfsQuarryMakesThreeBoarsThatLeaveFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.CastCatalogSpell(t, g, "Wolf's Quarry", "Sorcery", gapWolfsQuarryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var boars []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Boar" {
			boars = append(boars, c.InstanceID)
		}
	}
	if len(boars) != 3 {
		t.Fatalf("boars = %d, want 3", len(boars))
	}
	if p, tou := effectstest.PtOf(t, g, boars[0]); p != 1 || tou != 1 {
		t.Errorf("Boar = %d/%d, want 1/1", p, tou)
	}
	effectstest.P1Destroy(t, g, boars[0])
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Errorf("Food tokens = %d, want 1 after one Boar died", n)
	}
	effectstest.P1Destroy(t, g, boars[1])
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 2 {
		t.Errorf("Food tokens = %d, want 2 after two Boars died", n)
	}
}
