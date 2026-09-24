package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escape_test.go — S29. The game package pins the mechanic
// (escape_test.go there: the count, the distinctness, "other",
// "your"); this file pins the seven cards against the REAL catalog.
//
// That is the half a stubbed hook cannot check — that each card file
// declared BOTH the zone and the price, that Register accepted the
// pair, and that the offer the client would render is the offer the
// cast path accepts.

// seedGraveyardFodder puts n throwaway cards in the active seat's
// graveyard and returns their IDs. Escape's cost is the only one in
// the catalog that needs a stocked graveyard to be payable at all.
func seedGraveyardFodder(t *testing.T, g *game.Game, n int) []uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := game.NewCard("Fodder", active.ID)
		c.TypeLine = "Instant"
		active.Graveyard.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}

// escapeCatalogCards is the S29 escape roster, with the numbers a
// card file is most likely to get wrong: the cost string, the exile
// count, and the counters the creature escapes with.
var escapeCatalogCards = []struct {
	name     string
	oracle   string
	cost     string
	exile    int
	counters int
}{
	{"Voracious Typhon", "f62784c0-9c67-4a05-a09b-aabf03a9390f", "{5}{G}{G}", 4, 3},
	{"Loathsome Chimera", "fe983088-6e0d-4544-90bd-f5ebb95c3418", "{4}{G}", 3, 1},
	{"Ox of Agonas", "22113051-c971-4108-953b-95356d21323a", "{R}{R}", 8, 1},
	{"Fruit of Tizerus", "8d6ad0a0-3b71-4fab-8874-470285c39299", "{3}{B}", 3, 0},
	{"Glimpse of Freedom", "02d78e76-8151-4851-bb4e-e0a1fa98756f", "{2}{U}", 5, 0},
	{"Sweet Oblivion", "c023538d-2feb-4af5-a44e-b355a190f081", "{3}{U}", 4, 0},
	{"Cling to Dust", "9c44e556-4c7a-48bf-a108-c584f69c2cfa", "{3}{B}", 5, 0},
	// #653: the three cards that needed the permanent to remember it
	// escaped (CR 400.7d).
	{"Phlage, Titan of Fire's Fury", "3407eb6e-b74d-4159-a801-d7163937953c", "{R}{R}{W}{W}", 5, 0},
	{"Uro, Titan of Nature's Wrath", "ee302659-59ed-4eef-babe-451b9ccf7f14", "{G}{G}{U}{U}", 5, 0},
	{"Pharika's Spawn", "a44955e7-f1ad-41b9-b93d-0a980ac341d8", "{5}{B}", 3, 2},
	{"Woe Strider", woeStriderOracle, "{3}{B}{B}", 4, 2},
}

// The declaration test. Catches a card file that named the cost but
// forgot the zone — Register panics on the opposite mistake, and
// this is the direction nothing else notices.
func TestEscapeCardsDeclareBothHalves(t *testing.T) {
	for _, c := range escapeCatalogCards {
		t.Run(c.name, func(t *testing.T) {
			if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Fatalf("%s does not declare the graveyard castable", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "escape" {
				t.Fatalf("%s graveyard offers = %+v, want one escape", c.name, offers)
			}
			o := offers[0]
			if o.ManaCost != c.cost {
				t.Errorf("%s escape cost: got %q, want %q", c.name, o.ManaCost, c.cost)
			}
			if o.ExileFromGraveyard == nil {
				t.Fatalf("%s escape offer has no exile component — it is a free cast", c.name)
			}
			if got := o.ExileFromGraveyard.Min; got != c.exile {
				t.Errorf("%s exiles %d cards, want %d", c.name, got, c.exile)
			}
			if o.ExileFromGraveyard.Min != o.ExileFromGraveyard.Max {
				t.Errorf("%s exile count is a range (%d..%d), not a price",
					c.name, o.ExileFromGraveyard.Min, o.ExileFromGraveyard.Max)
			}
			if got := o.EntersWithCounterCount; got != c.counters {
				t.Errorf("%s escapes with %d counters, want %d", c.name, got, c.counters)
			}
			// The flashback contrast, asserted on every card so a
			// copy-paste from lingering_souls.go cannot slip through:
			// escape has no CR 702.34a clause.
			if o.ExileOnLeavingStack {
				t.Errorf("%s exiles itself on leaving the stack — that is flashback, not escape", c.name)
			}
			// And the offer is not claimable from hand.
			if got := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneHand); len(got) != 0 {
				t.Errorf("%s offers %+v from hand", c.name, got)
			}
		})
	}
}

// Voracious Typhon — the reference case. A 4/4 in the corner and a
// 7/7 out of the graveyard, and the card stays a card: it goes to
// the battlefield rather than to exile.
func TestVoraciousTyphonEscapesWithThreeCounters(t *testing.T) {
	const oracle = "f62784c0-9c67-4a05-a09b-aabf03a9390f"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Voracious Typhon", "Creature — Snake Beast", oracle)
	pay := seedGraveyardFodder(t, g, 4)

	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	for _, p := range pay {
		if !g.Exile.Contains(p) {
			t.Errorf("an exiled-as-a-cost card is not in exile")
		}
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatalf("the escaped Typhon is not on the battlefield")
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("the escaped Typhon vanished")
	}
	if got := c.Counters["+1/+1"]; got != 3 {
		t.Errorf("+1/+1 counters: got %d, want 3", got)
	}
}

