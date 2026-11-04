package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turtles_forever_test.go — #4681: "search your library and/or outside
// the game for exactly four legendary creature cards you own with
// different names".

const turtlesForeverOracle = "cff5aca2-72c5-4053-a32c-c0bd3174c52d"

const turtleType = "Legendary Creature — Mutant Ninja Turtle"

// castTurtlesForever casts the card for seat 0 and resolves it to the
// zone question, which it answers with `zones`.
func castTurtlesForever(t *testing.T, g *game.Game, zones int) {
	t.Helper()
	me := g.Seats[0]
	castCatalogSpell(t, g, "Turtles Forever", "Instant", turtlesForeverOracle, nil)
	passPriorityAroundTable(t, g)
	answerTurtlesZones(t, g, me.ID, zones)
}

func answerTurtlesZones(t *testing.T, g *game.Game, me uuid.UUID, zones int) {
	t.Helper()
	q := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me)
	if q == nil || q.PickThen != turtlesForeverZones.Key() || len(q.PickOptions) != 3 {
		t.Fatalf("no zone question: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionPick(q.ID, me, zones); err != nil {
		t.Fatalf("zone question: %v", err)
	}
}

// opponentChooses names `opp` when the caster is asked, then returns the
// opponent's reveal pick.
func opponentChooses(t *testing.T, g *game.Game, me, opp uuid.UUID) *game.PendingChoice {
	t.Helper()
	if q := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me); q != nil {
		answerChosenPlayer(t, g, me, opp)
	}
	rp := latestRevealPickFor(g, opp)
	if rp == nil || rp.PickThen != turtlesForeverTake.Key() {
		t.Fatalf("the opponent is not choosing: %+v", g.PendingChoices)
	}
	return rp
}

func libraryIDNamed(p *game.Player, name string) uuid.UUID {
	for _, c := range p.Library.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func libraryCardKnownTo(p *game.Player, id, viewer uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return c.IsKnownTo(viewer)
		}
	}
	return false
}

func TestTurtlesForeverIsFull(t *testing.T) {
	spec, ok := Lookup(turtlesForeverOracle)
	if !ok || spec.Name != "Turtles Forever" || spec.Completeness != CompletenessFull {
		t.Fatalf("Turtles Forever: %+v %v", spec.Completeness, ok)
	}
}

