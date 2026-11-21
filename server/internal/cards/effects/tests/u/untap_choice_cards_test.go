package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winter Orb: two tapped lands become a choice, and the creature that
// no cap counts untaps without being asked about.
func TestWinterOrbAsksWhichLandUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	forest := effectstest.PushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := effectstest.PushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	bear := effectstest.PushTappedForTest(g, owner.ID, "Grizzly Bears", "", "Creature — Bear")
	effectstest.PushPermanentForTest(g, owner.ID, "Winter Orb", effectstest.OracleWinterOrb, "Artifact")

	choice := effectstest.AdvanceToUntapChoiceOf(t, g, 1)
	if choice.ChooseMin != 1 || choice.ChooseMax != 1 {
		t.Fatalf("Winter Orb bounds = %d..%d, want 1..1", choice.ChooseMin, choice.ChooseMax)
	}
	if len(choice.ChooseCards) != 2 {
		t.Fatalf("Winter Orb offers %d permanents, want the two lands", len(choice.ChooseCards))
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("answer Winter Orb: %v", err)
	}
	if effectstest.TappedForTest(t, g, forest) {
		t.Error("the chosen land untapped")
	}
	if !effectstest.TappedForTest(t, g, island) {
		t.Error("the other land stays tapped under Winter Orb")
	}
	if effectstest.TappedForTest(t, g, bear) {
		t.Error("a creature is not a land and untaps as normal")
	}
}

// A tapped Winter Orb caps nothing: "as long as this artifact is
// untapped" is read when the step asks.
func TestWinterOrbTappedDoesNotCap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	forest := effectstest.PushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := effectstest.PushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	effectstest.PushTappedForTest(g, owner.ID, "Winter Orb", effectstest.OracleWinterOrb, "Artifact")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TappedForTest(t, g, forest) || effectstest.TappedForTest(t, g, island) {
		t.Error("a tapped Winter Orb capped the untap step")
	}
}

// Static Orb counts every permanent, and composes with Winter Orb:
// two permanents, at most one of them a land.
func TestStaticOrbAndWinterOrbCompose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	forest := effectstest.PushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := effectstest.PushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	bear := effectstest.PushTappedForTest(g, owner.ID, "Grizzly Bears", "", "Creature — Bear")
	effectstest.PushTappedForTest(g, owner.ID, "Ornithopter", "", "Artifact Creature — Thopter")
	effectstest.PushPermanentForTest(g, owner.ID, "Winter Orb", effectstest.OracleWinterOrb, "Artifact")
	effectstest.PushPermanentForTest(g, owner.ID, "Static Orb", effectstest.OracleStaticOrb, "Artifact")

	choice := effectstest.AdvanceToUntapChoiceOf(t, g, 1)
	if choice.ChooseMin != 1 || choice.ChooseMax != 2 {
		t.Fatalf("composed bounds = %d..%d, want 1..2", choice.ChooseMin, choice.ChooseMax)
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest, island}); err == nil {
		t.Fatal("two lands under Winter Orb was accepted")
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest, bear}); err != nil {
		t.Fatalf("one land and one creature was refused: %v", err)
	}
	if effectstest.TappedForTest(t, g, forest) || effectstest.TappedForTest(t, g, bear) {
		t.Error("the chosen permanents untapped")
	}
	if !effectstest.TappedForTest(t, g, island) {
		t.Error("the rest stay tapped under Static Orb")
	}
}

// Winter Moon counts NONBASIC lands, and caps whether it is tapped or
// untapped.
func TestWinterMoonCapsNonbasicLandsWhileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	basic := effectstest.PushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	nonbasic1 := effectstest.PushTappedForTest(g, owner.ID, "Command Tower", "", "Land")
	nonbasic2 := effectstest.PushTappedForTest(g, owner.ID, "Reliquary Tower", "", "Land")
	effectstest.PushTappedForTest(g, owner.ID, "Winter Moon", effectstest.OracleWinterMoon, "Artifact")

	choice := effectstest.AdvanceToUntapChoiceOf(t, g, 1)
	if len(choice.ChooseCards) != 2 {
		t.Fatalf("Winter Moon offers %d, want the two nonbasic lands", len(choice.ChooseCards))
	}
	for _, id := range choice.ChooseCards {
		if id == basic {
			t.Fatal("a basic land was counted by Winter Moon")
		}
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{nonbasic1}); err != nil {
		t.Fatalf("answer Winter Moon: %v", err)
	}
	if effectstest.TappedForTest(t, g, basic) {
		t.Error("the basic land untapped as normal")
	}
	if !effectstest.TappedForTest(t, g, nonbasic2) {
		t.Error("the second nonbasic land stays tapped")
	}
}

// "You may choose not to untap this" — both cards, both answers.
func TestOptOutCardsCanStayTapped(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
	}{
		{"Rust Tick", effectstest.OracleRustTick, "Artifact Creature — Insect"},
		{"Amber Prison", effectstest.OracleAmberPrison, "Artifact"},
	} {
		t.Run(tc.name+" declines", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner := g.Seats[1]
			id := effectstest.PushTappedForTest(g, owner.ID, tc.name, tc.oracle, tc.typeLine)
			other := effectstest.PushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")

			choice := effectstest.AdvanceToUntapChoiceOf(t, g, 1)
			if choice.ChooseMin != 0 {
				t.Fatalf("an opt-out-only prompt has floor %d, want 0", choice.ChooseMin)
			}
			if err := g.ResolveUntapChoice(choice.ID, owner.ID, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			if !effectstest.TappedForTest(t, g, id) {
				t.Errorf("%s untapped after its controller chose not to", tc.name)
			}
			if effectstest.TappedForTest(t, g, other) {
				t.Error("the permanent that was never in question untapped")
			}
		})
		t.Run(tc.name+" accepts", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner := g.Seats[1]
			id := effectstest.PushTappedForTest(g, owner.ID, tc.name, tc.oracle, tc.typeLine)

			choice := effectstest.AdvanceToUntapChoiceOf(t, g, 1)
			if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{id}); err != nil {
				t.Fatalf("accept: %v", err)
			}
			if effectstest.TappedForTest(t, g, id) {
				t.Errorf("%s stayed tapped after its controller chose to untap it", tc.name)
			}
		})
	}
}

// An untapped opt-out permanent is not in the CR 502.3 set at all, so
// it never asks — the commonest board for these cards.
func TestOptOutCardDoesNotAskWhenUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Amber Prison", effectstest.OracleAmberPrison, "Artifact")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceUntapChoice {
			t.Fatal("an untapped Amber Prison asked about untapping")
		}
	}
}
