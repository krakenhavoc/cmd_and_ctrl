package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const discipleOfBolasOracle = "8f2c8498-5b10-4ec7-8678-598c90987556"

// TestDiscipleOfBolasSacrificesForLifeAndCards casts Disciple of
// Bolas over a 5-power creature and checks the sacrifice prompt, the
// life gain and the draw all land off the sacrificed creature's
// power, not off Disciple's own.
func TestDiscipleOfBolasSacrificesForLifeAndCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})

	beforeLife, beforeHand := me.Life, me.Hand.Size()
	disciple := castCatalogSpell(t, g, "Disciple of Bolas", "Creature — Human Wizard", discipleOfBolasOracle, nil)
	passPriorityAroundTable(t, g)

	sac := sacrificeChoiceFor(g, me.ID)
	if sac == nil {
		t.Fatal("Disciple of Bolas should ask which other creature to sacrifice")
	}
	if len(sac.SacrificeOptions) != 1 || sac.SacrificeOptions[0] != beast {
		t.Errorf("sacrifice offered %v, want the one other creature %v", sac.SacrificeOptions, beast)
	}
	answerSacrifice(t, g, me.ID, beast)

	if !g.Battlefield.Contains(disciple) {
		t.Fatal("Disciple of Bolas is on the battlefield")
	}
	if !me.Graveyard.Contains(beast) {
		t.Error("the sacrificed creature is in the graveyard")
	}
	if got := me.Life; got != beforeLife+5 {
		t.Errorf("life %d, want %d (gain 5)", got, beforeLife+5)
	}
	if got := me.Hand.Size(); got != beforeHand+5 {
		t.Errorf("hand %d, want %d (draw 5)", got, beforeHand+5)
	}
}

// TestDiscipleOfBolasWithNoOtherCreatureDoesNothing — the trigger
// must not error when there is nothing else to sacrifice.
func TestDiscipleOfBolasWithNoOtherCreatureDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	beforeLife, beforeHand := me.Life, me.Hand.Size()
	disciple := castCatalogSpell(t, g, "Disciple of Bolas", "Creature — Human Wizard", discipleOfBolasOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(disciple) {
		t.Fatal("Disciple of Bolas is on the battlefield")
	}
	if got := me.Life; got != beforeLife {
		t.Errorf("life %d, want %d (nothing to sacrifice, nothing gained)", got, beforeLife)
	}
	if got := me.Hand.Size(); got != beforeHand {
		t.Errorf("hand %d, want %d (nothing to sacrifice, nothing drawn)", got, beforeHand)
	}
}
