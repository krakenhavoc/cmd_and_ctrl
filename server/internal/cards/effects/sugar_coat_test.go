package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const sugarCoatOracle = "1caa8c8a-ff45-4eee-a834-6e5e1fb54721"

func TestSugarCoatIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(sugarCoatOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Sugar Coat: %+v %v", spec, ok)
	}
}

func scCast(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sugar Coat", TypeLine: "Enchantment — Aura", OracleID: sugarCoatOracle,
			ManaCost: "{2}{U}", Colors: []string{"U"}, Keywords: []string{"flash"}, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "UUU")
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
}

func TestSugarCoatMakesTheCreatureAColorlessFoodArtifactWithTheFoodAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Flyer", TypeLine: "Legendary Creature — Dragon", ManaCost: "{4}{R}", Colors: []string{"R"},
		Keywords: []string{"flying", "haste"}, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if _, err := scCast(t, g, me, host); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, host)
	if !ok {
		t.Fatal("the host left")
	}
	e := c.Effective()
	if len(e.Types) != 1 || e.Types[0] != "Artifact" || c.IsCreature() {
		t.Errorf("types %v, want only Artifact", e.Types)
	}
	if !c.HasSubtype("Food") || c.HasSubtype("Dragon") {
		t.Errorf("subtypes %v, want only Food", e.Subtypes)
	}
	if !c.IsLegendary() {
		t.Error("the supertype was lost")
	}
	if len(e.Colors) != 0 {
		t.Errorf("colors %v, want colorless", e.Colors)
	}
	if game.HasKeyword(&c, "flying") || game.HasKeyword(&c, "haste") {
		t.Error("it kept its keywords")
	}
	if c.Name != "Red Flyer" || c.ManaValue() != 5 {
		t.Errorf("name %q mana value %d, want the name and mana value kept", c.Name, c.ManaValue())
	}
	// It is a Food: {2}, {T}, sacrifice: gain 3 life.
	floatForTest(g, me, "GG")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("eat the Food: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d life, want 3", got)
	}
}

func TestSugarCoatCanEnchantAFoodButNotAnotherKindOfPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	food := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Food", TypeLine: "Token Artifact — Food", Owner: me.ID, Controller: me.ID,
	})
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	advanceToMain(t, g)
	if _, err := scCast(t, g, me, land); err == nil {
		t.Error("enchanted a land")
	}
	if _, err := scCast(t, g, me, food); err != nil {
		t.Errorf("could not enchant a Food: %v", err)
	}
}