// Four from the library: only legendary creature cards are offered, two
// with the same name are refused, nothing is revealed until the search is
// answered, and the two chosen go to the hand while the other two stay in
// the library, which is shuffled.
func TestTurtlesForeverFindsFourInTheLibrary(t *testing.T) {
	g := newSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	seedTypedLibrary(me, "Creature — Bear", "Not Legendary")
	ids := seedTypedLibrary(me, turtleType, "Leonardo", "Donatello", "Raphael", "Michelangelo", "Raphael")
	hand, lib := me.Hand.Size(), me.Library.Size()
	shuffles := shuffleCount(g, me.ID)

	castTurtlesForever(t, g, turtlesForeverLibraryOnly)
	search := searchPromptFor(g, me.ID)
	if search == nil {
		t.Fatalf("no library search: %+v", g.PendingChoices)
	}
	if len(search.SearchCards) != 5 || search.SearchMin != 0 || search.SearchMax != 4 {
		t.Errorf("search offers %d cards, %d..%d; want the five legends, up to four", len(search.SearchCards), search.SearchMin, search.SearchMax)
	}
	if libraryCardKnownTo(me, ids[0], opp.ID) {
		t.Error("the opponent sees a library candidate")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{ids[0], ids[2], ids[4]}); err == nil {
		t.Error("two Raphaels were accepted")
	}
	four := ids[:4]
	if err := g.ResolveSearchLibrary(search.ID, me.ID, four); err != nil {
		t.Fatalf("four with different names: %v", err)
	}
	for _, id := range four {
		if !libraryCardKnownTo(me, id, opp.ID) {
			t.Errorf("found card %v was not revealed", id)
		}
	}
	if libraryCardKnownTo(me, ids[4], opp.ID) {
		t.Error("the card left behind was revealed")
	}
	rp := opponentChooses(t, g, me.ID, opp.ID)
	if rp.ChooseMin != 2 || rp.ChooseMax != 2 || len(rp.ChooseCards) != 4 {
		t.Errorf("opponent picks %d..%d of %d, want two of four", rp.ChooseMin, rp.ChooseMax, len(rp.ChooseCards))
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{ids[0], ids[1]}); err != nil {
		t.Fatalf("opponent's choice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Leonardo") || !outsideHandHas(me, "Donatello") || me.Hand.Size() != hand+2 {
		t.Errorf("hand %d (had %d): the chosen two go to the hand", me.Hand.Size(), hand)
	}
	if me.Library.Size() != lib-2 || !libraryHas(me, "Michelangelo") {
		t.Errorf("library %d (had %d): the rest stay in it", me.Library.Size(), lib)
	}
	if shuffleCount(g, me.ID) != shuffles+1 {
		t.Errorf("shuffled %d times, want once", shuffleCount(g, me.ID)-shuffles)
	}
}

// A mix with the sideboard: the library half first, then the outside half
// with the room left and none of the library's names, everything revealed
// together, and the rest, the one from outside the game included,
// shuffled into the library.
func TestTurtlesForeverMixesTheLibraryAndTheSideboard(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("April", turtleType), outsideCard("Casey", turtleType),
		outsideCard("Leonardo", turtleType), outsideCard("Splinter", "Creature — Rat"),
		outsideCard("Shredder", turtleType))
	me, opp := g.Seats[0], g.Seats[1]
	ids := seedTypedLibrary(me, turtleType, "Leonardo", "Donatello")
	hand, lib := me.Hand.Size(), me.Library.Size()
	shuffles := shuffleCount(g, me.ID)

	castTurtlesForever(t, g, turtlesForeverBoth)
	search := searchPromptFor(g, me.ID)
	if search == nil {
		t.Fatalf("no library search: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, ids); err != nil {
		t.Fatal(err)
	}
	if libraryCardKnownTo(me, ids[0], opp.ID) {
		t.Error("the library's finds were revealed before the outside half")
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.PickThen != turtlesForeverOutside.Key() {
		t.Fatalf("no outside pick: %+v", g.PendingChoices)
	}
	if got := candidateNames(g, pick); !equalNames(got, []string{"April", "Casey", "Shredder"}) {
		t.Errorf("outside candidates %v: legends only, and no Leonardo (found in the library)", got)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("outside pick %d..%d, want up to the two the library left room for", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{sideCardID(me, "April"), sideCardID(me, "Casey"), sideCardID(me, "Shredder")}); err == nil {
		t.Error("five cards in all were accepted")
	}
	april, casey := sideCardID(me, "April"), sideCardID(me, "Casey")
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{april, casey}); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if !libraryCardKnownTo(me, id, opp.ID) {
			t.Errorf("library find %v not revealed", id)
		}
	}
	for _, id := range []uuid.UUID{april, casey} {
		if !cardKnownToAll(g, me, id) {
			t.Errorf("outside find %v not revealed", id)
		}
	}
	if cardKnownTo(me, sideCardID(me, "Shredder"), opp.ID) {
		t.Error("a card left outside the game was revealed")
	}
	rp := opponentChooses(t, g, me.ID, opp.ID)
	if len(rp.ChooseCards) != 4 {
		t.Fatalf("opponent picks from %d cards, want four", len(rp.ChooseCards))
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{april, ids[1]}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "April") || !outsideHandHas(me, "Donatello") || me.Hand.Size() != hand+2 {
		t.Errorf("hand %d (had %d): April and Donatello", me.Hand.Size(), hand)
	}
	if !libraryHas(me, "Casey") || !libraryHas(me, "Leonardo") || me.Library.Size() != lib {
		t.Errorf("library %d (had %d): Casey shuffled in, Leonardo stays", me.Library.Size(), lib)
	}
	if got := sideNamed(me); !equalNames(got, []string{"Leonardo", "Splinter", "Shredder"}) {
		t.Errorf("sideboard %v", got)
	}
	if shuffleCount(g, me.ID) != shuffles+1 {
		t.Errorf("shuffled %d times, want once", shuffleCount(g, me.ID)-shuffles)
	}
}

// Fewer than four: one legend in the library and a player with no
// sideboard and no card database. The opponent "chooses two" of one,
// which is a choice of that one.
func TestTurtlesForeverFindsFewerThanFour(t *testing.T) {
	prev := game.OutsideCollection
	game.OutsideCollection = nil
	t.Cleanup(func() { game.OutsideCollection = prev })
	g := newSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	ids := seedTypedLibrary(me, turtleType, "Leonardo")
	hand := me.Hand.Size()

	castTurtlesForever(t, g, turtlesForeverBoth)
	search := searchPromptFor(g, me.ID)
	if search == nil {
		t.Fatalf("no library search: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, ids); err != nil {
		t.Fatal(err)
	}
	if c := latestChooseCardsFor(g, me.ID); c != nil {
		t.Errorf("a player with nothing outside the game was asked: %+v", c)
	}
	rp := opponentChooses(t, g, me.ID, opp.ID)
	if rp.ChooseMin != 1 || rp.ChooseMax != 1 {
		t.Errorf("opponent picks %d..%d of one", rp.ChooseMin, rp.ChooseMax)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, ids); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Leonardo") || me.Hand.Size() != hand+1 {
		t.Errorf("hand %d (had %d)", me.Hand.Size(), hand)
	}
}

// Finding nothing is legal (CR 701.23b, 701.23j): the library is still
// shuffled (CR 701.24d), and nobody is asked to choose.
func TestTurtlesForeverMayFindNothing(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("April", turtleType))
	me := g.Seats[0]
	seedTypedLibrary(me, turtleType, "Leonardo")
	hand, lib := me.Hand.Size(), me.Library.Size()
	shuffles := shuffleCount(g, me.ID)

	castTurtlesForever(t, g, turtlesForeverBoth)
	search := searchPromptFor(g, me.ID)
	if search == nil {
		t.Fatal("no library search")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing in the library: %v", err)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 || len(pick.ChooseCards) != 1 {
		t.Fatalf("outside pick: %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing outside the game: %v", err)
	}
	if rp := pendingOfKind(g, game.PendingChoiceRevealPick); rp != nil {
		t.Errorf("an opponent was asked to choose from nothing: %+v", rp)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Library.Size() != lib || len(me.SideboardCards()) != 1 {
		t.Errorf("hand %d/%d, library %d/%d, sideboard %v", me.Hand.Size(), hand, me.Library.Size(), lib, sideNamed(me))
	}
	if shuffleCount(g, me.ID) != shuffles+1 {
		t.Errorf("shuffled %d times, want once", shuffleCount(g, me.ID)-shuffles)
	}
}

// Outside the game only: the library is not searched (no search prompt,
// no search event), and the four come from the sideboard.
func TestTurtlesForeverOutsideOnlyDoesNotSearchTheLibrary(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("April", turtleType), outsideCard("Casey", turtleType),
		outsideCard("Shredder", turtleType), outsideCard("Krang", turtleType))
	me, opp := g.Seats[0], g.Seats[1]
	seedTypedLibrary(me, turtleType, "Leonardo")
	searches := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Label != "shuffle" {
			searches++
		}
	}
	castTurtlesForever(t, g, turtlesForeverOutsideOnly)
	if s := searchPromptFor(g, me.ID); s != nil {
		t.Fatalf("the library was searched: %+v", s)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 4 || len(pick.ChooseCards) != 4 {
		t.Fatalf("outside pick: %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards); err != nil {
		t.Fatal(err)
	}
	rp := opponentChooses(t, g, me.ID, opp.ID)
	if err := g.ResolveRevealPick(rp.ID, opp.ID, rp.ChooseCards[:2]); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	after := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Label != "shuffle" {
			after++
		}
	}
	if after != searches {
		t.Errorf("%d library searches announced, want none", after-searches)
	}
	if len(me.SideboardCards()) != 0 || !libraryHas(me, "Shredder") && !libraryHas(me, "Krang") {
		t.Errorf("sideboard %v: the two not chosen go into the library", sideNamed(me))
	}
}

