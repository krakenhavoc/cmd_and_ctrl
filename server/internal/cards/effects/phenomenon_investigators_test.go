package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const phenomenonInvestigatorsOracle = "cc9088c4-6f6b-4a1a-90f1-794eb1e938c6"

func TestPhenomenonInvestigatorsIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(phenomenonInvestigatorsOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Phenomenon Investigators: %+v %v", spec, ok)
	}
}

func pushPhenomenonInvestigators(g *game.Game, owner uuid.UUID, option string) uuid.UUID {
	id := pushCatalogPermanent(g, owner, "Phenomenon Investigators", "Creature — Human Detective", phenomenonInvestigatorsOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].ChosenOption = option
			}
		}
	})
	return id
}

func horrorEnchantmentTokens(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype("Horror") && c.IsEnchantment() && c.IsCreature() {
			n++
		}
	}
	return n
}

func TestPhenomenonInvestigatorsBelieveMakesAHorrorWhenANontokenCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPhenomenonInvestigators(g, me.ID, "Believe")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := horrorEnchantmentTokens(g, me.ID); got != 1 {
		t.Fatalf("Horror tokens = %d, want 1", got)
	}
	// The token dying is a token: no second Horror.
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Horror") {
			killCreature(t, g, me.ID, c.InstanceID)
			break
		}
	}
	passPriorityAroundTable(t, g)
	if got := horrorEnchantmentTokens(g, me.ID); got != 0 {
		t.Fatalf("Horror tokens after a token died = %d, want 0", got)
	}
}

func TestPhenomenonInvestigatorsBelieveIgnoresOpponentsCreaturesAndDoubt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPhenomenonInvestigators(g, me.ID, "Believe")
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := horrorEnchantmentTokens(g, me.ID); got != 0 {
		t.Fatalf("an opponent's creature died: Horror tokens = %d, want 0", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushPhenomenonInvestigators(g2, me2.ID, "Doubt")
	bear := pushVanillaCreature(g2, me2.ID, "Bear", 2, 2)
	killCreature(t, g2, me2.ID, bear)
	passPriorityAroundTable(t, g2)
	if got := horrorEnchantmentTokens(g2, me2.ID); got != 0 {
		t.Fatalf("Doubt was chosen: Horror tokens = %d, want 0", got)
	}
}

func TestPhenomenonInvestigatorsDoubtReturnsAPermanentAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushPhenomenonInvestigators(g, me.ID, "Doubt")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	land := pushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no permanent choice: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	if hasID(pick.ChooseCards, land) {
		t.Error("a land is offered")
	}
	if !hasID(pick.ChooseCards, bear) {
		t.Error("a nonland permanent you own is not offered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the chosen permanent did not return to hand")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2 (the bounced card and a draw)", hand, me.Hand.Size())
	}
}

func TestPhenomenonInvestigatorsDoubtChoosingNothingDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushPhenomenonInvestigators(g, me.ID, "Doubt")
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); pick != nil {
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want unchanged", hand, me.Hand.Size())
	}
}
