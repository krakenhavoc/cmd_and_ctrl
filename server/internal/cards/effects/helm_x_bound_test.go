package effects

import (
	"fmt"
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
// it is a clause about cards that have been put into a graveyard, so
// Bruvac cannot double the BOUND. Before #1161 X was the mill's amount
// and Bruvac doubled it, and X=2 milled four cards.
//
// X=2 is the board where the two readings agree on the total and
// disagree about why (#1176). The run mills a card and repeats; Bruvac
// doubles the first repetition to two; two cards land; the clause is
// satisfied and the run stops. Two cards, from ONE doubled repetition
// — which is also the paper answer. Where they part is X=3, and that
// is the test below.
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

// TestBruvacDoublesEachOfHelmsRepetitions is the paper number #1176
// filed for, on the smallest board that shows it.
//
//	Helm at X=3, Bruvac on the battlefield → FOUR cards.
//
// "Target opponent mills a card, then repeats this process until a
// creature card or X cards have been put into their graveyard this
// way" gives one instruction — "mills a card" — and repeats it. Each
// repetition is its own mill, so Bruvac replaces each of them: two
// cards land, three is not reached, the process repeats, two more land
// and the run stops at four, overshooting its own bound by a card.
//
// The engine used to model the whole run as one instruction that named
// no number. That kept the bound safe from Bruvac (the test above) and
// made the AMOUNT invisible, so this board milled three.
//
// The other two rows of #1176's table are in the two tests around
// this one: X=1 mills two (one doubled repetition already passes the
// bound) and X=2 mills two.
func TestBruvacDoublesEachOfHelmsRepetitions(t *testing.T) {
	for _, tc := range []struct {
		x    int
		want int
	}{
		{x: 1, want: 2},
		{x: 2, want: 2},
		{x: 3, want: 4},
		{x: 4, want: 4},
		{x: 5, want: 6},
	} {
		t.Run(fmt.Sprintf("X=%d", tc.x), func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			fillPool(me, 10)
			pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent",
				"Legendary Creature — Human Advisor", bruvacOracle)
			helm := pushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", helmOfObedienceOracle, false)
			// No creature card anywhere in the run, so the only clause
			// that can end it is the X half.
			for i := 0; i < 12; i++ {
				libraryCardFor(opp, "Filler", "Sorcery")
			}
			graveBefore := opp.Graveyard.Size()

			helmActivation(t, g, me, helm, opp, tc.x)

			if got := opp.Graveyard.Size() - graveBefore; got != tc.want {
				t.Errorf("the Helm put %d cards into the graveyard at X=%d, want %d — each "+
					"repetition is its own one-card mill and Bruvac doubles each of them",
					got, tc.x, tc.want)
			}
			if !g.Battlefield.Contains(helm) {
				t.Error("no creature card was milled, so the Helm stays")
			}
		})
	}
}

// TestHelmStopsOnTheFirstRepetitionThatLandsACreature — the other half
// of the clause still ends the run, and it is still asked about what
// ARRIVED. With the mill doubled, the repetition that finds a creature
// card puts BOTH of its cards into the graveyard before the clause is
// asked, because a repetition is one instruction.
func TestHelmStopsOnTheFirstRepetitionThatLandsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fillPool(me, 10)
	pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent",
		"Legendary Creature — Human Advisor", bruvacOracle)
	helm := pushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", helmOfObedienceOracle, false)

	// PushTop, so the LAST push is milled FIRST: a filler, then the
	// Beast, then more fillers behind them.
	for i := 0; i < 4; i++ {
		libraryCardFor(opp, "Deep Filler", "Sorcery")
	}
	beast := libraryCardFor(opp, "Their Beast", "Creature — Beast")
	libraryCardFor(opp, "Top Filler", "Sorcery")
	graveBefore := opp.Graveyard.Size()

	helmActivation(t, g, me, helm, opp, 9)

	if got := opp.Graveyard.Size() - graveBefore; got != 1 {
		t.Errorf("%d cards stayed in the graveyard, want 1 — the Beast was reanimated out of it "+
			"and the filler beside it stayed", got)
	}
	if g.Battlefield.Contains(helm) {
		t.Error("a creature card was put into that graveyard, so the Helm is sacrificed")
	}
	if !g.Battlefield.Contains(beast) {
		t.Error("the milled creature card is on the battlefield")
	}
	if own := cardController(g, beast); own != me.ID {
		t.Errorf("the Beast is under %v, want the Helm's controller — \"under your control\"", own)
	}
}

