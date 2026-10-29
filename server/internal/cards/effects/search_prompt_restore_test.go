package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// search_prompt_restore_test.go — #4253 slice 5: the library-search prompt
// itself is a restore point. Each test takes a STRICT restore (the boot
// path) while a search prompt is open, checks the restored prompt offers
// what the live one did, and answers it on the restored table: the take,
// the floor (CR 701.23b and 701.23d), the set rule and the card's "then"
// all run off the disk.

// restoreAtSearch takes a strict restore with chooser's search prompt
// open, after checking the prompt was written (its marker on PickThen),
// and returns the restored table and its prompt, which offers what the
// live one did.
func restoreAtSearch(t *testing.T, g *game.Game, chooser uuid.UUID) (*game.Game, *game.PendingChoice) {
	t.Helper()
	before := searchChoiceFor(g, chooser)
	if before == nil {
		t.Fatalf("no search prompt for %s before the restore", chooser)
	}
	written := false
	for _, c := range g.CaptureSnapshot().PendingChoices {
		if c.ID == before.ID && c.PickThen == game.SearchPromptKey && c.SearchPrompt != nil {
			written = true
		}
	}
	if !written {
		t.Fatal("the search prompt was not written as data")
	}
	g = restoreRoundTrip(t, g, true)
	after := searchChoiceFor(g, chooser)
	if after == nil || after.ID != before.ID {
		t.Fatalf("the restored table has no search prompt %s for %s", before.ID, chooser)
	}
	if !slices.Equal(after.SearchCards, before.SearchCards) || after.SearchMin != before.SearchMin ||
		after.SearchMax != before.SearchMax || after.PickThen != "" {
		t.Fatalf("the restored prompt offers %v (%d..%d, pickThen %q), want %v (%d..%d)",
			after.SearchCards, after.SearchMin, after.SearchMax, after.PickThen,
			before.SearchCards, before.SearchMin, before.SearchMax)
	}
	if after.SearchIsOptional() != before.SearchIsOptional() || after.SearchIsUndecided() != before.SearchIsUndecided() {
		t.Fatalf("the restored prompt is optional %v undecided %v, want %v %v",
			after.SearchIsOptional(), after.SearchIsUndecided(), before.SearchIsOptional(), before.SearchIsUndecided())
	}
	return g, after
}

// enumeratedSearchPicks is how many pick moves (not the always-legal
// answer) the enumerator offers seat for its open search prompt.
func enumeratedSearchPicks(g *game.Game, seat uuid.UUID) int {
	n := 0
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind == legal.KindChoice && !m.AlwaysLegal {
			n++
		}
	}
	return n
}

// --- a fetchland ------------------------------------------------------------

// Polluted Delta: the search prompt is a restore point. The restored prompt
// offers the Island and the Swamp cards and not the Forest; a Forest pick is
// refused, and taking Watery Grave goes on to its entry question and
// finishes the fetch (one search event, shuffled).
func TestAFetchlandSearchPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	delta := pushCatalogPermanent(g, me.ID, "Polluted Delta", "Land", pollutedDeltaOracle, false)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Watery Grave", TypeLine: "Land — Island Swamp", OracleID: shocklandOracleIDs["Watery Grave"]},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	grave, island, forest := ids[0], ids[1], ids[2]
	if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)

	g, c := restoreAtSearch(t, g, me.ID)
	me = g.Seats[0]
	if !slices.Contains(c.SearchCards, grave) || !slices.Contains(c.SearchCards, island) || slices.Contains(c.SearchCards, forest) {
		t.Fatalf("the restored prompt offers %v, want the Grave and the Island and not the Forest", c.SearchCards)
	}
	if c.SearchMin != 0 {
		t.Errorf("SearchMin %d: a search for a land type may find nothing (CR 701.23b)", c.SearchMin)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a Forest pick on the restored prompt: %v, want ErrInvalidParam", err)
	}
	answerSearchByID(t, g, me.ID, grave)
	answerEntryPayLife(t, g, me.ID, false)
	card, ok := battlefieldCard(g, grave)
	if !ok || !card.Tapped {
		t.Fatalf("the Grave is on the battlefield %v, tapped %v, want an unpaid shockland (tapped)", ok, card.Tapped)
	}
	if n, shuffled := searchEventsSince(g, 0); n != 1 || !shuffled {
		t.Errorf("%d search events (shuffled %v), want one that shuffled", n, shuffled)
	}
}

