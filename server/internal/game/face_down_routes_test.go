package game

import (
	"testing"

	"github.com/google/uuid"
)

// ADR 0189 decision 4: a manifested or cloaked card whose CARD also
// prints morph or disguise has two ways back up (CR 701.40b/c,
// CR 701.58b/c), and they can cost the same. These tests pin the
// offers (their order, route and label), the route-aware performer, and
// the cases that keep one offer or none.

const routeOracle = "oracle-route-test"

// routeCard seats a face-down permanent of `kind` whose card is a
// creature (or not) printing `declared` as its face-down cast, with the
// given printed mana cost.
func routeCard(t *testing.T, g *Game, owner *Player, typeLine, manaCost string, kind FaceDownKind, declared *AlternativeCost) uuid.UUID {
	t.Helper()
	if declared != nil {
		withCatalogAlternativeCosts(t, altCostFor(routeOracle, *declared))
	} else {
		withCatalogAlternativeCosts(t, altCostFor(routeOracle))
	}
	id := faceDownPermanent(t, g, owner, "Den Protector", typeLine, kind)
	c := findBattlefieldCard(g, id)
	c.OracleID = routeOracle
	c.ManaCost = manaCost
	return id
}

func routeOffers(g *Game, id uuid.UUID) []SpecialAction {
	return TurnFaceUpOffers(*findBattlefieldCard(g, id))
}

func summarize(offers []SpecialAction) []string {
	var out []string
	for _, o := range offers {
		out = append(out, string(o.Route)+"|"+o.Cost+"|"+o.Label)
	}
	return out
}

func wantOffers(t *testing.T, got []SpecialAction, want ...string) {
	t.Helper()
	g := summarize(got)
	if len(g) != len(want) {
		t.Fatalf("offers = %q, want %q", g, want)
	}
	for i := range want {
		if g[i] != want[i] {
			t.Errorf("offer %d = %q, want %q", i, g[i], want[i])
		}
	}
}

func greenPool(g *Game, p *Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.ManaPool.AddMana(ManaToken{Color: "G"})
		}
	})
}

// A manifested Den Protector (megamorph {1}{G}, mana cost {1}{G})
// offers both routes at the same price, the keyword route first.
func TestManifestedMegamorphOffersBothRoutes(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := routeCard(t, g, me, "Creature — Human Warrior", "{1}{G}", FaceDownManifested,
		ptrAltCost(morphOffer(FaceDownMorphed, "{1}{G}", true)))
	wantOffers(t, routeOffers(g, id),
		"morph|{1}{G}|Turn face up {1}{G} (megamorph, +1/+1 counter)",
		"mana_cost|{1}{G}|Turn face up {1}{G} (mana cost)",
	)
}

func TestManifestedMegamorphTurnsUpByEachRoute(t *testing.T) {
	for _, tc := range []struct {
		route       TurnFaceUpRoute
		wantCounter int
	}{
		{TurnFaceUpByMorph, 1},
		{TurnFaceUpByManaCost, 0},
		{"", 1}, // no route takes the first offer
	} {
		t.Run(string(tc.route), func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			id := routeCard(t, g, me, "Creature — Human Warrior", "{1}{G}", FaceDownManifested,
				ptrAltCost(morphOffer(FaceDownMorphed, "{1}{G}", true)))
			greenPool(g, me, 2)
			if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, Route: tc.route}); err != nil {
				t.Fatalf("turn face up: %v", err)
			}
			up := findBattlefieldCard(g, id)
			if up.FaceDown {
				t.Fatal("still face down")
			}
			if n := up.Counters["+1/+1"]; n != tc.wantCounter {
				t.Errorf("+1/+1 counters = %d, want %d", n, tc.wantCounter)
			}
		})
	}
}

