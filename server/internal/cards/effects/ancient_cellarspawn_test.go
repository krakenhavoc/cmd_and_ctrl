package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ancientCellarspawnOracle = "b8510b34-0e5f-4639-9db1-1ba40d84b800"

func TestAncientCellarspawnIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(ancientCellarspawnOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Ancient Cellarspawn: %+v %v", spec, ok)
	}
}

// castUnderCellarspawn casts a spell with the given type line and cost,
// paying `pool`, and reports the first opponent's life change once the
// trigger has resolved, and the cast error.
func castUnderCellarspawn(t *testing.T, typeLine, cost, pool string) (lifeLost int, err error) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Ancient Cellarspawn", "Enchantment Creature — Horror", ancientCellarspawnOracle, false)
	advanceToMain(t, g)
	floatForTest(g, me, pool)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test spell", TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	})
	before := opp.Life
	err = g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	if err != nil {
		return 0, err
	}
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if e := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); e != nil {
			t.Fatalf("pick target: %v", e)
		}
	}
	passPriorityAroundTable(t, g)
	return before - opp.Life, nil
}

func TestAncientCellarspawnDiscountsDemonsHorrorsAndNightmares(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, pool string
		wantErr                    bool
	}{
		{"demon", "Creature — Demon", "{2}{B}", "BB", false},
		{"horror", "Creature — Horror", "{2}{B}", "BB", false},
		{"nightmare", "Creature — Nightmare", "{2}{B}", "BB", false},
		{"horror at full price is still castable", "Creature — Horror", "{2}{B}", "BBB", false},
		{"other creature gets no discount", "Creature — Bear", "{2}{B}", "BB", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := castUnderCellarspawn(t, tc.typeLine, tc.cost, tc.pool)
			if (err != nil) != tc.wantErr {
				t.Fatalf("cast err = %v, want error %v", err, tc.wantErr)
			}
		})
	}
}

func TestAncientCellarspawnDrainsTheDifference(t *testing.T) {
	// A {2}{B} Horror cast for {B}{B}: spent 2, mana value 3 → 1 life.
	if lost, err := castUnderCellarspawn(t, "Creature — Horror", "{2}{B}", "BB"); err != nil || lost != 1 {
		t.Fatalf("discounted Horror: err %v, lost %d, want nil and 1", err, lost)
	}
	// Paid in full: no trigger.
	if lost, err := castUnderCellarspawn(t, "Creature — Bear", "{2}{B}", "BBB"); err != nil || lost != 0 {
		t.Fatalf("full price: err %v, lost %d, want nil and 0", err, lost)
	}
}
