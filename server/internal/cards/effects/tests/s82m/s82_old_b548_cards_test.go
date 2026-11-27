package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b548_cards_test.go — S82 OLD batch 548 (#5593, Lorwyn: changeling):
// Wings of Velis Vel, Woodland Changeling, Ghostly Changeling, Fire-Belly
// Changeling, Changeling Berserker, Blades of Velis Vel, Ego Erasure,
// Changeling Titan and Changeling Hero. Each test plays the real catalog entry
// through the engine.

const (
	oldB548Wings     = "0bb2b3cc-d3e6-40e9-a5da-d589e8e8f0bd"
	oldB548Woodland  = "338f0ab6-f744-4e5a-a358-05e71b93758d"
	oldB548Ghostly   = "4e779540-e667-457e-9071-4cbbd777d9a9"
	oldB548FireBelly = "65c1cac7-87d1-4687-86ba-f7c63c4f8549"
	oldB548Berserker = "6a3dee57-9912-4c51-872e-2ec47acdaf52"
	oldB548Blades    = "9955885b-fad7-4ee6-bd62-48d8a7731044"
	oldB548Ego       = "a7a47969-b29a-4f99-9500-29c59af43c48"
	oldB548Titan     = "c1409a11-3780-40e4-abe6-df53d26b2dd6"
	oldB548Hero      = "fc893d4a-debd-4df0-ae83-d3ebd2f0fefe"
)

func TestOldB548EveryCardIsRegisteredFullAndChangeling(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB548Wings: "Wings of Velis Vel", oldB548Woodland: "Woodland Changeling", oldB548Ghostly: "Ghostly Changeling", oldB548FireBelly: "Fire-Belly Changeling", oldB548Berserker: "Changeling Berserker", oldB548Blades: "Blades of Velis Vel", oldB548Ego: "Ego Erasure", oldB548Titan: "Changeling Titan", oldB548Hero: "Changeling Hero",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if !slices.Contains(spec.PrintedKeywords, game.KeywordChangeling) {
			t.Errorf("%s lacks the changeling keyword", name)
		}
	}
}

func oldB548Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB548Changeling(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Changeling", TypeLine: "Creature — Shapeshifter", Power: power, Toughness: toughness,
		Keywords: []string{game.KeywordChangeling}, Owner: owner, Controller: owner,
	})
}

func oldB548Types(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return effectstest.EffectiveCardForTest(t, g, id)
}

// --- Woodland Changeling ---------------------------------------------------------

func TestOldB548WoodlandChangelingIsEveryCreatureType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Woodland Changeling", TypeLine: "Creature — Shapeshifter", OracleID: oldB548Woodland,
		Power: 2, Toughness: 2, Keywords: []string{game.KeywordChangeling}, Owner: me.ID, Controller: me.ID,
	})
	c := oldB548Types(t, g, id)
	for _, st := range []string{"Elf", "Goblin", "Kithkin", "Treefolk"} {
		if !c.HasSubtype(st) {
			t.Errorf("a changeling is not a %s", st)
		}
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 2 {
		t.Errorf("it is %d/%d, want 2/2", p, tough)
	}
}

// --- Ghostly Changeling / Fire-Belly Changeling ------------------------------------

func TestOldB548GhostlyChangelingPumpsPlusOnePlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := effectstest.B12Push(g, me.ID, "Ghostly Changeling", "Creature — Shapeshifter", oldB548Ghostly, 2, 2)
	effectstest.B06AddMana(me, "B", "B", "B", "B")
	effectstest.ApaActivateNoPoolCheck(t, g, me, c, 0)
	effectstest.ApaActivateNoPoolCheck(t, g, me, c, 0)
	if p, tough := effectstest.PtOf(t, g, c); p != 4 || tough != 4 {
		t.Errorf("it is %d/%d, want 4/4", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if p, tough := effectstest.PtOf(t, g, c); p != 2 || tough != 2 {
		t.Errorf("the pump outlasted the turn: %d/%d", p, tough)
	}
}

func TestOldB548FireBellyChangelingActivatesNoMoreThanTwiceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := effectstest.B12Push(g, me.ID, "Fire-Belly Changeling", "Creature — Shapeshifter", oldB548FireBelly, 1, 1)
	effectstest.B06AddMana(me, "R", "R", "R")
	effectstest.ApaActivateNoPoolCheck(t, g, me, c, 0)
	effectstest.ApaActivateNoPoolCheck(t, g, me, c, 0)
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a third activation in one turn was allowed")
	}
	if p, _ := effectstest.PtOf(t, g, c); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
}

// --- Champion: Changeling Titan / Hero / Berserker -----------------------------------

