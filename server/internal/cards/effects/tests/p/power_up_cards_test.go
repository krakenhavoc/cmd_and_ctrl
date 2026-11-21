package testsp

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPowerUpCardsAreFull(t *testing.T) {
	for _, oracle := range []string{effectstest.BraveBrawlerOracle, effectstest.HulkGammaGoliathOracle, effectstest.AdvancingTheSpiritOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v, want Full and none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// Brave Brawler cast this turn: {4}{W} less its {1}{W} is {3}, paid with
// three colourless mana; it puts two +1/+1 counters on it and cannot be
// activated again.
func TestPowerUpBraveBrawlerCostsLessTheTurnItEnteredAndOnlyOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	brawler := effectstest.CastPowerUpCreature(t, g, me, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", "{C}{W}", effectstest.BraveBrawlerOracle, 2)

	if got := effectstest.ChargedPowerUp(t, g, me.ID, brawler); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3}", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, brawler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up for {3}: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, brawler); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, brawler, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

// The turn after, the same Brawler pays the printed {4}{W}.
func TestPowerUpBraveBrawlerPaysFullPriceTheNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	brawler := effectstest.CastPowerUpCreature(t, g, me, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", "{C}{W}", effectstest.BraveBrawlerOracle, 2)
	effectstest.PowerUpPassTo(t, g, 0)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, brawler); got != "{4}{W}" {
		t.Errorf("the next turn: %q, want {4}{W}", got)
	}
}

// Hulk's static discounts another creature's power-up by {3} and not his
// own, and only for creatures his controller controls. His own costs
// {3} the turn he entered ({6}{R}{G} less {3}{R}{G}).
func TestPowerUpHulkDiscountsOtherCreaturesPowerUps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", effectstest.BraveBrawlerOracle)
	theirs := effectstest.PushPowerUpPermanent(g, opp.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", effectstest.BraveBrawlerOracle)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, mine); got != "{4}{W}" {
		t.Fatalf("before Hulk: %q, want {4}{W}", got)
	}
	hulk := effectstest.CastPowerUpCreature(t, g, me, "Hulk, Gamma Goliath", "Legendary Creature — Gamma Berserker Hero", "{3}{R}{G}", "{C}{C}{C}{R}{G}", effectstest.HulkGammaGoliathOracle, 6)

	if got := effectstest.ChargedPowerUp(t, g, me.ID, mine); got != "{1}{W}" {
		t.Errorf("my Brawler under Hulk: %q, want {1}{W}", got)
	}
	if got := effectstest.ChargedPowerUp(t, g, opp.ID, theirs); got != "{4}{W}" {
		t.Errorf("an opponent's Brawler: %q, want {4}{W} — Hulk discounts creatures you control", got)
	}
	if got := effectstest.ChargedPowerUp(t, g, me.ID, hulk); got != "{3}" {
		t.Errorf("Hulk's own power-up the turn he entered: %q, want {3} — his static is for OTHER creatures", got)
	}
	effectstest.PowerUpPassTo(t, g, 0)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, hulk); got != "{6}{R}{G}" {
		t.Errorf("Hulk's own power-up the next turn: %q, want {6}{R}{G}", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate my Brawler for {1}{W}: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, mine); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

// Advancing the Spirit: the first power-up I activate during my turn is
// priced from {0}; the second that turn pays its cost; on an opponent's
// turn there is no alternative; and my next turn opens it again.
func TestPowerUpAdvancingTheSpiritPaysZeroForTheFirstPowerUpOfEachOfYourTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	effectstest.PushPowerUpPermanent(g, me.ID, "Advancing the Spirit", "Enchantment", "{2}{G}", effectstest.AdvancingTheSpiritOracle)
	a := effectstest.PushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", effectstest.BraveBrawlerOracle)
	b := effectstest.PushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", effectstest.BraveBrawlerOracle)
	c := effectstest.PushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", effectstest.BraveBrawlerOracle)

	if got := effectstest.ChargedPowerUp(t, g, me.ID, a); got != "" {
		t.Fatalf("the first power-up of my turn: %q, want {0}", got)
	}
	if me.ManaPool.Total() != 0 {
		t.Fatalf("setup: pool = %v, want empty", me.ManaPool)
	}
	if err := g.ActivateCatalogAbility(me.ID, a, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up for {0} with an empty pool: %v", err)
	}
	// Announced, not resolved: the alternative is already spent.
	if got := effectstest.ChargedPowerUp(t, g, me.ID, b); got != "{4}{W}" {
		t.Errorf("the second power-up of my turn: %q, want {4}{W}", got)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, a); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}

	effectstest.PowerUpPassTo(t, g, 1)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, c); got != "{4}{W}" {
		t.Errorf("on an opponent's turn: %q, want {4}{W} — the clause is for each of YOUR turns", got)
	}
	effectstest.PowerUpPassTo(t, g, 0)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, c); got != "" {
		t.Errorf("the first power-up of my next turn: %q, want {0}", got)
	}
}
