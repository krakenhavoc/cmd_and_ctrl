package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// search_chooser_test.go — the S22 "you choose" search prompt, plus
// the #263 regression suite for the CR 614 entry pipeline the search
// path used to skip.
//
// Helpers first: nearly every tutor / fetch test in this package now
// has to answer a prompt, so the shapes live here rather than being
// re-spelled per file.

// searchChoiceFor returns the newest queued search prompt addressed
// to a player, or nil.
func searchChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// searchOptionNamed returns the candidate with the given name from a
// search prompt, or uuid.Nil.
func searchOptionNamed(g *game.Game, c *game.PendingChoice, name string) uuid.UUID {
	for _, id := range c.SearchCards {
		var found bool
		g.ReadSnapshot(func() {
			card, ok := g.LookupCardForEffect(id)
			found = ok && card.Name == name
		})
		if found {
			return id
		}
	}
	return uuid.Nil
}

// answerSearchNamed answers a player's open search prompt by taking
// the named cards. Fails the test when no prompt is open or a name
// is not among the candidates — "the prompt didn't appear" and "the
// prompt didn't offer the card" are different bugs and both want to
// be loud.
func answerSearchNamed(t *testing.T, g *game.Game, chooser uuid.UUID, names ...string) {
	t.Helper()
	c := searchChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no search prompt for %s", chooser)
	}
	picks := make([]uuid.UUID, 0, len(names))
	for _, n := range names {
		id := searchOptionNamed(g, c, n)
		if id == uuid.Nil {
			t.Fatalf("search prompt did not offer %q", n)
		}
		picks = append(picks, id)
	}
	if err := g.ResolveSearchLibrary(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
}

// answerSearchByID answers a player's open search prompt with
// specific instance IDs — for fixtures that seed two cards with the
// same name and need to tell them apart.
func answerSearchByID(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := searchChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no search prompt for %s", chooser)
	}
	if err := g.ResolveSearchLibrary(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
}

// answerSearchFailToFind answers an open search prompt by taking nothing (CR
// 701.19c).
func answerSearchFailToFind(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := searchChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no search prompt for %s", chooser)
	}
	if err := g.ResolveSearchLibrary(c.ID, chooser, nil); err != nil {
		t.Fatalf("ResolveSearchLibrary (fail to find): %v", err)
	}
}

// seedSearchLibrary replaces a player's library with exactly the
// named cards, bottom-first. Returns their instance IDs in the same
// order.
func seedSearchLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		ids = append(ids, pushLibraryCardForTest(p, c))
	}
	return ids
}

func searchTestLand(name, typeLine string) game.Card {
	return game.Card{Name: name, TypeLine: typeLine}
}

// --- #263: the search path runs the CR 614 entry pipeline ---------

// The probe from #263, promoted to a regression test. Darkslick
// Shores is "enters tapped unless you control two or fewer other
// lands"; fetched with three other lands out it must enter TAPPED.
// Before the fix the search path never called applyReplacementsLocked
// and it entered untapped — strictly better than printed, for every
// conditional dual cycle in the catalog.
func TestSearchToBattlefieldRunsEntersTappedReplacement(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	}
	seedSearchLibrary(me, game.Card{
		Name:     "Darkslick Shores",
		TypeLine: "Land",
		OracleID: darkslickShoresOracle,
	})

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffectWithOptions(me.ID,
			func(c game.Card) bool { return c.Name == "Darkslick Shores" },
			game.ZoneBattlefield, 1, false, false, false)
	})

	card, ok := searchFetchedCard(g, "Darkslick Shores")
	if !ok {
		t.Fatal("Darkslick Shores was not fetched onto the battlefield")
	}
	if !card.Tapped {
		t.Error("fastland fetched with 3 other lands entered UNTAPPED; want tapped")
	}
	if card.Controller != me.ID {
		t.Errorf("controller = %s, want the searcher %s", card.Controller, me.ID)
	}
}

