package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exiled_doomsayer_test.go — #5844, against the real catalog: "All morph
// costs cost {2} more. (This doesn't affect the cost to cast creature
// spells face down.)" Every payment below is STRICT, so a price the
// engine did not raise would be paid from a pool sized for the raise and
// the refusal half would fail.

const exiledDoomsayerOracle = "7a4c2d47-3489-4e56-b38b-de2dddd942ad"

func doomsayerOnBattlefield(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, controller, "Exiled Doomsayer", "Creature — Human Cleric", exiledDoomsayerOracle, false)
}

// doomsayerMorphed casts `name` face down by its `key` (morph, megamorph
// or disguise) for the active seat, resolves it, and empties the pool.
func doomsayerMorphed(t *testing.T, g *game.Game, name, typeLine, oracle, key string) uuid.UUID {
	t.Helper()
	id := castWithAltCost(t, g, name, typeLine, oracle, key)
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	g.WithWriteLock(func() { g.Seats[g.Turn.ActiveSeat].ManaPool = nil })
	return id
}

// doomsayerManifested seats a manifested creature card for `p`.
func doomsayerManifested(g *game.Game, p *game.Player, name, oracle, manaCost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.Controller = p.ID
	c.TypeLine = "Creature — Human Warrior"
	c.OracleID = oracle
	c.ManaCost = manaCost
	c.Power, c.Toughness = 2, 1
	c.SetFaceDown(game.FaceDownManifested)
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

// turnUpFor tries to turn `id` face up, strictly, from a pool of exactly
// `mana`, and reports the error. The pool is emptied first.
func turnUpFor(g *game.Game, p *game.Player, id uuid.UUID, route game.TurnFaceUpRoute, mana ...string) error {
	g.WithWriteLock(func() {
		p.ManaPool = nil
		p.ManaPool.AddMana(manaTokens(mana...)...)
	})
	return g.PerformSpecialAction(p.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true, Route: route})
}

func colorless(n int, plus ...string) []string {
	out := append([]string(nil), plus...)
	for i := 0; i < n; i++ {
		out = append(out, "C")
	}
	return out
}

func TestExiledDoomsayerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(exiledDoomsayerOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Exiled Doomsayer: %+v %v", spec, ok)
	}
}

// "All morph costs": the controller's own and an opponent's Doomsayer
// both raise Ainok Tracker's morph {4}{R} to {6}{R}.
func TestExiledDoomsayerRaisesEveryPlayersMorphCostByTwo(t *testing.T) {
	for _, tc := range []struct {
		name     string
		opponent bool
	}{
		{"its controller's morph", false},
		{"an opponent's morph", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			owner := me
			if tc.opponent {
				owner = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			}
			doomsayerOnBattlefield(g, owner.ID)
			id := doomsayerMorphed(t, g, "Ainok Tracker", "Creature — Dog Scout", ainokTrackerOracle, "morph")
			if err := turnUpFor(g, me, id, "", colorless(5, "R")...); err == nil {
				t.Fatal("turned face up for {5}{R}: the morph cost is {6}{R} under Exiled Doomsayer")
			}
			if err := turnUpFor(g, me, id, "", colorless(6, "R")...); err != nil {
				t.Fatalf("turn face up for {6}{R}: %v", err)
			}
			if findCardAnywhere(t, g, id).FaceDown {
				t.Fatal("still face down")
			}
			if n := len(me.ManaPool); n != 0 {
				t.Errorf("%d mana left over, want the whole {6}{R} spent", n)
			}
		})
	}
}

// CR 702.37b: "A megamorph cost is a morph cost." Aerie Bowmasters'
// megamorph {5}{G} is {7}{G}, and the counter still rides it.
func TestExiledDoomsayerRaisesAMegamorphCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	doomsayerOnBattlefield(g, me.ID)
	id := doomsayerMorphed(t, g, "Aerie Bowmasters", "Creature — Hound Archer", aerieBowmastersOracle, "megamorph")
	if err := turnUpFor(g, me, id, "", colorless(6, "G")...); err == nil {
		t.Fatal("turned face up for {6}{G}: the megamorph cost is {7}{G} under Exiled Doomsayer")
	}
	if err := turnUpFor(g, me, id, "", colorless(7, "G")...); err != nil {
		t.Fatalf("turn face up for {7}{G}: %v", err)
	}
	if n := findCardAnywhere(t, g, id).Counters["+1/+1"]; n != 1 {
		t.Errorf("+1/+1 counters = %d, want megamorph's one", n)
	}
}

func TestTwoExiledDoomsayersRaiseAMorphCostByFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doomsayerOnBattlefield(g, me.ID)
	doomsayerOnBattlefield(g, opp.ID)
	id := doomsayerMorphed(t, g, "Ainok Tracker", "Creature — Dog Scout", ainokTrackerOracle, "morph")
	if err := turnUpFor(g, me, id, "", colorless(7, "R")...); err == nil {
		t.Fatal("turned face up for {7}{R}: two Doomsayers make it {8}{R}")
	}
	if err := turnUpFor(g, me, id, "", colorless(8, "R")...); err != nil {
		t.Fatalf("turn face up for {8}{R}: %v", err)
	}
}

// The reminder text: casting a card face down still costs {3}, even
// under two Doomsayers.
func TestExiledDoomsayerLeavesTheFaceDownCastAtThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	doomsayerOnBattlefield(g, me.ID)
	doomsayerOnBattlefield(g, me.ID)
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Ainok Tracker", TypeLine: "Creature — Dog Scout",
			OracleID: ainokTrackerOracle, Owner: me.ID, Controller: me.ID,
		})
		me.ManaPool = nil
		me.ManaPool.AddMana(manaTokens("C", "C", "C")...)
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "morph", Strict: true}); err != nil {
		t.Fatalf("cast face down for {3} under two Doomsayers: %v", err)
	}
	if n := len(me.ManaPool); n != 0 {
		t.Errorf("%d mana left over, want exactly {3} spent", n)
	}
}