// A fetchland whose land word Magical Hack changed: the candidates were
// listed with the word its text now says (CR 612.2), and the restored
// prompt keeps that list though the predicate is not written.
func TestATextChangedFetchlandSearchPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	delta := pushCatalogPermanent(g, me.ID, "Polluted Delta", "Land", twhPollutedDelta, false)
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	island := pushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	twhHack(t, g, delta, twhIsland, twhForest)
	if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	g, c := restoreAtSearch(t, g, me.ID)
	me = g.Seats[g.Turn.ActiveSeat]
	if !slices.Contains(c.SearchCards, forest) || slices.Contains(c.SearchCards, island) {
		t.Fatalf("the restored prompt offers %v, want the Forest card and not the Island card", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{island}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("an Island pick on the restored prompt: %v, want ErrInvalidParam", err)
	}
	answerSearchByID(t, g, me.ID, forest)
	if _, ok := battlefieldCard(g, forest); !ok {
		t.Error("the Forest was not fetched")
	}
}

// --- Demonic Tutor: an unconditional search ----------------------------------

// Demonic Tutor names no quality, so it must find a card (CR 701.23d): the
// restored prompt still refuses an empty answer (the floor rides the
// prompt, not the predicate) and puts the pick into the hand.
func TestDemonicTutorSearchPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Needle", TypeLine: "Artifact"},
		game.Card{Name: "Haystack", TypeLine: "Artifact"},
	)
	castCatalogSpell(t, g, "Demonic Tutor", "Sorcery", "82004860-e589-4e38-8d61-8c0210e4ea39", nil)
	passPriorityAroundTable(t, g)

	g, c := restoreAtSearch(t, g, me.ID)
	me = g.Seats[0]
	if c.SearchMin != 1 {
		t.Fatalf("SearchMin %d, want 1: an unconditional search must find a card", c.SearchMin)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("finding nothing on the restored prompt: %v, want ErrInvalidParam (CR 701.23d)", err)
	}
	answerSearchByID(t, g, me.ID, ids[0])
	if !me.Hand.Contains(ids[0]) {
		t.Error("the tutored card is not in the hand")
	}
}

// --- Green Sun's Zenith: a predicate that read X ------------------------------

// Green Sun's Zenith for X = 1: the candidates were listed with X (green
// creature cards with mana value 1 or less), the restored prompt keeps
// them, and the answer puts the creature onto the battlefield and shuffles
// the Zenith into the library (its registered tail).
func TestGreenSunsZenithSearchPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1},
		game.Card{Name: "Big Elf", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2},
		game.Card{Name: "Merfolk", TypeLine: "Creature — Merfolk", ManaCost: "{U}", Colors: []string{"U"}, Power: 1, Toughness: 1},
	)
	elf := ids[0]
	zenith := castXSpell(t, g, "Green Sun's Zenith", "Sorcery", sweepCGreenSunsZenithOracle, "{X}{G}", 1, nil)
	passPriorityAroundTable(t, g)

	g, c := restoreAtSearch(t, g, me.ID)
	me = g.Seats[0]
	if !slices.Equal(c.SearchCards, []uuid.UUID{elf}) {
		t.Fatalf("the restored prompt offers %v, want only the one-mana green creature", c.SearchCards)
	}
	if c.SearchMin != 0 {
		t.Errorf("SearchMin %d: a search for a green creature card may find nothing", c.SearchMin)
	}
	answerSearchByID(t, g, me.ID, elf)
	if !g.Battlefield.Contains(elf) {
		t.Error("the Elf did not enter")
	}
	if !me.Library.Contains(zenith) {
		t.Error("the Zenith was not shuffled into the library")
	}
}

// --- Cultivate: two searches, each restored ----------------------------------

