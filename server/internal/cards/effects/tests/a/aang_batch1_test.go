package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Authority of the Consuls ------------------------------------

// An opponent's creature enters tapped and the controller gains 1.
func TestAuthorityOfTheConsulsTapsOpponentAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0 := g.Seats[0]
	_ = effectstest.SeedReplacementPermanent(g, effectstest.AuthorityOfTheConsulsOracle, "Authority of the Consuls", p0.ID)
	lifeBefore := p0.Life

	effectstest.AangAdvanceToMain(t, g, 1)
	active := g.Seats[1]
	bear := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: bear, Name: "Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	got, ok := effectstest.AangCardOnBF(g, bear)
	if !ok {
		t.Fatal("bear never reached the battlefield")
	}
	if !got.Tapped {
		t.Error("opponent's creature entered untapped; Authority should tap it")
	}
	if p0.Life != lifeBefore+1 {
		t.Errorf("controller life %d, want %d", p0.Life, lifeBefore+1)
	}
}

// Your own creatures are untouched — neither tapped nor lifegain.
func TestAuthorityOfTheConsulsIgnoresYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0 := g.Seats[0]
	_ = effectstest.SeedReplacementPermanent(g, effectstest.AuthorityOfTheConsulsOracle, "Authority of the Consuls", p0.ID)
	lifeBefore := p0.Life

	bear := effectstest.CastCatalogSpell(t, g, "Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)

	got, ok := effectstest.AangCardOnBF(g, bear)
	if !ok {
		t.Fatal("bear never reached the battlefield")
	}
	if got.Tapped {
		t.Error("controller's own creature entered tapped")
	}
	if p0.Life != lifeBefore {
		t.Errorf("controller gained life off their own creature: %d → %d", lifeBefore, p0.Life)
	}
}

// --- Peregrine Drake ---------------------------------------------

// "Untap up to five lands" is the controller's choice, made on
// resolution, among EVERY tapped land at the table: an opponent's land
// is offered and untaps when picked, the bound is five, and unpicked
// lands stay tapped.
func TestPeregrineDrakeUntapsTheLandsItsControllerChooses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]

	mine := make([]uuid.UUID, 6)
	for i := range mine {
		mine[i] = effectstest.AangPushLand(g, p0.ID, "Island", true)
	}
	theirs := effectstest.AangPushLand(g, p1.ID, "Forest", true)
	alreadyUp := effectstest.AangPushLand(g, p0.ID, "Plains", false)

	effectstest.CastCatalogSpell(t, g, "Peregrine Drake", "Creature — Drake", effectstest.PeregrineDrakeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, p0.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt: the lands are not the controller's choice")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 5 {
		t.Errorf("bounds %d..%d, want 0..5 (\"up to five\")", pick.ChooseMin, pick.ChooseMax)
	}
	if !effectstest.HasID(pick.ChooseCards, theirs) {
		t.Error("an opponent's tapped land is not offered — the clause says no \"you control\"")
	}
	if effectstest.HasID(pick.ChooseCards, alreadyUp) {
		t.Error("an untapped land is offered; untapping it would do nothing")
	}
	if err := g.ResolveChooseCards(pick.ID, p0.ID, append(append([]uuid.UUID{}, mine...), theirs)[:6]); err == nil {
		t.Fatal("six lands accepted for \"up to five\"")
	}

	effectstest.AnswerChooseCards(t, g, p0.ID, mine[0], mine[1], mine[2], mine[3], theirs)
	effectstest.PassPriorityAroundTable(t, g)

	for i, id := range mine {
		c, _ := effectstest.AangCardOnBF(g, id)
		if want := i >= 4; c.Tapped != want {
			t.Errorf("my land %d tapped=%v, want %v", i, c.Tapped, want)
		}
	}
	if c, _ := effectstest.AangCardOnBF(g, theirs); c.Tapped {
		t.Error("the chosen opponent's land did not untap")
	}
}

// --- Brineborn Cutthroat -----------------------------------------

// Casting on an opponent's turn grows it; casting on your own does not.
func TestBrinebornCutthroatGrowsOnlyOnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p1 := g.Seats[1]
	cutthroat := effectstest.PushCatalogPermanent(g, p1.ID, "Brineborn Cutthroat",
		"Creature — Merfolk Pirate", effectstest.BrinebornCutthroatOracle, false)

	// Seat 0 is active: seat 1 flashing in a spell is "during an
	// opponent's turn".
	effectstest.AangAdvanceToMain(t, g, 0)
	instant := uuid.New()
	p1.Hand.PushTop(game.Card{
		InstanceID: instant, Name: "Impulse", TypeLine: "Instant",
		Owner: p1.ID, Controller: p1.ID,
	})
	if err := g.CastSpell(p1.ID, instant, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell on opponent's turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.AangCounterCount(g, cutthroat, "+1/+1"); got != 1 {
		t.Fatalf("counters after casting on an opponent's turn = %d, want 1", got)
	}

	// Now seat 1's own turn — a cast there must not trigger.
	effectstest.AangAdvanceToMain(t, g, 1)
	own := uuid.New()
	p1.Hand.PushTop(game.Card{
		InstanceID: own, Name: "Impulse", TypeLine: "Instant",
		Owner: p1.ID, Controller: p1.ID,
	})
	if err := g.CastSpell(p1.ID, own, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell on own turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.AangCounterCount(g, cutthroat, "+1/+1"); got != 1 {
		t.Errorf("counters after casting on own turn = %d, want still 1", got)
	}
}

// --- Archivist of Oghma ------------------------------------------

