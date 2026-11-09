package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gatheringStoneOracle = "9f6d6e90-7462-46e5-bf2f-725b4badc394"

func TestGatheringStoneIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(gatheringStoneOracle)
	if !ok || spec.Name != "Gathering Stone" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// gatheringStoneTop puts a card on top of p's library.
func gatheringStoneTop(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// Cast: the type is named as it enters, and the enters trigger looks at
// the top card; a card of the chosen type is revealed and taken.
func TestGatheringStoneEntersNamesATypeAndTakesAMatchingTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := gatheringStoneTop(me, "Kinsbaile Cavalier", "Creature — Kithkin Knight")
	rfdCast(t, g, "Gathering Stone", "Artifact", gatheringStoneOracle, 0, 0)
	answerCreatureType(t, g, me.ID, "Kithkin")
	monSettle(t, g)
	answerMayChoice(t, g, me.ID, true)
	monSettle(t, g)
	if !me.Hand.Contains(top) {
		t.Fatal("the Kithkin on top did not go to the hand")
	}
}

// A card not of the chosen type is only offered to the graveyard.
func TestGatheringStoneUpkeepMillsANonMatchingCard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushNamedTribePermanent(t, g, me.ID, "Gathering Stone", "Artifact", gatheringStoneOracle, "Kithkin")
	advanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	top := gatheringStoneTop(me, "Boggart Brute", "Creature — Goblin Warrior")
	advanceToStepOf(t, g, seat, game.StepUpkeep)
	monSettle(t, g)
	c := latestConfirmFor(g, me.ID)
	if c == nil {
		t.Fatal("no offer at the upkeep")
	}
	answerMayChoice(t, g, me.ID, true)
	monSettle(t, g)
	if !me.Graveyard.Contains(top) {
		t.Fatal("the declined Goblin did not go to the graveyard")
	}
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("a second offer followed the graveyard answer")
	}
}

// Declining the hand moves on to the graveyard offer; declining that too
// leaves the card on top.
func TestGatheringStoneDeclinedTwiceLeavesTheCardOnTop(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushNamedTribePermanent(t, g, me.ID, "Gathering Stone", "Artifact", gatheringStoneOracle, "Kithkin")
	advanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	top := gatheringStoneTop(me, "Kinsbaile Cavalier", "Creature — Kithkin Knight")
	advanceToStepOf(t, g, seat, game.StepUpkeep)
	monSettle(t, g)
	answerMayChoice(t, g, me.ID, false)
	monSettle(t, g)
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatal("declining the hand did not offer the graveyard")
	}
	answerMayChoice(t, g, me.ID, false)
	monSettle(t, g)
	if got, err := me.Library.Top(); err != nil || got.InstanceID != top || me.Hand.Contains(top) {
		t.Error("the twice-declined card is not on top of the library")
	}
}

// Spells of the chosen type, of any card type, cost {1} less; others do
// not.
func TestGatheringStoneDiscountsSpellsOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushNamedTribePermanent(t, g, me.ID, "Gathering Stone", "Artifact", gatheringStoneOracle, "Kithkin")
	toMainForCost(t, g)

	other := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: other, Name: "Goblin Instant", TypeLine: "Kindred Instant — Goblin", ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	var short *game.InsufficientManaError
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true}); !errors.As(err, &short) {
		t.Fatalf("a Goblin spell for {W}: err = %v, want InsufficientManaError", err)
	}
	kith := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: kith, Name: "Kithkin Instant", TypeLine: "Kindred Instant — Kithkin", ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, kith, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a Kithkin spell for {W}: %v", err)
	}
}
