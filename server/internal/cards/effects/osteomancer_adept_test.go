package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const osteomancerOracle = "791cbb7c-7935-4902-b050-9ea9d030e9fa"

func TestOsteomancerAdeptIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(osteomancerOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Osteomancer Adept: %+v %v", spec, ok)
	}
}

// oaBoard seats the Adept, a creature card and `filler` other cards in the
// graveyard, and activates the Adept's tap ability.
func oaBoard(t *testing.T, g *game.Game, me *game.Player, filler int) (creature uuid.UUID, others []uuid.UUID) {
	t.Helper()
	adept := dskB1Creature(g, me.ID, "Osteomancer Adept", "Creature — Squirrel Warlock", osteomancerOracle, 2, 2)
	creature = uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{B}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	for i := 0; i < filler; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		others = append(others, id)
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	return
}

func TestOsteomancerAdeptCastsACreatureFromTheGraveyardByExilingThreeAndGetsAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature, others := oaBoard(t, g, me, 4)
	floatForTest(g, me, "B")
	forages := countEvents(g, game.EventForage)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "osteomancer-forage-exile", AltCostIDs: others[:3]}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if counterOn(g, creature, game.CounterFinality) != 1 {
		t.Errorf("finality counters: %d, want 1", counterOn(g, creature, game.CounterFinality))
	}
	for _, id := range others[:3] {
		if !inExile(g, id) {
			t.Error("a forage card was not exiled")
		}
	}
	if inExile(g, others[3]) {
		t.Error("a fourth card was exiled")
	}
	if got := countEvents(g, game.EventForage) - forages; got != 1 {
		t.Errorf("forage events: %d, want 1", got)
	}
}

func TestOsteomancerAdeptCastsByForagingAFoodToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature, _ := oaBoard(t, g, me, 0)
	food := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Food", TypeLine: "Token Artifact — Food", Owner: me.ID, Controller: me.ID,
	})
	floatForTest(g, me, "B")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "osteomancer-forage-food", AltCostIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, food); ok {
		t.Error("the Food was not sacrificed")
	}
	if counterOn(g, creature, game.CounterFinality) != 1 {
		t.Error("no finality counter")
	}
}

func TestOsteomancerAdeptNeedsThreeOtherCardsOrAFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature, others := oaBoard(t, g, me, 2)
	floatForTest(g, me, "B")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "osteomancer-forage-exile", AltCostIDs: others}); err == nil {
		t.Error("cast by exiling only two cards")
	}
	// Not without the Adept's ability either.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	dead := uuid.New()
	me2.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{B}", Power: 2, Toughness: 2, Owner: me2.ID, Controller: me2.ID})
	advanceToMain(t, g2)
	floatForTest(g2, me2, "B")
	if err := g2.CastSpell(me2.ID, dead, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "osteomancer-forage-exile"}); err == nil {
		t.Error("cast from the graveyard with no Adept activation")
	}
}
