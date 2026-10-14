package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vraska_betrayals_sting_test.go — compleated (ADR 0183, #2946) and
// Vraska, Betrayal's Sting (#2910).

const vraskaBetrayalsStingOracle = "05d09c55-f551-4cb4-a8d2-4811dccc848d"

// castVraska casts Vraska from the active player's hand for a strict
// {4}{B}{B/P}, paying `phyrexianLife` of its Phyrexian symbols with life.
func castVraska(t *testing.T, g *game.Game, phyrexianLife int, pool string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Vraska, Betrayal's Sting", TypeLine: "Legendary Planeswalker — Vraska",
		ManaCost: "{4}{B}{B/P}", Colors: []string{"B"}, StartingLoyalty: 6,
		OracleID: vraskaBetrayalsStingOracle, Owner: me.ID, Controller: me.ID,
	})
	toMainForCost(t, g)
	floatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, PhyrexianLife: phyrexianLife}); err != nil {
		t.Fatalf("cast Vraska: %v", err)
	}
	return id
}

// Paid with {B}: printed loyalty. Paid with 2 life: two fewer.
func TestVraskaCompleatedEntersWithTwoFewerOnlyWhenLifeWasPaid(t *testing.T) {
	for _, tc := range []struct {
		name    string
		life    int
		pool    string
		loyalty int
	}{
		{"paid with mana", 0, "CCCCBBB", 6},
		{"paid with life", 1, "CCCCBB", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			life := me.Life
			v := castVraska(t, g, tc.life, tc.pool)
			passPriorityAroundTable(t, g)
			if got := loyaltyOf(g, v); got != tc.loyalty {
				t.Errorf("loyalty %d, want %d", got, tc.loyalty)
			}
			if got := life - me.Life; got != 2*tc.life {
				t.Errorf("paid %d life, want %d", got, 2*tc.life)
			}
		})
	}
}

// Vraska put onto the battlefield without being cast enters with her
// printed loyalty: nobody paid anything.
func TestVraskaNotCastEntersWithPrintedLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Vraska, Betrayal's Sting", TypeLine: "Legendary Planeswalker — Vraska",
		ManaCost: "{4}{B}{B/P}", StartingLoyalty: 6,
		OracleID: vraskaBetrayalsStingOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("put Vraska onto the battlefield: %v", err)
	}
	if got := loyaltyOf(g, id); got != 6 {
		t.Errorf("loyalty %d, want the printed 6", got)
	}
}

// Doubling Season beside compleated is a CR 616 question for Vraska's
// controller, and the order changes the result.
func TestVraskaCompleatedIsOrderedWithDoublingSeason(t *testing.T) {
	for _, tc := range []struct {
		name          string
		compleatFirst bool
		want          int
	}{
		{"compleated first", true, 8}, // (6-2)*2
		{"doubling first", false, 10}, // 6*2-2
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
			v := castVraska(t, g, 1, "CCCCBB")
			passPriorityAroundTable(t, g)
			if got := loyaltyOf(g, v); got != 0 {
				t.Fatalf("loyalty %d before the CR 616 prompt was answered", got)
			}
			if tc.compleatFirst {
				answerOrderBySource(t, g, v, ds)
			} else {
				answerOrderBySource(t, g, ds, v)
			}
			if got := loyaltyOf(g, v); got != tc.want {
				t.Errorf("loyalty %d, want %d", got, tc.want)
			}
		})
	}
}

// Compleated touches only the starting loyalty: the 0 ability's
// loyalty, and Vraska's later counters, are untouched.
func TestVraskaZeroDrawsLosesOneAndProliferates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	v := castVraska(t, g, 1, "CCCCBB")
	passPriorityAroundTable(t, g)
	awbPadLibrary(me)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 2) })
	life, hand := me.Life, me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	settleAnsweringProliferate(t, g)

	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d; want one card and one life lost", hand, me.Hand.Size(), life, me.Life)
	}
	if got := loyaltyOf(g, v); got != 5 {
		t.Errorf("loyalty %d, want 4 + 1 proliferated", got)
	}
	if got := opp.Counters[game.CounterPoison]; got != 3 {
		t.Errorf("opponent poison %d, want 3 (proliferated)", got)
	}
}

// −2: the creature is now only a Treasure artifact with the Treasure
// mana ability, keeps its supertypes and loses its other abilities.
func TestVraskaMinusTwoMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	v := pushWalkerForTest(g, me.ID, "Vraska, Betrayal's Sting", vraskaBetrayalsStingOracle, 6)
	nighthawk := pushDiesCreatureForTest(g, opp.ID, "Legendary Nighthawk", vampireNighthawkOracle, "Legendary Creature — Vampire Shaman", 2, 3)
	toMainForCost(t, g)

	if err := g.ActivateCatalogAbility(me.ID, v, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: nighthawk}},
	}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)

	ch := gaLayered(t, g, nighthawk)
	if ch.IsCreature() {
		t.Error("it is no longer a creature")
	}
	if !ch.IsArtifact() || !ch.HasSubtype("Treasure") {
		t.Errorf("type line %q, want a Treasure artifact", ch.TypeLine)
	}
	if !ch.HasSupertype("Legendary") {
		t.Error("it keeps its supertypes (ruling)")
	}
	if ch.HasSubtype("Vampire") {
		t.Error("it loses its creature types")
	}
	for _, kw := range []string{"flying", "deathtouch", "lifelink"} {
		if game.HasKeyword(&ch, kw) {
			t.Errorf("it still has %s", kw)
		}
	}
	shapes, _ := game.ManaAbilitiesWithOrigins(ch)
	if len(shapes) != 1 {
		t.Fatalf("%d mana abilities, want the Treasure's one", len(shapes))
	}
	if got := loyaltyOf(g, v); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

// −9 tops a player up to nine poison counters, and does nothing at
// nine (ten is already a loss, CR 704.5c).
func TestVraskaMinusNineTopsUpPoisonToNine(t *testing.T) {
	for _, tc := range []struct{ have, want int }{{0, 9}, {3, 9}, {8, 9}, {9, 9}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
		v := pushWalkerForTest(g, me.ID, "Vraska, Betrayal's Sting", vraskaBetrayalsStingOracle, 9)
		if tc.have > 0 {
			g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, tc.have) })
		}
		toMainForCost(t, g)
		if err := g.ActivateCatalogAbility(me.ID, v, 2, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("−9: %v", err)
		}
		passPriorityAroundTable(t, g)
		if got := opp.Counters[game.CounterPoison]; got != tc.want {
			t.Errorf("from %d poison: %d, want %d", tc.have, got, tc.want)
		}
	}
}