// The mirror of the above: inside the fastland's window the same
// fetch must produce an UNTAPPED land, so the fix is running the
// condition rather than just forcing everything tapped.
func TestSearchToBattlefieldRespectsTheUntappedCondition(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedSearchLibrary(me, game.Card{
		Name:     "Darkslick Shores",
		TypeLine: "Land",
		OracleID: darkslickShoresOracle,
	})

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffectWithOptions(me.ID,
			func(c game.Card) bool { return c.Name == "Darkslick Shores" },
			game.ZoneBattlefield, 1, false, false, false)
	})

	card, ok := searchFetchedCard(g, "Darkslick Shores")
	if !ok {
		t.Fatal("Darkslick Shores was not fetched onto the battlefield")
	}
	if card.Tapped {
		t.Error("fastland fetched with 1 other land entered tapped; want untapped")
	}
}

// The fetching effect's own "tapped" clause still wins on a land
// whose printed text says nothing about entering tapped — the
// TappedOnEntry flag and the pipeline are OR-ed, not either/or.
func TestSearchTappedOnEntryStillForcesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffectWithOptions(me.ID, IsBasicLand,
			game.ZoneBattlefield, 1, false, false, true)
	})

	card, ok := searchFetchedCard(g, "Forest")
	if !ok {
		t.Fatal("Forest was not fetched onto the battlefield")
	}
	if !card.Tapped {
		t.Error("TappedOnEntry fetch entered untapped")
	}
}

// An OPPONENT-facing static replacement has to see a fetched
// permanent too. Authority of the Consuls taps creatures entering
// under an opponent's control; before #263 a creature the opponent
// tutored out walked straight past it.
func TestSearchToBattlefieldHitsAuthorityOfTheConsuls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = seedReplacementPermanent(g, authorityOfTheConsulsOracle, "Authority of the Consuls", me.ID)
	seedSearchLibrary(opp, game.Card{
		Name:      "Grizzly Bears",
		TypeLine:  "Creature — Bear",
		Power:     2,
		Toughness: 2,
	})

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(opp.ID,
			func(c game.Card) bool { return c.Name == "Grizzly Bears" },
			game.ZoneBattlefield, 1, false, false)
	})

	card, ok := searchFetchedCard(g, "Grizzly Bears")
	if !ok {
		t.Fatal("Grizzly Bears was not fetched onto the battlefield")
	}
	if !card.Tapped {
		t.Error("a creature an opponent fetched dodged Authority of the Consuls")
	}
}

// Reanimation is the same bug in a different zone, and the same fix:
// a creature put onto the battlefield from a graveyard under an
// opponent's Authority enters tapped.
func TestReanimationRunsTheEntryPipeline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = seedReplacementPermanent(g, authorityOfTheConsulsOracle, "Authority of the Consuls", me.ID)
	corpse := pushGraveyardCardForTest(opp, "Reanimated Bear")

	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardForEffect(corpse, game.ZoneBattlefield)
	})

	card, ok := searchFetchedCard(g, "Reanimated Bear")
	if !ok {
		t.Fatal("the reanimated creature is not on the battlefield")
	}
	if !card.Tapped {
		t.Error("a reanimated creature dodged Authority of the Consuls")
	}
}

// searchFetchedCard finds a battlefield card by name.
func searchFetchedCard(g *game.Game, name string) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				out, ok = c, true
				return
			}
		}
	})
	return out, ok
}

// --- the chooser --------------------------------------------------

// More matches than the effect may take ⇒ a prompt, and the card the
// player picks is the card they get. This is the whole point: before
// S22 the engine took the first match in library order, so the pick
// was a property of the deck list.
func TestSearchPromptsWhenThereIsAChoiceAndHonoursIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
	)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, IsBasicLand,
			game.ZoneBattlefield, 1, false, true)
	})

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("three matches and a limit of one produced no prompt")
	}
	if len(c.SearchCards) != 3 {
		t.Errorf("prompt offered %d candidates, want all 3", len(c.SearchCards))
	}
	if c.SearchMax != 1 {
		t.Errorf("SearchMax = %d, want 1", c.SearchMax)
	}
	// Nothing has moved while the prompt is open.
	if _, ok := searchFetchedCard(g, "Island"); ok {
		t.Error("a card reached the battlefield before the prompt was answered")
	}

	answerSearchNamed(t, g, me.ID, "Island")

	if _, ok := searchFetchedCard(g, "Island"); !ok {
		t.Error("the chosen Island did not reach the battlefield")
	}
	if _, ok := searchFetchedCard(g, "Forest"); ok {
		t.Error("a card the player did not choose was fetched")
	}
}

