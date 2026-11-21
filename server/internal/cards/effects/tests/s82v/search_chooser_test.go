package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- #263: the search path runs the CR 614 entry pipeline ---------

// The probe from #263, promoted to a regression test. Darkslick
// Shores is "enters tapped unless you control two or fewer other
// lands"; fetched with three other lands out it must enter TAPPED.
// Before the fix the search path never called applyReplacementsLocked
// and it entered untapped — strictly better than printed, for every
// conditional dual cycle in the catalog.
func TestSearchToBattlefieldRunsEntersTappedReplacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	}
	effectstest.SeedSearchLibrary(me, game.Card{
		Name:     "Darkslick Shores",
		TypeLine: "Land",
		OracleID: effectstest.DarkslickShoresOracle,
	})

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffectWithOptions(me.ID,
			func(c game.Card) bool { return c.Name == "Darkslick Shores" },
			game.ZoneBattlefield, 1, false, false, false)
	})
	effectstest.TakeSearchPrompts(t, g)

	card, ok := effectstest.SearchFetchedCard(g, "Darkslick Shores")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.SeedSearchLibrary(me, game.Card{
		Name:     "Darkslick Shores",
		TypeLine: "Land",
		OracleID: effectstest.DarkslickShoresOracle,
	})

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffectWithOptions(me.ID,
			func(c game.Card) bool { return c.Name == "Darkslick Shores" },
			game.ZoneBattlefield, 1, false, false, false)
	})
	effectstest.TakeSearchPrompts(t, g)

	card, ok := effectstest.SearchFetchedCard(g, "Darkslick Shores")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffectWithOptions(me.ID, effects.IsBasicLand,
			game.ZoneBattlefield, 1, false, false, true)
	})
	effectstest.TakeSearchPrompts(t, g)

	card, ok := effectstest.SearchFetchedCard(g, "Forest")
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = effectstest.SeedReplacementPermanent(g, effectstest.AuthorityOfTheConsulsOracle, "Authority of the Consuls", me.ID)
	effectstest.SeedSearchLibrary(opp, game.Card{
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
	effectstest.TakeSearchPrompts(t, g)

	card, ok := effectstest.SearchFetchedCard(g, "Grizzly Bears")
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = effectstest.SeedReplacementPermanent(g, effectstest.AuthorityOfTheConsulsOracle, "Authority of the Consuls", me.ID)
	corpse := effectstest.PushGraveyardCardForTest(opp, "Reanimated Bear")

	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardForEffect(corpse, game.ZoneBattlefield)
	})

	card, ok := effectstest.SearchFetchedCard(g, "Reanimated Bear")
	if !ok {
		t.Fatal("the reanimated creature is not on the battlefield")
	}
	if !card.Tapped {
		t.Error("a reanimated creature dodged Authority of the Consuls")
	}
}

// --- the chooser --------------------------------------------------

// More matches than the effect may take ⇒ a prompt, and the card the
// player picks is the card they get. This is the whole point: before
// S22 the engine took the first match in library order, so the pick
// was a property of the deck list.
func TestSearchPromptsWhenThereIsAChoiceAndHonoursIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
	)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, effects.IsBasicLand,
			game.ZoneBattlefield, 1, false, true)
	})

	c := effectstest.SearchChoiceFor(g, me.ID)
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
	if _, ok := effectstest.SearchFetchedCard(g, "Island"); ok {
		t.Error("a card reached the battlefield before the prompt was answered")
	}

	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")

	if _, ok := effectstest.SearchFetchedCard(g, "Island"); !ok {
		t.Error("the chosen Island did not reach the battlefield")
	}
	if _, ok := effectstest.SearchFetchedCard(g, "Forest"); ok {
		t.Error("a card the player did not choose was fetched")
	}
}

// A search the rules force — an unconditional "a card" (CR 701.23d)
// against a library of exactly one — is not a choice, so no prompt: a
// modal with a single button is worse than no modal. (A search for a
// stated quality is different since #3053: it may find nothing, so it
// always asks. fail_to_find_cards_test.go pins that side.)
func TestSearchDoesNotPromptWhenTheRulesForceThePick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, nil,
			game.ZoneBattlefield, 1, false, true)
	})

	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		t.Error("a forced pick still queued a prompt")
	}
	if _, ok := effectstest.SearchFetchedCard(g, "Forest"); !ok {
		t.Error("the only card was not found")
	}
}

// CR 701.23b — a player may fail to find. The search still happened,
// so the library is still shuffled and the event still fires.
func TestSearchFailToFindTakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)
	before := me.Library.Size()

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, effects.IsBasicLand,
			game.ZoneBattlefield, 1, false, true)
	})
	effectstest.AnswerSearchFailToFind(t, g, me.ID)

	if me.Library.Size() != before {
		t.Errorf("library size %d after failing to find, want %d", me.Library.Size(), before)
	}
	if _, ok := effectstest.SearchFetchedCard(g, "Forest"); ok {
		t.Error("failing to find still fetched a card")
	}
}

// A "you may search" clause prompts even when the pick is forced, so
// the searcher can decline the card AND the shuffle.
func TestOptionalSearchPromptsEvenWithOneMatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))

	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player:   me.ID,
			Pred:     effects.IsBasicLand,
			Dest:     game.ZoneBattlefield,
			Limit:    1,
			Optional: true,
			Shuffle:  true,
		})
	})

	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("an optional search with one match did not prompt")
	}
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if _, ok := effectstest.SearchFetchedCard(g, "Forest"); ok {
		t.Error("declining an optional search still fetched the land")
	}
}

