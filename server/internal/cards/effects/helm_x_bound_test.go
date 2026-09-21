package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// helm_x_bound_test.go — #1161, the OTHER half of Helm of Obedience's
// printed sentence.
//
//	"Target opponent mills a card, then repeats this process until a
//	 creature card OR X CARDS have been put into their graveyard this
//	 way, whichever comes first."
//
// #1159 made the creature half read what ARRIVED (CR 400.7). The X
// half still counted what came OFF THE LIBRARY, because it was written
// as the mill's AMOUNT — so a card a replacement diverted on the way
// spent one of the X without ever being put into that graveyard, and
// the run stopped one card short of what the card says.
//
// It is one clause with two stop conditions now, both over the landed
// list: UntilAny(UntilCard(creature), UntilCount(X)). These are the
// three boards where the two readings disagree.

// helmActivation activates the Helm for X against `victim` and lets the
// ability resolve.
func helmActivation(t *testing.T, g *game.Game, me *game.Player, helm uuid.UUID, victim *game.Player, x int) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		XValue:  x,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("activate the Helm at X=%d: %v", x, err)
	}
	passPriorityAroundTable(t, g)
}

// TestHelmsXDoesNotCountACommanderThatTookTheCommandZone is the bug on
// the smallest board that shows it: X=2, and the first card off the
// library never reaches the graveyard.
//
// CR 903.9 takes the commander to the command zone, so it was not put
// into that graveyard this way — it does not end the run (#1159) and
// it does not spend one of the X (#1161). The Helm mills another card
// in its place and two real cards are in the graveyard when it stops.
// Before the fix the bound was the mill's amount, the plan was two
// cards long, and the run ended one card short with ONE card in the
// graveyard.
func TestHelmsXDoesNotCountACommanderThatTookTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fillPool(me, 5)
	helm := pushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", helmOfObedienceOracle, false)

	// PushTop, so the LAST push is milled FIRST: the commander comes
	// off the top, then the two fillers behind it.
	second := libraryCardFor(opp, "Filler B", "Sorcery")
	first := libraryCardFor(opp, "Filler A", "Sorcery")
	commander := libraryCardFor(opp, "Their Commander", "Legendary Creature — Beast")
	markCommanderCard(t, g, opp, commander)
	graveBefore := opp.Graveyard.Size()

	helmActivation(t, g, me, helm, opp, 2)
	b21AcceptCommandZone(t, g, opp.ID)

	if !opp.Command.Contains(commander) {
		t.Fatal("setup: accepting did not put the commander into the command zone")
	}
	if got := opp.Graveyard.Size() - graveBefore; got != 2 {
		t.Errorf("%d cards reached the graveyard, want X=2 — the diverted commander was never "+
			"put into that graveyard this way, so it does not count toward X", got)
	}
	if !opp.Graveyard.Contains(first) || !opp.Graveyard.Contains(second) {
		t.Error("the two cards that count are the two that arrived")
	}
	if !g.Battlefield.Contains(helm) {
		t.Error("no CREATURE card reached the graveyard, so the Helm is not sacrificed — " +
			"the commander that left the library was not put there")
	}
}

// TestHelmUnderRestInPeaceMillsTheWholeLibrary is the famous combo, and
// it is the same reading taken to its end: "if a card would be put into
// a graveyard from anywhere, exile it instead" means NOTHING is ever put
// into that graveyard this way, so the run never reaches X and never
// finds its creature card, and it ends where every mill ends — at the
// bottom of the library (CR 701.13b, no loss).
//
// Before the fix X was the mill's amount and the Helm milled exactly
// two cards with Rest in Peace on the battlefield.
func TestHelmUnderRestInPeaceMillsTheWholeLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castRestInPeace(t, g)

	fillPool(me, 5)
	helm := pushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", helmOfObedienceOracle, false)
	// A creature sits on top: under Rest in Peace it is exiled rather
	// than put into the graveyard, so it does not end the run either.
	beast := libraryCardFor(opp, "Their Beast", "Creature — Beast")
	size := opp.Library.Size()
	if size < 4 {
		t.Fatalf("setup: the victim's library holds %d cards", size)
	}

	helmActivation(t, g, me, helm, opp, 2)

	if n := opp.Library.Size(); n != 0 {
		t.Errorf("the victim's library holds %d cards, want 0 — with every card exiled on the "+
			"way, the run never reaches X and mills the whole library", n)
	}
	if n := opp.Graveyard.Size(); n != 0 {
		t.Errorf("%d cards are in the graveyard under Rest in Peace, want none", n)
	}
	if !g.Exile.Contains(beast) {
		t.Error("the milled cards are in exile")
	}
	if !g.Battlefield.Contains(helm) {
		t.Error("no creature card was put into that graveyard, so the Helm is not sacrificed")
	}
	if opp.Eliminated {
		t.Error("milling a library out is not a draw and loses nobody the game (CR 701.13b)")
	}
}

// TestBruvacDoublesAMillAmountAndNotHelmsBound is the line between the
// two rules, on one board.
//
// "Mill three" names a NUMBER and Bruvac the Grandiloquent doubles it
// (CR 701.13b, the RepEventMill window). Helm's X is not that number:
// it is a clause about cards that have been put into a graveyard, the
// run names no amount at all, and there is nothing for a mill-amount
// replacement to double. Before the fix X was the amount, so Bruvac
// doubled the BOUND and X=2 milled four cards.
func TestBruvacDoublesAMillAmountAndNotHelmsBound(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fillPool(me, 5)
	pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent", "Legendary Creature — Human Advisor", bruvacOracle)
	helm := pushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", helmOfObedienceOracle, false)
	for i := 0; i < 5; i++ {
		libraryCardFor(opp, "Filler", "Sorcery")
	}
	graveBefore := opp.Graveyard.Size()

	helmActivation(t, g, me, helm, opp, 2)

	if got := opp.Graveyard.Size() - graveBefore; got != 2 {
		t.Errorf("the Helm put %d cards into the graveyard at X=2, want 2 — Bruvac doubles a "+
			"mill's amount, and an until-run's bound is not one", got)
	}
	if !g.Battlefield.Contains(helm) {
		t.Error("no creature card was milled, so the Helm stays")
	}
	// The same board, the same opponent, an ordinary numbered mill: the
	// amount half still doubles.
	if got := millSeat(t, g, opp, 1); got != 2 {
		t.Errorf("an ordinary mill of 1 put %d cards into the graveyard, want 2 — Bruvac is on "+
			"the battlefield and that half is untouched", got)
	}
}
