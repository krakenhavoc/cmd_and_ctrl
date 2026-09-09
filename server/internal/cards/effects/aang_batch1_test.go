package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aang_batch1_test.go — the Azorius flash/blink batch. Every card
// here is expressible with existing machinery; the deck's airbend /
// flicker / waterbend cards are tracked separately as engine work.

const (
	authorityOfTheConsulsOracle = "55f3c721-e13a-406e-bc8e-d6cdc91ac477"
	peregrineDrakeOracle        = "0bd67481-6bd9-48d6-92bd-8933b5ea1eae"
	brinebornCutthroatOracle    = "916cb70f-3b06-48ed-972d-75f805aa0892"
	archivistOfOghmaOracle      = "08b13e1f-27ca-40a8-b5ed-88ac933d24bf"
	loyalWarhoundOracle         = "cc6a83c7-e645-4a53-9550-be79b42cd851"
	slithermuseOracle           = "4b6512aa-535e-4edf-8797-43df2b8463de"
	deputyOfAcquittalsOracle    = "3cbb5045-8566-4279-b7d3-3e599b11ccc5"
	thoughtVesselOracle         = "9965d9c5-2ebf-4a6c-930e-55c5890979be"
	fellwarStoneOracle          = "95560508-7ac9-4be9-8a3f-3c7d5b52807b"
)

// aangPushLand puts a land onto the battlefield under owner, tapped
// or untapped, and returns its instance ID.
func aangPushLand(g *game.Game, owner uuid.UUID, name string, tapped bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Land",
		Owner: owner, Controller: owner, Tapped: tapped,
	})
	return id
}