// The prompt is a look, not a reveal: only the searcher becomes a
// knower of the candidates.
func TestSearchCandidatesAreKnownOnlyToTheSearcher(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, effects.IsBasicLand,
			game.ZoneHand, 1, true, true)
	})

	c := effectstest.SearchChoiceFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)

	g.WithWriteLock(func() {
		// No shuffle, so the untaken candidate's knower set survives
		// for inspection.
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID,
			Pred:   effects.IsBasicLand,
			Dest:   game.ZoneHand,
			Limit:  1,
			Reveal: true,
		})
	})
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)
	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, effects.IsBasicLand,
			game.ZoneBattlefield, 1, false, true)
	})
	c := effectstest.SearchChoiceFor(g, me.ID)
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
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Error("a rejected answer consumed the prompt")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
}

// --- the Then continuation, taken through the prompt ---------------

// Fabled Passage's "then ... untap THAT LAND" names the card the
// SEARCHER chose, several client round-trips after the ability
// resolved. The card used to peek at the library and re-derive the
// first-match pick; with a chooser there is nothing to re-derive, so
// the untap rides the search's Then continuation.
func TestFabledPassageUntapsTheLandTheSearcherChose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	// Three lands plus the fetched one makes four.
	effectstest.SeedPermanentFor(g, me.ID, "Forest A", "Basic Land — Forest")
	effectstest.SeedPermanentFor(g, me.ID, "Forest B", "Basic Land — Forest")
	effectstest.SeedPermanentFor(g, me.ID, "Forest C", "Basic Land — Forest")
	passage := effectstest.PushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", effectstest.FabledPassageOracle, false)
	decoy := effectstest.StapleLibraryCard(me, "Swamp", "Basic Land — Swamp")
	want := effectstest.StapleLibraryCard(me, "Island", "Basic Land — Island")

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, want)

	if _, wrong := effectstest.BattlefieldCard(g, decoy); wrong {
		t.Error("the land the searcher did not pick was fetched")
	}
	got, ok := effectstest.BattlefieldCard(g, want)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	prize := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Prize", TypeLine: "Sorcery"})

	effectstest.CastCatalogSpell(t, g, "Gamble", "Sorcery", effectstest.GambleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// The spell has resolved but the discard has not happened yet —
	// the hand is still empty and the prize is still in the library.
	if me.Graveyard.Contains(prize) {
		t.Error("Gamble discarded before the searcher had picked anything")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, prize)

	if me.Hand.Size() != 0 {
		t.Errorf("hand = %d, want 0 (fetched then pitched)", me.Hand.Size())
	}
	if !me.Graveyard.Contains(prize) {
		t.Error("the fetched card should have been the random discard")
	}
}

// --- the declared limit, closed -----------------------------------

// A fetched shockland asks its owner to pay, and the search waits.
//
// This test used to pin the declared simplification on
// searchEnterBattlefieldLocked: the CR 614 pipeline ran, so the entry
// replacement was consulted, but the site was not entryResumable, so
// the pipeline could not pause to ask and took the un-paid branch. The
// land entered tapped — weaker than printed, never stronger.
//
// #478 closed it. The search's own duties (the shuffle,
// EventSearchLibrary, the Then continuation) ride across the pause on
// the entry tail, so the entry is resumable and the question is asked;
// the land enters untapped when it is paid. Both branches are pinned
// here, because "offered" and "answered correctly" are different
// claims.
func TestFetchedShocklandOffersItsPaymentAndTheSearchWaits(t *testing.T) {
	for _, tc := range []struct {
		name       string
		pay        bool
		wantTapped bool
	}{
		{"pay 2 life", true, false},
		{"decline", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			lifeBefore := me.Life
			effectstest.SeedSearchLibrary(me, game.Card{
				Name:     "Blood Crypt",
				TypeLine: "Land — Swamp Mountain",
				OracleID: effectstest.BloodCryptOracle,
			})
			before := len(g.Events)

			g.WithWriteLock(func() {
				_ = g.SearchLibraryForEffect(me.ID,
					func(c game.Card) bool { return c.Name == "Blood Crypt" },
					game.ZoneBattlefield, 1, false, true)
			})
			effectstest.TakeSearchPrompts(t, g)

			if _, ok := effectstest.SearchFetchedCard(g, "Blood Crypt"); ok {
				t.Fatal("nothing moves while the pay-life question is open")
			}
			if effectstest.PlSearchFired(g, before) {
				t.Error("the search is not finished until the entry is")
			}
			prompt := effectstest.EntryPayLifeChoiceFor(g, me.ID)
			if prompt == nil {
				t.Fatal("the fetched shockland's entry offers the payment")
			}
			if err := g.ResolveEntryPayLife(prompt.ID, me.ID, tc.pay); err != nil {
				t.Fatalf("ResolveEntryPayLife: %v", err)
			}

			card, ok := effectstest.SearchFetchedCard(g, "Blood Crypt")
			if !ok {
				t.Fatal("answering finishes the fetch")
			}
			if card.Tapped != tc.wantTapped {
				t.Errorf("tapped = %v, want %v", card.Tapped, tc.wantTapped)
			}
			wantLife := lifeBefore
			if tc.pay {
				wantLife -= 2
			}
			if me.Life != wantLife {
				t.Errorf("life %d -> %d, want %d", lifeBefore, me.Life, wantLife)
			}
			if !effectstest.PlSearchFired(g, before) {
				t.Error("the search event fires once the entry has settled")
			}
		})
	}
}
