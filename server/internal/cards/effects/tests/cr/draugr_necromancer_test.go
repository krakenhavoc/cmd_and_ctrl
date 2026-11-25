package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// draugrOffered reports whether the move enumerator offers me a cast of id.
func draugrOffered(g *game.Game, me, id uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, me) {
		if m.Type == legal.TypeCastSpell && m.Source == id {
			return true
		}
	}
	return false
}

const draugrOracle = "ccdf3399-b836-47ab-802d-af5a9c24d759"

func draugrSnowLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Snow-Covered "+subtype, "Basic Snow Land — "+subtype, "", 0, 0)
}

func draugrPlainLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, subtype, "Basic Land — "+subtype, "", 0, 0)
}

// draugrIcedCard puts an opponent-owned {1}{U} instant into exile with an
// ice counter on it.
func draugrIcedCard(g *game.Game, owner uuid.UUID, counter bool) uuid.UUID {
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: "Iced Instant", TypeLine: "Instant", ManaCost: "{1}{U}",
		Owner: owner, Controller: owner}
	if counter {
		c.Counters = map[string]int{"ice": 1}
	}
	g.WithWriteLock(func() { g.Exile.PushTop(c) })
	return id
}

// The replacement (CR 614): an opponent's nontoken creature that would die
// is exiled with an ice counter; my own creature dies as usual.
func TestDraugrNecromancerExilesOpponentsDyingCreaturesWithIce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.BlbB1Push(g, me.ID, "Draugr Necromancer", "Snow Creature — Zombie Cleric", draugrOracle, 4, 4)
	theirs := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatalf("destroy theirs: %v", err)
		}
		if err := g.DestroyPermanentForEffect(mine); err != nil {
			t.Fatalf("destroy mine: %v", err)
		}
	})
	var exiled *game.Card
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == theirs {
			exiled = &g.Exile.Cards[i]
		}
	}
	if exiled == nil || exiled.Counters["ice"] != 1 {
		t.Fatalf("their creature was not exiled with an ice counter: %+v", exiled)
	}
	if !me.Graveyard.Contains(mine) {
		t.Fatal("my creature did not go to my graveyard")
	}
}

// #6161 acceptance (CR 609.4b, CR 107.4h): an opponent's exiled {1}{U}
// card with an ice counter is cast off two Snow-Covered Swamps (the {U}
// paid with snow black mana), while two plain Swamps cannot pay the {U};
// a card without an ice counter is not castable at all.
func TestDraugrNecromancerCastsIcedCardsWithSnowManaAsAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.BlbB1Push(g, me.ID, "Draugr Necromancer", "Snow Creature — Zombie Cleric", draugrOracle, 4, 4)
	iced := draugrIcedCard(g, opp.ID, true)
	plain := draugrIcedCard(g, opp.ID, false)
	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "exile"})
	}

	swamps := []uuid.UUID{draugrPlainLand(g, me.ID, "Swamp"), draugrPlainLand(g, me.ID, "Swamp")}
	if draugrOffered(g, me.ID, iced) {
		t.Error("the enumerator offers the iced card off two nonsnow Swamps")
	}
	if err := cast(iced); err == nil {
		t.Fatal("the {U} was paid with a nonsnow Swamp's mana")
	}
	for _, s := range swamps {
		if c := effectstest.FindBattlefieldCardByID(g, s); c == nil || c.Tapped {
			t.Fatal("a refused cast tapped a Swamp")
		}
	}
	snow := []uuid.UUID{draugrSnowLand(g, me.ID, "Swamp"), draugrSnowLand(g, me.ID, "Swamp")}
	if !draugrOffered(g, me.ID, iced) {
		t.Error("the enumerator does not offer the iced card off two Snow-Covered Swamps")
	}
	if err := cast(plain); err == nil {
		t.Fatal("a card without an ice counter was cast from exile")
	}
	if err := cast(iced); err != nil {
		t.Fatalf("cast the iced {1}{U} off two Snow-Covered Swamps: %v", err)
	}
	tapped := 0
	for _, s := range snow {
		if c := effectstest.FindBattlefieldCardByID(g, s); c != nil && c.Tapped {
			tapped++
		}
	}
	if tapped < 1 {
		t.Fatal("the plan paid the {U} without a snow land")
	}
}

// The pool path: floating snow black mana pays the {U}; floating nonsnow
// black mana does not.
func TestDraugrNecromancerFloatingSnowManaPaysTheBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.BlbB1Push(g, me.ID, "Draugr Necromancer", "Snow Creature — Zombie Cleric", draugrOracle, 4, 4)
	iced := draugrIcedCard(g, opp.ID, true)
	cast := func() error {
		return g.CastSpell(me.ID, iced, game.CastSpellParams{Strict: true, FromZone: "exile"})
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B", Source: uuid.New(), SourceKinds: game.ManaSourceLand},
		game.ManaToken{Color: "B", Source: uuid.New(), SourceKinds: game.ManaSourceLand})
	if err := cast(); err == nil {
		t.Fatal("nonsnow {B}{B} paid {1}{U}")
	}
	me.ManaPool = nil
	me.ManaPool.AddMana(game.ManaToken{Color: "B", Source: uuid.New(), SourceKinds: game.ManaSourceLand | game.ManaSourceSnow},
		game.ManaToken{Color: "B", Source: uuid.New(), SourceKinds: game.ManaSourceLand})
	if err := cast(); err != nil {
		t.Fatalf("snow {B} and nonsnow {B} for {1}{U}: %v", err)
	}
}
