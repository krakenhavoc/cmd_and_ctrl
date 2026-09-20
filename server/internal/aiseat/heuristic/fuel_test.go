package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// fuel_test.go pins #1013's half of the fix: the heuristic prices a
// card the seat would SPEND to a cost.
//
// The enumerator's half is legal/cost_fuel_test.go. This one is the
// evaluation the issue says has to come first — "the cap is not the
// real constraint, the missing evaluation is" — so the questions here
// are all of the form "would the bot rather keep A or B".

// withGraveyard puts cards in a seat's graveyard.
func withGraveyard(cards ...protocol.CardView) seatOpt {
	return func(p *protocol.PlayerView) {
		p.Graveyard = protocol.ZoneView{Kind: "graveyard", Owner: p.ID, Count: len(cards), Cards: cards}
	}
}

// castable marks a graveyard or exile card as one this seat can still
// cast out of the zone it is in — an escape or flashback card, or a
// granted permission. The server stamps it; a fixture sets it the way
// the wire would.
func castable() cardOpt {
	return func(c *protocol.CardView) { c.CastableHere = true }
}

// graveyardCardView is a card sitting in a graveyard, by type line.
func graveyardCardView(id string, owner int, name, typeLine, cost string, opts ...cardOpt) protocol.CardView {
	c := protocol.CardView{
		InstanceID: id,
		Name:       name,
		Owner:      seatID(owner).String(),
		Controller: seatID(owner).String(),
		TypeLine:   typeLine,
		ManaCost:   cost,
		KnownByYou: true,
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

// fuelPrices runs the policy's CostFuelPrice over a view and returns a
// price per instance ID.
func fuelPrices(t *testing.T, v protocol.GameView, seat int, ids ...string) map[string]float64 {
	t.Helper()
	p := heuristic.New()
	pricer, ok := any(p).(aiseat.CostFuelPricer)
	if !ok {
		t.Fatal("the heuristic is not an aiseat.CostFuelPricer")
	}
	order := pricer.CostFuelPrice(aiseat.Input{View: v, Seat: seatID(seat)})
	out := map[string]float64{}
	for _, id := range ids {
		out[id] = order(legal.TargetCandidate{ID: uuid.MustParse(id)})
	}
	return out
}

// TestTheHeuristicIsACostFuelPricer is #687's compile-time guard with a
// second hook: without it the runner's type assertion would silently
// answer false after a signature change and the ordering would quietly
// stop happening, which is the failure the hook exists to prevent.
//
// It asks the question TWICE, and the second one is the one that
// matters. A bare heuristic.New() is a policy no seat is ever given:
// the lobby seats what tiers.Factory builds, which is this policy
// inside a rules.Filter or a model.Policy. This test passed for a
// month while every shipped tier answered false one layer up — #1060
// — so it now asks the factory as well, exactly the way the runner
// asks (aiseat.Capability, through the wrapper chain).
func TestTheHeuristicIsACostFuelPricer(t *testing.T) {
	if _, ok := any(heuristic.New()).(aiseat.CostFuelPricer); !ok {
		t.Fatal("the heuristic is no longer an aiseat.CostFuelPricer; " +
			"legal.Options.OrderCostFuel would go unset and every escape would eat " +
			"the oldest cards in the graveyard again (#1013)")
	}
	for _, tier := range seatedTiers() {
		p := factoryPolicy(t, tier)
		if _, ok := aiseat.Capability[aiseat.CostFuelPricer](p); !ok {
			t.Errorf("the %s seat the lobby builds (%T) is not an aiseat.CostFuelPricer; "+
				"legal.Options.OrderCostFuel goes unset on every table (#1060)", tier, p)
		}
	}
}

// TestALandInTheGraveyardIsTheCheapestFuel is the issue's own example,
// stated as a comparison. An Uro escaping over a graveyard holding a
// second Uro, a Snapcaster target and three lands must rank the lands
// below the spells — that is the whole behaviour, and everything else
// in this file is a corner of it.
func TestALandInTheGraveyardIsTheCheapestFuel(t *testing.T) {
	const (
		idLand  = "10000000-0000-0000-0000-000000000001"
		idSpell = "10000000-0000-0000-0000-000000000002"
		idUro   = "10000000-0000-0000-0000-000000000003"
	)
	me := newSeat(0, withGraveyard(
		graveyardCardView(idLand, 0, "Forest", "Basic Land — Forest", ""),
		graveyardCardView(idSpell, 0, "Lightning Bolt", "Instant", "{R}"),
		graveyardCardView(idUro, 0, "Uro", "Legendary Creature — Giant Horror", "{1}{G}{U}",
			castable()),
	))
	v := newView([]protocol.PlayerView{me, newSeat(1)})

	price := fuelPrices(t, v, 0, idLand, idSpell, idUro)
	if !(price[idLand] < price[idSpell]) {
		t.Errorf("a land in the graveyard (%.2f) must be cheaper fuel than a spell (%.2f): "+
			"the land is the one the seat cannot use", price[idLand], price[idSpell])
	}
	if !(price[idSpell] < price[idUro]) {
		t.Errorf("a spell the seat cannot recast (%.2f) must be cheaper fuel than a creature it "+
			"could escape (%.2f)", price[idSpell], price[idUro])
	}
	if price[idLand] <= 0 {
		t.Errorf("a land in the graveyard is worth %.2f — a floor of zero makes every "+
			"unreadable card the first thing a cost eats", price[idLand])
	}
}

// TestARecastableGraveyardCardIsWorthLessThanTheSameCardInHand. The
// discount IS the behaviour: a card in the graveyard with escape is a
// real card and it is not a card in hand — it needs its own cost, its
// own window, and it can be exiled out from under the plan.
func TestARecastableGraveyardCardIsWorthLessThanTheSameCardInHand(t *testing.T) {
	const (
		idHand = "10000000-0000-0000-0000-000000000011"
		idYard = "10000000-0000-0000-0000-000000000012"
	)
	body := func(id string, opts ...cardOpt) protocol.CardView {
		return graveyardCardView(id, 0, "Uro", "Legendary Creature — Giant Horror", "{1}{G}{U}", opts...)
	}
	me := newSeat(0, withHand(body(idHand)), withGraveyard(body(idYard, castable())))
	v := newView([]protocol.PlayerView{me, newSeat(1)})

	price := fuelPrices(t, v, 0, idHand, idYard)
	if !(price[idYard] < price[idHand]) {
		t.Errorf("the same card is worth %.2f in the graveyard and %.2f in hand — "+
			"a recastable graveyard card must be the cheaper of the two",
			price[idYard], price[idHand])
	}
	if price[idYard] <= 0 {
		t.Errorf("a recastable graveyard card is worth %.2f", price[idYard])
	}
}

// TestACardTheSeatCannotReadIsNotFreeFuel. An unreadable card — an
// opponent's, a face-down exile, one in a zone this seat may not see —
// is priced as Unknown: small and positive. Zero would make it the
// first thing every cost ate, which is the opposite of careful.
func TestACardTheSeatCannotReadIsNotFreeFuel(t *testing.T) {
	const idUnknown = "10000000-0000-0000-0000-000000000021"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)})

	if got := fuelPrices(t, v, 0, idUnknown)[idUnknown]; got <= 0 {
		t.Errorf("a card the seat cannot see is worth %.2f as fuel, want a small positive "+
			"value — something rather than nothing", got)
	}
}

// TestDazesIslandIsPricedAsThePermanentItIs. The card half of an
// alternative cost is not always a card off the battlefield: Daze
// returns an Island. It is priced by the same boardValue the rest of
// the evaluation uses, so an untapped land costs more to give up than
// a tapped one.
func TestDazesIslandIsPricedAsThePermanentItIs(t *testing.T) {
	const (
		idUntapped = "10000000-0000-0000-0000-000000000031"
		idTapped   = "10000000-0000-0000-0000-000000000032"
	)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(land(idUntapped, 0), land(idTapped, 0, tapped())))

	price := fuelPrices(t, v, 0, idUntapped, idTapped)
	if !(price[idTapped] < price[idUntapped]) {
		t.Errorf("a tapped land (%.2f) must be cheaper to bounce than an untapped one (%.2f)",
			price[idTapped], price[idUntapped])
	}
}