// Costs that differ: the cheaper route first, each paid for its own
// price.
func TestManifestedRoutesDifferingCostsCheaperFirst(t *testing.T) {
	cases := []struct {
		name, manaCost, morphCost string
		want                      []string
		route                     TurnFaceUpRoute
		pool                      int
	}{
		{
			name: "morph cheaper", manaCost: "{3}{G}", morphCost: "{1}{G}", route: TurnFaceUpByMorph, pool: 2,
			want: []string{
				"morph|{1}{G}|Turn face up {1}{G} (morph)",
				"mana_cost|{3}{G}|Turn face up {3}{G} (mana cost)",
			},
		},
		{
			name: "mana cost cheaper", manaCost: "{1}{G}", morphCost: "{3}{G}", route: TurnFaceUpByManaCost, pool: 2,
			want: []string{
				"mana_cost|{1}{G}|Turn face up {1}{G} (mana cost)",
				"morph|{3}{G}|Turn face up {3}{G} (morph)",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			id := routeCard(t, g, me, "Creature — Human Warrior", tc.manaCost, FaceDownManifested,
				ptrAltCost(morphOffer(FaceDownMorphed, tc.morphCost, false)))
			wantOffers(t, routeOffers(g, id), tc.want...)
			// Only the chosen route's price is available, and it is
			// paid: the cheaper route works with the small pool.
			greenPool(g, me, tc.pool)
			if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, Route: tc.route}); err != nil {
				t.Fatalf("turn face up by %s: %v", tc.route, err)
			}
			if findBattlefieldCard(g, id).FaceDown {
				t.Fatal("still face down")
			}
		})
	}
}

// Each route is charged ITS price: the dearer route cannot be paid with
// the cheaper one's pool.
func TestManifestedRouteChargesItsOwnPrice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := routeCard(t, g, me, "Creature — Human Warrior", "{3}{G}", FaceDownManifested,
		ptrAltCost(morphOffer(FaceDownMorphed, "{1}{G}", false)))
	greenPool(g, me, 2)
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, Route: TurnFaceUpByManaCost}); err == nil {
		t.Fatal("the {3}{G} mana-cost route was paid with two mana")
	}
	if !findBattlefieldCard(g, id).FaceDown {
		t.Fatal("a refused route turned the card up")
	}
}

// A cloaked disguise card offers disguise and mana cost.
func TestCloakedDisguiseOffersBothRoutes(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := routeCard(t, g, me, "Creature — Human Rogue", "{2}{U}", FaceDownCloaked,
		ptrAltCost(morphOffer(FaceDownDisguised, "{2}{U}", false)))
	wantOffers(t, routeOffers(g, id),
		"disguise|{2}{U}|Turn face up {2}{U} (disguise)",
		"mana_cost|{2}{U}|Turn face up {2}{U} (mana cost)",
	)
}

// A manifested noncreature card with no morph offers nothing; with
// morph, only the morph route, under the unchanged label.
func TestManifestedNoncreatureOffersOnlyItsMorph(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := routeCard(t, g, me, "Sorcery", "{1}{G}", FaceDownManifested, nil)
	wantOffers(t, routeOffers(g, id))

	g = newActiveGame(t)
	me = g.Seats[0]
	id = routeCard(t, g, me, "Sorcery", "{1}{G}", FaceDownManifested,
		ptrAltCost(morphOffer(FaceDownMorphed, "{2}{G}", false)))
	wantOffers(t, routeOffers(g, id), "morph|{2}{G}|Turn face up {2}{G}")
}

// A morphed (cast face down) card has one route, and its label is
// unchanged.
func TestMorphedCardOffersOneRouteWithTheOldLabel(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := routeCard(t, g, me, "Creature — Human Warrior", "{1}{G}", FaceDownMorphed,
		ptrAltCost(morphOffer(FaceDownMorphed, "{1}{G}", true)))
	wantOffers(t, routeOffers(g, id), "morph|{1}{G}|Turn face up {1}{G}")
}

// A route the permanent does not offer is refused, and a route on any
// other kind is refused before anything is paid.
func TestRouteIsRefusedWhereItMeansNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := routeCard(t, g, me, "Creature — Human Warrior", "{1}{G}", FaceDownMorphed,
		ptrAltCost(morphOffer(FaceDownMorphed, "{1}{G}", false)))
	greenPool(g, me, 2)
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, Route: TurnFaceUpByManaCost}); err != ErrSpecialActionNotOffered {
		t.Fatalf("a mana_cost route on a morphed card = %v, want ErrSpecialActionNotOffered", err)
	}
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionForetell, SpecialActionParams{Route: TurnFaceUpByMorph}); err != ErrInvalidParam {
		t.Fatalf("a route on foretell = %v, want ErrInvalidParam", err)
	}
}

func TestParseTurnFaceUpRoute(t *testing.T) {
	for in, ok := range map[string]bool{"": true, "mana_cost": true, "morph": true, "disguise": true, "megamorph": false, "Morph": false} {
		if _, got := ParseTurnFaceUpRoute(in); got != ok {
			t.Errorf("ParseTurnFaceUpRoute(%q) ok = %v, want %v", in, got, ok)
		}
	}
}