// Ox of Agonas — the escape cost that is CHEAPER than the printed
// cost, and a wheel on entry. Both halves in one cast.
func TestOxOfAgonasEscapesAndWheels(t *testing.T) {
	const oracle = "22113051-c971-4108-953b-95356d21323a"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Ox of Agonas", "Creature — Ox", oracle)
	pay := seedGraveyardFodder(t, g, 8)

	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatalf("the escaped Ox is not on the battlefield")
	}
	if got := active.Hand.Size(); got != 3 {
		t.Errorf("hand after the wheel: got %d, want 3", got)
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("the escaped Ox vanished")
	}
	if got := c.Counters["+1/+1"]; got != 1 {
		t.Errorf("+1/+1 counters: got %d, want 1", got)
	}
}

// Fruit of Tizerus — a sorcery, and the contrast the whole mechanic
// turns on: an escaped spell returns to the graveyard and can escape
// again. A flashed-back one never can.
func TestFruitOfTizerusEscapeReturnsToTheGraveyard(t *testing.T) {
	const oracle = "8d6ad0a0-3b71-4fab-8874-470285c39299"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := seedGraveyardCard(t, g, "Fruit of Tizerus", "Sorcery", oracle)
	pay := seedGraveyardFodder(t, g, 3)
	lifeBefore := victim.Life

	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "escape",
		AltCostIDs:      pay,
		Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := victim.Life; got != lifeBefore-2 {
		t.Errorf("target life: got %d, want %d", got, lifeBefore-2)
	}
	if g.Exile.Contains(id) {
		t.Fatalf("an escaped sorcery was exiled — escape has no flashback clause")
	}
	if !active.Graveyard.Contains(id) {
		t.Errorf("an escaped sorcery did not return to the graveyard")
	}
}

// Cling to Dust — the escape card whose cost and effect pull in
// opposite directions: it exiles one card from ANY graveyard and
// five from your own.
func TestClingToDustEscapeExilesOneAndPaysFive(t *testing.T) {
	const oracle = "9c44e556-4c7a-48bf-a108-c584f69c2cfa"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := seedGraveyardCard(t, g, "Cling to Dust", "Instant", oracle)
	pay := seedGraveyardFodder(t, g, 5)

	// A creature card in the opponent's graveyard: the "if it was a
	// creature card" branch gains 3 life rather than drawing.
	prey := game.NewCard("Grizzly Bears", opp.ID)
	prey.TypeLine = "Creature — Bear"
	opp.Graveyard.PushTop(prey)

	lifeBefore := active.Life
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "escape",
		AltCostIDs:      pay,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: prey.InstanceID}},
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(prey.InstanceID) {
		t.Errorf("the targeted creature card was not exiled")
	}
	if got := active.Life; got != lifeBefore+3 {
		t.Errorf("life after exiling a creature card: got %d, want %d", got, lifeBefore+3)
	}
	for _, p := range pay {
		if !g.Exile.Contains(p) {
			t.Errorf("a card paid to the escape cost is not in exile")
		}
	}
}

// The real-card version of the game package's count test: a cast
// that names four cards for a five-card cost is refused, and the
// card file's number is what says five.
func TestClingToDustRefusesAShortPayment(t *testing.T) {
	const oracle = "9c44e556-4c7a-48bf-a108-c584f69c2cfa"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Cling to Dust", "Instant", oracle)
	pay := seedGraveyardFodder(t, g, 5)

	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay[:4],
	})
	if err == nil {
		t.Fatalf("a four-card payment bought Cling to Dust's five-card escape cost")
	}
	if !active.Graveyard.Contains(id) {
		t.Errorf("the rejected cast moved the spell out of the graveyard")
	}
}

// An escape card cast out of HAND is an ordinary cast: printed cost,
// no exiles, no counters. The zone is the place and the cost is the
// price, and neither is implied by the other.
func TestEscapeCardCastFromHandIsOrdinary(t *testing.T) {
	const oracle = "f62784c0-9c67-4a05-a09b-aabf03a9390f"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	fodder := seedGraveyardFodder(t, g, 4)

	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Voracious Typhon", TypeLine: "Creature — Snake Beast",
		Power: 4, Toughness: 4, OracleID: oracle, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(id)
	if !ok || !g.Battlefield.Contains(id) {
		t.Fatalf("the hard-cast Typhon is not on the battlefield")
	}
	if got := c.Counters["+1/+1"]; got != 0 {
		t.Errorf("a hard-cast Typhon entered with %d escape counters", got)
	}
	for _, f := range fodder {
		if !active.Graveyard.Contains(f) {
			t.Errorf("a hand cast paid the escape cost")
		}
	}
}
