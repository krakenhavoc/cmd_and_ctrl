package testsg

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gorex_the_tombshell_test.go — #5895: "As an additional cost to cast this
// spell, you may exile any number of creature cards from your graveyard.
// This spell costs {2} less to cast for each card exiled this way." and
// "Whenever Gorex attacks or dies, choose a card at random exiled with
// Gorex and put that card into its owner's hand" (CR 601.2b, 601.2f,
// 607.2q).

const gorexOracle = "04238d49-941e-4298-8ef6-8e8e723f4034"

// castGorex puts Gorex in the active seat's hand and casts it in a main
// phase with `params`.
func castGorex(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Gorex, the Tombshell", TypeLine: "Legendary Creature — Zombie Turtle",
		ManaCost: "{6}{B}{B}", OracleID: gorexOracle, Power: 4, Toughness: 4, Colors: []string{"B"},
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	return id, g.CastSpell(me.ID, id, params)
}

// gorexLinked is the cards in exile linked to Gorex as the object it is now.
func gorexLinked(g *game.Game, gorex uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		c := effectstest.FindBattlefieldCardForTest(g, gorex)
		if c == nil {
			return
		}
		for _, e := range g.ExiledWithForEffect(game.PermissionCardRef{ID: gorex, Epoch: c.ObjectEpoch}) {
			out = append(out, e.InstanceID)
		}
	})
	return out
}

// gorexCountIn is how many of `ids` the zone-test `in` holds.
func gorexCountIn(ids []uuid.UUID, in func(uuid.UUID) bool) int {
	n := 0
	for _, id := range ids {
		if in(id) {
			n++
		}
	}
	return n
}

func TestGorexIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gorexOracle)
	if !ok {
		t.Fatal("Gorex, the Tombshell is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// Three creature cards take {6} off: Gorex is cast for {B}{B}, the cards
// are exiled, and the permanent it becomes is linked to all three. The
// count is not part of its mana value (CR 202.3e).
func TestGorexExilesCreatureCardsForTwoEachAndLinksThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fuel := effectstest.ExileXFuel(me, "Creature — Zombie", 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}")

	gorex, err := castGorex(t, g, game.CastSpellParams{XValue: 3, ExileIDs: fuel, Strict: true})
	if err != nil {
		t.Fatalf("three cards exiled with {B}{B} floating: %v", err)
	}
	if got := gorexCountIn(fuel, g.Exile.Contains); got != 3 {
		t.Fatalf("%d of the three named cards were exiled, want 3", got)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool has %d mana left, want {B}{B} spent", got)
	}
	var mv int
	g.ReadSnapshot(func() {
		for _, c := range g.Stack.Cards {
			if c.InstanceID == gorex {
				mv = c.ManaValue()
			}
		}
	})
	if mv != 8 {
		t.Errorf("Gorex's mana value on the stack = %d, want 8", mv)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(gorex) {
		t.Fatal("Gorex did not resolve")
	}
	if got := len(gorexLinked(g, gorex)); got != 3 {
		t.Errorf("%d cards linked to Gorex, want 3 (CR 607.2q)", got)
	}
}

// The price at each count: {2} a card off the generic part, never below
// {B}{B}; one card with only {B}{B} floating is refused and moves nothing.
func TestGorexDiscountIsTwoPerCardAndStopsAtItsColouredMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fuel := effectstest.ExileXFuel(me, "Creature — Zombie", 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}")

	gorex, err := castGorex(t, g, game.CastSpellParams{XValue: 1, ExileIDs: fuel[:1], Strict: true})
	if err == nil {
		t.Fatal("Gorex with one card exiled was cast for {B}{B}")
	}
	if g.Exile.Size() != 0 || me.Graveyard.Size() != 5 {
		t.Errorf("a refused cast moved cards: exile %d, graveyard %d", g.Exile.Size(), me.Graveyard.Size())
	}
	var card game.Card
	for _, c := range me.Hand.Cards {
		if c.InstanceID == gorex {
			card = c
		}
	}
	for x, want := range map[int]int{0: 6, 1: 4, 2: 2, 3: 0, 5: 0} {
		var price game.CastPrice
		var perr error
		g.ReadSnapshot(func() {
			price, perr = g.PriceCastForEffect(me.ID, card, game.CastSpellParams{XValue: x, ExileIDs: fuel[:x]})
		})
		if perr != nil {
			t.Fatalf("price at %d: %v", x, perr)
		}
		if got := price.Total.GenericWithX(x); got != want {
			t.Errorf("generic owed with %d cards exiled = %d, want %d", x, got, want)
		}
		if got := len(price.Total.Required); got != 2 {
			t.Errorf("coloured symbols owed with %d cards exiled = %d, want {B}{B}", x, got)
		}
	}
}

// None is a legal payment: Gorex is cast for its full cost and links
// nothing, and its triggers then find nothing to return.
func TestGorexWithNothingExiledPaysFullAndReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fuel := effectstest.ExileXFuel(me, "Creature — Zombie", 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{B}{B}")
	gorex, err := castGorex(t, g, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("Gorex with nothing exiled and {6}{B}{B} floating: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Size() != 0 || gorexCountIn(fuel, me.Graveyard.Contains) != 2 {
		t.Fatal("a Gorex cast with no exile moved graveyard cards")
	}
	hand := me.Hand.Size()
	effectstest.B27Kill(g, gorex)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand = %d, want %d: nothing was exiled with Gorex", got, hand)
	}
	if gorexCountIn(fuel, me.Graveyard.Contains) != 2 {
		t.Error("Gorex's death trigger touched cards not exiled with it")
	}
}

// Only creature cards in the caster's own graveyard pay, and the count must
// match the announced number.
func TestGorexRefusesAnIllegalExile(t *testing.T) {
	for name, build := range map[string]func(me, opp *game.Player) game.CastSpellParams{
		"a noncreature card": func(me, _ *game.Player) game.CastSpellParams {
			return game.CastSpellParams{XValue: 1, ExileIDs: effectstest.ExileXFuel(me, "Sorcery", 1)}
		},
		"an opponent's creature card": func(_, opp *game.Player) game.CastSpellParams {
			return game.CastSpellParams{XValue: 1, ExileIDs: effectstest.ExileXFuel(opp, "Creature — Zombie", 1)}
		},
		"more than the graveyard holds": func(me, _ *game.Player) game.CastSpellParams {
			return game.CastSpellParams{XValue: 2, ExileIDs: effectstest.ExileXFuel(me, "Creature — Zombie", 1)}
		},
		"fewer named than announced": func(me, _ *game.Player) game.CastSpellParams {
			return game.CastSpellParams{XValue: 2, ExileIDs: effectstest.ExileXFuel(me, "Creature — Zombie", 2)[:1]}
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			params := build(me, opp)
			gorex, err := castGorex(t, g, params)
			if !errors.Is(err, game.ErrInvalidParam) && !errors.Is(err, game.ErrCardNotFound) {
				t.Fatalf("cast = %v, want a refusal", err)
			}
			if g.Stack.Contains(gorex) || g.Exile.Size() != 0 {
				t.Errorf("a refused cast moved cards: stack %v, exile %d", g.Stack.Contains(gorex), g.Exile.Size())
			}
		})
	}
}