func aangCardOnBF(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

func aangAdvanceToMain(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepPrecombatMain {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached seat %d precombat main", seat)
}

// --- Authority of the Consuls ------------------------------------

// An opponent's creature enters tapped and the controller gains 1.
func TestAuthorityOfTheConsulsTapsOpponentAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	_ = seedReplacementPermanent(g, authorityOfTheConsulsOracle, "Authority of the Consuls", p0.ID)
	lifeBefore := p0.Life

	aangAdvanceToMain(t, g, 1)
	active := g.Seats[1]
	bear := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: bear, Name: "Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	got, ok := aangCardOnBF(g, bear)
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
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	_ = seedReplacementPermanent(g, authorityOfTheConsulsOracle, "Authority of the Consuls", p0.ID)
	lifeBefore := p0.Life

	bear := castCatalogSpell(t, g, "Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	got, ok := aangCardOnBF(g, bear)
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

// Untaps five of six tapped lands — capped at five, and only the
// controller's.
func TestPeregrineDrakeUntapsUpToFiveOfYourLands(t *testing.T) {
	g := newCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]

	mine := make([]uuid.UUID, 6)
	for i := range mine {
		mine[i] = aangPushLand(g, p0.ID, "Island", true)
	}
	theirs := aangPushLand(g, p1.ID, "Island", true)

	castCatalogSpell(t, g, "Peregrine Drake", "Creature — Drake", peregrineDrakeOracle, nil)
	passPriorityAroundTable(t, g)

	untapped := 0
	for _, id := range mine {
		if c, ok := aangCardOnBF(g, id); ok && !c.Tapped {
			untapped++
		}
	}
	if untapped != 5 {
		t.Errorf("untapped %d of the controller's lands, want exactly 5", untapped)
	}
	if c, ok := aangCardOnBF(g, theirs); ok && !c.Tapped {
		t.Error("Peregrine Drake untapped an opponent's land")
	}
}

// --- Brineborn Cutthroat -----------------------------------------

// Casting on an opponent's turn grows it; casting on your own does not.
func TestBrinebornCutthroatGrowsOnlyOnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	p1 := g.Seats[1]
	cutthroat := pushCatalogPermanent(g, p1.ID, "Brineborn Cutthroat",
		"Creature — Merfolk Pirate", brinebornCutthroatOracle, false)

	// Seat 0 is active: seat 1 flashing in a spell is "during an
	// opponent's turn".
	aangAdvanceToMain(t, g, 0)
	instant := uuid.New()
	p1.Hand.PushTop(game.Card{
		InstanceID: instant, Name: "Impulse", TypeLine: "Instant",
		Owner: p1.ID, Controller: p1.ID,
	})
	if err := g.CastSpell(p1.ID, instant, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell on opponent's turn: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := aangCounterCount(g, cutthroat, "+1/+1"); got != 1 {
		t.Fatalf("counters after casting on an opponent's turn = %d, want 1", got)
	}

	// Now seat 1's own turn — a cast there must not trigger.
	aangAdvanceToMain(t, g, 1)
	own := uuid.New()
	p1.Hand.PushTop(game.Card{
		InstanceID: own, Name: "Impulse", TypeLine: "Instant",
		Owner: p1.ID, Controller: p1.ID,
	})
	if err := g.CastSpell(p1.ID, own, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell on own turn: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := aangCounterCount(g, cutthroat, "+1/+1"); got != 1 {
		t.Errorf("counters after casting on own turn = %d, want still 1", got)
	}
}

func aangCounterCount(g *game.Game, cardID uuid.UUID, kind string) int {
	c, ok := aangCardOnBF(g, cardID)
	if !ok {
		return -1
	}
	return c.Counters[kind]
}

// --- Archivist of Oghma ------------------------------------------

// An opponent's search gains 1 life and draws; your own search does not.
func TestArchivistOfOghmaTriggersOnlyOnOpponentSearch(t *testing.T) {
	g := newCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, p0.ID, "Archivist of Oghma",
		"Creature — Halfling Cleric", archivistOfOghmaOracle, false)

	lifeBefore, handBefore := p0.Life, len(p0.Hand.Cards)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(p1.ID, func(game.Card) bool { return true },
			game.ZoneHand, 1, false, false)
	})
	passPriorityAroundTable(t, g)

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
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	pushLibraryCardForTest(p0, game.Card{
		InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
		Owner: p0.ID, Controller: p0.ID,
	})
	// Opponent ahead 2-0 on lands.
	aangPushLand(g, p1.ID, "Island", false)
	aangPushLand(g, p1.ID, "Island", false)

	castCatalogSpell(t, g, "Loyal Warhound", "Creature — Dog", loyalWarhoundOracle, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	pushLibraryCardForTest(p0, game.Card{
		InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
		Owner: p0.ID, Controller: p0.ID,
	})
	aangPushLand(g, p0.ID, "Plains", false) // 1-0 up

	castCatalogSpell(t, g, "Loyal Warhound", "Creature — Dog", loyalWarhoundOracle, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	muse := pushCatalogPermanent(g, p0.ID, "Slithermuse",
		"Creature — Elemental", slithermuseOracle, false)

	// Put the opponent three cards up.
	for i := 0; i < 3; i++ {
		p1.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "filler",
			Owner: p1.ID, Controller: p1.ID})
	}
	want := len(p0.Hand.Cards) + (len(p1.Hand.Cards) - len(p0.Hand.Cards))

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(muse) })
	passPriorityAroundTable(t, g)

	if len(p0.Hand.Cards) != want {
		t.Errorf("hand %d, want %d (the three-card difference)", len(p0.Hand.Cards), want)
	}
}

// --- Deputy of Acquittals ----------------------------------------

// The optional ETB returns the chosen creature to hand.
func TestDeputyOfAcquittalsReturnsChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	bear := pushCatalogPermanent(g, p0.ID, "Bears", "Creature — Bear", "", false)

	castCatalogSpell(t, g, "Deputy of Acquittals", "Creature — Human Wizard",
		deputyOfAcquittalsOracle, nil)
	passPriorityAroundTable(t, g)

	answerLatestTriggerPrompt(t, g, p0.ID, true)
	pickCard(t, g, p0.ID, bear)
	passPriorityAroundTable(t, g)

	if _, stillOut := aangCardOnBF(g, bear); stillOut {
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
		{"Thought Vessel", thoughtVesselOracle},
		{"Fellwar Stone", fellwarStoneOracle},
	} {
		abilities := game.ManaAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abilities) != 1 {
			t.Errorf("%s: %d mana abilities, want 1", tc.name, len(abilities))
		}
	}
}