func oldB548CastChampion(t *testing.T, g *game.Game, name, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Shapeshifter", OracleID: oracle,
		Power: p, Toughness: tough, Keywords: []string{game.KeywordChangeling}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB548ChampionExilesAnyCreatureAndReturnsItWhenTheChampionLeaves(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		p, tough     int
	}{
		{"Changeling Titan", oldB548Titan, 7, 7},
		{"Changeling Hero", oldB548Hero, 4, 4},
		{"Changeling Berserker", oldB548Berserker, 5, 3},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
		champ := oldB548CastChampion(t, g, tc.name, tc.oracle, tc.p, tc.tough)
		pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatalf("%s: no champion prompt", tc.name)
		}
		if slices.Contains(pick.ChooseCards, champ) || slices.Contains(pick.ChooseCards, theirs) || !slices.Contains(pick.ChooseCards, bear) {
			t.Fatalf("%s: the prompt offers %v, want only my other creature", tc.name, pick.ChooseCards)
		}
		effectstest.AnswerChooseCards(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(champ) {
			t.Fatalf("%s was sacrificed despite championing a creature", tc.name)
		}
		if g.Battlefield.Contains(bear) {
			t.Errorf("%s: the championed Bear is still on the battlefield", tc.name)
		}
		effectstest.KillCreature(t, g, me.ID, champ)
		effectstest.PassPriorityAroundTable(t, g)
		returned := 0
		for _, bc := range g.Battlefield.Cards {
			if bc.Name == "Bear" && bc.Controller == me.ID {
				returned++
			}
		}
		if returned != 1 {
			t.Errorf("%s: %d Bears on the battlefield after the champion left, want the championed one back", tc.name, returned)
		}
	}
}

func TestOldB548ChampionIsSacrificedWithNoOtherCreatureOrWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	titan := oldB548CastChampion(t, g, "Changeling Titan", oldB548Titan, 7, 7)
	if g.Battlefield.Contains(titan) {
		t.Error("the Titan stayed with no other creature to champion")
	}
	g2 := effectstest.NewCatalogGame(t)
	me := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	bear := effectstest.B12Creature(g2, me.ID, "Bear", "Creature — Bear", 2, 2)
	hero := oldB548CastChampion(t, g2, "Changeling Hero", oldB548Hero, 4, 4)
	effectstest.AnswerChooseCards(t, g2, me.ID)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(hero) {
		t.Error("the Hero stayed although its controller exiled nothing")
	}
	if !g2.Battlefield.Contains(bear) {
		t.Error("the Bear left although it was not championed")
	}
}

// --- Wings of Velis Vel ------------------------------------------------------------

func TestOldB548WingsMakesAFourByFourFlyingEveryTypeUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mouse := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Wings of Velis Vel", "Kindred Instant — Shapeshifter", oldB548Wings, oldB548Card(mouse))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mouse); p != 4 || tough != 4 {
		t.Errorf("the Mouse is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOf(g, mouse, "flying") {
		t.Error("the Mouse did not gain flying")
	}
	if c := oldB548Types(t, g, mouse); !c.HasSubtype("Elf") || !c.HasSubtype("Goblin") {
		t.Error("the Mouse did not gain all creature types")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if p, tough := effectstest.PtOf(t, g, mouse); p != 1 || tough != 1 {
		t.Errorf("the Mouse is %d/%d after the turn, want 1/1", p, tough)
	}
	if effectstest.HasKeywordOf(g, mouse, "flying") || oldB548Types(t, g, mouse).HasSubtype("Elf") {
		t.Error("flying or the types outlasted the turn")
	}
}

// --- Blades of Velis Vel -----------------------------------------------------------

func TestOldB548BladesPumpsUpToTwoCreaturesAndGivesThemEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blades of Velis Vel", "Kindred Instant — Shapeshifter", oldB548Blades,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 2 {
			t.Errorf("a target is %d/%d, want 4/2", p, tough)
		}
		if !oldB548Types(t, g, id).HasSubtype("Merfolk") {
			t.Error("a target did not gain all creature types")
		}
	}
	if p, _ := effectstest.PtOf(t, g, c); p != 2 || oldB548Types(t, g, c).HasSubtype("Merfolk") {
		t.Error("an untargeted creature was affected")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Blades of Velis Vel", "Kindred Instant — Shapeshifter", oldB548Blades,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err == nil {
		t.Fatal("cast with three targets")
	}
}

func TestOldB548BladesCanBeCastWithOneTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blades of Velis Vel", "Kindred Instant — Shapeshifter", oldB548Blades, oldB548Card(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}

// --- Ego Erasure -------------------------------------------------------------------

func TestOldB548EgoErasureShrinksATableOfCreaturesAndStripsTheirTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 3, 3)
	shifter := oldB548Changeling(g, opp.ID, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf Warrior", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Ego Erasure", "Kindred Instant — Shapeshifter", oldB548Ego, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, elf); p != 1 || tough != 3 {
		t.Errorf("their Elf is %d/%d, want 1/3", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, shifter); p != 0 {
		t.Errorf("their changeling has power %d, want 0", p)
	}
	if oldB548Types(t, g, elf).HasSubtype("Elf") {
		t.Error("their Elf is still an Elf")
	}
	if c := oldB548Types(t, g, shifter); c.HasSubtype("Elf") || c.HasSubtype("Goblin") {
		t.Error("their changeling is still every creature type")
	}
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 || !oldB548Types(t, g, mine).HasSubtype("Elf") {
		t.Error("my own creature was affected")
	}
	later := effectstest.B12Creature(g, opp.ID, "Later Elf", "Creature — Elf Warrior", 3, 3)
	if p, _ := effectstest.PtOf(t, g, later); p != 3 || !oldB548Types(t, g, later).HasSubtype("Elf") {
		t.Error("a creature that arrived afterwards was affected")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if p, _ := effectstest.PtOf(t, g, elf); p != 3 || !oldB548Types(t, g, elf).HasSubtype("Elf") {
		t.Error("the effect outlasted the turn")
	}
}
