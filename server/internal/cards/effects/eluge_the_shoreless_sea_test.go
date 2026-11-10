package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const elugeOracle = "027989e7-20c4-4773-b6b4-20337cc5fe8c"

func TestElugeIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(elugeOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Eluge: %+v %v", spec, ok)
	}
}

func elugeBoard(t *testing.T, g *game.Game, me *game.Player) (eluge uuid.UUID) {
	t.Helper()
	pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	return pushCatalogPermanent(g, me.ID, "Eluge, the Shoreless Sea", "Legendary Creature — Elemental Fish", elugeOracle, false)
}

func TestElugeSizeIsTheIslandsYouControlAndFloodMakesALandAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eluge := elugeBoard(t, g, me)
	forest := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.RunStateChecksForTest()
	c, _ := battlefieldCard(g, eluge)
	if e := c.Effective(); e.Power != 2 || e.Toughness != 2 {
		t.Fatalf("Eluge %d/%d with two Islands, want 2/2", e.Power, e.Toughness)
	}
	g.WithWriteLock(func() {
		ctx := NewContext(g, game.NewTriggeredItem(&c, "test"))
		ctx.Item.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: forest}}
		if err := FloodTargetLand(ctx, "Eluge"); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	c, _ = battlefieldCard(g, eluge)
	if e := c.Effective(); e.Power != 3 || e.Toughness != 3 {
		t.Errorf("Eluge %d/%d with a flooded Forest, want 3/3", e.Power, e.Toughness)
	}
}

func TestElugeEntersAndFloodsATargetLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Eluge, the Shoreless Sea", TypeLine: "Legendary Creature — Elemental Fish", OracleID: elugeOracle,
			ManaCost: "{1}{U}{U}{U}", Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "UUUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, forest)
	passPriorityAroundTable(t, g)
	if counterOn(g, forest, "flood") != 1 {
		t.Fatalf("the Forest has %d flood counters, want 1", counterOn(g, forest, "flood"))
	}
	f, _ := battlefieldCard(g, forest)
	if !f.HasSubtype("Island") || !f.HasSubtype("Forest") {
		t.Errorf("flooded land subtypes %v, want Forest and Island", f.Effective().Subtypes)
	}
}

func elugeCast(t *testing.T, g *game.Game, me *game.Player, typeLine, cost, pool string) error {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test spell", TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, pool)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}

func TestElugeDiscountsTheFirstInstantOrSorceryEachTurnPerFloodedLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elugeBoard(t, g, me)
	for i := 0; i < 2; i++ {
		l := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
		g.WithWriteLock(func() { g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Counters = map[string]int{"flood": 1} })
		_ = l
	}
	advanceToMain(t, g)
	// {1}{U}{U}: two flooded lands take off both {U}, leaving {1}.
	if err := elugeCast(t, g, me, "Sorcery", "{1}{U}{U}", "G"); err != nil {
		t.Fatalf("first sorcery should cost {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The second one pays full price.
	if err := elugeCast(t, g, me, "Instant", "{1}{U}", "G"); err == nil {
		t.Error("the second instant or sorcery this turn was discounted")
	}
}

func TestElugeDiscountReducesGenericWhenThereIsNoBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elugeBoard(t, g, me)
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() { g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Counters = map[string]int{"flood": 1} })
	advanceToMain(t, g)
	if err := elugeCast(t, g, me, "Instant", "{1}{R}", "R"); err != nil {
		t.Fatalf("{1}{R} with one flooded land should cost {R}: %v", err)
	}
}

func TestElugeDoesNotDiscountACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elugeBoard(t, g, me)
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() { g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Counters = map[string]int{"flood": 1} })
	advanceToMain(t, g)
	if err := elugeCast(t, g, me, "Creature — Bear", "{1}{G}", "G"); err == nil {
		t.Error("a creature spell was discounted")
	}
}