// Cultivate's first search (a basic onto the battlefield tapped) and its
// second (a basic into the hand) are each a restore point.
func TestCultivateSearchPromptsAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	forest, island, bear := ids[0], ids[1], ids[2]
	castCatalogSpell(t, g, "Cultivate", "Sorcery", cultivateOracle, nil)
	passPriorityAroundTable(t, g)

	g, c := restoreAtSearch(t, g, me.ID)
	me = g.Seats[g.Turn.ActiveSeat]
	if slices.Contains(c.SearchCards, bear) {
		t.Fatalf("the restored prompt offers the Bear: %v", c.SearchCards)
	}
	answerSearchByID(t, g, me.ID, forest)
	if card, ok := battlefieldCard(g, forest); !ok || !card.Tapped {
		t.Fatalf("the Forest is on the battlefield %v, tapped %v, want tapped", ok, card.Tapped)
	}

	g, _ = restoreAtSearch(t, g, me.ID)
	me = g.Seats[g.Turn.ActiveSeat]
	answerSearchByID(t, g, me.ID, island)
	if !me.Hand.Contains(island) {
		t.Error("the Island did not reach the hand")
	}
	if n, shuffled := searchEventsSince(g, 0); n != 2 || !shuffled {
		t.Errorf("%d search events (shuffled %v), want two and the last shuffled", n, shuffled)
	}
}

// --- Myriad Landscape: a two-card search with a set rule ----------------------

// Myriad Landscape's "two basic land cards that share a land type" is a
// registered set rule (search/basics-share-a-land-type). The restored
// prompt refuses a mismatched pair, the enumerator offers what it offered
// before the restore and judges the pairs by the same rule, and a matched
// pair is fetched tapped.
func TestMyriadLandscapeSearchPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	landscape := pushCatalogPermanent(g, me.ID, "Myriad Landscape", "Land", myriadLandscapeOracle, false)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
	)
	forest, plainsA, plainsB := ids[0], ids[1], ids[2]
	if err := g.ActivateCatalogAbility(me.ID, landscape, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	picksBefore := enumeratedSearchPicks(g, me.ID)
	for _, c := range g.CaptureSnapshot().PendingChoices {
		if c.SearchPrompt != nil && c.SearchPrompt.Rule != searchRuleBasicsShareALandType.Key() {
			t.Fatalf("the prompt's rule is %q, want %q", c.SearchPrompt.Rule, searchRuleBasicsShareALandType.Key())
		}
	}

	g, c := restoreAtSearch(t, g, me.ID)
	me = top100ActiveSeat(g)
	if got := enumeratedSearchPicks(g, me.ID); got != picksBefore || got == 0 {
		t.Fatalf("the enumerator offers %d picks after the restore, %d before", got, picksBefore)
	}
	var mixed, matched bool
	g.ReadSnapshot(func() {
		mixed = g.SearchPickLegalLocked(c, []uuid.UUID{forest, plainsA})
		matched = g.SearchPickLegalLocked(c, []uuid.UUID{plainsA, plainsB})
	})
	if mixed || !matched {
		t.Fatalf("the enumerator's window judges a Forest and a Plains %v and two Plains %v, want false and true", mixed, matched)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest, plainsA}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a Forest and a Plains on the restored prompt: %v, want ErrInvalidParam", err)
	}
	answerSearchByID(t, g, me.ID, plainsA, plainsB)
	for _, id := range []uuid.UUID{plainsA, plainsB} {
		if card, ok := battlefieldCard(g, id); !ok || !card.Tapped {
			t.Errorf("a Plains is on the battlefield %v, tapped %v, want tapped", ok, card.Tapped)
		}
	}
}

// A search prompt whose set rule this binary does not know is refused
// rather than restored into a prompt that judges nothing.
func TestASearchPromptWithAnUnknownRuleIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	landscape := pushCatalogPermanent(g, me.ID, "Myriad Landscape", "Land", myriadLandscapeOracle, false)
	seedSearchLibrary(me,
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
	)
	if err := g.ActivateCatalogAbility(me.ID, landscape, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if sp := snap.PendingChoices[i].SearchPrompt; sp != nil {
			sp.Rule = "search/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatal("the search prompt was not written")
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown search rule: %v, want ErrUnknownEffectKey", err)
	}
}

// --- Tempt with Discovery: the caster's and an opponent's searches ------------