// A disguise cost is not a morph cost: Fugitive Codebreaker turns up for
// its printed disguise {5}{R}.
func TestExiledDoomsayerLeavesADisguiseCostAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	doomsayerOnBattlefield(g, me.ID)
	id := doomsayerMorphed(t, g, "Fugitive Codebreaker", "Creature — Goblin Rogue", stdCodebreakerOracle, "disguise")
	if err := turnUpFor(g, me, id, "", colorless(5, "R")...); err != nil {
		t.Fatalf("turn a disguised card face up for {5}{R} under Exiled Doomsayer: %v", err)
	}
}

// A manifested creature card's mana cost is not a morph cost
// (CR 701.40b). A manifested Den Protector has both routes (CR 701.40c):
// its mana cost {1}{G} is untouched, its megamorph {1}{G} is {3}{G}.
func TestExiledDoomsayerLeavesAManifestedCardsManaCostAlone(t *testing.T) {
	t.Run("a plain creature card", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		doomsayerOnBattlefield(g, me.ID)
		advanceToMain(t, g)
		id := doomsayerManifested(g, me, "Hill Giant", "oracle-doomsayer-manifested-giant", "{3}{R}")
		if err := turnUpFor(g, me, id, "", colorless(3, "R")...); err != nil {
			t.Fatalf("turn a manifested card face up for its {3}{R}: %v", err)
		}
	})
	for _, tc := range []struct {
		route   game.TurnFaceUpRoute
		short   []string
		enough  []string
		counter int
	}{
		{game.TurnFaceUpByManaCost, nil, colorless(1, "G"), 0},
		{game.TurnFaceUpByMorph, colorless(2, "G"), colorless(3, "G"), 1},
	} {
		t.Run("Den Protector by "+string(tc.route), func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			doomsayerOnBattlefield(g, me.ID)
			advanceToMain(t, g)
			id := doomsayerManifested(g, me, "Den Protector", s36DenProtectorOracle, "{1}{G}")
			if tc.short != nil {
				if err := turnUpFor(g, me, id, tc.route, tc.short...); err == nil {
					t.Fatal("turned face up by the megamorph route for {2}{G}: it is {3}{G} under Exiled Doomsayer")
				}
			}
			if err := turnUpFor(g, me, id, tc.route, tc.enough...); err != nil {
				t.Fatalf("turn face up by %s: %v", tc.route, err)
			}
			if n := findCardAnywhere(t, g, id).Counters["+1/+1"]; n != tc.counter {
				t.Errorf("+1/+1 counters = %d, want %d", n, tc.counter)
			}
		})
	}
}

// A static of the permanent: once Doomsayer is gone the printed {4}{R}
// is the price again.
func TestExiledDoomsayersTaxEndsWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	doomsayer := doomsayerOnBattlefield(g, me.ID)
	id := doomsayerMorphed(t, g, "Ainok Tracker", "Creature — Dog Scout", ainokTrackerOracle, "morph")
	if err := turnUpFor(g, me, id, "", colorless(4, "R")...); err == nil {
		t.Fatal("turned face up for {4}{R} with Exiled Doomsayer on the battlefield")
	}
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(doomsayer) })
	if err != nil {
		t.Fatalf("destroy Exiled Doomsayer: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == doomsayer {
			t.Fatal("Exiled Doomsayer is still on the battlefield")
		}
	}
	if err := turnUpFor(g, me, id, "", colorless(4, "R")...); err != nil {
		t.Fatalf("turn face up for the printed {4}{R} once Doomsayer is gone: %v", err)
	}
}

// The raised price is the one the player's row shows (ChargedCost) and
// the one the bot's enumerator prices the move at.
func TestExiledDoomsayersRaiseIsWhatTheViewAndTheEnumeratorPrice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	doomsayerOnBattlefield(g, me.ID)
	id := doomsayerMorphed(t, g, "Ainok Tracker", "Creature — Dog Scout", ainokTrackerOracle, "morph")

	row := doomsayerTurnUpRow(t, g, me.ID, id)
	if row.Cost != "{4}{R}" {
		t.Errorf("printed cost = %q, want the card's {4}{R}", row.Cost)
	}
	if row.ChargedCost == nil || *row.ChargedCost != "{6}{R}" {
		t.Errorf("charged cost = %v, want {6}{R}", row.ChargedCost)
	}

	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens(colorless(5, "R")...)...) })
	if doomsayerTurnUpOffered(t, g, me.ID, id) {
		t.Error("the enumerator offered the turn face up with {5}{R} available: it costs {6}{R}")
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C")...) })
	if !doomsayerTurnUpOffered(t, g, me.ID, id) {
		t.Error("the enumerator withheld the turn face up with {6}{R} available")
	}
}

func doomsayerTurnUpRow(t *testing.T, g *game.Game, viewer, id uuid.UUID) protocol.SpecialActionView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID != id.String() {
			continue
		}
		for _, sa := range cv.SpecialActions {
			if sa.Kind == string(game.SpecialActionTurnFaceUp) {
				return sa
			}
		}
	}
	t.Fatalf("no turn_face_up row on %s for its controller", id)
	return protocol.SpecialActionView{}
}

func doomsayerTurnUpOffered(t *testing.T, g *game.Game, seat, id uuid.UUID) bool {
	t.Helper()
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeSpecialAction || m.Source != id {
			continue
		}
		var p struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if p.Kind == string(game.SpecialActionTurnFaceUp) {
			return true
		}
	}
	return false
}
