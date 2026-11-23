package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b143_test.go — S82 GAP batch 143 (Dominaria Remastered), #5948:
// the cards through the real catalog.

const (
	gapTrevasRuins        = "7b2c7758-2b89-49ff-8838-8dc9880c7209"
	gapNantukoMonastery   = "c6de0ee9-785d-4cd8-8a7f-5bb715763131"
	gapSmolderingCrater   = "be09e83f-6486-4f42-8d2e-416cb95173f9"
	gapIchorSlick         = "7a36fc2e-bb02-4222-a335-4c80e380e690"
	gapUrborgUprising     = "882d3be7-1c0f-4d2b-8f2e-32369488ef82"
	gapSaprolingSymbiosis = "69d22037-fbf7-44d2-81a7-7ce8e32f06ec"
)

func gapB143PT(g *game.Game, id uuid.UUID) (p, tough int) {
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			p, tough = c.PowerForComparison(), c.CurrentToughness()
		}
	})
	return
}

func gapB143PlayLand(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

func TestGapB143Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		gapTrevasRuins: "Treva's Ruins", gapNantukoMonastery: "Nantuko Monastery",
		gapSmolderingCrater: "Smoldering Crater", gapIchorSlick: "Ichor Slick",
		gapUrborgUprising: "Urborg Uprising", gapSaprolingSymbiosis: "Saproling Symbiosis",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestGapB143TrevasRuinsReturnsANonLairLandOrIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	ruins := gapB143PlayLand(t, g, "Treva's Ruins", "Land — Lair", gapTrevasRuins)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the trigger must ask which land to return")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != other {
		t.Fatalf("offered %v, want only the Forest", prompt.ChooseCards)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{other}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) || !me.Hand.Contains(other) || !g.Battlefield.Contains(ruins) {
		t.Error("the Forest should be in hand and the Lair should stay")
	}
	spec, _ := effects.Lookup(gapTrevasRuins)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{G|W|U}" {
		t.Errorf("produces %+v, want {G|W|U}", spec.ManaAbilities)
	}

	// Only another Lair to return: the new Lair is sacrificed.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	otherLair := effectstest.SeedLandOnBattlefield(g2, me2.ID, "Rith's Grove", "Land — Lair")
	ruins2 := gapB143PlayLand(t, g2, "Treva's Ruins", "Land — Lair", gapTrevasRuins)
	effectstest.PassPriorityAroundTable(t, g2)
	if pick := effectstest.LatestChooseCardsFor(g2, me2.ID); pick != nil {
		if len(pick.ChooseCards) != 0 {
			t.Fatalf("offered %v: a Lair must not be offered", pick.ChooseCards)
		}
		if err := g2.ResolveChooseCards(pick.ID, me2.ID, nil); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g2)
	}
	if g2.Battlefield.Contains(ruins2) || !g2.Battlefield.Contains(otherLair) {
		t.Error("with only a Lair to return the new Lair should be sacrificed and the other left alone")
	}
}

func TestGapB143SmolderingCraterEntersTappedAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	crater := gapB143PlayLand(t, g, "Smoldering Crater", "Land", gapSmolderingCrater)
	effectstest.Top100AssertEnteredTapped(t, g, crater, "Smoldering Crater")

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.SeedLibrary(me2, "A", "B")
	n := len(me2.Library.Cards)
	effectstest.CycleFromHand(t, g2, "Smoldering Crater", "Land", gapSmolderingCrater, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g2)
	if len(me2.Library.Cards) != n-1 {
		t.Errorf("cycling drew %d cards, want 1", n-len(me2.Library.Cards))
	}
}

func TestGapB143NantukoMonasteryNeedsThresholdAndBecomesAFourFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monastery := effectstest.PushCatalogPermanent(g, me.ID, "Nantuko Monastery", "Land", gapNantukoMonastery, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "W")
	if err := g.ActivateCatalogAbility(me.ID, monastery, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with fewer than seven cards in the graveyard")
	}
	for i := 0; i < 7; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	if err := g.ActivateCatalogAbility(me.ID, monastery, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with threshold: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := gapB143PT(g, monastery); p != 4 || th != 4 {
		t.Errorf("got %d/%d, want 4/4", p, th)
	}
	if !effectstest.P1g2HasKeyword(g, monastery, "first strike") {
		t.Error("the land lacks first strike")
	}
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(monastery)
		if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Insect") || !c.HasSubtype("Monk") || !c.HasColor("G") || !c.HasColor("W") {
			t.Errorf("not a green and white Insect Monk creature land: type line %q colors %v", c.TypeLine, c.Colors)
		}
	})
}

func TestGapB143IchorSlickShrinksAndHasCyclingAndMadness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Ichor Slick", "Sorcery", gapIchorSlick, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := gapB143PT(g, bear); p != 1 || th != 1 {
		t.Errorf("got %d/%d, want 1/1", p, th)
	}
	spec, _ := effects.Lookup(gapIchorSlick)
	if spec.Madness != "{3}{B}" || len(spec.Activated) != 1 || !spec.Activated[0].Cycling {
		t.Errorf("madness %q, activated %d", spec.Madness, len(spec.Activated))
	}
}

func TestGapB143UrborgUprisingReturnsUpToTwoAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	var gy []uuid.UUID
	for _, n := range []string{"One", "Two", "Three"} {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		gy = append(gy, id)
	}
	effectstest.SeedLibrary(me, "Draw")
	if err := effectstest.CastCatalogSpellErr(t, g, "Urborg Uprising", "Sorcery", gapUrborgUprising, effectstest.CardRefs(gy...)); err == nil {
		t.Fatal("three targets were accepted for 'up to two'")
	}
	// The refused cast left its card in hand; count from here.
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Urborg Uprising", "Sorcery", gapUrborgUprising, effectstest.CardRefs(gy[0], gy[1]))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gy[0]) || !me.Hand.Contains(gy[1]) || me.Hand.Contains(gy[2]) {
		t.Error("the two targeted creatures should be in hand and the third left")
	}
	if got := len(me.Hand.Cards); got != hand+3 {
		t.Errorf("hand %d, want %d (two returned plus a draw)", got, hand+3)
	}
}

func TestGapB143SaprolingSymbiosisMakesOnePerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.BlbB1Push(g, me.ID, "Cub", "Creature — Bear", "", 1, 1)
	effectstest.BlbB1Push(g, opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Saproling Symbiosis", "Sorcery", gapSaprolingSymbiosis, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.HasSubtype("Saproling") {
				n++
			}
		}
	})
	if n != 2 {
		t.Errorf("%d Saprolings, want 2 (one per creature you control)", n)
	}
}

func TestGapB143SaprolingSymbiosisFlashesOnlyForTheSurcharge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Saproling Symbiosis", TypeLine: "Sorcery", ManaCost: "{3}{G}",
		OracleID: gapSaprolingSymbiosis, Owner: me.ID, Controller: me.ID, Colors: []string{"G"},
	})
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G", "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast a sorcery in combat without paying for flash")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with the {2} surcharge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var n int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.HasSubtype("Saproling") {
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
}
