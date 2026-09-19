package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_choice_cards_test.go — #826 / ADR 0070 through the real
// catalog: Winter Orb, Static Orb, Winter Moon (caps) and Rust Tick /
// Amber Prison (opt-outs).

const (
	oracleWinterOrb   = "1dcbd583-3388-4b34-a7cd-131648aa6abd"
	oracleStaticOrb   = "0004ebd0-dfd6-4276-b4a6-de0003e94237"
	oracleWinterMoon  = "b922f057-1c91-43eb-b74a-8a933b1be2d2"
	oracleRustTick    = "c7f20899-2625-4b3a-8bd8-0bcee07ed86e"
	oracleAmberPrison = "1c69fdcf-ba87-480a-88df-70aa4ec9fff0"
)

// pushTappedForTest parks a tapped permanent on the battlefield.
func pushTappedForTest(g *game.Game, owner uuid.UUID, name, oracleID, typeLine string) uuid.UUID {
	id := pushPermanentForTest(g, owner, name, oracleID, typeLine)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	return id
}

// advanceToUntapChoiceOf walks the cursor until the named seat's untap
// step stops for its CR 502.3 prompt. AdvanceStep is refused once the
// prompt is open (the gate), so the loop stops on the prompt itself.
func advanceToUntapChoiceOf(t *testing.T, g *game.Game, seat int) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 300; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceUntapChoice {
				if g.Turn.ActiveSeat != seat {
					t.Fatalf("untap prompt opened on seat %d, want %d", g.Turn.ActiveSeat, seat)
				}
				return c
			}
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward seat %d's untap choice: %v", seat, err)
		}
	}
	t.Fatalf("seat %d's untap step never asked", seat)
	return nil
}

func tappedForTest(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	return c.Tapped
}

// Winter Orb: two tapped lands become a choice, and the creature that
// no cap counts untaps without being asked about.
func TestWinterOrbAsksWhichLandUntaps(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	forest := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := pushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	bear := pushTappedForTest(g, owner.ID, "Grizzly Bears", "", "Creature — Bear")
	pushPermanentForTest(g, owner.ID, "Winter Orb", oracleWinterOrb, "Artifact")

	choice := advanceToUntapChoiceOf(t, g, 1)
	if choice.ChooseMin != 1 || choice.ChooseMax != 1 {
		t.Fatalf("Winter Orb bounds = %d..%d, want 1..1", choice.ChooseMin, choice.ChooseMax)
	}
	if len(choice.ChooseCards) != 2 {
		t.Fatalf("Winter Orb offers %d permanents, want the two lands", len(choice.ChooseCards))
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("answer Winter Orb: %v", err)
	}
	if tappedForTest(t, g, forest) {
		t.Error("the chosen land untapped")
	}
	if !tappedForTest(t, g, island) {
		t.Error("the other land stays tapped under Winter Orb")
	}
	if tappedForTest(t, g, bear) {
		t.Error("a creature is not a land and untaps as normal")
	}
}

// A tapped Winter Orb caps nothing: "as long as this artifact is
// untapped" is read when the step asks.
func TestWinterOrbTappedDoesNotCap(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	forest := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := pushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	pushTappedForTest(g, owner.ID, "Winter Orb", oracleWinterOrb, "Artifact")

	advanceToUpkeepOf(t, g, 1)
	if tappedForTest(t, g, forest) || tappedForTest(t, g, island) {
		t.Error("a tapped Winter Orb capped the untap step")
	}
}

// Static Orb counts every permanent, and composes with Winter Orb:
// two permanents, at most one of them a land.
func TestStaticOrbAndWinterOrbCompose(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	forest := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := pushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	bear := pushTappedForTest(g, owner.ID, "Grizzly Bears", "", "Creature — Bear")
	pushTappedForTest(g, owner.ID, "Ornithopter", "", "Artifact Creature — Thopter")
	pushPermanentForTest(g, owner.ID, "Winter Orb", oracleWinterOrb, "Artifact")
	pushPermanentForTest(g, owner.ID, "Static Orb", oracleStaticOrb, "Artifact")

	choice := advanceToUntapChoiceOf(t, g, 1)
	if choice.ChooseMin != 1 || choice.ChooseMax != 2 {
		t.Fatalf("composed bounds = %d..%d, want 1..2", choice.ChooseMin, choice.ChooseMax)
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest, island}); err == nil {
		t.Fatal("two lands under Winter Orb was accepted")
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest, bear}); err != nil {
		t.Fatalf("one land and one creature was refused: %v", err)
	}
	if tappedForTest(t, g, forest) || tappedForTest(t, g, bear) {
		t.Error("the chosen permanents untapped")
	}
	if !tappedForTest(t, g, island) {
		t.Error("the rest stay tapped under Static Orb")
	}
}

// Winter Moon counts NONBASIC lands, and caps whether it is tapped or
// untapped.
func TestWinterMoonCapsNonbasicLandsWhileTapped(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	basic := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	nonbasic1 := pushTappedForTest(g, owner.ID, "Command Tower", "", "Land")
	nonbasic2 := pushTappedForTest(g, owner.ID, "Reliquary Tower", "", "Land")
	pushTappedForTest(g, owner.ID, "Winter Moon", oracleWinterMoon, "Artifact")

	choice := advanceToUntapChoiceOf(t, g, 1)
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
	if tappedForTest(t, g, basic) {
		t.Error("the basic land untapped as normal")
	}
	if !tappedForTest(t, g, nonbasic2) {
		t.Error("the second nonbasic land stays tapped")
	}
}

// "You may choose not to untap this" — both cards, both answers.
func TestOptOutCardsCanStayTapped(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
	}{
		{"Rust Tick", oracleRustTick, "Artifact Creature — Insect"},
		{"Amber Prison", oracleAmberPrison, "Artifact"},
	} {
		t.Run(tc.name+" declines", func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[1]
			id := pushTappedForTest(g, owner.ID, tc.name, tc.oracle, tc.typeLine)
			other := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")

			choice := advanceToUntapChoiceOf(t, g, 1)
			if choice.ChooseMin != 0 {
				t.Fatalf("an opt-out-only prompt has floor %d, want 0", choice.ChooseMin)
			}
			if err := g.ResolveUntapChoice(choice.ID, owner.ID, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			if !tappedForTest(t, g, id) {
				t.Errorf("%s untapped after its controller chose not to", tc.name)
			}
			if tappedForTest(t, g, other) {
				t.Error("the permanent that was never in question untapped")
			}
		})
		t.Run(tc.name+" accepts", func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[1]
			id := pushTappedForTest(g, owner.ID, tc.name, tc.oracle, tc.typeLine)

			choice := advanceToUntapChoiceOf(t, g, 1)
			if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{id}); err != nil {
				t.Fatalf("accept: %v", err)
			}
			if tappedForTest(t, g, id) {
				t.Errorf("%s stayed tapped after its controller chose to untap it", tc.name)
			}
		})
	}
}

// An untapped opt-out permanent is not in the CR 502.3 set at all, so
// it never asks — the commonest board for these cards.
func TestOptOutCardDoesNotAskWhenUntapped(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Amber Prison", oracleAmberPrison, "Artifact")
	advanceToUpkeepOf(t, g, 1)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceUntapChoice {
			t.Fatal("an untapped Amber Prison asked about untapping")
		}
	}
}