// An opponent's search gains 1 life and draws; your own search does not.
func TestArchivistOfOghmaTriggersOnlyOnOpponentSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, p0.ID, "Archivist of Oghma",
		"Creature — Halfling Cleric", effectstest.ArchivistOfOghmaOracle, false)

	lifeBefore, handBefore := p0.Life, len(p0.Hand.Cards)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(p1.ID, func(game.Card) bool { return true },
			game.ZoneHand, 1, false, false)
	})
	// S22: an accept-all predicate over a real library always offers
	// a choice, so the searcher answers before the search finishes
	// and EventSearchLibrary fires.
	effectstest.AnswerSearchByID(t, g, p1.ID, p1.Library.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)

	if p0.Life != lifeBefore+1 {
		t.Errorf("life %d, want %d after opponent searched", p0.Life, lifeBefore+1)
	}
	if len(p0.Hand.Cards) != handBefore+1 {
		t.Errorf("hand %d, want %d after opponent searched", len(p0.Hand.Cards), handBefore+1)
	}

	lifeMid, handMid := p0.Life, len(p0.Hand.Cards)
	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(p0.ID, func(game.Card) bool { return true },
			game.ZoneHand, 1, false, false)
	})
	effectstest.AnswerSearchByID(t, g, p0.ID, p0.Library.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)

	if p0.Life != lifeMid {
		t.Errorf("own search gained life: %d → %d", lifeMid, p0.Life)
	}
	// +1 from the search itself, never +2.
	if len(p0.Hand.Cards) != handMid+1 {
		t.Errorf("own search drew an extra card: %d → %d", handMid, len(p0.Hand.Cards))
	}
}

// --- Loyal Warhound ----------------------------------------------

// Fetches only while an opponent is ahead on lands.
func TestLoyalWarhoundFetchesOnlyWhenBehindOnLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	effectstest.PushLibraryCardForTest(p0, game.Card{
		InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
		Owner: p0.ID, Controller: p0.ID,
	})
	// Opponent ahead 2-0 on lands.
	effectstest.AangPushLand(g, p1.ID, "Island", false)
	effectstest.AangPushLand(g, p1.ID, "Island", false)

	effectstest.CastCatalogSpell(t, g, "Loyal Warhound", "Creature — Dog", effectstest.LoyalWarhoundOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plains" && c.Controller == p0.ID {
			found = true
			if !c.Tapped {
				t.Error("fetched Plains entered untapped; card says tapped")
			}
		}
	}
	if !found {
		t.Error("Loyal Warhound did not fetch while behind on lands")
	}
}

func TestLoyalWarhoundDoesNotFetchWhenAhead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0 := g.Seats[0]
	effectstest.PushLibraryCardForTest(p0, game.Card{
		InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
		Owner: p0.ID, Controller: p0.ID,
	})
	effectstest.AangPushLand(g, p0.ID, "Plains", false) // 1-0 up

	effectstest.CastCatalogSpell(t, g, "Loyal Warhound", "Creature — Dog", effectstest.LoyalWarhoundOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	plains := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plains" && c.Controller == p0.ID {
			plains++
		}
	}
	if plains != 1 {
		t.Errorf("controller controls %d Plains, want 1 (no fetch while ahead)", plains)
	}
}

// --- Slithermuse -------------------------------------------------

// Leaving the battlefield draws the hand-size difference.
func TestSlithermuseDrawsHandSizeDifference(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	muse := effectstest.PushCatalogPermanent(g, p0.ID, "Slithermuse",
		"Creature — Elemental", effectstest.SlithermuseOracle, false)

	// Put the opponent three cards up.
	for i := 0; i < 3; i++ {
		p1.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "filler",
			Owner: p1.ID, Controller: p1.ID})
	}
	want := len(p0.Hand.Cards) + (len(p1.Hand.Cards) - len(p0.Hand.Cards))

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(muse) })
	effectstest.PassPriorityAroundTable(t, g)
	// #929: "choose an opponent" is a prompt now, not an auto-pick.
	effectstest.AnswerChoosePlayer(t, g, p0.ID, p1)

	if len(p0.Hand.Cards) != want {
		t.Errorf("hand %d, want %d (the three-card difference)", len(p0.Hand.Cards), want)
	}
}

// --- Deputy of Acquittals ----------------------------------------

// The optional ETB returns the chosen creature to hand.
func TestDeputyOfAcquittalsReturnsChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0 := g.Seats[0]
	bear := effectstest.PushCatalogPermanent(g, p0.ID, "Bears", "Creature — Bear", "", false)

	effectstest.CastCatalogSpell(t, g, "Deputy of Acquittals", "Creature — Human Wizard",
		effectstest.DeputyOfAcquittalsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.PickCard(t, g, p0.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, p0.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if _, stillOut := effectstest.AangCardOnBF(g, bear); stillOut {
		t.Error("chosen creature is still on the battlefield")
	}
	inHand := false
	for _, c := range p0.Hand.Cards {
		if c.InstanceID == bear {
			inHand = true
		}
	}
	if !inHand {
		t.Error("chosen creature did not return to its owner's hand")
	}
}

// --- Mana rocks ---------------------------------------------------

func TestAangBatchManaRocksAreWired(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Thought Vessel", effectstest.ThoughtVesselOracle},
		{"Fellwar Stone", effectstest.FellwarStoneOracle},
	} {
		abilities := game.ManaAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abilities) != 1 {
			t.Errorf("%s: %d mana abilities, want 1", tc.name, len(abilities))
		}
	}
}
