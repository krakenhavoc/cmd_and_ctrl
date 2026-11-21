package testsg

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gatheringStoneOracle = "9f6d6e90-7462-46e5-bf2f-725b4badc394"

func TestGatheringStoneIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(gatheringStoneOracle)
	if !ok || spec.Name != "Gathering Stone" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := gatheringStoneTop(me, "Kinsbaile Cavalier", "Creature — Kithkin Knight")
	effectstest.RfdCast(t, g, "Gathering Stone", "Artifact", gatheringStoneOracle, 0, 0)
	effectstest.AnswerCreatureType(t, g, me.ID, "Kithkin")
	effectstest.MonSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	if !me.Hand.Contains(top) {
		t.Fatal("the Kithkin on top did not go to the hand")
	}
}

// A card not of the chosen type is only offered to the graveyard.
func TestGatheringStoneUpkeepMillsANonMatchingCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushNamedTribePermanent(t, g, me.ID, "Gathering Stone", "Artifact", gatheringStoneOracle, "Kithkin")
	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	top := gatheringStoneTop(me, "Boggart Brute", "Creature — Goblin Warrior")
	effectstest.AdvanceToStepOf(t, g, seat, game.StepUpkeep)
	effectstest.MonSettle(t, g)
	c := effectstest.LatestConfirmFor(g, me.ID)
	if c == nil {
		t.Fatal("no offer at the upkeep")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	if !me.Graveyard.Contains(top) {
		t.Fatal("the declined Goblin did not go to the graveyard")
	}
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("a second offer followed the graveyard answer")
	}
}

// Declining the hand moves on to the graveyard offer; declining that too
// leaves the card on top.
func TestGatheringStoneDeclinedTwiceLeavesTheCardOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushNamedTribePermanent(t, g, me.ID, "Gathering Stone", "Artifact", gatheringStoneOracle, "Kithkin")
	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	top := gatheringStoneTop(me, "Kinsbaile Cavalier", "Creature — Kithkin Knight")
	effectstest.AdvanceToStepOf(t, g, seat, game.StepUpkeep)
	effectstest.MonSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.MonSettle(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) == nil {
		t.Fatal("declining the hand did not offer the graveyard")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.MonSettle(t, g)
	if got, err := me.Library.Top(); err != nil || got.InstanceID != top || me.Hand.Contains(top) {
		t.Error("the twice-declined card is not on top of the library")
	}
}

// Spells of the chosen type, of any card type, cost {1} less; others do
// not.
func TestGatheringStoneDiscountsSpellsOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushNamedTribePermanent(t, g, me.ID, "Gathering Stone", "Artifact", gatheringStoneOracle, "Kithkin")
	effectstest.ToMainForCost(t, g)

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
