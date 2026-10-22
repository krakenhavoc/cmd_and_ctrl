package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// research_development_test.go — Research // Development (#3525):
// Research shuffles cards from outside the game into the library, and
// Development is three rounds of an opponent's choice.

const researchDevelopmentOracle = "fa38394b-6ac4-460c-ae3a-54d50037715b"

// shuffleCount is how many times `player` shuffled their own library.
func shuffleCount(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == player && ev.Label == "shuffle" {
			n++
		}
	}
	return n
}

func libraryHas(p *game.Player, name string) bool {
	for _, c := range p.Library.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func castResearch(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Research", "Instant", researchDevelopmentOracle, nil)
	passPriorityAroundTable(t, g)
}

// Up to four cards from outside the game go into the library, unrevealed,
// and the library is shuffled; the rest stay outside.
func TestResearchShufflesCardsFromOutsideIntoTheLibrary(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("Alpha", "Instant"), outsideCard("Beta", "Sorcery"), outsideCard("Gamma", "Creature — Bear"),
		outsideCard("Delta", "Artifact"), outsideCard("Epsilon", "Land"))
	me, opp := g.Seats[0], g.Seats[1]
	libBefore := me.Library.Size()
	shuffles := shuffleCount(g, me.ID)

	castResearch(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 4 || len(c.ChooseCards) != 5 {
		t.Errorf("bounds %d..%d over %d cards, want 0..4 over all five", c.ChooseMin, c.ChooseMax, len(c.ChooseCards))
	}
	five := []uuid.UUID{sideCardID(me, "Alpha"), sideCardID(me, "Beta"), sideCardID(me, "Gamma"), sideCardID(me, "Delta"), sideCardID(me, "Epsilon")}
	if err := g.ResolveChooseCards(c.ID, me.ID, five); err == nil {
		t.Fatal("Research accepted five cards")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, five[:1]); err == nil {
		t.Error("an opponent answered my Research")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, five[:4]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := sideNamed(me); !equalNames(got, []string{"Epsilon"}) {
		t.Errorf("sideboard %v, want only the card not chosen", got)
	}
	if me.Library.Size() != libBefore+4 {
		t.Errorf("library %d -> %d, want +4", libBefore, me.Library.Size())
	}
	for _, name := range []string{"Alpha", "Beta", "Gamma", "Delta"} {
		if !libraryHas(me, name) {
			t.Errorf("%s did not go into the library", name)
		}
		if outsideHandHas(me, name) {
			t.Errorf("%s went into the hand", name)
		}
	}
	if shuffleCount(g, me.ID) != shuffles+1 {
		t.Errorf("the library was shuffled %d times, want once", shuffleCount(g, me.ID)-shuffles)
	}
	for _, card := range me.Library.Cards {
		if card.IsKnownTo(opp.ID) || card.IsKnownTo(me.ID) {
			t.Errorf("%s is known to a seat after the shuffle", card.Name)
		}
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.OldZone == game.ZoneOutside {
			t.Error("Research revealed a card; it does not")
		}
	}
}

// "Up to": zero is a legal answer, and the library is still shuffled
// (CR 701.24d).
func TestResearchMayChooseNothingAndStillShuffles(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Alpha", "Instant"))
	me := g.Seats[0]
	shuffles := shuffleCount(g, me.ID)
	castResearch(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if len(me.SideboardCards()) != 1 {
		t.Error("declining took a card")
	}
	if shuffleCount(g, me.ID) != shuffles+1 {
		t.Errorf("shuffled %d times, want once", shuffleCount(g, me.ID)-shuffles)
	}
}

// With nothing outside the game (and no database), nothing is asked, and
// the library is shuffled anyway: an empty set is still shuffled in.
func TestResearchWithNothingOutsideJustShuffles(t *testing.T) {
	prev := game.OutsideCollection
	game.OutsideCollection = nil
	t.Cleanup(func() { game.OutsideCollection = prev })
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	shuffles := shuffleCount(g, me.ID)
	castResearch(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a prompt went up: %+v", g.PendingChoices)
	}
	if shuffleCount(g, me.ID) != shuffles+1 {
		t.Errorf("shuffled %d times, want once", shuffleCount(g, me.ID)-shuffles)
	}
}

// ADR 0198: a deck with no sideboard names cards from the database, and
// they are minted into the library as real cards it owns.
func TestResearchTakesCardsFromTheDatabaseWithoutASideboard(t *testing.T) {
	installTestCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	libBefore := me.Library.Size()
	castResearch(t, g)
	c := latestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no collection pick: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 4 {
		t.Errorf("bounds %d..%d, want up to four", c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Sol Ring", "Forest", "Forest"}); err != nil {
		t.Fatalf("ResolveOutsideCollection: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != libBefore+3 || !libraryHas(me, "Sol Ring") {
		t.Errorf("library %d -> %d, Sol Ring in it %v; want +3", libBefore, me.Library.Size(), libraryHas(me, "Sol Ring"))
	}
	for _, card := range me.Library.Cards {
		if card.Name == "Sol Ring" && card.Owner != me.ID {
			t.Error("the minted card is not owned by its player")
		}
	}
	if me.HasSideboard {
		t.Error("taking from the database gave the deck a sideboard")
	}
}

// A restore point before Research restores a table that can run it.
func TestResearchRunsOnARestoredTable(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Alpha", "Instant"))
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	restored := restoreRoundTripFrom(t, snap)
	castResearch(t, restored)
	me := restored.Seats[0]
	c := latestChooseCardsFor(restored, me.ID)
	if c == nil {
		t.Fatal("no prompt after the restore")
	}
	if err := restored.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sideCardID(me, "Alpha")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, restored)
	if !libraryHas(me, "Alpha") {
		t.Error("Alpha did not go into the library after the restore")
	}
}

// --- Development -------------------------------------------------------

func castDevelopment(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpellFace(t, g, "Research // Development", researchDevelopmentOracle, 1)
	passPriorityAroundTable(t, g)
}

// castCatalogSpellFace casts the second half of a split card.
func castCatalogSpellFace(t *testing.T, g *game.Game, name, oracle string, face int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Research", TypeLine: "Instant", ManaCost: "{G}{U}", Colors: []string{"G", "U"}},
			{Name: "Development", TypeLine: "Instant", ManaCost: "{3}{U}{R}", Colors: []string{"U", "R"}},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}

func elementalTokens(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == owner && c.Power == 3 && c.Toughness == 1 {
			n++
		}
	}
	return n
}

const (
	haveThemDraw = 0
	letThemToken = 1
)

// Three rounds; every opponent lets the token be made each time.
func TestDevelopmentMakesThreeTokensWhenNoOpponentHasYouDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	handBefore := me.Hand.Size()
	castDevelopment(t, g)
	asked := 0
	for round := 1; round <= 3; round++ {
		// Opponents are asked in turn order, starting after the caster.
		for _, seat := range []int{1, 2, 3} {
			opp := g.Seats[seat]
			if latestOptionPickFor(g, opp.ID) == nil {
				t.Fatalf("round %d: seat %d was not asked: %+v", round, seat, g.PendingChoices)
			}
			for other := 1; other <= 3; other++ {
				if other > seat && latestOptionPickFor(g, g.Seats[other].ID) != nil {
					t.Fatalf("round %d: seat %d was asked before seat %d answered", round, other, seat)
				}
			}
			answerOptionPick(t, g, opp.ID, letThemToken)
			asked++
			passPriorityAroundTable(t, g)
		}
		if got := elementalTokens(g, me.ID); got != round {
			t.Fatalf("after round %d: %d tokens, want %d", round, got, round)
		}
	}
	if asked != 9 {
		t.Errorf("%d questions, want 9", asked)
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("the caster drew: hand %d -> %d", handBefore, me.Hand.Size())
	}
}

// The first opponent to say yes ends the round with a draw: nobody after
// them is asked, and no token is made that round.
func TestDevelopmentAnOpponentCanHaveYouDrawInstead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	handBefore := me.Hand.Size()
	castDevelopment(t, g)

	// Round 1: seat 1 declines, seat 2 has the caster draw; seat 3 is not asked.
	answerOptionPick(t, g, g.Seats[1].ID, letThemToken)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, g.Seats[2].ID, haveThemDraw)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d -> %d, want one drawn", handBefore, me.Hand.Size())
	}
	if elementalTokens(g, me.ID) != 0 {
		t.Error("a token was made in a round that ended in a draw")
	}
	if latestOptionPickFor(g, g.Seats[3].ID) != nil && latestOptionPickFor(g, g.Seats[1].ID) == nil {
		t.Error("seat 3 was asked in round 1 after seat 2 chose the draw")
	}

	// Round 2 starts with seat 1 again. Everyone declines: a token.
	for _, seat := range []int{1, 2, 3} {
		answerOptionPick(t, g, g.Seats[seat].ID, letThemToken)
		passPriorityAroundTable(t, g)
	}
	if elementalTokens(g, me.ID) != 1 {
		t.Errorf("%d tokens after round 2, want 1", elementalTokens(g, me.ID))
	}
	// Round 3: seat 1 has the caster draw.
	answerOptionPick(t, g, g.Seats[1].ID, haveThemDraw)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+2 || elementalTokens(g, me.ID) != 1 {
		t.Errorf("hand %d (want %d), tokens %d (want 1)", me.Hand.Size(), handBefore+2, elementalTokens(g, me.ID))
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left open: %+v", g.PendingChoices)
	}
}