// One match and a limit of one is not a choice, so no prompt — a
// modal with a single button is worse than no modal.
func TestSearchDoesNotPromptWhenThereIsNothingToDecide(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, IsBasicLand,
			game.ZoneBattlefield, 1, false, true)
	})

	if c := searchChoiceFor(g, me.ID); c != nil {
		t.Error("a forced pick still queued a prompt")
	}
	if _, ok := searchFetchedCard(g, "Forest"); !ok {
		t.Error("the only match was not fetched")
	}
}

// CR 701.23b — a player may fail to find. The search still happened,
// so the library is still shuffled and the event still fires.
func TestSearchFailToFindTakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	before := me.Library.Size()

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, IsBasicLand,
			game.ZoneBattlefield, 1, false, true)
	})
	answerSearchFailToFind(t, g, me.ID)

	if me.Library.Size() != before {
		t.Errorf("library size %d after failing to find, want %d", me.Library.Size(), before)
	}
	if _, ok := searchFetchedCard(g, "Forest"); ok {
		t.Error("failing to find still fetched a card")
	}
}

// A "you may search" clause prompts even when the pick is forced, so
// the searcher can decline the card AND the shuffle.
func TestOptionalSearchPromptsEvenWithOneMatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))

	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player:   me.ID,
			Pred:     IsBasicLand,
			Dest:     game.ZoneBattlefield,
			Limit:    1,
			Optional: true,
			Shuffle:  true,
		})
	})

	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("an optional search with one match did not prompt")
	}
	answerSearchFailToFind(t, g, me.ID)
	if _, ok := searchFetchedCard(g, "Forest"); ok {
		t.Error("declining an optional search still fetched the land")
	}
}

// The prompt is a look, not a reveal: only the searcher becomes a
// knower of the candidates.
func TestSearchCandidatesAreKnownOnlyToTheSearcher(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, IsBasicLand,
			game.ZoneHand, 1, true, true)
	})

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	g.ReadSnapshot(func() {
		for _, card := range me.Library.Cards {
			if !card.IsLand() {
				continue
			}
			if !card.KnownBy[me.ID] {
				t.Errorf("%s: the searcher is not a knower of a candidate", card.Name)
			}
			if card.KnownBy[opp.ID] {
				t.Errorf("%s: an opponent became a knower of a search candidate", card.Name)
			}
		}
	})
}

// Only the cards actually TAKEN by a "reveal those cards" search
// become public — not every card the searcher flipped past.
func TestSearchRevealsOnlyWhatWasTaken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)

	g.WithWriteLock(func() {
		// No shuffle, so the untaken candidate's knower set survives
		// for inspection.
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID,
			Pred:   IsBasicLand,
			Dest:   game.ZoneHand,
			Limit:  1,
			Reveal: true,
		})
	})
	answerSearchNamed(t, g, me.ID, "Island")

	g.ReadSnapshot(func() {
		for _, card := range me.Hand.Cards {
			if card.Name == "Island" && !card.KnownBy[opp.ID] {
				t.Error("a revealed, taken card is not public")
			}
		}
		for _, card := range me.Library.Cards {
			if card.Name == "Forest" && card.KnownBy[opp.ID] {
				t.Error("a candidate that was NOT taken leaked to an opponent")
			}
		}
	})
}

// The engine refuses a pick that was never a candidate, and a pick
// count above the limit — and refuses without consuming the prompt,
// so the client can try again.
func TestResolveSearchLibraryRejectsBadAnswers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, IsBasicLand,
			game.ZoneBattlefield, 1, false, true)
	})
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}

	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{uuid.New()}); err == nil {
		t.Error("a pick that was never a candidate was accepted")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards); err == nil {
		t.Error("picking two cards under a limit of one was accepted")
	}
	if err := g.ResolveSearchLibrary(c.ID, g.Seats[1].ID, nil); err == nil {
		t.Error("a player who is not the chooser answered the prompt")
	}
	if searchChoiceFor(g, me.ID) == nil {
		t.Error("a rejected answer consumed the prompt")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
}

