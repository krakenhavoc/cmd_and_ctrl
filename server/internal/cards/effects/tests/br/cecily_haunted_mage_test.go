package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cecilyTable seeds Cecily under seat 0 with `inHand` cards in hand (one
// of them the sorcery returned, the rest lands) and a card on top of the
// library, declares her attack, and settles the trigger in the declare
// attackers step.
func cecilyTable(t *testing.T, inHand int) (*game.Game, *game.Player, uuid.UUID, int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cecily := effectstest.PushCatalogPermanent(g, me.ID, "Cecily, Haunted Mage", "Legendary Creature — Human Wizard", effectstest.CecilyOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	sorcery := effectstest.HandCardFull(me, "Fodder Sorcery", "Sorcery", "{3}{B}", "", []string{"B"})
	for i := 1; i < inHand; i++ {
		effectstest.HandCardFull(me, "Swamp", "Basic Land — Swamp", "", "", nil)
	}
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Drawn", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, cecily)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, sorcery, life
}

func TestCecilyWithElevenCardsCastsASorceryFreeInCombat(t *testing.T) {
	g, me, sorcery, life := cecilyTable(t, 10)
	if me.Life != life-1 {
		t.Errorf("you lose 1 life: %d → %d", life, me.Life)
	}
	if me.Hand.Size() != 11 {
		t.Fatalf("drew to eleven: %d", me.Hand.Size())
	}
	// #4250: the free cast is offered inside the trigger's resolution.
	p := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, sorcery)
	if len(p.DiscardOptions) != 1 {
		t.Fatalf("only the instant or sorcery is offered: %v", p.DiscardOptions)
	}
	if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
		t.Fatalf("the test wants a combat step, got %v", g.Turn.Step)
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a sorcery cast free during combat: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, sorcery)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sorcery) {
		t.Error("the free sorcery resolved")
	}
}

func TestCecilyWithFewerThanElevenCardsOffersNothing(t *testing.T) {
	g, me, _, life := cecilyTable(t, 9)
	if me.Life != life-1 || me.Hand.Size() != 10 {
		t.Errorf("drew and lost 1: hand %d life %d → %d", me.Hand.Size(), life, me.Life)
	}
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); p != nil {
		t.Errorf("ten cards in hand offers no free cast: %+v", p)
	}
	if p := effectstest.ResolutionCastOfferFor(g, me.ID); p != nil {
		t.Errorf("ten cards in hand offers no free cast: %+v", p)
	}
}

func TestCecilyMaximumHandSizeIsEleven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Cecily, Haunted Mage", "Legendary Creature — Human Wizard", effectstest.CecilyOracle, false)
	if got := effectstest.MhMax(g, me); got != 11 {
		t.Errorf("maximum hand size = %d, want 11", got)
	}
}