// Tempt with Discovery: the caster's opening search, the opponent's search
// after accepting the offer, and the caster's search that pays them back
// are each a restore point.
func TestTemptWithDiscoverySearchPromptsAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	for _, p := range []*game.Player{me, opp1, opp2, opp3} {
		for i := 0; i < 2; i++ {
			b07LibraryLand(p, "Forest", "Basic Land — Forest")
		}
	}
	castCatalogSpell(t, g, "Tempt with Discovery", "Sorcery", b07TemptWithDiscoveryOracle, nil)
	passPriorityAroundTable(t, g)
	search := func(who uuid.UUID) {
		t.Helper()
		var c *game.PendingChoice
		g, c = restoreAtSearch(t, g, who)
		answerSearchByID(t, g, who, c.SearchCards[0])
	}
	search(me.ID)
	ask := latestConfirmFor(g, opp1.ID)
	if ask == nil {
		t.Fatal("the first opponent was not offered a search")
	}
	if err := g.ResolveConfirm(ask.ID, opp1.ID, true); err != nil {
		t.Fatal(err)
	}
	search(opp1.ID)
	search(me.ID)
	for _, opp := range []uuid.UUID{opp2.ID, opp3.ID} {
		ask = latestConfirmFor(g, opp)
		if ask == nil {
			t.Fatal("an opponent was not offered a search")
		}
		if err := g.ResolveConfirm(ask.ID, opp, false); err != nil {
			t.Fatal(err)
		}
	}
	lands := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == g.Seats[0].ID && c.IsLand() {
			lands++
		}
	}
	if lands != 2 {
		t.Errorf("the caster has %d lands, want 2: the opening search and one for the opponent who searched", lands)
	}
}

// --- a "you may search" -------------------------------------------------------

// Assassin's Trophy's victim is asked before the search begins (#4134).
// The undecided prompt is a restore point and is still undecided after the
// restore, so the victim may still decline; once begun it is restored as
// begun, and the basic is fetched.
func TestAnOptionalSearchPromptIsARestorePointBeforeAndAfterItBegins(t *testing.T) {
	g, _, victim, target, basic := declineVictimTable(t)
	castTrophy(t, g, target)

	g, c := restoreAtSearch(t, g, victim.ID)
	if !c.SearchIsOptional() || !c.SearchIsUndecided() {
		t.Fatalf("optional/undecided = %v/%v after the restore, want true/true", c.SearchIsOptional(), c.SearchIsUndecided())
	}
	if err := g.BeginSearchLibrary(c.ID, victim.ID); err != nil {
		t.Fatalf("BeginSearchLibrary on the restored prompt: %v", err)
	}

	g, c = restoreAtSearch(t, g, victim.ID)
	if c.SearchIsUndecided() {
		t.Fatal("the begun search was restored undecided")
	}
	if err := answerSearchApply(g, c, victim.ID, victim.ID, false); err == nil {
		t.Fatal("a begun search was declined after the restore")
	}
	answerSearchByID(t, g, victim.ID, basic)
	if _, ok := battlefieldCard(g, basic); !ok {
		t.Error("the basic was not fetched")
	}
}

// --- a TakeAll search: the graveyard's finds ride the prompt ------------------

// Eradicate: every graveyard copy is found with the searcher's picks from
// the hand and the library. The restored prompt still holds the graveyard's
// finds, so they are exiled with the picks.
func TestEradicateSearchPromptIsARestorePointWithItsGraveyardFinds(t *testing.T) {
	x := extractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	x.permanent("bears", "Grizzly Bears", "Creature — Bear", "G")
	castCatalogSpell(t, x.g, "Eradicate", "Sorcery", eradicateOracle, x.target("bears"))
	passPriorityAroundTable(t, x.g)

	x.g, _ = restoreAtSearch(t, x.g, x.me.ID)
	answerSearchByID(t, x.g, x.me.ID, x.cards["hand"], x.cards["lib"])
	for _, k := range []string{"gy", "gy 2", "hand", "lib"} {
		if !x.g.Exile.Contains(x.cards[k]) {
			t.Errorf("%s was not exiled after the restore", k)
		}
	}
}