// The card database (a deck with no sideboard, ADR 0198): the suggestions
// are legends only, a name found in the library is not one of them, and
// two copies of a name are refused.
func TestTurtlesForeverTakesFromTheDatabase(t *testing.T) {
	installTurtlesCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	ids := seedTypedLibrary(me, turtleType, "Leonardo")
	castTurtlesForever(t, g, turtlesForeverBoth)
	search := searchPromptFor(g, me.ID)
	if err := g.ResolveSearchLibrary(search.ID, me.ID, ids); err != nil {
		t.Fatal(err)
	}
	c := latestCollectionPickFor(g, me.ID)
	if c == nil || c.PickThen != turtlesForeverOutside.Key() || c.ChooseMax != 3 {
		t.Fatalf("collection pick: %+v", c)
	}
	if !equalNames(c.CollectionSuggestions, []string{"April", "Casey", "Shredder"}) {
		t.Errorf("suggestions %v: the legends, without Leonardo", c.CollectionSuggestions)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Leonardo"}); err == nil {
		t.Error("a second Leonardo was accepted")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"April", "April"}); err == nil {
		t.Error("two Aprils were accepted")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Splinter"}); err == nil {
		t.Error("a creature that is not legendary was accepted")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"April", "Casey", "Shredder"}); err != nil {
		t.Fatal(err)
	}
	rp := opponentChooses(t, g, me.ID, opp.ID)
	if len(rp.ChooseCards) != 4 {
		t.Fatalf("opponent picks from %d", len(rp.ChooseCards))
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{ids[0], sideCardID(me, "April")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Leonardo") || !outsideHandHas(me, "April") {
		t.Error("the chosen cards are not in the hand")
	}
	if !libraryHas(me, "Casey") || !libraryHas(me, "Shredder") || len(me.SideboardCards()) != 0 {
		t.Errorf("the rest were not shuffled in: sideboard %v", sideNamed(me))
	}
}