// cardController reads the battlefield controller of a card.
func cardController(g *game.Game, id uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Controller
		}
	}
	return uuid.Nil
}

// TestHelmAsksWhichCreatureWhenADoubledRepetitionLandsTwo — "put ONE
// OF THEM onto the battlefield" is a real choice exactly when the run
// can land two creature cards at once, which needs a mill-amount
// replacement doubling the repetition that found the first.
//
// Without one the run stops on the first repetition that lands a
// creature card and a repetition mills one card, so there is never
// more than one to choose from and no prompt is raised — which is what
// the test above this one already shows.
func TestHelmAsksWhichCreatureWhenADoubledRepetitionLandsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fillPool(me, 10)
	pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent",
		"Legendary Creature — Human Advisor", bruvacOracle)
	helm := pushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", helmOfObedienceOracle, false)

	// PushTop, so the LAST push is milled first: two creature cards on
	// top, taken by one doubled repetition.
	for i := 0; i < 4; i++ {
		libraryCardFor(opp, "Deep Filler", "Sorcery")
	}
	second := libraryCardFor(opp, "Their Ox", "Creature — Ox")
	first := libraryCardFor(opp, "Their Beast", "Creature — Beast")

	helmActivation(t, g, me, helm, opp, 9)

	pick := pendingChooseCards(t, g, me.ID)
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("the prompt offers %d cards, want both creature cards", len(pick.ChooseCards))
	}
	if g.Battlefield.Contains(helm) {
		t.Error("the Helm is sacrificed before the pick — printed order")
	}

	// Take the SECOND one, which is the pick the old engine could not
	// express: it always reanimated whichever came off the top first.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(second) {
		t.Error("the chosen creature card is not on the battlefield")
	}
	if g.Battlefield.Contains(first) {
		t.Error("the creature card that was NOT chosen came back too")
	}
	if !opp.Graveyard.Contains(first) {
		t.Error("the unchosen creature card stays in the graveyard")
	}
}

// TestHelmsRunResumesAfterEachRepetitionsOrderingPrompt — a repetition
// is a whole mill instruction, so it can PAUSE the way any other mill
// can: two mill-amount replacements in one window is a CR 616.1
// ordering question, asked of the milled player (#982).
//
// The run has to survive that once per repetition. Nothing of the
// repetition has happened when the prompt opens, the next repetition is
// started from the resumed one's continuation, and the clause is asked
// with everything that has landed by then — so a paused run neither
// double-mills nor stalls.
//
// Bruvac first is (1 x 2) + 4 = six cards a repetition, so X=20 takes
// four repetitions and four prompts and ends at 24, four past its own
// bound.
func TestHelmsRunResumesAfterEachRepetitionsOrderingPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fillPool(me, 25)
	pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent",
		"Legendary Creature — Human Advisor", bruvacOracle)
	pushMillReplacement(g, me.ID, "The Water Crystal", "Legendary Artifact", theWaterCrystalOracle)
	helm := pushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", helmOfObedienceOracle, false)
	// No creature card in the run, so only the X half can end it.
	for i := 0; i < 40; i++ {
		libraryCardFor(opp, "Filler", "Sorcery")
	}
	graveBefore := opp.Graveyard.Size()

	helmActivation(t, g, me, helm, opp, 20)

	prompts := 0
	for len(g.PendingChoices) > 0 {
		c := g.PendingChoices[0]
		if c.Kind != game.PendingChoiceReplacementOrder {
			t.Fatalf("prompt kind = %q, want %q", c.Kind, game.PendingChoiceReplacementOrder)
		}
		if c.Chooser != opp.ID {
			t.Fatalf("chooser = %s, want the milled opponent %s (CR 616.1)", c.Chooser, opp.ID)
		}
		prompts++
		if prompts > 10 {
			t.Fatal("the run is not finishing")
		}
		if err := g.ResolveReplacementOrder(c.ID, c.Chooser, replacementOrderByLabel(t, g, c,
			"Bruvac the Grandiloquent — mill twice that many",
			"The Water Crystal — mill that many plus four",
		)); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
	}

	if prompts != 4 {
		t.Errorf("%d ordering prompts, want 4 — one per repetition, each asked before any card "+
			"of that repetition moves", prompts)
	}
	if got := opp.Graveyard.Size() - graveBefore; got != 24 {
		t.Errorf("the Helm put %d cards into the graveyard at X=20, want 24 — four repetitions "+
			"of six, the last of them carrying the run past its bound", got)
	}
	if !g.Battlefield.Contains(helm) {
		t.Error("no creature card was milled, so the Helm stays")
	}
}
