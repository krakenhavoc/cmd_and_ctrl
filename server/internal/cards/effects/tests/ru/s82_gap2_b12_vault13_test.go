package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_vault13_test.go — S82 GAP2 batch 12: Vault 13: Dweller's
// Journey, driven through the real cast, chapter and choice paths.

const gap2B12Vault13 = "586fb987-cd77-420f-978c-2ce140bcdcfb"

func TestGap2B12Vault13IsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12Vault13)
	if !ok {
		t.Fatal("Vault 13: Dweller's Journey is not registered")
	}
	if spec.Name != "Vault 13: Dweller's Journey" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// gap2B12AnswerPickTargets fills each clause of chapter I with the listed
// pick, in order, then declines the rest.
func gap2B12AnswerPickTargets(t *testing.T, g *game.Game, me uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	for _, id := range picks {
		effectstest.PickCard(t, g, me, id)
	}
	for effectstest.LatestPickTarget(g, me) != nil {
		if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me).ID, me, nil); err != nil {
			t.Fatalf("declining a clause: %v", err)
		}
	}
}

func TestGap2B12Vault13ExilesPerPlayerThenReturnsTwoAndTucksTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp1.ID, "Their Bear", 2, 2)
	ench := effectstest.PushTestPermanent(g, opp2.ID, game.Card{Name: "Their Charm", TypeLine: "Enchantment"})
	land := effectstest.PushTestPermanent(g, opp1.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	effectstest.CastCatalogSpell(t, g, "Vault 13: Dweller's Journey", "Enchantment — Saga", gap2B12Vault13, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no target prompt for chapter I: %+v", g.PendingChoices)
	}
	saga := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vault 13: Dweller's Journey" {
			saga = c.InstanceID
		}
	}
	if effectstest.HasID(prompt.PickTargetCards, saga) {
		t.Error("the Saga itself was offered: it says other")
	}
	if effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered on your own clause")
	}
	gap2B12AnswerPickTargets(t, g, me.ID, mine, theirs, ench)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, ench} {
		if g.Battlefield.Contains(id) {
			t.Fatal("chapter I did not exile one permanent per player")
		}
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("a land was exiled")
	}

	// Chapter II: gain 2 and scry 2.
	life := me.Life
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryIfAsked(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d after chapter II, want %d", me.Life, life+2)
	}

	// Chapter III: choose two of the three to return; the third goes to
	// the bottom of its owner's library.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatalf("no choice for chapter III: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{mine}); err == nil {
		t.Error("chapter III accepted one card, want exactly two")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{mine, ench}); err != nil {
		t.Fatalf("choosing two: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "My Bear", me.ID); n != 1 {
		t.Errorf("My Bear: %d on the battlefield, want 1", n)
	}
	if n := effectstest.NamedOnBattlefield(g, "Their Charm", opp2.ID); n != 1 {
		t.Errorf("Their Charm: %d under its owner's control, want 1", n)
	}
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp1.ID); n != 0 {
		t.Errorf("Their Bear came back (%d), want it on the bottom of its owner's library", n)
	}
	lib := opp1.Library.Cards
	if len(lib) == 0 || lib[0].Name != "Their Bear" {
		t.Error("Their Bear is not on the bottom of its owner's library")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vault 13: Dweller's Journey" {
			t.Error("the Saga was not sacrificed after chapter III")
		}
	}
}

func TestGap2B12Vault13ReturnsEverythingWhenItLeavesEarly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Vault 13: Dweller's Journey", "Enchantment — Saga", gap2B12Vault13, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatalf("no target prompt for chapter I: %+v", g.PendingChoices)
	}
	gap2B12AnswerPickTargets(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("chapter I did not exile the creature")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vault 13: Dweller's Journey" {
			effectstest.Destroy(t, g, c.InstanceID)
			break
		}
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		z := g.FindCardZoneForEffect(theirs)
		t.Logf("zone %+v pending %+v stack %d", z, g.PendingChoices, len(g.Stack.Cards))
		t.Errorf("Their Bear: %d back, want 1 once the Saga left", n)
	}
}