// When Gorex dies, one linked card goes to its owner's hand, the rest stay
// in exile; the link survives a save and restore first.
func TestGorexDiesReturnsOneLinkedCardAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fuel := effectstest.ExileXFuel(me, "Creature — Zombie", 3)
	gorex, err := castGorex(t, g, game.CastSpellParams{XValue: 3, ExileIDs: fuel})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	r := effectstest.RestoreRoundTrip(t, g, true)
	if got := len(gorexLinked(r, gorex)); got != 3 {
		t.Fatalf("after a restore %d cards are linked to Gorex, want 3", got)
	}
	rme := r.Seats[r.Turn.ActiveSeat]
	effectstest.B27Kill(r, gorex)
	effectstest.PassPriorityAroundTable(t, r)
	if got := gorexCountIn(fuel, rme.Hand.Contains); got != 1 {
		t.Fatalf("%d linked cards went to hand when Gorex died, want 1", got)
	}
	if got := gorexCountIn(fuel, r.Exile.Contains); got != 2 {
		t.Errorf("%d linked cards still in exile, want 2", got)
	}
}

// Each attack returns one linked card, at random among those still there.
func TestGorexAttackReturnsOneLinkedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fuel := effectstest.ExileXFuel(me, "Creature — Zombie", 2)
	gorex, err := castGorex(t, g, game.CastSpellParams{XValue: 2, ExileIDs: fuel})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, gorex).SummonedThisTurn = false })

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(gorex, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gorexCountIn(fuel, me.Hand.Contains); got != 1 {
		t.Fatalf("%d linked cards went to hand on the attack, want 1", got)
	}
	if got := len(gorexLinked(g, gorex)); got != 1 {
		t.Errorf("%d cards still linked to Gorex, want 1", got)
	}
}

// A countered Gorex paid its cost: the cards stay in exile, linked to
// nothing that exists.
func TestGorexCounteredKeepsTheCardsInExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fuel := effectstest.ExileXFuel(me, "Creature — Zombie", 2)
	gorex, err := castGorex(t, g, game.CastSpellParams{XValue: 2, ExileIDs: fuel})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(gorex); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(gorex) {
		t.Fatal("the countered Gorex is not in its owner's graveyard")
	}
	if got := gorexCountIn(fuel, g.Exile.Contains); got != 2 {
		t.Errorf("%d cards still in exile, want both", got)
	}
}
