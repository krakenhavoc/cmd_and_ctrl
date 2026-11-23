package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2942: "up to two basic land cards" is zero, one or two. The prompt
// to the battlefield and the prompt to the hand are each optional.
func TestCultivateAndKodamasReachTakeZeroOneOrTwo(t *testing.T) {
	cards := []struct{ name, typeLine, oracle string }{
		{"Cultivate", "Sorcery", effectstest.C2875CultivateOracle},
		{"Kodama's Reach", "Sorcery — Arcane", effectstest.C2875KodamaOracle},
	}
	// bf and hand are indexes into the three Forests, -1 for declining.
	for _, card := range cards {
		for _, tc := range []struct {
			name         string
			bf, hand     int
			wantBf, want int
		}{
			{"zero", -1, -1, 0, 0},
			{"one to the battlefield", 0, -1, 1, 0},
			{"one to the hand", -1, 1, 0, 1},
			{"two", 0, 1, 1, 1},
		} {
			t.Run(card.name+" "+tc.name, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				me.Library.Cards = nil
				var forests []uuid.UUID
				for i := 0; i < 3; i++ {
					forests = append(forests, effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
				}
				effectstest.PushLibraryCardForTest(me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"})
				handBefore := me.Hand.Size()

				effectstest.CastCatalogSpell(t, g, card.name, card.typeLine, card.oracle, nil)
				effectstest.PassPriorityAroundTable(t, g)
				effectstest.AnswerOrDecline(t, g, me.ID, tc.bf, forests)
				effectstest.AnswerOrDecline(t, g, me.ID, tc.hand, forests)

				onField, inHand := 0, 0
				for _, id := range forests {
					if g.Battlefield.Contains(id) {
						onField++
						if !effectstest.IsTapped(g, id) {
							t.Error("a Cultivated land entered untapped")
						}
					}
					if me.Hand.Contains(id) {
						inHand++
					}
				}
				if onField != tc.wantBf || inHand != tc.want {
					t.Errorf("battlefield %d hand %d, want %d and %d", onField, inHand, tc.wantBf, tc.want)
				}
				if got := me.Hand.Size() - handBefore; got != tc.want {
					t.Errorf("hand grew by %d, want %d", got, tc.want)
				}
				if effectstest.SearchChoiceFor(g, me.ID) != nil {
					t.Error("a search prompt is still open")
				}
			})
		}
	}
}

// #2942: with exactly two basics the second used to be taken without
// being asked about; it can now be declined.
func TestCultivateTwoBasicsLetsYouKeepTheSecondInTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	b := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Cultivate", "Sorcery", effectstest.C2875CultivateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, a)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the second basic was taken without a prompt")
	}
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if !g.Battlefield.Contains(a) || me.Hand.Contains(b) || !me.Library.Contains(b) {
		t.Error("the declined Forest should have stayed in the library")
	}
}

// #2941: the card is chosen after the mill, so one the spell milled
// can be the one that comes back.
func TestGrappleWithThePastCanReturnACardItJustMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	rock := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: rock, Name: "Milled Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	old := effectstest.B17GraveyardCard(me, "Old Land", "Land", "")

	effectstest.CastCatalogSpell(t, g, "Grapple with the Past", "Instant", effectstest.B20GrappleWithThePastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt after the mill")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[milled] || !offered[old] {
		t.Errorf("offered %v, want the milled bear and the old land", pick.ChooseCards)
	}
	if offered[rock] {
		t.Error("an artifact card is not a legal pick")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (\"you may\", one card)", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
	if !me.Hand.Contains(milled) || me.Graveyard.Contains(milled) {
		t.Error("the milled creature card should be in hand")
	}
}

// #2941: "you may" — declining returns nothing; with nothing eligible
// there is no prompt; and the spell targets nothing.
func TestGrappleWithThePastMayDeclineAndTargetsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spec, _ := effects.Lookup(effectstest.B20GrappleWithThePastOracle)
	if spec.Completeness != effects.CompletenessFull || spec.Targets != nil {
		t.Errorf("Grapple should be Full and untargeted, got %v / %v", spec.Completeness, spec.Targets)
	}
	old := effectstest.B17GraveyardCard(me, "Old Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Grapple with the Past", "Instant", effectstest.B20GrappleWithThePastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if me.Hand.Contains(old) || !me.Graveyard.Contains(old) {
		t.Error("declining should leave the graveyard alone")
	}

	g2 := effectstest.NewCatalogGame(t)
	you := g2.Seats[0]
	you.Library.Cards = nil
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryCardForTest(you, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	effectstest.CastCatalogSpell(t, g2, "Grapple with the Past", "Instant", effectstest.B20GrappleWithThePastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.ChooseCardsChoiceFor(g2, you.ID) != nil {
		t.Error("no creature or land card to return, yet a prompt opened")
	}
}
