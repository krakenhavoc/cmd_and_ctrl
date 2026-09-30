package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const palantirOfOrthancOracle = "7798a4d9-1c1d-48e6-b9a2-ecd6aec1efa7"

func pushPalantir(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Palantír of Orthanc", TypeLine: "Legendary Artifact",
		OracleID: palantirOfOrthancOracle, Owner: owner, Controller: owner,
	})
}

// TestPalantirYesBranchDrawsACard: the opponent agrees to let the
// controller draw, so nothing is milled and nobody loses life.
func TestPalantirYesBranchDrawsACard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Bottom Me", "Bottom Too")
	palantir := pushPalantir(g, me.ID)
	hand := me.Hand.Size()
	oppLife := opp.Life

	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt for the trigger's own target opponent: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := counterCount(g, palantir, "influence"); got != 1 {
		t.Fatalf("an influence counter should land first: %d, want 1", got)
	}
	sc := scryChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatalf("no scry-2 prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveScry(sc.ID, me.ID, sc.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}

	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if ask == nil {
		t.Fatalf("no yes/no prompt for the opponent: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm(yes): %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+1 {
		t.Errorf("the yes branch should draw a card: hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if opp.Life != oppLife {
		t.Errorf("the yes branch should not cost the opponent any life: %d, want %d", opp.Life, oppLife)
	}
}

// TestPalantirNoBranchMillsAndBurns: the opponent declines, so the
// controller mills X (the influence-counter count) and the opponent
// loses life equal to the total mana value milled.
func TestPalantirNoBranchMillsAndBurns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Library.PushTop(game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{B}"})
	seedLibrary(me, "Bottom Me", "Bottom Too") // scry-2 fodder sits above the mill target
	palantir := pushPalantir(g, me.ID)
	lib := me.Library.Size()
	oppLife := opp.Life

	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt for the trigger's own target opponent: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	sc := scryChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatalf("no scry-2 prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveScry(sc.ID, me.ID, sc.ScryCards, nil); err != nil { // bottom both
		t.Fatalf("ResolveScry: %v", err)
	}

	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if ask == nil {
		t.Fatalf("no yes/no prompt for the opponent: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm(no): %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := counterCount(g, palantir, "influence"); got != 1 {
		t.Fatalf("influence counters after one activation: %d, want 1", got)
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("the no branch should mill X=1: library %d, want %d", me.Library.Size(), lib-1)
	}
	if opp.Life != oppLife-3 {
		t.Errorf("the opponent should lose life equal to the milled card's mana value: %d, want %d",
			opp.Life, oppLife-3)
	}
}
