package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestResolutionPickCardsAreRegistered(t *testing.T) {
	for oracleID, name := range map[string]string{
		effectstest.IntuitionOracle:       "Intuition",
		effectstest.GiftsUngivenOracle:    "Gifts Ungiven",
		effectstest.ScapeshiftOracle:      "Scapeshift",
		effectstest.TragicArroganceOracle: "Tragic Arrogance",
	} {
		spec, ok := effects.Lookup(oracleID)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracleID)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracleID, spec.Name)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s ships %q, want full", name, spec.Completeness)
		}
	}
}

// --- Intuition -------------------------------------------------------

// TestIntuitionRevealsThreeAndTheOpponentPicksOne is the whole card:
// the search reveals three, the TARGET picks one, that one goes to the
// caster's hand and the other two to their graveyard.
func TestIntuitionRevealsThreeAndTheOpponentPicksOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seeded := effectstest.SeedTypedLibrary(me, "Artifact", "Int One", "Int Two", "Int Three")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()

	effectstest.CastCatalogSpell(t, g, "Intuition", "Instant", effectstest.IntuitionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	search := effectstest.SearchPromptFor(g, me.ID)
	if search == nil {
		t.Fatalf("the caster searches their own library: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, seeded); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}

	pick := effectstest.LatestRevealPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the target opponent chooses: %+v", g.PendingChoices)
	}
	if pick.FromPlayer != me.ID {
		t.Errorf("the cards belong to the caster: FromPlayer %s", pick.FromPlayer)
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("chooses ONE: %d..%d", pick.ChooseMin, pick.ChooseMax)
	}
	// The reveal is what entitles the opponent to see them at all.
	for _, id := range seeded {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.KnownBy[opp.ID] {
			t.Fatalf("the reveal makes %s public to the chooser", id)
		}
	}

	chosen := pick.ChooseCards[0]
	if err := g.ResolveRevealPick(pick.ID, opp.ID, []uuid.UUID{chosen}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("the chosen card goes to hand: %d → %d", hand, me.Hand.Size())
	}
	// +2 for the unchosen cards and +1 for Intuition itself, which
	// reaches its owner's graveyard as it finishes resolving.
	if me.Graveyard.Size() != graves+3 {
		t.Errorf("the rest go to the graveyard: %d → %d", graves, me.Graveyard.Size())
	}
	if _, ok := effectstest.CardInZone(me.Hand, chosen); !ok {
		t.Errorf("the card in hand is the one the opponent chose")
	}
}

// TestIntuitionDoesNothingWithoutItsTarget — CR 608.2b: the only
// target is gone, so the spell does nothing at all. It is NOT "search
// and keep all three".
func TestIntuitionDoesNothingWithoutItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedTypedLibrary(me, "Artifact", "Int One", "Int Two", "Int Three")
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Intuition", "Instant", effectstest.IntuitionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.SearchPromptFor(g, me.ID) != nil {
		t.Error("the spell searched with no legal target")
	}
	if me.Hand.Size() != hand {
		t.Errorf("nothing moved: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Gifts Ungiven ---------------------------------------------------

// TestGiftsUngivenPutsTheChosenTwoInTheGraveyard — the inversion of
// Intuition's last sentence, and the set-level "with different names"
// rule on the search.
func TestGiftsUngivenPutsTheChosenTwoInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seeded := effectstest.SeedTypedLibrary(me, "Artifact", "Gift One", "Gift Two", "Gift Three", "Gift Four")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()

	effectstest.CastCatalogSpell(t, g, "Gifts Ungiven", "Instant", effectstest.GiftsUngivenOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	search := effectstest.SearchPromptFor(g, me.ID)
	if search == nil {
		t.Fatalf("the caster searches: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, seeded); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}

	pick := effectstest.LatestRevealPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the target opponent chooses two: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Errorf("chooses TWO: %d..%d", pick.ChooseMin, pick.ChooseMax)
	}
	chosen := pick.ChooseCards[:2]
	if err := g.ResolveRevealPick(pick.ID, opp.ID, chosen); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	// +2 for the chosen cards and +1 for Gifts Ungiven itself.
	if me.Graveyard.Size() != graves+3 {
		t.Errorf("the CHOSEN two go to the graveyard: %d → %d", graves, me.Graveyard.Size())
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("the rest go to hand: %d → %d", hand, me.Hand.Size())
	}
	for _, id := range chosen {
		if _, ok := effectstest.CardInZone(me.Graveyard, id); !ok {
			t.Errorf("%s was chosen and belongs in the graveyard", id)
		}
	}
}

// TestGiftsUngivenRefusesTwoCardsWithTheSameName — the set-level rule
// on the SEARCH, which no per-card predicate can state.
func TestGiftsUngivenRefusesTwoCardsWithTheSameName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seeded := effectstest.SeedTypedLibrary(me, "Artifact", "Same Name", "Same Name", "Other Name")

	effectstest.CastCatalogSpell(t, g, "Gifts Ungiven", "Instant", effectstest.GiftsUngivenOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.SearchPromptFor(g, me.ID)
	if search == nil {
		t.Fatalf("the caster searches: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, seeded[:2]); err == nil {
		t.Error("two cards with the same name are not a legal find")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{seeded[0], seeded[2]}); err != nil {
		t.Fatalf("two differently-named cards are a legal find: %v", err)
	}
}

// --- Scapeshift ------------------------------------------------------

// TestScapeshiftSacrificesAnyNumberThenFetchesThatMany — the
// own_permanents kind's card: the count is READ OFF what actually left
// the battlefield.
func TestScapeshiftSacrificesAnyNumberThenFetchesThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lands := make([]uuid.UUID, 0, 3)
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: "Scape Forest", TypeLine: "Basic Land — Forest",
				Owner: me.ID, Controller: me.ID,
			})
			lands = append(lands, id)
		}
	})
	fetchable := effectstest.SeedTypedLibrary(me, "Basic Land — Mountain", "Fetch One", "Fetch Two", "Fetch Three")

	effectstest.CastCatalogSpell(t, g, "Scapeshift", "Sorcery", effectstest.ScapeshiftOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatalf("the caster chooses which of their own lands to sacrifice: %+v", g.PendingChoices)
	}
	if sac.ChooseMin != 0 {
		t.Errorf("ANY NUMBER has a floor of zero, got %d", sac.ChooseMin)
	}
	if sac.ChooseMax != 3 {
		t.Errorf("every land is on offer: ceiling %d, want 3", sac.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, lands[:2]); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}

	search := effectstest.SearchPromptFor(g, me.ID)
	if search == nil {
		t.Fatalf("the search runs after the lands have gone: %+v", g.PendingChoices)
	}
	if search.SearchMax != 2 {
		t.Errorf("up to THAT MANY: ceiling %d, want 2", search.SearchMax)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, fetchable[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	tapped := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Fetch One" || c.Name == "Fetch Two" {
			if c.Tapped {
				tapped++
			}
		}
	}
	if tapped != 2 {
		t.Errorf("the fetched lands enter TAPPED: %d of 2", tapped)
	}
}