// A table waiting on an opponent's answer is a restore point, and the
// process carries on from the restored table: the round, and the
// opponents still to be asked, are data on the prompt.
func TestDevelopmentOpenQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	castDevelopment(t, g)
	answerOptionPick(t, g, g.Seats[1].ID, letThemToken)
	passPriorityAroundTable(t, g)
	// Now seat 2 is asked, with seat 3 still to come in round 1.
	restored := restoreRoundTrip(t, g, true)
	me := restored.Seats[0]
	if latestOptionPickFor(restored, restored.Seats[2].ID) == nil {
		t.Fatalf("seat 2's question did not survive the restore: %+v", restored.PendingChoices)
	}
	handBefore := me.Hand.Size()
	answerOptionPick(t, restored, restored.Seats[2].ID, letThemToken)
	passPriorityAroundTable(t, restored)
	if latestOptionPickFor(restored, restored.Seats[3].ID) == nil {
		t.Fatal("seat 3 was not asked after the restore")
	}
	answerOptionPick(t, restored, restored.Seats[3].ID, letThemToken)
	passPriorityAroundTable(t, restored)
	if elementalTokens(restored, me.ID) != 1 {
		t.Errorf("%d tokens after round 1, want 1", elementalTokens(restored, me.ID))
	}
	// Round 2 is asked of seat 1 again: the round count survived.
	if latestOptionPickFor(restored, restored.Seats[1].ID) == nil {
		t.Error("round 2 did not start with seat 1")
	}
	if me.Hand.Size() != handBefore {
		t.Error("the caster drew")
	}
}

// A seat that has left the game is skipped.
func TestDevelopmentSkipsAnEliminatedOpponent(t *testing.T) {
	g := newCatalogGame(t)
	g.Seats[1].Eliminated = true
	castDevelopment(t, g)
	if latestOptionPickFor(g, g.Seats[1].ID) != nil {
		t.Fatal("an eliminated opponent was asked")
	}
	if latestOptionPickFor(g, g.Seats[2].ID) == nil {
		t.Fatalf("seat 2 was not asked first: %+v", g.PendingChoices)
	}
}

func TestResearchAndDevelopmentAreBothFull(t *testing.T) {
	for _, id := range []string{researchDevelopmentOracle, researchDevelopmentOracle + "#1"} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if !slices.Contains(TokenKeys(), "3/1 red Elemental") {
		t.Error("the Elemental token is not in the table")
	}
}