// A restore with each prompt open: the zone question, the library search,
// the outside pick (which keeps the library's names) and the opponent's
// choice, each answered on the restored table.
func TestTurtlesForeverIsARestorePointAtEveryPrompt(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("Leonardo", turtleType), outsideCard("April", turtleType))
	castCatalogSpell(t, g, "Turtles Forever", "Instant", turtlesForeverOracle, nil)
	passPriorityAroundTable(t, g)
	seedTypedLibrary(g.Seats[0], turtleType, "Leonardo", "Donatello")

	g = restoreRoundTrip(t, g, true)
	answerTurtlesZones(t, g, g.Seats[0].ID, turtlesForeverBoth)

	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	search := searchPromptFor(g, me.ID)
	if search == nil || len(search.SearchCards) != 2 {
		t.Fatalf("restored search: %+v", search)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, search.SearchCards); err != nil {
		t.Fatal(err)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.PickThen != turtlesForeverOutside.Key() {
		t.Fatalf("restored outside pick: %+v", g.PendingChoices)
	}
	if got := candidateNames(g, pick); !equalNames(got, []string{"April"}) {
		t.Errorf("restored candidates %v, want April alone", got)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{sideCardID(me, "April")}); err != nil {
		t.Fatal(err)
	}

	if q := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID); q != nil {
		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		answerChosenPlayer(t, g, me.ID, g.Seats[1].ID)
	}
	g = restoreRoundTrip(t, g, true)
	me, opp := g.Seats[0], g.Seats[1]
	rp := latestRevealPickFor(g, opp.ID)
	if rp == nil || len(rp.ChooseCards) != 3 {
		t.Fatalf("restored reveal pick: %+v", g.PendingChoices)
	}
	donatello := libraryIDNamed(me, "Donatello")
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{donatello, sideCardID(me, "April")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Donatello") || !outsideHandHas(me, "April") || !libraryHas(me, "Leonardo") {
		t.Error("the restored table did not settle the choice")
	}
}

// The database pick keeps the library's names across a restore: the
// word rides the prompt.
func TestTurtlesForeverDatabasePickKeepsTheLibraryNamesAcrossARestore(t *testing.T) {
	installTurtlesCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	ids := seedTypedLibrary(me, turtleType, "Leonardo")
	castTurtlesForever(t, g, turtlesForeverBoth)
	if err := g.ResolveSearchLibrary(searchPromptFor(g, me.ID).ID, me.ID, ids); err != nil {
		t.Fatal(err)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := latestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("restored: %+v", g.PendingChoices)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Leonardo"}); err == nil {
		t.Error("after the restore a second Leonardo was accepted")
	}
	var legal bool
	g.WithWriteLock(func() { legal = g.CollectionPickLegalLocked(c, []string{"April", "Casey"}) })
	if !legal {
		t.Error("the enumerator's window refuses two legends with new names")
	}
}

func TestTurtlesForeverPicksWithAnUnknownKeyAreRefused(t *testing.T) {
	t.Run("zones", func(t *testing.T) {
		g := newSideboardCatalogGame(t, 2)
		castCatalogSpell(t, g, "Turtles Forever", "Instant", turtlesForeverOracle, nil)
		passPriorityAroundTable(t, g)
		refuseUnknownKey(t, g, turtlesForeverZones.Key())
	})
	t.Run("outside", func(t *testing.T) {
		g := newSideboardCatalogGame(t, 2, outsideCard("April", turtleType))
		castTurtlesForever(t, g, turtlesForeverOutsideOnly)
		refuseUnknownKey(t, g, turtlesForeverOutside.Key())
	})
}

// installTurtlesCollection installs a database of a few legends and one
// creature that is not legendary.
func installTurtlesCollection(t *testing.T) {
	t.Helper()
	prev := game.OutsideCollection
	cards := []game.Card{}
	for _, name := range []string{"April", "Casey", "Leonardo", "Shredder"} {
		cards = append(cards, dbCard(name, turtleType, ""))
	}
	cards = append(cards, dbCard("Splinter", "Creature — Rat", ""))
	slices.SortFunc(cards, func(a, b game.Card) int {
		if a.Name < b.Name {
			return -1
		}
		return 1
	})
	game.OutsideCollection = testCollection{cards: cards}
	t.Cleanup(func() { game.OutsideCollection = prev })
}