// TestScapeshiftSacrificingNothingStillShuffles — a floor of zero is a
// real answer, and the sentence printed after it still runs.
func TestScapeshiftSacrificingNothingStillShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Scape Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID,
		})
	})
	effectstest.CastCatalogSpell(t, g, "Scapeshift", "Sorcery", effectstest.ScapeshiftOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveOwnPermanents(empty): %v", err)
	}
	if effectstest.SearchPromptFor(g, me.ID) != nil {
		t.Error("a search for up to zero cards is not a prompt")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the card finished: %+v", g.PendingChoices)
	}
}

// --- Tragic Arrogance ------------------------------------------------

// TestTragicArroganceKeepsOnePerTypePerPlayer — the their_permanents
// kind's card: ONE seat answers one prompt per player, and the
// continuation runs once after the last of them.
func TestTragicArroganceKeepsOnePerTypePerPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var myBear, myOtherBear, theirBear, theirOtherBear, theirRelic, aLand uuid.UUID
	g.WithWriteLock(func() {
		myBear = effectstest.PushNamedPermanent(g, me.ID, "My Bear", "Creature — Bear")
		myOtherBear = effectstest.PushNamedPermanent(g, me.ID, "My Other Bear", "Creature — Bear")
		theirBear = effectstest.PushNamedPermanent(g, opp.ID, "Their Bear", "Creature — Bear")
		theirOtherBear = effectstest.PushNamedPermanent(g, opp.ID, "Their Other Bear", "Creature — Bear")
		theirRelic = effectstest.PushNamedPermanent(g, opp.ID, "Their Relic", "Artifact")
		aLand = effectstest.PushNamedPermanent(g, me.ID, "My Forest", "Basic Land — Forest")
	})

	effectstest.CastCatalogSpell(t, g, "Tragic Arrogance", "Sorcery", effectstest.TragicArroganceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// One prompt per player with something to keep, all addressed to
	// the caster — and the caster's own leg is the other kind.
	mine := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if mine == nil {
		t.Fatalf("the caster's own board is an own_permanents leg: %+v", g.PendingChoices)
	}
	if mine.ChooseMin != 1 || mine.ChooseMax != 1 {
		t.Errorf("two creatures and nothing else fills one slot: %d..%d", mine.ChooseMin, mine.ChooseMax)
	}
	theirs := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if theirs == nil {
		t.Fatalf("another player's board is a their_permanents leg: %+v", g.PendingChoices)
	}
	if theirs.ChooseMin != 2 || theirs.ChooseMax != 2 {
		t.Errorf("a creature and an artifact fill two slots: %d..%d", theirs.ChooseMin, theirs.ChooseMax)
	}
	// A pick of the RIGHT SIZE that doubles up on one slot is refused
	// by the set-level rule and by nothing else: two creatures are two
	// permanents, but they fill one slot between them, so this is the
	// assertion the bounds cannot make.
	if err := g.ResolveTheirPermanents(theirs.ID, me.ID, []uuid.UUID{theirBear, theirOtherBear}); err == nil {
		t.Error("two creatures cannot fill the creature and artifact slots")
	}
	if err := g.ResolveOwnPermanents(mine.ID, me.ID, []uuid.UUID{myBear, myOtherBear}); err == nil {
		t.Error("a board of two creatures keeps one permanent, not two")
	}

	// Answer every leg; the sacrifice runs once, after the last.
	for _, c := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
		if c == nil {
			continue
		}
		switch c.Kind {
		case game.PendingChoiceOwnPermanents:
			if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{myBear}); err != nil {
				t.Fatalf("ResolveOwnPermanents: %v", err)
			}
		case game.PendingChoiceTheirPermanents:
			keep := []uuid.UUID{theirBear, theirRelic}
			if c.FromPlayer != opp.ID {
				keep = nil
			}
			if len(keep) == 0 {
				t.Fatalf("an unexpected leg over %s: %+v", c.FromPlayer, c)
			}
			if err := g.ResolveTheirPermanents(c.ID, me.ID, keep); err != nil {
				t.Fatalf("ResolveTheirPermanents: %v", err)
			}
		}
	}

	survivors := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		survivors[c.InstanceID] = true
	}
	for _, id := range []uuid.UUID{myBear, theirBear, theirRelic, aLand} {
		if !survivors[id] {
			t.Errorf("%s was chosen (or is a land) and should have survived", id)
		}
	}
	if survivors[myOtherBear] {
		t.Error("the unchosen creature should have been sacrificed")
	}
}