// --- the Then continuation, taken through the prompt ---------------

// Fabled Passage's "then ... untap THAT LAND" names the card the
// SEARCHER chose, several client round-trips after the ability
// resolved. The card used to peek at the library and re-derive the
// first-match pick; with a chooser there is nothing to re-derive, so
// the untap rides the search's Then continuation.
func TestFabledPassageUntapsTheLandTheSearcherChose(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	// Three lands plus the fetched one makes four.
	seedPermanentFor(g, me.ID, "Forest A", "Basic Land — Forest")
	seedPermanentFor(g, me.ID, "Forest B", "Basic Land — Forest")
	seedPermanentFor(g, me.ID, "Forest C", "Basic Land — Forest")
	passage := pushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", fabledPassageOracle, false)
	decoy := stapleLibraryCard(me, "Swamp", "Basic Land — Swamp")
	want := stapleLibraryCard(me, "Island", "Basic Land — Island")

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, want)

	if _, wrong := battlefieldCard(g, decoy); wrong {
		t.Error("the land the searcher did not pick was fetched")
	}
	got, ok := battlefieldCard(g, want)
	if !ok {
		t.Fatal("the chosen land did not reach the battlefield")
	}
	if got.Tapped {
		t.Error("with four lands the CHOSEN land must be untapped by the Then continuation")
	}
}

// Gamble is "search, put that card into your hand, THEN discard a
// card at random". The discard has to wait for the search — with an
// otherwise empty hand the discarded card is the tutored one, which
// is the entire joke, and a discard that ran before the prompt was
// answered would take from an empty hand instead.
func TestGambleDiscardsAfterTheSearchPromptIsAnswered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	prize := pushLibraryCardForTest(me, game.Card{Name: "Prize", TypeLine: "Sorcery"})

	castCatalogSpell(t, g, "Gamble", "Sorcery", gambleOracle, nil)
	passPriorityAroundTable(t, g)

	// The spell has resolved but the discard has not happened yet —
	// the hand is still empty and the prize is still in the library.
	if me.Graveyard.Contains(prize) {
		t.Error("Gamble discarded before the searcher had picked anything")
	}
	answerSearchByID(t, g, me.ID, prize)

	if me.Hand.Size() != 0 {
		t.Errorf("hand = %d, want 0 (fetched then pitched)", me.Hand.Size())
	}
	if !me.Graveyard.Contains(prize) {
		t.Error("the fetched card should have been the random discard")
	}
}

// --- the declared limit, pinned -----------------------------------

// A fetched shockland enters TAPPED and nobody is asked to pay.
//
// This is the declared simplification on searchEnterBattlefieldLocked
// made executable. The search path now runs the CR 614 pipeline, so
// the shockland's entry replacement IS consulted — but the search
// entry site is not entryResumable, so the pipeline cannot pause
// there to ask the question, and takes the un-paid branch.
//
// Weaker than printed, never stronger, and a strict improvement on
// the pre-#263 behaviour where the clause was skipped entirely and
// the land arrived untapped for free. When the search path learns to
// carry its continuation through an entry prompt, this test flips to
// "a prompt is queued and the land enters untapped if you pay".
func TestFetchedShocklandEntersTappedWithNoPaymentOffered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lifeBefore := me.Life
	seedSearchLibrary(me, game.Card{
		Name:     "Blood Crypt",
		TypeLine: "Land — Swamp Mountain",
		OracleID: bloodCryptOracle,
	})

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID,
			func(c game.Card) bool { return c.Name == "Blood Crypt" },
			game.ZoneBattlefield, 1, false, false)
	})

	card, ok := searchFetchedCard(g, "Blood Crypt")
	if !ok {
		t.Fatal("Blood Crypt was not fetched onto the battlefield")
	}
	if !card.Tapped {
		t.Error("a fetched shockland entered UNTAPPED for free — the entry clause was skipped")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("a search prompt is open; the fetch had exactly one candidate")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryPayLife {
			t.Error("the fetched entry queued a pay-life prompt it has no resume for")
		}
	}
	if me.Life != lifeBefore {
		t.Errorf("life %d -> %d; nothing should have been paid", lifeBefore, me.Life)
	}
}
